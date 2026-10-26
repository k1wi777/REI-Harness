package view

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/REI-Harness/internal/meta"
	"github.com/k1wi777/REI-Harness/internal/paths"
)

// viewGit ejecuta git en dir con una identidad fija (para poder commitear).
func viewGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func headSHAView(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// newGitViewFixture construye un Work Item dentro de un repo git con la base de
// revisión fijada, requirements.md modificado y design.md nuevo. Se salta el
// test si no hay git.
func newGitViewFixture(t *testing.T) (*paths.Project, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("sin git")
	}
	p, id := newViewFixture(t, meta.StatusInProgress)
	// review.md queda rastreado y sin cambios tras el commit base (R5).
	writeTestFile(t, filepath.Join(p.WorkItemDir(id), "review.md"), "# Review\n\nsin cambios\n")
	viewGit(t, p.Root, "init", "-q")
	viewGit(t, p.Root, "add", "-A")
	viewGit(t, p.Root, "commit", "-q", "-m", "base")

	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		t.Fatal(err)
	}
	m.BaseCommit = headSHAView(t, p.Root)
	if err := m.Save(p.MetaFile(id)); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(p.SpecDir(id), "requirements.md"), "# Requisitos\n\n- uno\n- DOS\n")
	writeTestFile(t, filepath.Join(p.SpecDir(id), "design.md"), "# Diseño\n\nnuevo\n")
	return p, id
}

// TestBuildModelSinGit verifica la degradación sin repositorio git (R7).
func TestBuildModelSinGit(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusInProgress)
	vm, err := buildModel(p, id)
	if err != nil {
		t.Fatal(err)
	}
	if vm.GitAvailable || vm.HasBase {
		t.Errorf("sin git debe degradar: GitAvailable=%v HasBase=%v", vm.GitAvailable, vm.HasBase)
	}
	if vm.Docs == nil {
		t.Errorf("la página debe seguir mostrando documentos")
	}
}

// TestBuildModelConCambios verifica la composición de cambios por documento y
// su render (R1–R5, R8).
func TestBuildModelConCambios(t *testing.T) {
	p, id := newGitViewFixture(t)
	vm, err := buildModel(p, id)
	if err != nil {
		t.Fatal(err)
	}
	if !vm.GitAvailable || !vm.HasBase {
		t.Fatalf("debe haber base: GitAvailable=%v HasBase=%v", vm.GitAvailable, vm.HasBase)
	}
	if vm.BaseOrigin != "base_commit" || vm.BaseSHA == "" {
		t.Errorf("base inesperada: origin=%q sha=%q", vm.BaseOrigin, vm.BaseSHA)
	}

	byName := map[string]docView{}
	for _, d := range vm.Docs {
		byName[d.Name] = d
	}
	if d := byName["requirements.md"]; !d.Changed || d.New || d.Diff == "" {
		t.Errorf("requirements.md debería cambiar (no ser nuevo): %+v", d)
	}
	if d := byName["design.md"]; !d.Changed || !d.New {
		t.Errorf("design.md debería ser nuevo: %+v", d)
	}
	if d := byName["review.md"]; !d.Present || d.Changed || d.New {
		t.Errorf("review.md no debería cambiar: %+v", d)
	}
	if d := byName["plan.md"]; d.Changed || d.New {
		t.Errorf("plan.md ausente no debería cambiar: %+v", d)
	}

	var page strings.Builder
	if err := renderPage(&page, vm); err != nil {
		t.Fatal(err)
	}
	out := page.String()
	for _, want := range []string{
		`class="doc-body has-changes"`, `class="changes-toggle"`,
		`class="doc-diff"`, "dl-add", "dl-del", "cambios desde", "sin cambios",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("la página debería contener %q", want)
		}
	}
}
