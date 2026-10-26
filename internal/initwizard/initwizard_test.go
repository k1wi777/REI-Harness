package initwizard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/REI-Harness/internal/paths"
)

func writeProjectFile(t *testing.T, p *paths.Project, rel, content string) {
	t.Helper()
	full := filepath.Join(p.Root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func setupTemplates(t *testing.T, p *paths.Project) {
	t.Helper()
	for _, f := range []string{"current.md", "history.md"} {
		writeProjectFile(t, p, filepath.Join(".rei", "templates", f), "plantilla "+f)
	}
}

func TestPendingDeteccion(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeProjectFile(t, p, "AGENTS.md", "# AGENTS\n"+marker+"\n")
	writeProjectFile(t, p, ".rei/docs/project/architecture.md", "# Arquitectura\n")
	writeProjectFile(t, p, ".rei/docs/project/conventions.md", marker+"\n")
	writeProjectFile(t, p, ".rei/docs/project/verification.md", "# Verification\n")

	pending, err := Pending(p)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, d := range pending {
		ids[d.ID] = true
	}
	if !ids["agentes"] {
		t.Error("AGENTS.md con un marcador debe estar pendiente")
	}
	if ids["arquitectura"] {
		t.Error("architecture.md sin marcador no debe estar pendiente")
	}
	if !ids["convenciones"] {
		t.Error("conventions.md con marcador debe estar pendiente")
	}
	if ids["verificacion"] {
		t.Error("verification.md sin marcador no debe estar pendiente")
	}
}

func TestPendingArchivoInexistente(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	pending, err := Pending(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != len(Docs) {
		t.Fatalf("esperaba %d pendientes, obtuve %d", len(Docs), len(pending))
	}
}

func TestInitCreaEstructura(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	setupTemplates(t, p)
	var buf bytes.Buffer
	if code := Init(p, &buf); code != 0 {
		t.Fatalf("Init = %d, salida:\n%s", code, buf.String())
	}
	for _, dir := range []string{p.SpecsDir(), p.WorkItemsDir()} {
		if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
			t.Errorf("no se creó %s (err=%v)", dir, err)
		}
	}
	for _, f := range []string{p.CurrentFile(), p.HistoryFile()} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("no se creó %s (err=%v)", f, err)
		}
	}
	if !strings.Contains(buf.String(), "Plan de pasos:") {
		t.Errorf("Init debe mostrar el plan de pasos:\n%s", buf.String())
	}
}

func TestInitIdempotente(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	setupTemplates(t, p)
	var buf bytes.Buffer
	if code := Init(p, &buf); code != 0 {
		t.Fatalf("primera ejecución = %d, salida:\n%s", code, buf.String())
	}
	if err := os.WriteFile(p.CurrentFile(), []byte("modificado"), 0o644); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if code := Init(p, &buf); code != 0 {
		t.Fatalf("segunda ejecución = %d, salida:\n%s", code, buf.String())
	}
	data, err := os.ReadFile(p.CurrentFile())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "modificado" {
		t.Errorf("Init sobrescribió current.md: %q", string(data))
	}
}

func TestStatusConPendientes(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	var buf bytes.Buffer
	if code := Status(p, &buf); code != 1 {
		t.Fatalf("Status = %d, esperaba 1, salida:\n%s", code, buf.String())
	}
	if _, err := os.Stat(p.SpecsDir()); !os.IsNotExist(err) {
		t.Errorf("Status no debe crear .rei/specs/ (err=%v)", err)
	}
	if _, err := os.Stat(p.CurrentFile()); !os.IsNotExist(err) {
		t.Errorf("Status no debe crear current.md (err=%v)", err)
	}
}

func TestStatusSinPendientes(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	for _, d := range Docs {
		writeProjectFile(t, p, d.Path, "# documento sin marcador\n")
	}
	var buf bytes.Buffer
	if code := Status(p, &buf); code != 0 {
		t.Fatalf("Status = %d, esperaba 0, salida:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "personalizada") {
		t.Errorf("Status debería indicar que está personalizada:\n%s", buf.String())
	}
}

func TestInitMencionaPersonalize(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	setupTemplates(t, p)
	var buf bytes.Buffer
	if code := Init(p, &buf); code != 0 {
		t.Fatalf("Init = %d, salida:\n%s", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "/personalize") {
		t.Errorf("Init debe indicar cómo iniciar la entrevista con /personalize:\n%s", out)
	}
	if !strings.Contains(out, InitializerPath) {
		t.Errorf("Init debe mencionar el rol initializer (%s):\n%s", InitializerPath, out)
	}
}
