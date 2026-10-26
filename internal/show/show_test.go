package show

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeMeta(t *testing.T, p *paths.Project, m *meta.Meta) {
	t.Helper()
	if err := os.MkdirAll(p.SpecDir(m.ID), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := m.Save(p.MetaFile(m.ID)); err != nil {
		t.Fatal(err)
	}
}

func writeSession(t *testing.T, p *paths.Project, id, typ, status, agent string) {
	t.Helper()
	content := "# Sesión actual\n\n" +
		"- **Work Item:** " + id + "\n" +
		"- **Tipo:** " + typ + "\n" +
		"- **Estado:** " + status + "\n" +
		"- **Agente activo:** " + agent + "\n"
	writeFile(t, p.CurrentFile(), content)
}

func TestWorkItemTaskActiva(t *testing.T) {
	root := t.TempDir()
	p := &paths.Project{Root: root}
	id := "2026-10-02_01-55__demo"
	writeMeta(t, p, &meta.Meta{
		ID:          id,
		Title:       "demo task",
		Description: "desc",
		Type:        "task",
		Status:      "review",
		CreatedAt:   "2026-10-02T01:55:00-05:00",
		BaseCommit:  "abc123",
	})
	writeFile(t, filepath.Join(p.SpecDir(id), "plan.md"), "plan")
	writeFile(t, filepath.Join(p.WorkItemDir(id), "impl.md"), "impl")
	writeSession(t, p, id, "task", "review", "implementer")

	var buf bytes.Buffer
	if code := WorkItem(p, id, &buf); code != 0 {
		t.Fatalf("WorkItem = %d, salida:\n%s", code, buf.String())
	}
	out := buf.String()
	for _, want := range []string{id, "status:", "review", "plan.md", "impl.md", "sí", "Sesión: activa", "Resultado: OK"} {
		if !strings.Contains(out, want) {
			t.Errorf("salida no contiene %q:\n%s", want, out)
		}
	}
}

func TestWorkItemFeatureSinPlanificacion(t *testing.T) {
	root := t.TempDir()
	p := &paths.Project{Root: root}
	id := "2026-10-02_02-00__demo-feature"
	writeMeta(t, p, &meta.Meta{
		ID:          id,
		Title:       "demo feature",
		Description: "desc",
		Type:        "feature",
		Status:      "ready",
		CreatedAt:   "2026-10-02T02:00:00-05:00",
	})

	var buf bytes.Buffer
	if code := WorkItem(p, id, &buf); code != 1 {
		t.Fatalf("WorkItem = %d, esperaba 1, salida:\n%s", code, buf.String())
	}
	out := buf.String()
	for _, want := range []string{"[FAIL]", "requirements.md", "design.md", "tasks.md"} {
		if !strings.Contains(out, want) {
			t.Errorf("salida no contiene %q:\n%s", want, out)
		}
	}
}

func TestWorkItemInexistente(t *testing.T) {
	root := t.TempDir()
	p := &paths.Project{Root: root}

	var buf bytes.Buffer
	if code := WorkItem(p, "no-existe", &buf); code != 1 {
		t.Fatalf("WorkItem = %d, esperaba 1", code)
	}
}

func TestWorkItemMetaInvalido(t *testing.T) {
	root := t.TempDir()
	p := &paths.Project{Root: root}
	id := "2026-10-02_03-00__invalido"
	writeFile(t, p.MetaFile(id), "{esto no es json")

	var buf bytes.Buffer
	if code := WorkItem(p, id, &buf); code != 1 {
		t.Fatalf("WorkItem = %d, esperaba 1", code)
	}
}

// newLinksFixture prepara un Work Item válido en review con links.json opcional.
func newLinksFixture(t *testing.T, linksJSON string) (*paths.Project, string) {
	t.Helper()
	p := &paths.Project{Root: t.TempDir()}
	id := "2026-10-02_04-00__enlaces"
	writeMeta(t, p, &meta.Meta{
		ID:          id,
		Title:       "enlaces",
		Description: "desc",
		Type:        "task",
		Status:      "review",
		CreatedAt:   "2026-10-02T04:00:00-05:00",
	})
	writeFile(t, filepath.Join(p.SpecDir(id), "plan.md"), "plan")
	writeFile(t, filepath.Join(p.WorkItemDir(id), "impl.md"), "impl")
	if linksJSON != "" {
		writeFile(t, filepath.Join(p.SpecDir(id), "links.json"), linksJSON)
	}
	return p, id
}

func TestWorkItemEnlacesValidos(t *testing.T) {
	p, id := newLinksFixture(t, `{
  "provider": "github",
  "repo": "owner/name",
  "issue": 123,
  "url": "https://github.com/owner/name/issues/123",
  "base": {"branch": "dev", "sha": "0a1b2c3d4e5f"},
  "epic": {"slug": "modulo-x", "issue": 100, "url": "https://github.com/owner/name/issues/100"}
}`)
	var buf bytes.Buffer
	if code := WorkItem(p, id, &buf); code != 0 {
		t.Fatalf("WorkItem = %d, salida:\n%s", code, buf.String())
	}
	out := buf.String()
	for _, want := range []string{"Enlaces:", "issue:", "#123", "url:", "base:", "dev @ 0a1b2c3d4e5f", "epic:", "modulo-x (#100)", "https://github.com/owner/name/issues/100"} {
		if !strings.Contains(out, want) {
			t.Errorf("salida no contiene %q:\n%s", want, out)
		}
	}
}

func TestWorkItemSinEnlaces(t *testing.T) {
	p, id := newLinksFixture(t, "")
	var buf bytes.Buffer
	if code := WorkItem(p, id, &buf); code != 0 {
		t.Fatalf("WorkItem = %d, salida:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "Enlaces: (sin links.json)") {
		t.Fatalf("falta la nota de ausencia:\n%s", buf.String())
	}
}

func TestWorkItemEnlacesInvalidos(t *testing.T) {
	cases := []struct {
		name, json, want string
	}{
		{"json roto", `{"provider":`, "JSON"},
		{"forma inválida", `{"provider":"github","repo":"owner","issue":0,"url":"nope"}`, "repo"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, id := newLinksFixture(t, tc.json)
			var buf bytes.Buffer
			if code := WorkItem(p, id, &buf); code != 0 {
				t.Fatalf("WorkItem = %d, el enlace inválido no debe cambiar el código:\n%s", code, buf.String())
			}
			out := buf.String()
			if !strings.Contains(out, "nota: links.json inválido") || !strings.Contains(out, tc.want) {
				t.Fatalf("falta la nota con la causa %q:\n%s", tc.want, out)
			}
		})
	}
}
