package initwizard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/REI-Harness/internal/paths"
)

func TestHashBytes(t *testing.T) {
	// SHA-256 de "abc".
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got := hashBytes([]byte("abc")); got != want {
		t.Fatalf("hashBytes(abc) = %q, want %q", got, want)
	}
	// Determinista y sensible al contenido.
	if hashBytes([]byte("abc")) == hashBytes([]byte("abd")) {
		t.Fatal("hashBytes debería diferir para contenidos distintos")
	}
}

func TestSaveLoadManifest(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	entries := map[string]string{
		".rei/agents/leader.md":  "aaaa",
		".rei/docs/harness/x.md": "bbbb",
	}
	if err := saveManifest(p, entries); err != nil {
		t.Fatalf("saveManifest: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p.Root, filepath.FromSlash(ManifestPath))); err != nil {
		t.Fatalf("no se creó %s: %v", ManifestPath, err)
	}
	got, err := loadManifest(p)
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	for k, v := range entries {
		if got[k] != v {
			t.Errorf("loadManifest[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestSaveManifestOrdenadoEIndentado(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	if err := saveManifest(p, map[string]string{"z.md": "1", "a.md": "2"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(p.Root, filepath.FromSlash(ManifestPath)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "\n    \"a.md\"") {
		t.Errorf("el manifiesto debe estar indentado:\n%s", raw)
	}
	if strings.Index(string(raw), `"a.md"`) > strings.Index(string(raw), `"z.md"`) {
		t.Errorf("las claves deben ir ordenadas:\n%s", raw)
	}
	var m manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("manifiesto no es JSON válido: %v", err)
	}
}

func TestLoadManifestInexistente(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	got, err := loadManifest(p)
	if err != nil {
		t.Fatalf("loadManifest sin archivo: %v", err)
	}
	if got != nil {
		t.Fatalf("loadManifest sin archivo = %v, esperaba nil", got)
	}
}

func TestLoadManifestInvalido(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeProjectFile(t, p, ManifestPath, "{no es json")
	got, err := loadManifest(p)
	if err != nil {
		t.Fatalf("loadManifest inválido no debe devolver error: %v", err)
	}
	if got != nil {
		t.Fatalf("loadManifest inválido = %v, esperaba nil", got)
	}
}
