package adapter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/REI-Harness/internal/paths"
)

const testOpenCodeToolsJSON = `{
  "read": { "permission": ["read"] },
  "write": { "permission": ["edit"] },
  "edit": { "permission": ["edit"] },
  "search": { "permission": ["glob", "grep"] },
  "shell": { "permission": ["bash"] },
  "subagent": { "permission": ["task"] }
}
`

const testOpenCodeAgentTmpl = `---
description: {{.Description}}
mode: {{.Mode}}
{{if .Model}}model: {{.Model}}
{{end}}permission:
{{range .Permissions}}  {{.Key}}: {{.Value}}
{{end}}---

{{.Mark}}

{{.Contract}}
`

// testCommandTmpl es la plantilla genérica del comando `/personalize` usada por
// los tests de ambos runtimes: solo consume la marca GENERATED.
const testCommandTmpl = `---
description: Comando de prueba.
---

{{.Mark}}

Cuerpo del comando de prueba.
`

// setupRoles crea los 5 roles canónicos en un proyecto temporal.
func setupRoles(t *testing.T, p *paths.Project) {
	t.Helper()
	agentsDir := filepath.Join(p.ReiDir(), "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range Roles {
		mode, tools := "subagent", "[read, write, edit, search, shell]"
		if name == "leader" {
			mode, tools = "primary", "[read, search, shell, write, edit, subagent]"
		}
		writeTestFile(t, filepath.Join(agentsDir, name+".md"), canonicalRole(name, mode, tools))
	}
}

// writeAdapter instala tools.json, agent.tmpl y command.tmpl de un runtime en p.
func writeAdapter(t *testing.T, p *paths.Project, rt runtime, toolsJSON, tmpl string) {
	t.Helper()
	dir := adapterDir(p, rt)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(dir, "tools.json"), toolsJSON)
	writeTestFile(t, filepath.Join(dir, "agent.tmpl"), tmpl)
	writeTestFile(t, filepath.Join(dir, "command.tmpl"), testCommandTmpl)
}

// setupProject crea un proyecto temporal con los 5 roles canónicos y el
// adaptador OpenCode para ejercitar la generación sin depender del esqueleto.
func setupProject(t *testing.T) *paths.Project {
	t.Helper()
	p := &paths.Project{Root: t.TempDir()}
	setupRoles(t, p)
	writeAdapter(t, p, opencodeRuntime, testOpenCodeToolsJSON, testOpenCodeAgentTmpl)
	return p
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func testToolMap() toolMap {
	return toolMap{
		"read":     {"permission": []string{"read"}},
		"write":    {"permission": []string{"edit"}},
		"edit":     {"permission": []string{"edit"}},
		"search":   {"permission": []string{"glob", "grep"}},
		"shell":    {"permission": []string{"bash"}},
		"subagent": {"permission": []string{"task"}},
	}
}

func TestResolveTargetsOrdenYDedup(t *testing.T) {
	got, err := resolveTargets([]string{"read", "write", "edit", "search", "shell"}, testToolMap(), opencodeRuntime)
	if err != nil {
		t.Fatalf("resolveTargets: %v", err)
	}
	want := []string{"read", "edit", "glob", "grep", "bash"}
	if len(got) != len(want) {
		t.Fatalf("destinos = %v, esperaba %v", got, want)
	}
	for i, target := range got {
		if target != want[i] {
			t.Errorf("destinos[%d] = %q, esperaba %q", i, target, want[i])
		}
	}
}

func TestResolveTargetsSubagent(t *testing.T) {
	got, err := resolveTargets([]string{"subagent"}, testToolMap(), opencodeRuntime)
	if err != nil {
		t.Fatalf("resolveTargets: %v", err)
	}
	if len(got) != 1 || got[0] != "task" {
		t.Fatalf("subagent debe mapear a task: %+v", got)
	}
}

func TestValidateToolMapIncompleto(t *testing.T) {
	tm := testToolMap()
	delete(tm, "subagent")
	if err := validateToolMap(tm, opencodeRuntime); err == nil {
		t.Fatal("se esperaba error por mapa de herramientas incompleto")
	}
}

func TestValidateToolMapDestinoInvalido(t *testing.T) {
	tm := testToolMap()
	tm["read"] = toolEntry{"permission": []string{"superpoder"}}
	if err := validateToolMap(tm, opencodeRuntime); err == nil {
		t.Fatal("se esperaba error por clave de permiso desconocida")
	}
}

func TestGenerateAllCreaArchivos(t *testing.T) {
	p := setupProject(t)
	var buf bytes.Buffer

	if code := generateAll(p, opencodeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d, esperaba 0:\n%s", code, buf.String())
	}
	for _, name := range Roles {
		got := readTestFile(t, agentPath(p, opencodeRuntime, name))
		if !strings.Contains(got, generatedMark) {
			t.Errorf("%s.md no lleva la marca GENERATED", name)
		}
		if !strings.Contains(got, "## Contrato") || !strings.Contains(got, "### Identidad") {
			t.Errorf("%s.md no contiene el Contrato", name)
		}
		if !strings.Contains(got, "permission:") || !strings.Contains(got, "  read: allow") {
			t.Errorf("%s.md no contiene el frontmatter nativo:\n%s", name, got)
		}
	}
	leader := readTestFile(t, agentPath(p, opencodeRuntime, "leader"))
	if !strings.Contains(leader, "mode: primary") || !strings.Contains(leader, "  task: allow") {
		t.Errorf("leader.md debe ser primary y permitir task:\n%s", leader)
	}
}

func TestGenerateAllIdempotente(t *testing.T) {
	p := setupProject(t)
	var first bytes.Buffer
	if code := generateAll(p, opencodeRuntime, &first); code != 0 {
		t.Fatalf("primera ejecución = %d:\n%s", code, first.String())
	}
	before := readTestFile(t, agentPath(p, opencodeRuntime, "implementer"))

	var second bytes.Buffer
	if code := generateAll(p, opencodeRuntime, &second); code != 0 {
		t.Fatalf("segunda ejecución = %d:\n%s", code, second.String())
	}
	after := readTestFile(t, agentPath(p, opencodeRuntime, "implementer"))
	if before != after {
		t.Error("la segunda ejecución alteró el archivo generado")
	}
	if strings.Contains(second.String(), "[UPD]") || strings.Contains(second.String(), "[WARN]") {
		t.Errorf("la segunda ejecución debe ser idempotente:\n%s", second.String())
	}
}

func TestGenerateAllNoSobrescribeSinMarca(t *testing.T) {
	p := setupProject(t)
	custom := "# agente manual del usuario\n"
	dest := agentPath(p, opencodeRuntime, "leader")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, dest, custom)

	var buf bytes.Buffer
	if code := generateAll(p, opencodeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d, esperaba 0:\n%s", code, buf.String())
	}
	if got := readTestFile(t, dest); got != custom {
		t.Errorf("el archivo sin marca fue sobrescrito:\n%s", got)
	}
	if !strings.Contains(buf.String(), "[WARN]") {
		t.Errorf("debe avisar del archivo sin marca:\n%s", buf.String())
	}
	if _, err := os.Stat(agentPath(p, opencodeRuntime, "reviewer")); err != nil {
		t.Errorf("los demás roles deben generarse: %v", err)
	}
}

func TestCheckDetectaDeriva(t *testing.T) {
	p := setupProject(t)
	var buf bytes.Buffer
	if code := generateAll(p, opencodeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d:\n%s", code, buf.String())
	}

	buf.Reset()
	if code := check(p, opencodeRuntime, &buf); code != 0 {
		t.Fatalf("check con todo generado = %d, esperaba 0:\n%s", code, buf.String())
	}

	// Falta un archivo.
	if err := os.Remove(agentPath(p, opencodeRuntime, "reviewer")); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if code := check(p, opencodeRuntime, &buf); code != 1 {
		t.Fatalf("check con un archivo ausente = %d, esperaba 1:\n%s", code, buf.String())
	}

	// Contenido alterado.
	dest := agentPath(p, opencodeRuntime, "implementer")
	writeTestFile(t, dest, readTestFile(t, dest)+"\nlínea manual\n")
	buf.Reset()
	if code := check(p, opencodeRuntime, &buf); code != 1 {
		t.Fatalf("check con contenido alterado = %d, esperaba 1:\n%s", code, buf.String())
	}
}

// commandPath devuelve la ruta destino del comando `/personalize` de un runtime.
func commandPath(p *paths.Project, rt runtime) string {
	return filepath.Join(p.Root, filepath.FromSlash(rt.artifacts[0].rel))
}

func TestGenerateAllCreaCommand(t *testing.T) {
	p := setupProject(t)
	var buf bytes.Buffer
	if code := generateAll(p, opencodeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d:\n%s", code, buf.String())
	}
	got := readTestFile(t, commandPath(p, opencodeRuntime))
	if !strings.Contains(got, generatedMark) {
		t.Errorf("el comando no lleva la marca GENERATED:\n%s", got)
	}
	if !strings.Contains(got, "Cuerpo del comando de prueba.") {
		t.Errorf("el comando no contiene el cuerpo de la plantilla:\n%s", got)
	}
}

func TestGenerateAllNoSobrescribeCommandSinMarca(t *testing.T) {
	p := setupProject(t)
	custom := "# comando manual del usuario\n"
	dest := commandPath(p, opencodeRuntime)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, dest, custom)

	var buf bytes.Buffer
	if code := generateAll(p, opencodeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d, esperaba 0:\n%s", code, buf.String())
	}
	if got := readTestFile(t, dest); got != custom {
		t.Errorf("el comando sin marca fue sobrescrito:\n%s", got)
	}
	if !strings.Contains(buf.String(), "[WARN]") {
		t.Errorf("debe avisar del comando sin marca:\n%s", buf.String())
	}
}

func TestCheckDetectaDerivaCommand(t *testing.T) {
	p := setupProject(t)
	var buf bytes.Buffer
	if code := generateAll(p, opencodeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d:\n%s", code, buf.String())
	}
	if code := check(p, opencodeRuntime, &buf); code != 0 {
		t.Fatalf("check con todo generado = %d:\n%s", code, buf.String())
	}
	if err := os.Remove(commandPath(p, opencodeRuntime)); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if code := check(p, opencodeRuntime, &buf); code != 1 {
		t.Fatalf("check sin el comando = %d, esperaba 1:\n%s", code, buf.String())
	}
}

// TestInstallGeneraComandoYMencionaPersonalize ejercita el camino completo con
// el esqueleto embebido real (plantillas incluidas) en un proyecto temporal.
func TestInstallGeneraComandoYMencionaPersonalize(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	var buf bytes.Buffer
	if code := install(p, opencodeRuntime, &buf); code != 0 {
		t.Fatalf("install = %d:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "/personalize") {
		t.Errorf("el cierre de install debe mencionar /personalize:\n%s", buf.String())
	}
	got := readTestFile(t, commandPath(p, opencodeRuntime))
	if !strings.Contains(got, "agent: initializer") || !strings.Contains(got, "subtask: true") {
		t.Errorf("el comando generado debe declarar agent: initializer y subtask: true:\n%s", got)
	}

	buf.Reset()
	if code := check(p, opencodeRuntime, &buf); code != 0 {
		t.Fatalf("check tras install = %d, esperaba 0:\n%s", code, buf.String())
	}
}
