package initwizard

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/REI-Harness/internal/check"
	"github.com/k1wi777/REI-Harness/internal/config"
	"github.com/k1wi777/REI-Harness/internal/paths"
)

func TestInstallSkeletonCreaEsqueleto(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	var buf bytes.Buffer

	if code := InstallSkeleton(p, &buf); code != 0 {
		t.Fatalf("InstallSkeleton = %d, esperaba 0:\n%s", code, buf.String())
	}

	for _, f := range append(append([]string{}, check.RequiredFiles...), ".rei/config.json") {
		if _, err := os.Stat(filepath.Join(p.Root, filepath.FromSlash(f))); err != nil {
			t.Errorf("no se creó %s: %v", f, err)
		}
	}

	for _, d := range []string{".rei/specs", ".rei/progress"} {
		if _, err := os.Stat(filepath.Join(p.Root, filepath.FromSlash(d))); !os.IsNotExist(err) {
			t.Errorf("%s no debe desplegarse desde el esqueleto (err=%v)", d, err)
		}
	}
}

func TestInitConEsqueletoCreaEstructura(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	var buf bytes.Buffer

	if code := Init(p, &buf); code != 0 {
		t.Fatalf("Init = %d, esperaba 0:\n%s", code, buf.String())
	}

	for _, d := range []string{p.SpecsDir(), p.WorkItemsDir()} {
		if fi, err := os.Stat(d); err != nil || !fi.IsDir() {
			t.Errorf("no se creó %s (err=%v)", d, err)
		}
	}
	for _, f := range []string{p.CurrentFile(), p.HistoryFile()} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("no se creó %s (err=%v)", f, err)
		}
	}
}

func TestInstallSkeletonNoSobrescribeYAvisaAgents(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	var buf bytes.Buffer

	if code := InstallSkeleton(p, &buf); code != 0 {
		t.Fatalf("primera ejecución = %d:\n%s", code, buf.String())
	}

	// Personalizaciones manuales que no deben perderse.
	agents := "# AGENTS propio del usuario\n"
	config := "{\n  \"checks\": [{\"id\": \"x\"}]\n}\n"
	architecture := "# Arquitectura propia\n"
	writeProjectFile(t, p, "AGENTS.md", agents)
	writeProjectFile(t, p, ".rei/config.json", config)
	writeProjectFile(t, p, ".rei/docs/project/architecture.md", architecture)

	buf.Reset()
	if code := InstallSkeleton(p, &buf); code != 0 {
		t.Fatalf("segunda ejecución = %d, esperaba 0:\n%s", code, buf.String())
	}
	out := buf.String()

	if got := readProjectFile(t, p, "AGENTS.md"); got != agents {
		t.Errorf("AGENTS.md fue sobrescrito:\n%s", got)
	}
	if got := readProjectFile(t, p, ".rei/config.json"); got != config {
		t.Errorf(".rei/config.json fue sobrescrito:\n%s", got)
	}
	if got := readProjectFile(t, p, ".rei/docs/project/architecture.md"); got != architecture {
		t.Errorf(".rei/docs/project/architecture.md fue sobrescrito:\n%s", got)
	}
	if !strings.Contains(out, "[WARN]") || !strings.Contains(out, "AGENTS.md") {
		t.Errorf("la omisión de AGENTS.md debe emitir un aviso:\n%s", out)
	}
	if !strings.Contains(out, "[SKIP]") {
		t.Errorf("la segunda ejecución debe reportar archivos omitidos:\n%s", out)
	}
}

func TestInstallSkeletonResumen(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	var buf bytes.Buffer

	if code := InstallSkeleton(p, &buf); code != 0 {
		t.Fatalf("primera ejecución = %d:\n%s", code, buf.String())
	}
	first := buf.String()
	created := strings.Count(first, "[OK]")
	if want := fmt.Sprintf("%d archivo(s) creado(s), 0 omitido(s).", created); !strings.Contains(first, want) {
		t.Errorf("resumen inicial debe contener %q:\n%s", want, first)
	}

	buf.Reset()
	if code := InstallSkeleton(p, &buf); code != 0 {
		t.Fatalf("segunda ejecución = %d:\n%s", code, buf.String())
	}
	second := buf.String()
	skipped := strings.Count(second, "[SKIP]")
	if want := fmt.Sprintf("0 archivo(s) creado(s), %d omitido(s).", skipped); !strings.Contains(second, want) {
		t.Errorf("resumen de la segunda ejecución debe contener %q:\n%s", want, second)
	}
}

func TestInstalacionLimpiaIntegra(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	var buf bytes.Buffer

	if code := Init(p, &buf); code != 0 {
		t.Fatalf("Init = %d, esperaba 0:\n%s", code, buf.String())
	}

	for _, f := range append(append([]string{}, check.RequiredFiles...), ".rei/config.json") {
		if _, err := os.Stat(filepath.Join(p.Root, filepath.FromSlash(f))); err != nil {
			t.Errorf("tras instalar falta %s: %v", f, err)
		}
	}

	// `rei check` no debe reportar faltantes ni fallos de integridad.
	buf.Reset()
	if code := check.Run(p, true, &buf); code != 0 {
		t.Fatalf("check.Run tras instalación limpia = %d:\n%s", code, buf.String())
	}
	if out := buf.String(); strings.Contains(out, "[FAIL]") || strings.Contains(out, "Falta") {
		t.Errorf("check.Run no debe reportar faltantes:\n%s", out)
	}

	// Los marcadores se conservan: la personalización sigue pendiente.
	buf.Reset()
	if code := Status(p, &buf); code != 1 {
		t.Fatalf("Status tras instalación limpia = %d, esperaba 1:\n%s", code, buf.String())
	}
}

func TestInstalacionLimpiaNeutral(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	var buf bytes.Buffer

	if code := Init(p, &buf); code != 0 {
		t.Fatalf("Init = %d, esperaba 0:\n%s", code, buf.String())
	}

	// El esqueleto es neutral: los cuatro documentos de personalización
	// conservan el marcador.
	for _, d := range Docs {
		if got := readProjectFile(t, p, d.Path); !strings.Contains(got, marker) {
			t.Errorf("%s debe conservar el marcador %q tras una instalación limpia", d.Path, marker)
		}
	}

	// Config neutral: sin checks, sin integración GitHub y visor con apertura
	// automática (default aprobado).
	cfgPath := filepath.Join(p.Root, ".rei", "config.json")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if len(cfg.Checks) != 0 {
		t.Errorf("una instalación limpia no debe declarar checks: %+v", cfg.Checks)
	}
	if cfg.Integrations.GitHub != nil {
		t.Errorf("una instalación limpia no debe integrar GitHub: %+v", cfg.Integrations.GitHub)
	}
	if !cfg.Viewer.AutoOpenOnReady {
		t.Error("viewer.autoOpenOnReady debe quedar activado")
	}

	pending, err := Pending(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != len(Docs) {
		t.Errorf("esperaba %d documentos pendientes, obtuve %d", len(Docs), len(pending))
	}

	if out := buf.String(); !strings.Contains(out, "/personalize") || !strings.Contains(out, InitializerPath) {
		t.Errorf("Init debe indicar /personalize y el rol initializer:\n%s", out)
	}
}

func TestResolveRootFromConReiAncestro(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".rei"), 0o755); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}

	p, err := ResolveRootFrom(child)
	if err != nil {
		t.Fatalf("ResolveRootFrom(%s): %v", child, err)
	}
	if p.Root != root {
		t.Errorf("ResolveRootFrom = %q, esperaba %q", p.Root, root)
	}
}

func TestResolveRootFromSinRei(t *testing.T) {
	dir := t.TempDir()

	p, err := ResolveRootFrom(dir)
	if err != nil {
		t.Fatalf("ResolveRootFrom(%s): %v", dir, err)
	}
	if p.Root != dir {
		t.Errorf("ResolveRootFrom = %q, esperaba el propio directorio %q", p.Root, dir)
	}
}

func readProjectFile(t *testing.T, p *paths.Project, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(p.Root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
