package view

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
	"github.com/k1wi777/my-harness-SDD/internal/state"
)

const cookieName = "rei_view"

// Server sirve el visor de un Work Item. Todas las rutas (salvo la carga
// inicial con token válido) exigen autenticación por cookie o token.
type Server struct {
	p     *paths.Project
	id    string
	token string

	background          bool
	closeOnStatusChange bool
	pollInterval        time.Duration
	idleTimeout         time.Duration

	srv        *http.Server
	stopServer func()

	mu           sync.Mutex
	lastActivity time.Time

	closed    chan struct{}
	closeOnce sync.Once
	onClose   func()
}

// newServer construye un servidor para el Work Item id con un token secreto por
// instancia.
func newServer(p *paths.Project, id string) (*Server, error) {
	token, err := newToken()
	if err != nil {
		return nil, err
	}
	return &Server{
		p:            p,
		id:           id,
		token:        token,
		pollInterval: time.Second,
		idleTimeout:  30 * time.Minute,
		lastActivity: time.Now(),
		closed:       make(chan struct{}),
	}, nil
}

// newToken genera 32 bytes aleatorios en hex.
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Listen abre un listener en 127.0.0.1 con puerto efímero.
func (s *Server) Listen() (net.Listener, error) {
	return net.Listen("tcp", "127.0.0.1:0")
}

// Port devuelve el puerto del listener.
func portOf(ln net.Listener) int {
	if addr, ok := ln.Addr().(*net.TCPAddr); ok {
		return addr.Port
	}
	return 0
}

// URL construye la URL inicial del visor (con el token).
func (s *Server) URL(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d/?token=%s", port, s.token)
}

// Handler devuelve el mux de rutas, envuelto para registrar la actividad.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/content", s.handleContent)
	mux.HandleFunc("/events", s.handleEvents)
	mux.HandleFunc("/approve", s.handleApprove)
	mux.HandleFunc("/static/", s.handleStatic)
	return s.trackActivity(mux)
}

// Serve sirve hasta que el servidor se cierra. Si es de fondo, arranca el
// monitor de inactividad.
func (s *Server) Serve(ln net.Listener) error {
	s.srv = &http.Server{Handler: s.Handler(), ReadHeaderTimeout: 5 * time.Second}
	s.stopServer = func() { _ = s.srv.Close() }
	if s.background {
		go s.monitor()
		if s.closeOnStatusChange {
			go s.watchStatus()
		}
	}
	err := s.srv.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// requestClose cierra el servidor una sola vez e invoca onClose.
func (s *Server) requestClose() {
	s.closeOnce.Do(func() {
		close(s.closed)
		if s.stopServer != nil {
			go s.stopServer()
		}
		if s.onClose != nil {
			s.onClose()
		}
	})
}

// watchStatus cierra el servidor de fondo en cuanto el Work Item deja de estar
// en ready (R30), sin depender de que haya un cliente SSE conectado.
func (s *Server) watchStatus() {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.closed:
			return
		case <-ticker.C:
			if m, err := meta.Load(s.p.MetaFile(s.id)); err == nil && m.Status != meta.StatusReady {
				s.requestClose()
				return
			}
		}
	}
}

// monitor cierra el servidor de fondo tras idleTimeout sin peticiones.
func (s *Server) monitor() {
	interval := s.idleTimeout / 2
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.closed:
			return
		case <-ticker.C:
			s.mu.Lock()
			last := s.lastActivity
			s.mu.Unlock()
			if time.Since(last) >= s.idleTimeout {
				s.requestClose()
				return
			}
		}
	}
}

func (s *Server) trackActivity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.lastActivity = time.Now()
		s.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

// authorized valida el token por cookie o por query (carga inicial).
func (s *Server) authorized(r *http.Request) bool {
	if c, err := r.Cookie(cookieName); err == nil {
		if subtle.ConstantTimeCompare([]byte(c.Value), []byte(s.token)) == 1 {
			return true
		}
	}
	if q := r.URL.Query().Get("token"); q != "" {
		if subtle.ConstantTimeCompare([]byte(q), []byte(s.token)) == 1 {
			return true
		}
	}
	return false
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if !s.authorized(r) {
		http.Error(w, "no autorizado", http.StatusForbidden)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    s.token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	vm, err := buildModel(s.p, s.id)
	if err != nil {
		http.Error(w, "no se pudo cargar el Work Item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := renderPage(w, vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleContent(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		http.Error(w, "no autorizado", http.StatusForbidden)
		return
	}
	vm, err := buildModel(s.p, s.id)
	if err != nil {
		http.Error(w, "no se pudo cargar el Work Item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := renderContent(w, vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorized(r) {
		http.Error(w, "no autorizado", http.StatusForbidden)
		return
	}
	from, forced, err := state.Transition(s.p, s.id, meta.StatusInProgress, false)
	switch {
	case errors.Is(err, state.ErrTransitionNotAllowed):
		writeJSON(w, http.StatusConflict, map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("transición no permitida: %s -> in_progress", from),
		})
	case err != nil:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
	default:
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"from":    from,
			"status":  meta.StatusInProgress,
			"forced":  forced,
			"message": "Aprobado: ready -> in_progress. Reanuda la conversación en el chat con una palabra breve (p. ej. «listo»).",
		})
	}
}

// handleEvents es el endpoint SSE: emite `data: reload` cuando cambia el mtime
// de algún documento del Work Item. En segundo plano, además, cierra el
// servidor cuando el Work Item deja de estar en ready.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		http.Error(w, "no autorizado", http.StatusForbidden)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming no soportado", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	last := s.snapshot()
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-s.closed:
			return
		case <-ticker.C:
			cur := s.snapshot()
			if changedSnapshot(last, cur) {
				last = cur
				if _, err := fmt.Fprint(w, "data: reload\n\n"); err != nil {
					return
				}
				flusher.Flush()
			}
			if s.background && s.closeOnStatusChange {
				if m, err := meta.Load(s.p.MetaFile(s.id)); err == nil && m.Status != meta.StatusReady {
					s.requestClose()
					return
				}
			}
		}
	}
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/static/")
	if name == "" || strings.Contains(name, "..") {
		http.NotFound(w, r)
		return
	}
	data, err := viewAssets.ReadFile("assets/" + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", staticContentType(name))
	_, _ = w.Write(data)
}

// snapshot recoge el mtime de los documentos mostrados del Work Item.
func (s *Server) snapshot() map[string]time.Time {
	snap := make(map[string]time.Time, len(documentos))
	for _, name := range documentos {
		if fi, err := os.Stat(docPath(s.p, s.id, name)); err == nil {
			snap[name] = fi.ModTime()
		}
	}
	return snap
}

// changedSnapshot indica si el conjunto de documentos o sus mtimes cambiaron.
func changedSnapshot(a, b map[string]time.Time) bool {
	if len(a) != len(b) {
		return true
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || !w.Equal(v) {
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func staticContentType(name string) string {
	switch path.Ext(name) {
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}
