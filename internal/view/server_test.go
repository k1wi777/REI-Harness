package view

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
)

func doRequest(t *testing.T, h http.Handler, method, target string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// TestServerTokenYContent cubre autenticación por token/cookie, /content y
// /static (R27, R28).
func TestServerTokenYContent(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	h := s.Handler()

	if rr := doRequest(t, h, http.MethodGet, "/", nil); rr.Code != http.StatusForbidden {
		t.Errorf("sin token = %d, want 403", rr.Code)
	}
	if rr := doRequest(t, h, http.MethodGet, "/?token=malo", nil); rr.Code != http.StatusForbidden {
		t.Errorf("token inválido = %d, want 403", rr.Code)
	}

	rr := doRequest(t, h, http.MethodGet, "/?token="+s.token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("token válido = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), id) {
		t.Errorf("la página debe mostrar el Work Item")
	}

	resp := rr.Result()
	var cookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == cookieName {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("no se fijó la cookie de sesión")
	}
	if !cookie.HttpOnly {
		t.Errorf("la cookie debe ser HttpOnly")
	}

	rrContent := doRequest(t, h, http.MethodGet, "/content", cookie)
	if rrContent.Code != http.StatusOK {
		t.Errorf("/content con cookie = %d, want 200", rrContent.Code)
	}
	// El colofón informa siempre del estado de los cambios (aquí, sin git).
	if !strings.Contains(rrContent.Body.String(), "cambios") {
		t.Errorf("/content debe incluir la información de cambios:\n%s", rrContent.Body.String())
	}
	if rr := doRequest(t, h, http.MethodGet, "/content", nil); rr.Code != http.StatusForbidden {
		t.Errorf("/content sin cookie = %d, want 403", rr.Code)
	}
	if rr := doRequest(t, h, http.MethodGet, "/static/app.js", nil); rr.Code != http.StatusOK {
		t.Errorf("/static/app.js = %d, want 200", rr.Code)
	}
	if rr := doRequest(t, h, http.MethodGet, "/no-existe", nil); rr.Code != http.StatusNotFound {
		t.Errorf("ruta desconocida = %d, want 404", rr.Code)
	}
}

// TestApproveRutaInexistente confirma que el visor es de solo lectura: la ruta
// /approve ya no existe y no cambia meta.json.
func TestApproveRutaInexistente(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	rr := doRequest(t, s.Handler(), http.MethodPost, "/approve?token="+s.token, nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("/approve = %d, want 404; body: %s", rr.Code, rr.Body.String())
	}
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		t.Fatal(err)
	}
	if m.Status != meta.StatusReady {
		t.Errorf("meta.json no debería cambiar: %q", m.Status)
	}
}

// TestEventsEmiteReload verifica que el SSE emite `data: reload` al cambiar un
// documento mostrado (R19, R20).
func TestEventsEmiteReload(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	s.pollInterval = 10 * time.Millisecond

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/events?token="+s.token, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/events = %d, want 200", resp.StatusCode)
	}

	// Da margen a que el handler tome la instantánea inicial, luego cambia el doc.
	time.Sleep(100 * time.Millisecond)
	writeTestFile(t, filepath.Join(p.SpecDir(id), "requirements.md"), "# cambiado\n")

	lines := make(chan string, 8)
	go func() {
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			lines <- sc.Text()
		}
	}()

	deadline := time.After(3 * time.Second)
	for {
		select {
		case l := <-lines:
			if strings.TrimSpace(l) == "data: reload" {
				return
			}
		case <-deadline:
			t.Fatal("no se recibió 'data: reload' por SSE")
		}
	}
}

// TestFondoSeCierraEnDone cubre el ciclo de vida del servidor de fondo: al
// pasar a in_progress y durante review NO se cierra; solo se cierra al llegar
// a done.
func TestFondoSeCierraEnDone(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	s.background = true
	s.closeOnDone = true
	s.pollInterval = 10 * time.Millisecond

	ln, err := s.Listen()
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- s.Serve(ln) }()

	setStatus := func(status string) {
		t.Helper()
		m, err := meta.Load(p.MetaFile(id))
		if err != nil {
			t.Fatal(err)
		}
		m.Status = status
		if err := m.Save(p.MetaFile(id)); err != nil {
			t.Fatal(err)
		}
	}

	// ready -> in_progress: el visor sigue abierto para leer los documentos.
	setStatus(meta.StatusInProgress)
	select {
	case <-s.closed:
		t.Fatal("el servidor no debe cerrarse en in_progress")
	case <-time.After(150 * time.Millisecond):
	}

	// in_progress -> review: tampoco.
	setStatus(meta.StatusReview)
	select {
	case <-s.closed:
		t.Fatal("el servidor no debe cerrarse en review")
	case <-time.After(150 * time.Millisecond):
	}

	// review -> done: trabajo terminado, el servidor se cierra.
	setStatus(meta.StatusDone)
	select {
	case <-s.closed:
	case <-time.After(3 * time.Second):
		t.Fatal("el servidor debe cerrarse al llegar a done")
	}

	select {
	case err := <-served:
		if err != nil {
			t.Fatalf("Serve devolvió error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve no terminó tras cerrarse el servidor")
	}
}
