package check

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

func TestRunOK(t *testing.T) {
	root := t.TempDir()
	p := &paths.Project{Root: root}
	for _, f := range RequiredFiles {
		full := filepath.Join(root, f)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var buf bytes.Buffer
	if code := Run(p, true, &buf); code != 0 {
		t.Fatalf("Run = %d, salida:\n%s", code, buf.String())
	}
}

func TestRunMissingFile(t *testing.T) {
	root := t.TempDir()
	p := &paths.Project{Root: root}
	var buf bytes.Buffer
	if code := Run(p, true, &buf); code == 0 {
		t.Fatalf("esperaba fallo por archivos faltantes")
	}
}

func setupTemplates(t *testing.T, p *paths.Project) {
	t.Helper()
	if err := os.MkdirAll(p.TemplatesDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"current.md", "history.md"} {
		if err := os.WriteFile(filepath.Join(p.TemplatesDir(), f), []byte("plantilla "+f), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEnsureStructureOK(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	setupTemplates(t, p)
	var buf bytes.Buffer
	if code := EnsureStructure(p, &buf); code != 0 {
		t.Fatalf("EnsureStructure = %d, salida:\n%s", code, buf.String())
	}
	for _, dir := range []string{p.SpecsDir(), p.WorkItemsDir()} {
		if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
			t.Errorf("no se creó %s (err=%v)", dir, err)
		}
	}
	for _, f := range []string{p.CurrentFile(), p.HistoryFile()} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("no se creó %s (err=%v)", f, err)
		}
	}
}

func TestEnsureStructureIdempotente(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	setupTemplates(t, p)
	var buf bytes.Buffer
	if code := EnsureStructure(p, &buf); code != 0 {
		t.Fatalf("primera llamada = %d, salida:\n%s", code, buf.String())
	}
	if err := os.WriteFile(p.CurrentFile(), []byte("modificado"), 0o644); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if code := EnsureStructure(p, &buf); code != 0 {
		t.Fatalf("segunda llamada = %d, salida:\n%s", code, buf.String())
	}
	data, err := os.ReadFile(p.CurrentFile())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "modificado" {
		t.Errorf("current.md se sobrescribió: %q", string(data))
	}
}

func writeConfig(t *testing.T, p *paths.Project, body string) {
	t.Helper()
	if err := os.MkdirAll(p.ReiDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.ReiDir(), "config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunDeclaredSinChecks(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	var buf bytes.Buffer
	if code := RunDeclared(p, &buf); code != 0 {
		t.Fatalf("RunDeclared = %d, esperaba 0, salida:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "Sin checks configurados") {
		t.Fatalf("salida inesperada:\n%s", buf.String())
	}
}

func TestRunDeclaredChecksVacios(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeConfig(t, p, `{"checks":[]}`)
	var buf bytes.Buffer
	if code := RunDeclared(p, &buf); code != 0 {
		t.Fatalf("RunDeclared = %d, esperaba 0, salida:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "Sin checks configurados") {
		t.Fatalf("salida inesperada:\n%s", buf.String())
	}
}

func TestRunDeclaredCheckOK(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeConfig(t, p, `{"checks":[{"id":"ok","description":"check-ok","command":["sh","-c","exit 0"]}]}`)
	var buf bytes.Buffer
	if code := RunDeclared(p, &buf); code != 0 {
		t.Fatalf("RunDeclared = %d, esperaba 0, salida:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[OK]") || !strings.Contains(out, "check-ok") {
		t.Fatalf("salida inesperada:\n%s", out)
	}
}

func TestRunDeclaredCheckFail(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeConfig(t, p, `{"checks":[{"id":"fail","description":"check-fail","command":["sh","-c","exit 1"]}]}`)
	var buf bytes.Buffer
	if code := RunDeclared(p, &buf); code != 1 {
		t.Fatalf("RunDeclared = %d, esperaba 1, salida:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[FAIL]") || !strings.Contains(out, "check-fail") {
		t.Fatalf("salida inesperada:\n%s", out)
	}
}

func TestRunDeclaredSinEfectos(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeConfig(t, p, `{"checks":[{"id":"ok","description":"check-ok","command":["sh","-c","exit 0"]}]}`)
	var buf bytes.Buffer
	if code := RunDeclared(p, &buf); code != 0 {
		t.Fatalf("RunDeclared = %d, esperaba 0, salida:\n%s", code, buf.String())
	}
	if _, err := os.Stat(p.SpecsDir()); !os.IsNotExist(err) {
		t.Errorf("RunDeclared creó .rei/specs/ (err=%v)", err)
	}
	if _, err := os.Stat(p.CurrentFile()); !os.IsNotExist(err) {
		t.Errorf("RunDeclared creó current.md (err=%v)", err)
	}
}

// writeWorkItem registra un Work Item con el estado indicado.
func writeWorkItem(t *testing.T, p *paths.Project, id, status string) {
	t.Helper()
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

func TestRunUnInProgressOK(t *testing.T) {
	p := setupValidProject(t)
	writeWorkItem(t, p, "2026-01-01_00-00__a", meta.StatusInProgress)
	var buf bytes.Buffer
	if code := Run(p, true, &buf); code != 0 {
		t.Fatalf("Run = %d, esperaba 0, salida:\n%s", code, buf.String())
	}
}

func TestRunDosInProgressFail(t *testing.T) {
	p := setupValidProject(t)
	writeWorkItem(t, p, "2026-01-01_00-00__a", meta.StatusInProgress)
	writeWorkItem(t, p, "2026-01-02_00-00__b", meta.StatusInProgress)
	var buf bytes.Buffer
	if code := Run(p, true, &buf); code != 1 {
		t.Fatalf("Run = %d, esperaba 1, salida:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[FAIL]") || !strings.Contains(out, "2026-01-01_00-00__a") ||
		!strings.Contains(out, "2026-01-02_00-00__b") {
		t.Fatalf("salida inesperada:\n%s", out)
	}
}

// setupValidProject crea un harness válido con los archivos requeridos.
func setupValidProject(t *testing.T) *paths.Project {
	t.Helper()
	root := t.TempDir()
	p := &paths.Project{Root: root}
	for _, f := range RequiredFiles {
		full := filepath.Join(root, f)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return p
}
