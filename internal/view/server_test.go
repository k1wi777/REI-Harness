package view

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/k1wi777/REI-Harness/internal/meta"
	"github.com/k1wi777/REI-Harness/internal/paths"
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

// setViewStatus cambia el estado del Work Item en meta.json.
func setViewStatus(t *testing.T, p *paths.Project, id, status string) {
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

// scanSSE convierte el cuerpo SSE en líneas por canal hasta que el stream se
// cierra.
func scanSSE(body io.Reader, lines chan<- string) {
	sc := bufio.NewScanner(body)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		lines <- sc.Text()
	}
	close(lines)
}

// waitForSSE espera y devuelve la primera línea SSE con el prefijo dado.
func waitForSSE(t *testing.T, lines <-chan string, prefix string) string {
	t.Helper()
	deadline := time.After(3 * time.Second)
	for {
		select {
		case l, ok := <-lines:
			if !ok {
				t.Fatalf("el stream SSE se cerró esperando %q", prefix)
			}
			if strings.HasPrefix(strings.TrimSpace(l), prefix) {
				return strings.TrimSpace(l)
			}
		case <-deadline:
			t.Fatalf("no se recibió %q por SSE", prefix)
		}
	}
}

// openEvents abre una conexión SSE contra base y devuelve sus líneas y una
// función de cierre que libera la conexión.
func openEvents(t *testing.T, base, token string) (<-chan string, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/events?token="+token, nil)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		cancel()
		_ = resp.Body.Close()
		t.Fatalf("/events = %d, want 200", resp.StatusCode)
	}
	lines := make(chan string, 32)
	go func() {
		scanSSE(resp.Body, lines)
		_ = resp.Body.Close()
	}()
	return lines, func() {
		cancel()
		_ = resp.Body.Close()
	}
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
	s.gracePeriod = 200 * time.Millisecond
	s.minFlushDelay = 0

	ln, err := s.Listen()
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- s.Serve(ln) }()

	setStatus := func(status string) {
		t.Helper()
		setViewStatus(t, p, id, status)
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

// TestEventsEmiteReloadAlCambiarEstadoYEnlaces verifica que meta.json y
// links.json forman parte de la instantánea vigilada: sus cambios disparan
// `reload` y el fragmento refleja el estado y la ficha nuevos (R3, R4).
func TestEventsEmiteReloadAlCambiarEstadoYEnlaces(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	s.pollInterval = 10 * time.Millisecond

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	lines, closeEvents := openEvents(t, ts.URL, s.token)
	defer closeEvents()
	time.Sleep(100 * time.Millisecond)

	setViewStatus(t, p, id, meta.StatusInProgress)
	waitForSSE(t, lines, "data: reload")
	if got := renderContentFor(t, p, id); !strings.Contains(got, "en curso") {
		t.Errorf("el contenido debe reflejar el nuevo estado:\n%s", got)
	}

	writeTestFile(t, filepath.Join(p.SpecDir(id), "links.json"),
		`{"provider":"github","repo":"owner/name","issue":7,"url":"https://github.com/owner/name/issues/7"}`)
	waitForSSE(t, lines, "data: reload")
	if got := renderContentFor(t, p, id); !strings.Contains(got, "#7") {
		t.Errorf("el contenido debe reflejar la ficha de enlaces:\n%s", got)
	}
}

// TestFinalAckCierre cubre la secuencia completa: observación de done, evento
// final con «terminado» y HTML status-done, confirmación por /ack y cierre
// posterior respetando el intervalo mínimo de vaciado (R9–R11).
func TestFinalAckCierre(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReview)
	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	s.background = true
	s.closeOnDone = true
	s.pollInterval = 10 * time.Millisecond
	s.gracePeriod = 2 * time.Second
	s.minFlushDelay = 300 * time.Millisecond

	ln, err := s.Listen()
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- s.Serve(ln) }()
	base := fmt.Sprintf("http://127.0.0.1:%d", portOf(ln))

	lines, closeEvents := openEvents(t, base, s.token)
	defer closeEvents()
	time.Sleep(100 * time.Millisecond)
	setViewStatus(t, p, id, meta.StatusDone)

	dataLine := waitForSSE(t, lines, "data: {")
	var payload finalPayload
	if err := json.Unmarshal([]byte(strings.TrimPrefix(dataLine, "data: ")), &payload); err != nil {
		t.Fatalf("payload final inválido: %v (%q)", err, dataLine)
	}
	if payload.Status != meta.StatusDone {
		t.Errorf("status = %q, want %q", payload.Status, meta.StatusDone)
	}
	if payload.Label != "terminado" {
		t.Errorf("label = %q, want %q", payload.Label, "terminado")
	}
	if !strings.Contains(payload.HTML, "status-done") || !strings.Contains(payload.HTML, "terminado") {
		t.Errorf("el HTML final debe mostrar done: %s", payload.HTML)
	}

	// El servidor no debe haberse cerrado antes de que la página confirme.
	select {
	case <-s.closed:
		t.Fatal("el servidor no debe cerrar antes de la confirmación")
	default:
	}

	resp, err := http.Post(base+"/ack?token="+s.token, "application/json", nil)
	if err != nil {
		t.Fatalf("POST /ack: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("/ack = %d, want 204", resp.StatusCode)
	}

	ackAt := time.Now()
	select {
	case <-s.closed:
		if elapsed := time.Since(ackAt); elapsed < 150*time.Millisecond {
			t.Errorf("cierre demasiado pronto tras el ack: %v", elapsed)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("el servidor debe cerrarse tras la confirmación")
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

// TestCierrePorGraciaSinAck verifica el respaldo sin cliente: el cierre ocurre
// tras la gracia y nunca antes del intervalo mínimo de vaciado (R10).
func TestCierrePorGraciaSinAck(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReview)
	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	s.background = true
	s.closeOnDone = true
	s.pollInterval = 10 * time.Millisecond
	s.gracePeriod = 100 * time.Millisecond
	s.minFlushDelay = 250 * time.Millisecond

	ln, err := s.Listen()
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- s.Serve(ln) }()

	start := time.Now()
	setViewStatus(t, p, id, meta.StatusDone)
	select {
	case <-s.closed:
		if elapsed := time.Since(start); elapsed < 200*time.Millisecond {
			t.Errorf("cierre antes del vaciado mínimo: %v", elapsed)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("el servidor debe cerrarse por gracia sin confirmación")
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

// TestPrimerPlanoNoCierraEnDone comprueba que en primer plano el servidor sigue
// sirviendo tras done hasta Ctrl+C (R13).
func TestPrimerPlanoNoCierraEnDone(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReview)
	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	s.background = false
	s.closeOnDone = false
	s.pollInterval = 10 * time.Millisecond

	ln, err := s.Listen()
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- s.Serve(ln) }()

	setViewStatus(t, p, id, meta.StatusDone)
	select {
	case <-s.closed:
		t.Fatal("en primer plano no debe cerrarse por done")
	case <-time.After(300 * time.Millisecond):
	}

	s.requestClose()
	select {
	case err := <-served:
		if err != nil {
			t.Fatalf("Serve devolvió error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve no terminó tras requestClose")
	}
}

// TestAckAutorizadoSinMutaciones cubre /ack: método y autorización, 204 y
// ausencia de escrituras sobre meta.json o los documentos (R14).
func TestAckAutorizadoSinMutaciones(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReview)
	s, err := newServer(p, id)
	if err != nil {
		t.Fatal(err)
	}
	h := s.Handler()

	if rr := doRequest(t, h, http.MethodPost, "/ack", nil); rr.Code != http.StatusForbidden {
		t.Errorf("/ack sin token = %d, want 403", rr.Code)
	}
	if rr := doRequest(t, h, http.MethodGet, "/ack?token="+s.token, nil); rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /ack = %d, want 405", rr.Code)
	}

	docPath := filepath.Join(p.SpecDir(id), "requirements.md")
	before, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatal(err)
	}
	if rr := doRequest(t, h, http.MethodPost, "/ack?token="+s.token, nil); rr.Code != http.StatusNoContent {
		t.Fatalf("/ack = %d, want 204; body: %s", rr.Code, rr.Body.String())
	}
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		t.Fatal(err)
	}
	if m.Status != meta.StatusReview {
		t.Errorf("/ack no debe cambiar el estado: %q", m.Status)
	}
	after, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("/ack no debe modificar los documentos")
	}
}
