package rei

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/REI-Harness/internal/config"
)

// projectDocs son los documentos de proyecto que el esqueleto despliega en la
// raíz del proyecto.
var projectDocs = []string{
	"skeleton/.rei/docs/project/README.md",
	"skeleton/.rei/docs/project/architecture.md",
	"skeleton/.rei/docs/project/conventions.md",
	"skeleton/.rei/docs/project/verification.md",
}

func readEmbedded(t *testing.T, fsys fs.FS, path string) string {
	t.Helper()
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		t.Fatalf("no se pudo leer %s: %v", path, err)
	}
	return string(data)
}

func TestSkeletonContieneArchivosRequeridos(t *testing.T) {
	paths := append([]string{"skeleton/AGENTS.md", "skeleton/.rei/config.json"}, projectDocs...)
	for _, path := range paths {
		if _, err := fs.Stat(Skeleton, path); err != nil {
			t.Errorf("Skeleton no contiene %s: %v", path, err)
		}
	}
}

func TestSkeletonNeutralTieneMarcadores(t *testing.T) {
	const marker = "<!-- REI:PENDIENTE -->"
	for _, path := range []string{
		"skeleton/AGENTS.md",
		"skeleton/.rei/docs/project/architecture.md",
		"skeleton/.rei/docs/project/conventions.md",
		"skeleton/.rei/docs/project/verification.md",
	} {
		if got := readEmbedded(t, Skeleton, path); !strings.Contains(got, marker) {
			t.Errorf("%s debe contener %q (esqueleto neutral)", path, marker)
		}
	}
}

func TestSkeletonConfigNeutral(t *testing.T) {
	data, err := fs.ReadFile(Skeleton, "skeleton/.rei/config.json")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("config.Load del esqueleto: %v", err)
	}
	if len(cfg.Checks) != 0 {
		t.Errorf("el config del esqueleto no debe declarar checks: %+v", cfg.Checks)
	}
	if cfg.Integrations.GitHub != nil {
		t.Errorf("el config del esqueleto no debe integrar GitHub: %+v", cfg.Integrations.GitHub)
	}
	if !cfg.Viewer.AutoOpenOnReady {
		t.Error("viewer.autoOpenOnReady debe estar activado en el esqueleto")
	}
}

func TestHarnessContieneMaterialDelArnes(t *testing.T) {
	for _, path := range []string{
		".rei/docs/harness/workflow.md",
		".rei/docs/harness/integrations/github.md",
		".rei/docs/harness/integrations/github/core.md",
		".rei/docs/harness/integrations/github/backlog.md",
		".rei/docs/harness/integrations/github/templates.md",
		".rei/docs/harness/integrations/github/errors.md",
		".rei/docs/usage.md",
		".rei/agents/leader.md",
		".rei/agents/initializer.md",
		".rei/templates/current.md",
		".rei/templates/meta.json",
	} {
		if _, err := fs.Stat(Harness, path); err != nil {
			t.Errorf("Harness no contiene %s: %v", path, err)
		}
	}
}

func TestHarnessContieneAdaptadores(t *testing.T) {
	for _, path := range []string{
		".rei/adapters/opencode/agent.tmpl",
		".rei/adapters/opencode/command.tmpl",
		".rei/adapters/opencode/tools.json",
		".rei/adapters/claude/agent.tmpl",
		".rei/adapters/claude/command.tmpl",
		".rei/adapters/claude/tools.json",
	} {
		if _, err := fs.Stat(Harness, path); err != nil {
			t.Errorf("Harness no contiene %s: %v", path, err)
		}
	}
}

func TestHarnessCommandTmplOpenCode(t *testing.T) {
	text := readEmbedded(t, Harness, ".rei/adapters/opencode/command.tmpl")
	for _, want := range []string{"agent: initializer", "subtask: true", "{{.Mark}}"} {
		if !strings.Contains(text, want) {
			t.Errorf("command.tmpl de OpenCode debe contener %q:\n%s", want, text)
		}
	}
}

func TestEmbedExcluyeArchivosVivosDelProyecto(t *testing.T) {
	live := append([]string{"AGENTS.md", ".rei/config.json"},
		".rei/docs/project/README.md",
		".rei/docs/project/architecture.md",
		".rei/docs/project/conventions.md",
		".rei/docs/project/verification.md",
	)
	for name, fsys := range map[string]fs.FS{"Skeleton": Skeleton, "Harness": Harness} {
		for _, path := range live {
			if _, err := fs.Stat(fsys, path); err == nil {
				t.Errorf("%s NO debe embeber %s sin el prefijo del esqueleto", name, path)
			} else if !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("%s: Stat(%s) devolvió un error inesperado: %v", name, path, err)
			}
		}
	}
}

func TestEmbedExcluyeEstadoDelProyecto(t *testing.T) {
	for name, fsys := range map[string]fs.FS{"Skeleton": Skeleton, "Harness": Harness} {
		for _, path := range []string{".rei/specs", ".rei/progress", "skeleton/.rei/specs", "skeleton/.rei/progress"} {
			if _, err := fs.Stat(fsys, path); err == nil {
				t.Errorf("%s NO debe contener %s", name, path)
			} else if !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("%s: Stat(%s) devolvió un error inesperado: %v", name, path, err)
			}
		}
	}
}
