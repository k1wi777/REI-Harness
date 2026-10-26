package workitem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1wi777/REI-Harness/internal/meta"
	"github.com/k1wi777/REI-Harness/internal/paths"
)

const metaTemplate = `{"id":"{{ID}}","title":"{{TITLE}}","description":"{{DESCRIPTION}}","type":"{{TYPE}}","status":"{{STATUS}}","created_at":"{{CREATED_AT}}","base_commit":"{{BASE_COMMIT}}","last_review_commit":"{{LAST_REVIEW_COMMIT}}"}`

func TestNewCreatesValidMeta(t *testing.T) {
	root := t.TempDir()
	p := &paths.Project{Root: root}
	if err := os.MkdirAll(p.TemplatesDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.TemplatesDir(), "meta.json"), []byte(metaTemplate), 0o644); err != nil {
		t.Fatal(err)
	}

	id := "2026-10-01_10-00__demo"
	if err := New(p, id, "feature", `Demo "x"/y`); err != nil {
		t.Fatal(err)
	}
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		t.Fatalf("meta inválido: %v", err)
	}
	if m.ID != id || m.Type != "feature" || m.Status != "pending" || m.Title != `Demo "x"/y` {
		t.Fatalf("meta inesperado: %+v", m)
	}
	if fi, err := os.Stat(p.WorkItemDir(id)); err != nil || !fi.IsDir() {
		t.Fatal("no se creó .rei/progress/work-items/<id>")
	}
}

func TestNewRejectsBadInput(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	if err := New(p, "malo", "task", ""); err == nil {
		t.Fatal("esperaba error por id inválido")
	}
	if err := New(p, "2026-10-01_10-00__ok", "x", ""); err == nil {
		t.Fatal("esperaba error por type inválido")
	}
}
