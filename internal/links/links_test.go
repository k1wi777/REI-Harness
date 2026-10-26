package links

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validJSON = `{
  "provider": "github",
  "repo": "owner/name",
  "issue": 123,
  "url": "https://github.com/owner/name/issues/123",
  "base": {"branch": "main", "sha": "0a1b2c3d4e5f"},
  "epic": {"slug": "modulo-x", "issue": 100, "url": "https://github.com/owner/name/issues/100"}
}`

func writeLinks(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "links.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadValido(t *testing.T) {
	l, err := Read(writeLinks(t, validJSON))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if l.Provider != "github" || l.Repo != "owner/name" || l.Issue != 123 {
		t.Fatalf("links inesperados: %+v", l)
	}
	if l.Base == nil || l.Base.Branch != "main" || l.Base.SHA != "0a1b2c3d4e5f" {
		t.Fatalf("base inesperada: %+v", l.Base)
	}
	if l.Epic == nil || l.Epic.Slug != "modulo-x" || l.Epic.Issue != 100 {
		t.Fatalf("epic inesperada: %+v", l.Epic)
	}
	if problems := Problems(l); len(problems) != 0 {
		t.Fatalf("no esperaba problemas: %v", problems)
	}
}

func TestReadMinimoValido(t *testing.T) {
	l, err := Read(writeLinks(t, `{"provider":"github","repo":"o/n","issue":1,"url":"https://github.com/o/n/issues/1","epic":null}`))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if l.Base != nil || l.Epic != nil {
		t.Fatalf("base/epic deben quedar ausentes: %+v", l)
	}
	if problems := Problems(l); len(problems) != 0 {
		t.Fatalf("no esperaba problemas: %v", problems)
	}
}

func TestReadAusente(t *testing.T) {
	_, err := Read(filepath.Join(t.TempDir(), "links.json"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v, want fs.ErrNotExist", err)
	}
}

func TestReadJSONRoto(t *testing.T) {
	_, err := Read(writeLinks(t, `{"provider":`))
	if err == nil {
		t.Fatal("esperaba error de JSON")
	}
	if errors.Is(err, fs.ErrNotExist) {
		t.Fatal("un JSON roto no es un archivo ausente")
	}
	if !strings.Contains(err.Error(), "JSON") {
		t.Fatalf("err = %q, debe mencionar JSON", err)
	}
}

func TestProblemsForma(t *testing.T) {
	base := func() Links {
		return Links{
			Provider: "github", Repo: "owner/name", Issue: 1,
			URL: "https://github.com/owner/name/issues/1",
		}
	}
	cases := []struct {
		name string
		mod  func(*Links)
		want string
	}{
		{"provider vacío", func(l *Links) { l.Provider = "" }, "provider"},
		{"provider distinto", func(l *Links) { l.Provider = "gitlab" }, "no soportado"},
		{"repo con host", func(l *Links) { l.Repo = "https://github.com/owner/name" }, "repo"},
		{"repo una parte", func(l *Links) { l.Repo = "owner" }, "repo"},
		{"repo tres partes", func(l *Links) { l.Repo = "a/b/c" }, "repo"},
		{"repo con .git", func(l *Links) { l.Repo = "owner/name.git" }, "repo"},
		{"issue cero", func(l *Links) { l.Issue = 0 }, "issue"},
		{"url relativa", func(l *Links) { l.URL = "/issues/1" }, "url"},
		{"url sin ruta", func(l *Links) { l.URL = "https://github.com" }, "url"},
		{"url ftp", func(l *Links) { l.URL = "ftp://github.com/o/n/1" }, "url"},
		{"base sin branch", func(l *Links) { l.Base = &Base{SHA: "0a1b2c3"} }, "base.branch"},
		{"base sha corto", func(l *Links) { l.Base = &Base{Branch: "main", SHA: "abc"} }, "base.sha"},
		{"base sha no hex", func(l *Links) { l.Base = &Base{Branch: "main", SHA: "zzzzzzz"} }, "base.sha"},
		{"epic sin slug", func(l *Links) { l.Epic = &Epic{Issue: 1, URL: "https://github.com/o/n/issues/1"} }, "epic.slug"},
		{"epic issue cero", func(l *Links) { l.Epic = &Epic{Slug: "x", URL: "https://github.com/o/n/issues/1"} }, "epic.issue"},
		{"epic url mala", func(l *Links) { l.Epic = &Epic{Slug: "x", Issue: 1, URL: "nope"} }, "epic.url"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := base()
			tc.mod(&l)
			problems := Problems(l)
			if len(problems) == 0 {
				t.Fatal("esperaba al menos un problema")
			}
			if !strings.Contains(strings.Join(problems, "; "), tc.want) {
				t.Fatalf("problemas = %v, esperaba mencionar %q", problems, tc.want)
			}
		})
	}
}

func TestProblemsNoCuentaCamposExtra(t *testing.T) {
	l, err := Read(writeLinks(t, `{"provider":"github","repo":"o/n","issue":1,"url":"https://github.com/o/n/issues/1","extra":true}`))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if problems := Problems(l); len(problems) != 0 {
		t.Fatalf("los campos extra se ignoran: %v", problems)
	}
}

func TestValidRepo(t *testing.T) {
	valid := []string{"owner/name", "k1wi777/REI-Harness", "a_b/c.d-e"}
	for _, s := range valid {
		if !ValidRepo(s) {
			t.Errorf("ValidRepo(%q) = false, want true", s)
		}
	}
	invalid := []string{"", "owner", "a/b/c", "https://github.com/a/b", "a/b.git", " /b", "a/ b", "@/b", "a:/b"}
	for _, s := range invalid {
		if ValidRepo(s) {
			t.Errorf("ValidRepo(%q) = true, want false", s)
		}
	}
}
