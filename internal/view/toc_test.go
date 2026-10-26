package view

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/REI-Harness/internal/meta"
)

// TestRenderDocSections verifica la extracción de encabezados (nivel, id y
// texto) del AST de goldmark (R9, R10).
func TestRenderDocSections(t *testing.T) {
	src := []byte("# Título\n\n## Uno\n\ntexto\n\n### Sub\n\n## Dos con *énfasis*\n")
	_, sections, err := renderDoc(src)
	if err != nil {
		t.Fatalf("renderDoc: %v", err)
	}
	want := []sectionView{
		{Level: 1, Text: "Título"},
		{Level: 2, Text: "Uno"},
		{Level: 3, Text: "Sub"},
		{Level: 2, Text: "Dos con énfasis"},
	}
	if len(sections) != len(want) {
		t.Fatalf("sections = %+v, want %d", sections, len(want))
	}
	for i, w := range want {
		if sections[i].Level != w.Level || sections[i].Text != w.Text {
			t.Errorf("section %d = %+v, want %+v", i, sections[i], w)
		}
		if sections[i].ID == "" {
			t.Errorf("section %d sin id: %+v", i, sections[i])
		}
	}
}

// TestTOCEnPagina comprueba que el índice de secciones se emite solo con dos o
// más encabezados (R9, R12).
func TestTOCEnPagina(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusInProgress)

	writeTestFile(t, filepath.Join(p.SpecDir(id), "requirements.md"), "# R\n\n## A\n\n## B\n")
	vm, err := buildModel(p, id)
	if err != nil {
		t.Fatal(err)
	}
	var page strings.Builder
	if err := renderPage(&page, vm); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(page.String(), `class="toc"`) {
		t.Errorf("con 2+ encabezados debe haber TOC:\n%s", page.String())
	}

	writeTestFile(t, filepath.Join(p.SpecDir(id), "requirements.md"), "# R\n\nsolo texto\n")
	vm2, err := buildModel(p, id)
	if err != nil {
		t.Fatal(err)
	}
	var page2 strings.Builder
	if err := renderPage(&page2, vm2); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(page2.String(), `class="toc"`) {
		t.Errorf("con un solo encabezado no debe haber TOC:\n%s", page2.String())
	}
}
