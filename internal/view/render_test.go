package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

// newViewFixture crea un proyecto mínimo con un Work Item `feature` en el
// estado dado y requirements.md presente (design/tasks ausentes a propósito).
func newViewFixture(t *testing.T, status string) (*paths.Project, string) {
	t.Helper()
	root := t.TempDir()
	p := &paths.Project{Root: root}
	id := "2026-01-01_00-00__demo"
	if err := os.MkdirAll(p.SpecDir(id), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.WorkItemDir(id), 0o755); err != nil {
		t.Fatal(err)
	}
	m := &meta.Meta{
		ID: id, Title: "Demo", Description: "d", Type: "feature",
		Status: status, CreatedAt: "2026-01-01T00:00:00-05:00",
	}
	if err := m.Save(p.MetaFile(id)); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(p.SpecDir(id), "requirements.md"), "# Requisitos\n\n- uno\n- dos\n")
	return p, id
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRenderDocNeutralizaHTML comprueba el modo seguro por defecto de goldmark:
// el HTML crudo no se interpreta y el Markdown sí se convierte.
func TestRenderDocNeutralizaHTML(t *testing.T) {
	src := "# Título\n\n<script>alert(1)</script>\n\n<img src=x onerror=alert(2)>\n\n**negrita** y `código`\n"
	html, _, err := renderDoc([]byte(src))
	if err != nil {
		t.Fatalf("renderDoc: %v", err)
	}
	got := string(html)
	// En modo seguro (sin html.WithUnsafe), goldmark no interpreta el HTML crudo:
	// ni etiquetas activas ni eventos inline deben llegar al HTML final.
	if strings.Contains(got, "<script") || strings.Contains(got, "<img") || strings.Contains(got, "onerror") {
		t.Fatalf("el HTML crudo no debe interpretarse: %s", got)
	}
	if !strings.Contains(got, "<h1") || !strings.Contains(got, "<strong>negrita</strong>") {
		t.Errorf("el Markdown no se renderizó: %s", got)
	}
}

// TestBuildModelDocsPresentesAusentes verifica la composición con documentos
// presentes y ausentes (R11, R12, R16).
func TestBuildModelDocsPresentesAusentes(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	vm, err := buildModel(p, id)
	if err != nil {
		t.Fatalf("buildModel: %v", err)
	}
	byName := map[string]docView{}
	for _, d := range vm.Docs {
		byName[d.Name] = d
	}
	if !byName["requirements.md"].Present || byName["requirements.md"].HTML == "" {
		t.Errorf("requirements.md debería estar presente y renderizado: %+v", byName["requirements.md"])
	}
	if byName["design.md"].Present || byName["tasks.md"].Present {
		t.Errorf("design/tasks no existen y no deben marcarse presentes")
	}
	if !vm.CanApprove {
		t.Errorf("en ready, CanApprove debe ser true")
	}
}

// TestRenderPageYCcontent verifica la página completa y el fragmento, y la
// ausencia del botón fuera de ready (R21).
func TestRenderPageYCcontent(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	vm, err := buildModel(p, id)
	if err != nil {
		t.Fatal(err)
	}
	var page strings.Builder
	if err := renderPage(&page, vm); err != nil {
		t.Fatalf("renderPage: %v", err)
	}
	if !strings.Contains(page.String(), "<!DOCTYPE html>") || !strings.Contains(page.String(), `id="approve"`) {
		t.Errorf("la página en ready debe incluir el botón aprobar:\n%s", page.String())
	}
	if !strings.Contains(page.String(), "requirements.md") {
		t.Errorf("la página debe listar los documentos")
	}

	var content strings.Builder
	if err := renderContent(&content, vm); err != nil {
		t.Fatalf("renderContent: %v", err)
	}
	if strings.Contains(content.String(), "<!DOCTYPE html>") {
		t.Errorf("el fragmento /content no debe ser un documento HTML completo")
	}
	if !strings.Contains(content.String(), `id="approve"`) {
		t.Errorf("el fragmento debe incluir el botón en ready")
	}

	// Fuera de ready no hay botón.
	p2, id2 := newViewFixture(t, meta.StatusInProgress)
	vm2, err := buildModel(p2, id2)
	if err != nil {
		t.Fatal(err)
	}
	var c2 strings.Builder
	if err := renderContent(&c2, vm2); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(c2.String(), `id="approve"`) {
		t.Errorf("no debe haber botón aprobar fuera de ready:\n%s", c2.String())
	}
}
