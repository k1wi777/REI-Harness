package initwizard

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	reiskel "github.com/k1wi777/my-harness-SDD"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

// newInstalledProject crea un proyecto limpio (esqueleto + estructura de
// estado + manifiesto) y devuelve su raíz.
func newInstalledProject(t *testing.T) *paths.Project {
	t.Helper()
	p := &paths.Project{Root: t.TempDir()}
	var buf bytes.Buffer
	if code := Init(p, &buf); code != 0 {
		t.Fatalf("Init = %d:\n%s", code, buf.String())
	}
	return p
}

func embeddedFile(t *testing.T, rel string) string {
	t.Helper()
	if data, err := reiskel.Skeleton.ReadFile(skeletonPrefix + rel); err == nil {
		return string(data)
	}
	data, err := reiskel.Harness.ReadFile(rel)
	if err != nil {
		t.Fatalf("no se pudo leer el embebido %s: %v", rel, err)
	}
	return string(data)
}

func TestUpdateSinCambios(t *testing.T) {
	p := newInstalledProject(t)
	var buf bytes.Buffer
	if code := Update(p, false, &buf); code != 0 {
		t.Fatalf("Update = %d:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "0 nuevo(s), 0 actualizado(s),") || !strings.Contains(out, "0 omitido(s).") {
		t.Errorf("una instalación limpia no debe crear ni omitir nada:\n%s", out)
	}
}

func TestUpdateIdempotente(t *testing.T) {
	p := newInstalledProject(t)

	var first bytes.Buffer
	if code := Update(p, false, &first); code != 0 {
		t.Fatalf("primera Update = %d:\n%s", code, first.String())
	}
	before := snapshotTree(t, p.Root)

	var second bytes.Buffer
	if code := Update(p, false, &second); code != 0 {
		t.Fatalf("segunda Update = %d:\n%s", code, second.String())
	}
	out := second.String()
	if !strings.Contains(out, "0 nuevo(s), 0 actualizado(s),") || !strings.Contains(out, "0 omitido(s).") {
		t.Errorf("la segunda ejecución debe ser idempotente:\n%s", out)
	}
	after := snapshotTree(t, p.Root)
	if before != after {
		t.Errorf("la segunda Update modificó el árbol:\nantes=%s\ndespués=%s", before, after)
	}
}

func TestUpdateConservaPersonalizacionYModificados(t *testing.T) {
	p := newInstalledProject(t)

	// Doc personalizado (sin marcador).
	conv := readProjectFile(t, p, ".rei/docs/project/conventions.md")
	personalized := strings.ReplaceAll(conv, marker, "personalizado")
	writeProjectFile(t, p, ".rei/docs/project/conventions.md", personalized)

	// Doc no personalizado que difiere del embebido pero conserva el marcador.
	arch := readProjectFile(t, p, ".rei/docs/project/architecture.md")
	writeProjectFile(t, p, ".rei/docs/project/architecture.md", arch+"\nedición con marcador\n")

	// Harness modificado a mano (sin tocar el manifiesto).
	leader := readProjectFile(t, p, ".rei/agents/leader.md")
	writeProjectFile(t, p, ".rei/agents/leader.md", leader+"\n# mod manual\n")
	config := readProjectFile(t, p, ".rei/config.json")
	writeProjectFile(t, p, ".rei/config.json", config+"\n")

	var buf bytes.Buffer
	if code := Update(p, false, &buf); code != 0 {
		t.Fatalf("Update = %d:\n%s", code, buf.String())
	}
	out := buf.String()
	for _, want := range []string{
		"[SKIP]  .rei/docs/project/conventions.md",
		"[SKIP]  .rei/agents/leader.md",
		"[SKIP]  .rei/config.json",
		"[UPD]   .rei/docs/project/architecture.md",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("la salida debe contener %q:\n%s", want, out)
		}
	}

	if got := readProjectFile(t, p, ".rei/docs/project/conventions.md"); got != personalized {
		t.Errorf("el doc personalizado se sobrescribió:\n%s", got)
	}
	if got := readProjectFile(t, p, ".rei/docs/project/architecture.md"); got != embeddedFile(t, ".rei/docs/project/architecture.md") {
		t.Errorf("el doc con marcador debe actualizarse al embebido:\n%s", got)
	}
	if got := readProjectFile(t, p, ".rei/agents/leader.md"); !strings.Contains(got, "# mod manual") {
		t.Errorf("el harness modificado se sobrescribió:\n%s", got)
	}
	if got := readProjectFile(t, p, ".rei/config.json"); !strings.HasSuffix(got, "\n\n") {
		t.Errorf("config.json modificado se sobrescribió:%q", got)
	}
}

func TestUpdateForceSobrescribe(t *testing.T) {
	p := newInstalledProject(t)

	conv := readProjectFile(t, p, ".rei/docs/project/conventions.md")
	writeProjectFile(t, p, ".rei/docs/project/conventions.md", strings.ReplaceAll(conv, marker, "personalizado"))
	leader := readProjectFile(t, p, ".rei/agents/leader.md")
	writeProjectFile(t, p, ".rei/agents/leader.md", leader+"\n# mod manual\n")

	var buf bytes.Buffer
	if code := Update(p, true, &buf); code != 0 {
		t.Fatalf("Update --force = %d:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "0 omitido(s).") {
		t.Errorf("--force no debe omitir archivos:\n%s", buf.String())
	}
	if got := readProjectFile(t, p, ".rei/docs/project/conventions.md"); got != embeddedFile(t, ".rei/docs/project/conventions.md") {
		t.Errorf("--force debe restaurar el doc personalizado:\n%s", got)
	}
	if got := readProjectFile(t, p, ".rei/agents/leader.md"); got != embeddedFile(t, ".rei/agents/leader.md") {
		t.Errorf("--force debe restaurar el harness modificado:\n%s", got)
	}
}

func TestUpdateHarnessConBaselineIntacto(t *testing.T) {
	p := newInstalledProject(t)

	const rel = ".rei/docs/harness/workflow.md"
	old := "versión antigua del harness\n"
	writeProjectFile(t, p, rel, old)
	// Simula que el disco era el baseline instalado (el embebido cambió).
	if err := saveManifest(p, map[string]string{rel: hashBytes([]byte(old))}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if code := Update(p, false, &buf); code != 0 {
		t.Fatalf("Update = %d:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "[UPD]   "+rel) {
		t.Errorf("el harness con baseline intacto debe actualizarse:\n%s", buf.String())
	}
	if got := readProjectFile(t, p, rel); got != embeddedFile(t, rel) {
		t.Errorf("%s debe actualizarse al embebido:\n%s", rel, got)
	}
}

func TestUpdateNoTocaSpecsNiProgreso(t *testing.T) {
	p := newInstalledProject(t)
	writeProjectFile(t, p, ".rei/specs/foo/meta.json", "{\"id\":\"foo\"}\n")
	writeProjectFile(t, p, ".rei/progress/current.md", "estado vivo\n")
	writeProjectFile(t, p, ".rei/progress/work-items/foo/impl.md", "notas\n")

	before := snapshotTree(t, filepath.Join(p.Root, ".rei", "specs")) + snapshotTree(t, filepath.Join(p.Root, ".rei", "progress"))

	var buf bytes.Buffer
	if code := Update(p, true, &buf); code != 0 {
		t.Fatalf("Update = %d:\n%s", code, buf.String())
	}
	after := snapshotTree(t, filepath.Join(p.Root, ".rei", "specs")) + snapshotTree(t, filepath.Join(p.Root, ".rei", "progress"))
	if before != after {
		t.Errorf("Update no debe tocar specs/ ni progress/:\nantes=%s\ndespués=%s", before, after)
	}
}

// snapshotTree resume un árbol de directorios como un mapa path->sha256.
func snapshotTree(t *testing.T, root string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		sum := sha256.Sum256(data)
		b.WriteString(rel + "=" + hex.EncodeToString(sum[:]) + "\n")
		return nil
	})
	if err != nil {
		t.Fatalf("snapshotTree(%s): %v", root, err)
	}
	return b.String()
}
