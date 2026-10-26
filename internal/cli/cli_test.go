package cli

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

func TestProjectMessage(t *testing.T) {
	const want = "REI no está inicializado aquí; ejecuta `rei init` para inicializarlo."
	if got := projectMessage(paths.ErrNotFound); got != want {
		t.Fatalf("projectMessage(ErrNotFound) = %q, want %q", got, want)
	}
	if got := projectMessage(errors.New("boom")); got != "boom" {
		t.Fatalf("projectMessage(other) = %q, want %q", got, "boom")
	}
}

func TestCmdInitUpdateUsoIncorrecto(t *testing.T) {
	cases := [][]string{
		{"--update", "extra"},
		{"--update", "--bogus"},
		{"opencode", "--update"},
		{"--force", "--update", "extra"},
	}
	for _, args := range cases {
		if code := cmdInit(args); code != 2 {
			t.Errorf("cmdInit(%v) = %d, esperaba 2", args, code)
		}
	}
}

// TestCmdViewUsoIncorrecto cubre modos excluyentes, modos desconocidos y dos
// Work Items (R4, R5). El parseo falla antes de resolver el proyecto.
func TestCmdViewUsoIncorrecto(t *testing.T) {
	cases := [][]string{
		{"--open", "--stop"},
		{"--enable", "--disable"},
		{"--open", "--bogus"},
		{"--wat"},
		{"id1", "id2"},
	}
	for _, args := range cases {
		if code := cmdView(args); code != 2 {
			t.Errorf("cmdView(%v) = %d, esperaba 2", args, code)
		}
	}
}

// TestCmdGitHubUsoIncorrecto comprueba que el uso se valida antes de resolver
// el proyecto (R2): sin subcomando, desconocido o con argumentos sobrantes.
func TestCmdGitHubUsoIncorrecto(t *testing.T) {
	cases := [][]string{
		nil,
		{},
		{"bogus"},
		{"enable", "extra"},
		{"disable", "--force"},
		{"status", "id"},
	}
	for _, args := range cases {
		if code := cmdGitHub(args); code != 2 {
			t.Errorf("cmdGitHub(%v) = %d, esperaba 2", args, code)
		}
	}
}

// newCLIFixture crea un proyecto temporal con los Work Items indicados (id ->
// estado), una sesión inactiva y cambia el directorio a la raíz del proyecto.
func newCLIFixture(t *testing.T, statuses map[string]string) *paths.Project {
	t.Helper()
	root := t.TempDir()
	p := &paths.Project{Root: root}
	if err := os.MkdirAll(p.ProgressDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	current := "# Sesión actual\n\n- **Work Item:** _ninguno_\n- **Estado:** _—_ \n"
	if err := os.WriteFile(p.CurrentFile(), []byte(current), 0o644); err != nil {
		t.Fatal(err)
	}
	for id, status := range statuses {
		if err := os.MkdirAll(p.SpecDir(id), 0o755); err != nil {
			t.Fatal(err)
		}
		m := &meta.Meta{
			ID: id, Title: "t", Description: "d", Type: "task",
			Status: status, CreatedAt: "2026-01-01T00:00:00-05:00",
		}
		if err := m.Save(p.MetaFile(id)); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(root)
	return p
}

// capture ejecuta fn capturando stdout y stderr.
func capture(t *testing.T, fn func() int) (code int, stdout, stderr string) {
	t.Helper()
	origOut, origErr := os.Stdout, os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = wOut, wErr
	code = fn()
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout, os.Stderr = origOut, origErr
	outBytes, _ := io.ReadAll(rOut)
	errBytes, _ := io.ReadAll(rErr)
	_ = rOut.Close()
	_ = rErr.Close()
	return code, string(outBytes), string(errBytes)
}

func TestCmdItemsStatusSinActivos(t *testing.T) {
	newCLIFixture(t, map[string]string{
		"2026-01-01_00-00__a": meta.StatusPending,
		"2026-01-02_00-00__b": meta.StatusReview,
	})
	code, out, _ := capture(t, func() int { return cmdItems([]string{"status"}) })
	if code != 0 {
		t.Fatalf("cmdItems = %d, want 0", code)
	}
	if strings.Contains(out, "AVISO") || strings.Contains(out, "<< ACTIVO") {
		t.Fatalf("sin activos no debe haber aviso ni marcador:\n%s", out)
	}
}

func TestCmdItemsStatusUnActivo(t *testing.T) {
	newCLIFixture(t, map[string]string{
		"2026-01-01_00-00__a": meta.StatusInProgress,
		"2026-01-02_00-00__b": meta.StatusReady,
	})
	code, out, _ := capture(t, func() int { return cmdItems([]string{"status"}) })
	if code != 0 {
		t.Fatalf("cmdItems = %d, want 0", code)
	}
	if strings.Contains(out, "AVISO") {
		t.Fatalf("con exactamente un activo no debe avisar:\n%s", out)
	}
	if !strings.Contains(out, "<< ACTIVO") {
		t.Fatalf("debe conservar el marcador << ACTIVO:\n%s", out)
	}
}

func TestCmdItemsStatusDosActivos(t *testing.T) {
	newCLIFixture(t, map[string]string{
		"2026-01-01_00-00__a": meta.StatusInProgress,
		"2026-01-02_00-00__b": meta.StatusInProgress,
	})
	code, _, stderr := capture(t, func() int { return cmdItems([]string{"status"}) })
	if code != 1 {
		t.Fatalf("cmdItems = %d, want 1", code)
	}
	if !strings.Contains(stderr, "solo puede haber uno") ||
		!strings.Contains(stderr, "2026-01-01_00-00__a") ||
		!strings.Contains(stderr, "2026-01-02_00-00__b") {
		t.Fatalf("mensaje inesperado: %q", stderr)
	}
}

func TestCmdStatusRechazaOtroInProgress(t *testing.T) {
	const id = "2026-01-01_00-00__a"
	const otro = "2026-01-02_00-00__b"
	p := newCLIFixture(t, map[string]string{
		id:   meta.StatusReady,
		otro: meta.StatusInProgress,
	})
	code, _, stderr := capture(t, func() int {
		return cmdStatus([]string{"set", id, meta.StatusInProgress})
	})
	if code != 1 {
		t.Fatalf("cmdStatus = %d, want 1", code)
	}
	if !strings.Contains(stderr, "ya está en curso") || !strings.Contains(stderr, otro) {
		t.Fatalf("mensaje inesperado: %q", stderr)
	}
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		t.Fatal(err)
	}
	if m.Status != meta.StatusReady {
		t.Fatalf("meta.json no debería cambiar: %q", m.Status)
	}
}
