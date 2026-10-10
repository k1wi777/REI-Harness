package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

func TestReadSessionActive(t *testing.T) {
	root := t.TempDir()
	progress := filepath.Join(root, ".rei", "progress")
	if err := os.MkdirAll(progress, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "# Sesión actual\n\n" +
		"- **Work Item:** 2026-10-01_10-00__demo\n" +
		"- **Tipo:** task\n" +
		"- **Estado:** in_progress\n" +
		"- **Inicio:** x\n" +
		"- **Agente activo:** implementer\n"
	if err := os.WriteFile(filepath.Join(progress, "current.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := &paths.Project{Root: root}
	s, err := ReadSession(p)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Active() {
		t.Fatal("la sesión debería estar activa")
	}
	if s.WorkItem != "2026-10-01_10-00__demo" || s.State != "in_progress" || s.Agent != "implementer" {
		t.Fatalf("sesión inesperada: %+v", s)
	}
}

func TestReadSessionIdle(t *testing.T) {
	root := t.TempDir()
	progress := filepath.Join(root, ".rei", "progress")
	if err := os.MkdirAll(progress, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "- **Work Item:** _ninguno_\n- **Estado:** _—_ \n"
	if err := os.WriteFile(filepath.Join(progress, "current.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	s, _ := ReadSession(&paths.Project{Root: root})
	if s.Active() {
		t.Fatal("no debería haber sesión activa")
	}
}

// newProjectFixture crea un proyecto con un meta.json en el estado dado y una
// sesión activa que apunta al Work Item.
func newProjectFixture(t *testing.T, id, status string, active bool) (*paths.Project, string) {
	t.Helper()
	root := t.TempDir()
	p := &paths.Project{Root: root}
	if err := os.MkdirAll(p.SpecDir(id), 0o755); err != nil {
		t.Fatal(err)
	}
	m := &meta.Meta{
		ID: id, Title: "t", Description: "d", Type: "feature",
		Status: status, CreatedAt: "2026-01-01T00:00:00-05:00",
	}
	if err := m.Save(p.MetaFile(id)); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.ProgressDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	workItem := "_ninguno_"
	if active {
		workItem = id
	}
	current := "# Sesión actual\n\n" +
		"- **Work Item:** " + workItem + "\n" +
		"- **Tipo:** feature\n" +
		"- **Estado:** " + status + "\n" +
		"- **Inicio:** x\n" +
		"- **Agente activo:** implementer\n"
	if err := os.WriteFile(p.CurrentFile(), []byte(current), 0o644); err != nil {
		t.Fatal(err)
	}
	return p, current
}

func TestTransitionValidaYAlinea(t *testing.T) {
	p, _ := newProjectFixture(t, "2026-01-01_00-00__demo", meta.StatusReady, true)

	from, forced, err := Transition(p, "2026-01-01_00-00__demo", meta.StatusInProgress, false)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if from != meta.StatusReady || forced {
		t.Fatalf("from=%q forced=%v, want ready/false", from, forced)
	}
	m, err := meta.Load(p.MetaFile("2026-01-01_00-00__demo"))
	if err != nil {
		t.Fatal(err)
	}
	if m.Status != meta.StatusInProgress {
		t.Fatalf("meta.json status = %q, want in_progress", m.Status)
	}
	s, err := ReadSession(p)
	if err != nil {
		t.Fatal(err)
	}
	if s.State != meta.StatusInProgress {
		t.Fatalf("current.md Estado = %q, want in_progress", s.State)
	}
}

func TestTransitionNoPermitidaNoEscribe(t *testing.T) {
	const id = "2026-01-01_00-00__demo"
	p, _ := newProjectFixture(t, id, meta.StatusPending, true)

	_, _, err := Transition(p, id, meta.StatusDone, false)
	if err != ErrTransitionNotAllowed {
		t.Fatalf("err = %v, want ErrTransitionNotAllowed", err)
	}
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		t.Fatal(err)
	}
	if m.Status != meta.StatusPending {
		t.Fatalf("meta.json no debería cambiar, status = %q", m.Status)
	}
}

func TestTransitionForzada(t *testing.T) {
	const id = "2026-01-01_00-00__demo"
	p, _ := newProjectFixture(t, id, meta.StatusPending, false)

	from, forced, err := Transition(p, id, meta.StatusDone, true)
	if err != nil {
		t.Fatalf("Transition: %v", err)
	}
	if from != meta.StatusPending || !forced {
		t.Fatalf("from=%q forced=%v, want pending/true", from, forced)
	}
	m, _ := meta.Load(p.MetaFile(id))
	if m.Status != meta.StatusDone {
		t.Fatalf("status = %q, want done", m.Status)
	}
	// No es la sesión activa: current.md no debe alinearse.
	s, _ := ReadSession(p)
	if s.State != meta.StatusPending {
		t.Fatalf("current.md no debería cambiar: Estado=%q", s.State)
	}
}

func TestTransitionMismoEstadoEsNoOp(t *testing.T) {
	const id = "2026-01-01_00-00__demo"
	p, original := newProjectFixture(t, id, meta.StatusReady, true)

	from, forced, err := Transition(p, id, meta.StatusReady, false)
	if err != nil || forced || from != meta.StatusReady {
		t.Fatalf("Transition no-op = (%q,%v,%v)", from, forced, err)
	}
	after, _ := os.ReadFile(p.CurrentFile())
	if string(after) != original {
		t.Fatal("current.md no debería cambiar en una transición no-op")
	}
}

func TestTransitionEstadoInvalido(t *testing.T) {
	const id = "2026-01-01_00-00__demo"
	p, _ := newProjectFixture(t, id, meta.StatusReady, true)
	if _, _, err := Transition(p, id, "no-existe", false); err == nil {
		t.Fatal("se esperaba error por estado inválido")
	}
}
