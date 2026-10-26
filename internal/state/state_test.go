package state

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

// addWorkItem crea un meta.json adicional en el estado dado.
func addWorkItem(t *testing.T, p *paths.Project, id, status string) {
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

// TestListWorkItemsIncluyeTipo comprueba que el modelo de listado expone el
// tipo de cada meta.json (R3).
func TestListWorkItemsIncluyeTipo(t *testing.T) {
	p, _ := newProjectFixture(t, "2026-01-01_00-00__a", meta.StatusReady, false)
	addWorkItem(t, p, "2026-01-02_00-00__b", meta.StatusDone)
	items, err := ListWorkItems(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
	if items[0].Type != "feature" || items[1].Type != "task" {
		t.Fatalf("tipos = %q, %q", items[0].Type, items[1].Type)
	}
}

func TestInProgressItems(t *testing.T) {
	p, _ := newProjectFixture(t, "2026-01-01_00-00__a", meta.StatusReady, false)
	addWorkItem(t, p, "2026-01-02_00-00__b", meta.StatusInProgress)
	addWorkItem(t, p, "2026-01-03_00-00__c", meta.StatusInProgress)
	addWorkItem(t, p, "2026-01-04_00-00__d", meta.StatusDone)

	ids, err := InProgressItems(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != "2026-01-02_00-00__b" || ids[1] != "2026-01-03_00-00__c" {
		t.Fatalf("InProgressItems = %v, want [b c]", ids)
	}
}

func TestTransitionRechazaOtroInProgress(t *testing.T) {
	const id = "2026-01-01_00-00__demo"
	const otro = "2026-01-02_00-00__otro"
	for _, force := range []bool{false, true} {
		p, original := newProjectFixture(t, id, meta.StatusReady, true)
		addWorkItem(t, p, otro, meta.StatusInProgress)

		_, _, err := Transition(p, id, meta.StatusInProgress, force)
		if !errors.Is(err, ErrAnotherInProgress) {
			t.Fatalf("force=%v: err = %v, want ErrAnotherInProgress", force, err)
		}
		if !strings.Contains(err.Error(), otro) {
			t.Errorf("el error debe indicar el Work Item activo: %v", err)
		}
		m, err := meta.Load(p.MetaFile(id))
		if err != nil {
			t.Fatal(err)
		}
		if m.Status != meta.StatusReady {
			t.Errorf("meta.json no debería cambiar: status = %q", m.Status)
		}
		after, err := os.ReadFile(p.CurrentFile())
		if err != nil {
			t.Fatal(err)
		}
		if string(after) != original {
			t.Error("current.md no debería cambiar al rechazar la transición")
		}
	}
}

func TestTransitionSinOtroInProgressPermitida(t *testing.T) {
	const id = "2026-01-01_00-00__demo"
	p, _ := newProjectFixture(t, id, meta.StatusReady, true)
	addWorkItem(t, p, "2026-01-02_00-00__cerrado", meta.StatusDone)

	if _, _, err := Transition(p, id, meta.StatusInProgress, false); err != nil {
		t.Fatalf("Transition: %v", err)
	}
	m, _ := meta.Load(p.MetaFile(id))
	if m.Status != meta.StatusInProgress {
		t.Fatalf("status = %q, want in_progress", m.Status)
	}
}

// TestTransitionDoneABlocked comprueba que done -> blocked no es estándar: sin
// --force se rechaza sin escribir y con --force se aplica como transición no
// estándar (R14).
func TestTransitionDoneABlocked(t *testing.T) {
	const id = "2026-01-01_00-00__demo"

	t.Run("rechazada sin force", func(t *testing.T) {
		p, original := newProjectFixture(t, id, meta.StatusDone, true)
		_, _, err := Transition(p, id, meta.StatusBlocked, false)
		if !errors.Is(err, ErrTransitionNotAllowed) {
			t.Fatalf("err = %v, want ErrTransitionNotAllowed", err)
		}
		m, _ := meta.Load(p.MetaFile(id))
		if m.Status != meta.StatusDone {
			t.Fatalf("meta.json no debería cambiar: %q", m.Status)
		}
		after, _ := os.ReadFile(p.CurrentFile())
		if string(after) != original {
			t.Fatal("current.md no debería cambiar")
		}
	})

	t.Run("permitida con force", func(t *testing.T) {
		p, _ := newProjectFixture(t, id, meta.StatusDone, true)
		from, forced, err := Transition(p, id, meta.StatusBlocked, true)
		if err != nil {
			t.Fatalf("Transition: %v", err)
		}
		if from != meta.StatusDone || !forced {
			t.Fatalf("from=%q forced=%v, want done/true", from, forced)
		}
		m, _ := meta.Load(p.MetaFile(id))
		if m.Status != meta.StatusBlocked {
			t.Fatalf("status = %q, want blocked", m.Status)
		}
	})
}

func TestTransitionChangesRequestedAInProgress(t *testing.T) {
	const id = "2026-01-01_00-00__demo"

	t.Run("permitida", func(t *testing.T) {
		p, _ := newProjectFixture(t, id, meta.StatusChangesRequested, false)
		if _, _, err := Transition(p, id, meta.StatusInProgress, false); err != nil {
			t.Fatalf("Transition: %v", err)
		}
		m, _ := meta.Load(p.MetaFile(id))
		if m.Status != meta.StatusInProgress {
			t.Fatalf("status = %q, want in_progress", m.Status)
		}
	})

	t.Run("rechazada con otro activo", func(t *testing.T) {
		p, _ := newProjectFixture(t, id, meta.StatusChangesRequested, false)
		addWorkItem(t, p, "2026-01-02_00-00__otro", meta.StatusInProgress)
		_, _, err := Transition(p, id, meta.StatusInProgress, false)
		if !errors.Is(err, ErrAnotherInProgress) {
			t.Fatalf("err = %v, want ErrAnotherInProgress", err)
		}
		m, _ := meta.Load(p.MetaFile(id))
		if m.Status != meta.StatusChangesRequested {
			t.Fatalf("meta.json no debería cambiar: %q", m.Status)
		}
	})
}
