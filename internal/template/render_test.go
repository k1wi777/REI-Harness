package template

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/REI-Harness/internal/meta"
	"github.com/k1wi777/REI-Harness/internal/paths"
)

func TestRenderFile(t *testing.T) {
	dir := t.TempDir()
	tpl := filepath.Join(dir, "t.md")
	if err := os.WriteFile(tpl, []byte("Hola {{NAME}} y {{OTRO}}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "out.md")
	if err := RenderFile(tpl, target, map[string]string{"NAME": "mundo"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(target)
	if got := string(data); !strings.Contains(got, "Hola mundo y {{OTRO}}") {
		t.Fatalf("got %q", got)
	}
}

// newProject prepara un proyecto temporal con la plantilla de current.md y el
// esqueleto de progreso.
func newProject(t *testing.T) *paths.Project {
	t.Helper()
	p := &paths.Project{Root: t.TempDir()}
	for _, dir := range []string{p.TemplatesDir(), p.ProgressDir()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	tpl := "- **Work Item:** {{WORK_ITEM}}\n- **Estado:** {{STATE}}\n"
	if err := os.WriteFile(filepath.Join(p.TemplatesDir(), "current.md"), []byte(tpl), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func writeMeta(t *testing.T, p *paths.Project, id, status string) {
	t.Helper()
	if err := os.MkdirAll(p.SpecDir(id), 0o755); err != nil {
		t.Fatal(err)
	}
	m := meta.Meta{ID: id, Type: "task", Status: status}
	if err := m.Save(p.MetaFile(id)); err != nil {
		t.Fatal(err)
	}
}

func writeSession(t *testing.T, p *paths.Project, id, status string) {
	t.Helper()
	content := "- **Work Item:** " + id + "\n- **Estado:** " + status + "\n"
	if err := os.WriteFile(p.CurrentFile(), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestStartAndArchive(t *testing.T) {
	p := newProject(t)
	id := "2026-10-01_10-00__x"
	writeMeta(t, p, id, meta.StatusPending)

	if err := StartSession(p, id, "task"); err != nil {
		t.Fatal(err)
	}
	cur, _ := os.ReadFile(p.CurrentFile())
	if !strings.Contains(string(cur), id) {
		t.Fatalf("StartSession: %s", cur)
	}

	// La sesión debe estar cerrada (done + current.md coherente) para archivar.
	writeMeta(t, p, id, meta.StatusDone)
	aligned := strings.Replace(string(cur), "**Estado:** pending", "**Estado:** done", 1)
	if err := os.WriteFile(p.CurrentFile(), []byte(aligned), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ArchiveSession(p)
	if err != nil {
		t.Fatal(err)
	}
	if got != id {
		t.Fatalf("ArchiveSession id = %q", got)
	}
	cur, _ = os.ReadFile(p.CurrentFile())
	if strings.Contains(string(cur), id) {
		t.Fatal("current.md no se restableció")
	}
	hist, _ := os.ReadFile(p.HistoryFile())
	if !strings.Contains(string(hist), id) {
		t.Fatal("history.md no contiene la entrada")
	}
	if !strings.Contains(string(hist), "**Estado:** done") {
		t.Fatalf("history.md no refleja el estado final: %s", hist)
	}
	if strings.Contains(string(hist), "# Sesión actual") {
		t.Fatal("history.md no debe incluir el encabezado de la plantilla")
	}
}

// TestStartSessionRechazos comprueba la validación de existencia, tipo y estado
// antes de escribir: todo rechazo deja current.md intacto (R17).
func TestStartSessionRechazos(t *testing.T) {
	const id = "2026-10-01_10-00__x"
	const seed = "# Sesión actual\n\n- **Work Item:** _ninguno_\n"

	cases := []struct {
		name       string
		metaStatus string
		metaType   string
		createMeta bool
		reqType    string
		wantText   string
	}{
		{"inexistente", "", "", false, "task", "no existe"},
		{"tipo distinto", meta.StatusPending, "feature", true, "task", "tipo"},
		{"ready", meta.StatusReady, "task", true, "task", "planificación"},
		{"in_progress", meta.StatusInProgress, "task", true, "task", "retoma"},
		{"review", meta.StatusReview, "task", true, "task", "retoma"},
		{"changes_requested", meta.StatusChangesRequested, "task", true, "task", "retoma"},
		{"done", meta.StatusDone, "task", true, "task", "cerrado"},
		{"blocked", meta.StatusBlocked, "task", true, "task", "cerrado"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := newProject(t)
			if tc.createMeta {
				writeMeta(t, p, id, tc.metaStatus)
				m, err := meta.Load(p.MetaFile(id))
				if err != nil {
					t.Fatal(err)
				}
				m.Type = tc.metaType
				if err := m.Save(p.MetaFile(id)); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(p.CurrentFile(), []byte(seed), 0o644); err != nil {
				t.Fatal(err)
			}

			err := StartSession(p, id, tc.reqType)
			if err == nil {
				t.Fatal("esperaba rechazo")
			}
			for _, want := range []string{tc.wantText, "Siguiente paso"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %q, debe contener %q", err, want)
				}
			}
			after, readErr := os.ReadFile(p.CurrentFile())
			if readErr != nil {
				t.Fatal(readErr)
			}
			if string(after) != seed {
				t.Fatalf("current.md cambió: %q", after)
			}
		})
	}
}

func TestArchiveSessionGuard(t *testing.T) {
	const id = "2026-10-01_10-00__x"
	const historySeed = "# Bitácora\n"

	tests := []struct {
		name       string
		metaStatus string
		metaExists bool
		estado     string // Estado de current.md; "" usa metaStatus
		wantErr    bool
		wantText   string
	}{
		{name: "done archiva", metaStatus: meta.StatusDone, metaExists: true},
		{name: "blocked archiva", metaStatus: meta.StatusBlocked, metaExists: true},
		{name: "review no archiva", metaStatus: meta.StatusReview, metaExists: true, wantErr: true, wantText: "review"},
		{name: "in_progress no archiva", metaStatus: meta.StatusInProgress, metaExists: true, wantErr: true, wantText: "in_progress"},
		{name: "ready no archiva", metaStatus: meta.StatusReady, metaExists: true, wantErr: true, wantText: "ready"},
		{name: "pending no archiva", metaStatus: meta.StatusPending, metaExists: true, wantErr: true, wantText: "pending"},
		{name: "changes_requested no archiva", metaStatus: meta.StatusChangesRequested, metaExists: true, wantErr: true, wantText: "changes_requested"},
		{name: "current.md incoherente done/review", metaStatus: meta.StatusDone, metaExists: true, estado: meta.StatusReview, wantErr: true, wantText: "review"},
		{name: "current.md incoherente review/done", metaStatus: meta.StatusReview, metaExists: true, estado: meta.StatusDone, wantErr: true, wantText: "done"},
		{name: "sin meta.json", metaStatus: "", metaExists: false, wantErr: true, wantText: "meta.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newProject(t)
			if tt.metaExists {
				writeMeta(t, p, id, tt.metaStatus)
			}
			estado := tt.estado
			if estado == "" {
				estado = tt.metaStatus
			}
			writeSession(t, p, id, estado)
			if err := os.WriteFile(p.HistoryFile(), []byte(historySeed), 0o644); err != nil {
				t.Fatal(err)
			}
			before, _ := os.ReadFile(p.CurrentFile())

			got, err := ArchiveSession(p)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("ArchiveSession: %v", err)
				}
				if got != id {
					t.Fatalf("ArchiveSession id = %q, want %q", got, id)
				}
				return
			}

			if !errors.Is(err, ErrArchiveNotAllowed) {
				t.Fatalf("err = %v, want ErrArchiveNotAllowed", err)
			}
			if got != "" {
				t.Fatalf("ArchiveSession id = %q, want \"\"", got)
			}
			if !strings.Contains(err.Error(), tt.wantText) {
				t.Fatalf("err = %q, debe mencionar %q", err, tt.wantText)
			}
			hist, _ := os.ReadFile(p.HistoryFile())
			if string(hist) != historySeed {
				t.Fatalf("history.md cambió: %q", hist)
			}
			after, _ := os.ReadFile(p.CurrentFile())
			if string(after) != string(before) {
				t.Fatalf("current.md cambió: %q", after)
			}
		})
	}
}

func TestArchiveSessionEstadoIncoherente(t *testing.T) {
	const id = "2026-10-01_10-00__x"
	p := newProject(t)
	writeMeta(t, p, id, meta.StatusDone)
	writeSession(t, p, id, meta.StatusReview)

	_, err := ArchiveSession(p)
	if !errors.Is(err, ErrArchiveNotAllowed) {
		t.Fatalf("err = %v, want ErrArchiveNotAllowed", err)
	}
	for _, want := range []string{meta.StatusReview, meta.StatusDone} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("err = %q, debe mencionar %q", err, want)
		}
	}
}

func TestArchiveSessionSinSesionActiva(t *testing.T) {
	p := newProject(t)
	if err := os.WriteFile(p.CurrentFile(), []byte("- **Work Item:** _ninguno_\n- **Estado:** _—_\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ArchiveSession(p)
	if err != nil || got != "" {
		t.Fatalf("ArchiveSession = (%q, %v), want (\"\", nil)", got, err)
	}
	if _, err := os.Stat(p.HistoryFile()); !os.IsNotExist(err) {
		t.Fatal("history.md no debe crearse sin sesión activa")
	}
}
