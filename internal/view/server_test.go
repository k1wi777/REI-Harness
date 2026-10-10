package view

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/state"
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

// TestApprovePermitido comprueba la transición ready -> in_progress y la
// alineación de current.md (R22, R23, R26).
func TestApprovePermitido(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	if err := os.MkdirAll(p.ProgressDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, p.CurrentFile(), "# Sesión actual\n\n- **Work Item:** "+id+"\n- **Estado:** ready\n")

	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	h := s.Handler()

	if rr := doRequest(t, h, http.MethodPost, "/approve", nil); rr.Code != http.StatusForbidden {
		t.Errorf("/approve sin token = %d, want 403", rr.Code)
	}
	if rr := doRequest(t, h, http.MethodGet, "/approve?token="+s.token, nil); rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("/approve GET = %d, want 405", rr.Code)
	}

	rr := doRequest(t, h, http.MethodPost, "/approve?token="+s.token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("/approve = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["ok"] != true || body["status"] != meta.StatusInProgress {
		t.Errorf("respuesta inesperada: %v", body)
	}
	if msg, _ := body["message"].(string); !strings.Contains(msg, "chat") {
		t.Errorf("el mensaje debe recordar reanudar en el chat: %q", msg)
	}

	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		t.Fatal(err)
	}
	if m.Status != meta.StatusInProgress {
		t.Errorf("meta.json = %q, want in_progress", m.Status)
	}
	sess, err := state.ReadSession(p)
	if err != nil {
		t.Fatal(err)
	}
	if sess.State != meta.StatusInProgress {
		t.Errorf("current.md Estado = %q, want in_progress", sess.State)
	}
}

// TestApproveDenegado comprueba que una transición no permitida no escribe
// meta.json y responde 409 (R24).
func TestApproveDenegado(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusPending)
	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	rr := doRequest(t, s.Handler(), http.MethodPost, "/approve?token="+s.token, nil)
	if rr.Code != http.StatusConflict {
		t.Fatalf("/approve no permitido = %d, want 409; body: %s", rr.Code, rr.Body.String())
	}
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		t.Fatal(err)
	}
	if m.Status != meta.StatusPending {
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
