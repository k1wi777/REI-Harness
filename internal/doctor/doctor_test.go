package doctor

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/my-harness-SDD/internal/check"
	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

// setupValidHarness crea un harness completo y válido en un directorio temporal,
// sin Work Items y sin repositorio git.
func setupValidHarness(t *testing.T) *paths.Project {
	t.Helper()
	root := t.TempDir()
	p := &paths.Project{Root: root}
	for _, f := range check.RequiredFiles {
		full := filepath.Join(root, f)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range []string{".rei/specs", ".rei/progress/work-items"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{".rei/progress/current.md", ".rei/progress/history.md"} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(f)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func TestRunOK(t *testing.T) {
	p := setupValidHarness(t)
	var buf bytes.Buffer
	if code := Run(p, &buf); code != 0 {
		t.Fatalf("Run = %d, salida:\n%s", code, buf.String())
	}
}

func TestRunMissingFile(t *testing.T) {
	root := t.TempDir()
	p := &paths.Project{Root: root}
	var buf bytes.Buffer
	if code := Run(p, &buf); code == 0 {
		t.Fatalf("esperaba fallo por archivos faltantes, salida:\n%s", buf.String())
	}
}

func TestRunPersonalizacionPendiente(t *testing.T) {
	p := setupValidHarness(t)
	doc := filepath.Join(p.Root, ".rei", "docs", "project", "architecture.md")
	if err := os.WriteFile(doc, []byte("<!-- REI:PENDIENTE -->"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if code := Run(p, &buf); code != 0 {
		t.Fatalf("Run = %d, esperaba 0 (los pendientes son WARN), salida:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "Personalización pendiente: .rei/docs/project/architecture.md") {
		t.Errorf("faltó el WARN de personalización pendiente:\n%s", buf.String())
	}
}

func TestRunPersonalizacionCompleta(t *testing.T) {
	p := setupValidHarness(t)
	var buf bytes.Buffer
	if code := Run(p, &buf); code != 0 {
		t.Fatalf("Run = %d, salida:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "Documentación del proyecto personalizada.") {
		t.Errorf("faltó el OK de documentación personalizada:\n%s", buf.String())
	}
}

func TestRunInvalidWorkItem(t *testing.T) {
	p := setupValidHarness(t)
	id := "2026-01-01_00-00__bad"
	if err := os.MkdirAll(p.SpecDir(id), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.MetaFile(id), []byte("{ inválido"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if code := Run(p, &buf); code == 0 {
		t.Fatalf("esperaba fallo por Work Item inválido, salida:\n%s", buf.String())
	}
}

func writeDoctorItem(t *testing.T, p *paths.Project, id, status string) {
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
	if err := os.MkdirAll(p.WorkItemDir(id), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.SpecDir(id), "plan.md"), []byte("# Plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunUnInProgressOK(t *testing.T) {
	p := setupValidHarness(t)
	writeDoctorItem(t, p, "2026-01-01_00-00__a", meta.StatusInProgress)
	writeDoctorItem(t, p, "2026-01-02_00-00__b", meta.StatusReady)
	var buf bytes.Buffer
	if code := Run(p, &buf); code != 0 {
		t.Fatalf("Run = %d, esperaba 0, salida:\n%s", code, buf.String())
	}
}

func TestRunDosInProgressFail(t *testing.T) {
	p := setupValidHarness(t)
	writeDoctorItem(t, p, "2026-01-01_00-00__a", meta.StatusInProgress)
	writeDoctorItem(t, p, "2026-01-02_00-00__b", meta.StatusInProgress)
	var buf bytes.Buffer
	if code := Run(p, &buf); code != 1 {
		t.Fatalf("Run = %d, esperaba 1, salida:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[FAIL]") || !strings.Contains(out, "Solo puede haber uno") ||
		!strings.Contains(out, "2026-01-01_00-00__a") || !strings.Contains(out, "2026-01-02_00-00__b") {
		t.Fatalf("salida inesperada:\n%s", out)
	}
}
