package view

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/k1wi777/REI-Harness/internal/meta"
	"github.com/k1wi777/REI-Harness/internal/paths"
)

const cookieName = "rei_view"

// Server sirve el visor de un Work Item. Todas las rutas (salvo la carga
// inicial con token válido) exigen autenticación por cookie o token.
type Server struct {
	p     *paths.Project
	id    string
	token string

	background   bool
	closeOnDone  bool
	pollInterval time.Duration
	idleTimeout  time.Duration

	// Protocolo de cierre en done: marca terminal idempotente, confirmación del
	// cliente y gracia de respaldo (R9–R13).
	doneOnce      sync.Once
	done          chan struct{}
	ack           chan struct{}
	gracePeriod   time.Duration
	minFlushDelay time.Duration

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
		p:             p,
		id:            id,
		token:         token,
		pollInterval:  time.Second,
		idleTimeout:   30 * time.Minute,
		gracePeriod:   2 * time.Second,
		minFlushDelay: 200 * time.Millisecond,
		done:          make(chan struct{}),
		ack:           make(chan struct{}, 1),
		lastActivity:  time.Now(),
		closed:        make(chan struct{}),
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
	mux.HandleFunc("/ack", s.handleAck)
	mux.HandleFunc("/static/", s.handleStatic)
	return s.trackActivity(mux)
}

// Serve sirve hasta que el servidor se cierra. Si es de fondo, arranca el
// monitor de inactividad y, si procede, el vigilante del estado terminal.
func (s *Server) Serve(ln net.Listener) error {
	s.srv = &http.Server{Handler: s.Handler(), ReadHeaderTimeout: 5 * time.Second}
	s.stopServer = func() { _ = s.srv.Close() }
	if s.background {
		go s.monitor()
		if s.closeOnDone {
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

// watchStatus es el único responsable de cerrar por done. Al detectarlo marca el
// estado terminal (idempotente) y deja que las conexiones SSE emitan su evento
// final antes de cerrar. Mientras avanza (in_progress, review) el servidor sigue
// vivo para poder leer los documentos que se generan; además siguen vigentes
// `rei view --stop` y el idle timeout.
func (s *Server) watchStatus() {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.closed:
			return
		case <-ticker.C:
			if m, err := meta.Load(s.p.MetaFile(s.id)); err == nil && m.Status == meta.StatusDone {
				s.markDone()
				return
			}
		}
	}
}

// markDone marca el estado terminal una sola vez y programa el cierre. Es
// idempotente: observar `done` de nuevo no reinicia la gracia.
func (s *Server) markDone() {
	s.doneOnce.Do(func() {
		close(s.done)
		go s.closeAfterDone()
	})
}

// closeAfterDone espera la confirmación de la página o, en su defecto, la
// gracia de respaldo, y nunca cierra antes del intervalo mínimo de vaciado para
// que las conexiones SSE alcancen a emitir el evento final (R10, R11).
func (s *Server) closeAfterDone() {
	start := time.Now()
	timer := time.NewTimer(s.gracePeriod)
	defer timer.Stop()
	select {
	case <-s.closed:
		return
	case <-s.ack:
	case <-timer.C:
	}
	if wait := s.minFlushDelay - time.Since(start); wait > 0 {
		drain := time.NewTimer(wait)
		defer drain.Stop()
		select {
		case <-s.closed:
			return
		case <-drain.C:
		}
	}
	s.requestClose()
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

// handleEvents es el endpoint SSE: emite `data: reload` cuando cambia el mtime
// de cualquier fichero vigilado (documentos, meta.json o links.json) y, cuando
// el servidor de fondo va a cerrarse por done, un evento final con el estado y
// el fragmento ya renderizado. Nunca cierra por su cuenta (R9–R11).
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
		case <-s.done:
			s.sendFinal(w, flusher)
			// El evento final ya se ha enviado: se mantiene el stream hasta que
			// el servidor cierre para que el cliente no vea un corte prematuro.
			select {
			case <-r.Context().Done():
			case <-s.closed:
			}
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
		}
	}
}

// finalPayload es el contenido del evento SSE `final`. El HTML viaja en el
// propio evento para que el cliente no dependa de un fetch que podría perder la
// carrera contra el cierre del servidor.
type finalPayload struct {
	Status string `json:"status"`
	Label  string `json:"label"`
	HTML   string `json:"html"`
}

// sendFinal renderiza el fragmento con el estado actual (ya done) y lo envía en
// un único evento SSE con JSON de una sola línea.
func (s *Server) sendFinal(w io.Writer, flusher http.Flusher) {
	vm, err := buildModel(s.p, s.id)
	if err != nil {
		return
	}
	var content bytes.Buffer
	if err := renderContent(&content, vm); err != nil {
		return
	}
	data, err := json.Marshal(finalPayload{
		Status: vm.Meta.Status,
		Label:  statusLabel(vm.Meta.Status),
		HTML:   content.String(),
	})
	if err != nil {
		return
	}
	if _, err := fmt.Fprintf(w, "event: final\ndata: %s\n\n", data); err != nil {
		return
	}
	flusher.Flush()
}

// handleAck recibe la confirmación de que la página aplicó el estado final.
// Solo valida la sesión y despierta al cierre: no lee ni escribe ningún
// documento, meta.json ni estado (R14).
func (s *Server) handleAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorized(r) {
		http.Error(w, "no autorizado", http.StatusForbidden)
		return
	}
	select {
	case s.ack <- struct{}{}:
	default:
	}
	w.WriteHeader(http.StatusNoContent)
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

// snapshot recoge el mtime de los documentos mostrados y de los metadatos que
// alimentan la cabecera y la FICHA (estado y enlaces), de modo que sus cambios
// también disparan el `reload` (R3, R4).
func (s *Server) snapshot() map[string]time.Time {
	snap := make(map[string]time.Time, len(documentos)+2)
	for _, name := range documentos {
		if fi, err := os.Stat(docPath(s.p, s.id, name)); err == nil {
			snap[name] = fi.ModTime()
		}
	}
	for _, extra := range []struct{ key, path string }{
		{"meta.json", s.p.MetaFile(s.id)},
		{"links.json", linksPath(s.p, s.id)},
	} {
		if fi, err := os.Stat(extra.path); err == nil {
			snap[extra.key] = fi.ModTime()
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
