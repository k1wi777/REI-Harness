package gitx

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/k1wi777/REI-Harness/internal/meta"
	"github.com/k1wi777/REI-Harness/internal/paths"
)

func TestFilterPaths(t *testing.T) {
	in := []string{
		"src/a.go",
		"AGENTS.md",
		".rei/progress/current.md",
		".rei/progress/work-items/x/impl.md",
		".rei/specs/2026-10-01_10-00__demo/tasks.md",
		".rei/specs/otro/tasks.md",
		".rei/docs/harness/specs.md",
		".rei/agents/leader.md",
		".rei/adapters/opencode/map.json",
		".rei/templates/current.md",
		".rei/config.json",
		"README.md",
	}
	got := filterPaths(in)
	want := []string{
		"src/a.go",
		"AGENTS.md",
		".rei/specs/2026-10-01_10-00__demo/tasks.md",
		".rei/specs/otro/tasks.md",
		".rei/docs/harness/specs.md",
		".rei/agents/leader.md",
		".rei/adapters/opencode/map.json",
		".rei/templates/current.md",
		".rei/config.json",
		"README.md",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterPaths = %v, want %v", got, want)
	}
}

// gitRun ejecuta git en dir con una identidad fija (para poder commitear).
func gitRun(t *testing.T, dir string, args ...string) {
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

// gitRepoBase crea un repo git temporal con dos archivos rastreados
// comprometidos y devuelve la raíz y el SHA del commit base.
func gitRepoBase(t *testing.T) (root, sha string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("sin git")
	}
	root = t.TempDir()
	gitRun(t, root, "init", "-q")
	if err := os.WriteFile(filepath.Join(root, "foo.md"), []byte("uno\ndos\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "stable.md"), []byte("fijo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-q", "-m", "base")
	return root, HeadSHA(root)
}

// TestReviewBase cubre la resolución de base: last_review_commit, respaldo
// base_commit y ausencia/invalidez de base.
func TestReviewBase(t *testing.T) {
	root, sha := gitRepoBase(t)
	p := &paths.Project{Root: root}
	id := "2026-01-01_00-00__demo"
	if err := os.MkdirAll(p.SpecDir(id), 0o755); err != nil {
		t.Fatal(err)
	}
	save := func(m *meta.Meta) {
		t.Helper()
		if err := m.Save(p.MetaFile(id)); err != nil {
			t.Fatal(err)
		}
	}

	save(&meta.Meta{ID: id, BaseCommit: sha})
	base, origin, ok, err := ReviewBase(p, id)
	if err != nil || !ok || base != sha || origin != "base_commit" {
		t.Fatalf("base_commit: got %q %q ok=%v err=%v", base, origin, ok, err)
	}

	save(&meta.Meta{ID: id, BaseCommit: sha, LastReviewCommit: sha})
	base, origin, ok, err = ReviewBase(p, id)
	if err != nil || !ok || base != sha || origin != "last_review_commit" {
		t.Fatalf("last_review_commit: got %q %q ok=%v err=%v", base, origin, ok, err)
	}

	save(&meta.Meta{ID: id})
	if _, _, ok, err := ReviewBase(p, id); err != nil || ok {
		t.Fatalf("sin base: ok=%v err=%v", ok, err)
	}

	save(&meta.Meta{ID: id, BaseCommit: "0000000000000000000000000000000000000000"})
	if _, _, ok, err := ReviewBase(p, id); err != nil || ok {
		t.Fatalf("commit inválido: ok=%v err=%v", ok, err)
	}
}

// TestFileChanges cubre el diff por archivo: cambiado, sin cambios, nuevo
// (sin rastrear) y ausente.
func TestFileChanges(t *testing.T) {
	root, sha := gitRepoBase(t)
	p := &paths.Project{Root: root}
	if err := os.WriteFile(filepath.Join(root, "foo.md"), []byte("uno\nDOS\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bar.md"), []byte("a\nb\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := FileChanges(p, sha, []string{"foo.md", "stable.md", "bar.md", "ghost.md"})
	if err != nil {
		t.Fatal(err)
	}
	if fc := got["foo.md"]; !fc.Changed || fc.New || fc.Added != 1 || fc.Removed != 1 || fc.Diff == "" {
		t.Errorf("foo.md = %+v", fc)
	}
	if fc := got["stable.md"]; fc.Changed || fc.New || fc.Diff != "" {
		t.Errorf("stable.md = %+v", fc)
	}
	if fc := got["bar.md"]; !fc.Changed || !fc.New || fc.Added != 2 || fc.Removed != 0 || fc.Diff == "" {
		t.Errorf("bar.md = %+v", fc)
	}
	if fc := got["ghost.md"]; fc.Changed || fc.New || fc.Diff != "" {
		t.Errorf("ghost.md = %+v", fc)
	}
}
