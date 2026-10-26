package adapter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	reiskel "github.com/k1wi777/REI-Harness"
	"github.com/k1wi777/REI-Harness/internal/paths"
)

const testClaudeToolsJSON = `{
  "read":     { "tools": ["Read"] },
  "write":    { "tools": ["Write", "Edit"] },
  "edit":     { "tools": ["Write", "Edit"] },
  "search":   { "tools": ["Grep", "Glob"] },
  "shell":    { "tools": ["Bash"] },
  "subagent": { "tools": ["Agent"] }
}
`

const testClaudeAgentTmpl = `---
name: {{.Name}}
description: {{.Description}}
tools: {{.Tools}}
{{if .Model}}model: {{.Model}}
{{end}}---

{{.Mark}}

{{.Contract}}
`

func testClaudeToolMap() toolMap {
	return toolMap{
		"read":     {"tools": []string{"Read"}},
		"write":    {"tools": []string{"Write", "Edit"}},
		"edit":     {"tools": []string{"Write", "Edit"}},
		"search":   {"tools": []string{"Grep", "Glob"}},
		"shell":    {"tools": []string{"Bash"}},
		"subagent": {"tools": []string{"Agent"}},
	}
}

func setupClaudeProject(t *testing.T) *paths.Project {
	t.Helper()
	p := &paths.Project{Root: t.TempDir()}
	setupRoles(t, p)
	writeAdapter(t, p, claudeRuntime, testClaudeToolsJSON, testClaudeAgentTmpl)
	return p
}

func TestClaudeResolveTargetsMapeoYDedup(t *testing.T) {
	got, err := resolveTargets(
		[]string{"read", "write", "edit", "search", "shell", "subagent"},
		testClaudeToolMap(),
		claudeRuntime,
	)
	if err != nil {
		t.Fatalf("resolveTargets: %v", err)
	}
	want := []string{"Read", "Write", "Edit", "Grep", "Glob", "Bash", "Agent"}
	if len(got) != len(want) {
		t.Fatalf("destinos = %v, esperaba %v", got, want)
	}
	for i, target := range got {
		if target != want[i] {
			t.Errorf("destinos[%d] = %q, esperaba %q", i, target, want[i])
		}
	}
}

func TestClaudeResolveTargetsRespetanOrdenDeclarado(t *testing.T) {
	got, err := resolveTargets([]string{"shell", "read", "search"}, testClaudeToolMap(), claudeRuntime)
	if err != nil {
		t.Fatalf("resolveTargets: %v", err)
	}
	want := []string{"Bash", "Read", "Grep", "Glob"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("destinos = %v, esperaba %v", got, want)
	}
}

func TestClaudeValidateToolMapIncompleto(t *testing.T) {
	tm := testClaudeToolMap()
	delete(tm, "subagent")
	if err := validateToolMap(tm, claudeRuntime); err == nil {
		t.Fatal("se esperaba error por mapa de herramientas incompleto")
	}
}

func TestClaudeValidateToolMapDestinoInvalido(t *testing.T) {
	tm := testClaudeToolMap()
	tm["read"] = toolEntry{"tools": []string{"Superpoder"}}
	if err := validateToolMap(tm, claudeRuntime); err == nil {
		t.Fatal("se esperaba error por herramienta nativa desconocida")
	}
}

func TestClaudeGenerateCreaArchivos(t *testing.T) {
	p := setupClaudeProject(t)
	var buf bytes.Buffer
	if code := generateAll(p, claudeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d, esperaba 0:\n%s", code, buf.String())
	}
	for _, name := range Roles {
		got := readTestFile(t, agentPath(p, claudeRuntime, name))
		if !strings.Contains(got, claudeMark) {
			t.Errorf("%s.md no lleva la marca GENERATED de Claude", name)
		}
		if !strings.Contains(got, "name: "+name) {
			t.Errorf("%s.md no declara 'name: %s':\n%s", name, name, got)
		}
		if !strings.Contains(got, "description: ") {
			t.Errorf("%s.md no declara description:\n%s", name, got)
		}
		if !strings.Contains(got, "tools: ") {
			t.Errorf("%s.md no declara tools:\n%s", name, got)
		}
		if !strings.Contains(got, "## Contrato") || !strings.Contains(got, "### Identidad") {
			t.Errorf("%s.md no contiene el Contrato", name)
		}
	}
	implementer := readTestFile(t, agentPath(p, claudeRuntime, "implementer"))
	if !strings.Contains(implementer, "tools: Read, Write, Edit, Grep, Glob, Bash") {
		t.Errorf("implementer.md no tiene el mapeo esperado:\n%s", implementer)
	}
	if strings.Contains(implementer, "model:") {
		t.Errorf("implementer.md no debe declarar model si el rol no lo trae:\n%s", implementer)
	}
}

func TestClaudeModelOpcional(t *testing.T) {
	p := setupClaudeProject(t)
	leaderPath := filepath.Join(p.ReiDir(), "agents", "leader.md")
	data := strings.Replace(
		readTestFile(t, leaderPath),
		"mode: primary\n",
		"mode: primary\nmodel: anthropic/claude\n",
		1,
	)
	writeTestFile(t, leaderPath, data)

	var buf bytes.Buffer
	if code := generateAll(p, claudeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d, esperaba 0:\n%s", code, buf.String())
	}
	leader := readTestFile(t, agentPath(p, claudeRuntime, "leader"))
	if !strings.Contains(leader, "model: anthropic/claude") {
		t.Errorf("leader.md debe declarar el model del rol:\n%s", leader)
	}
}

func TestClaudeGenerateIdempotente(t *testing.T) {
	p := setupClaudeProject(t)
	var first bytes.Buffer
	if code := generateAll(p, claudeRuntime, &first); code != 0 {
		t.Fatalf("primera ejecución = %d:\n%s", code, first.String())
	}
	before := readTestFile(t, agentPath(p, claudeRuntime, "implementer"))

	var second bytes.Buffer
	if code := generateAll(p, claudeRuntime, &second); code != 0 {
		t.Fatalf("segunda ejecución = %d:\n%s", code, second.String())
	}
	after := readTestFile(t, agentPath(p, claudeRuntime, "implementer"))
	if before != after {
		t.Error("la segunda ejecución alteró el archivo generado")
	}
	if strings.Contains(second.String(), "[UPD]") || strings.Contains(second.String(), "[WARN]") {
		t.Errorf("la segunda ejecución debe ser idempotente:\n%s", second.String())
	}
}

func TestClaudeNoSobrescribeSinMarca(t *testing.T) {
	p := setupClaudeProject(t)
	custom := "# agente manual del usuario\n"
	dest := agentPath(p, claudeRuntime, "leader")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, dest, custom)

	var buf bytes.Buffer
	if code := generateAll(p, claudeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d, esperaba 0:\n%s", code, buf.String())
	}
	if got := readTestFile(t, dest); got != custom {
		t.Errorf("el archivo sin marca fue sobrescrito:\n%s", got)
	}
	if !strings.Contains(buf.String(), "[WARN]") {
		t.Errorf("debe avisar del archivo sin marca:\n%s", buf.String())
	}
	if _, err := os.Stat(agentPath(p, claudeRuntime, "reviewer")); err != nil {
		t.Errorf("los demás roles deben generarse: %v", err)
	}
}

func TestClaudeCheckDetectaDeriva(t *testing.T) {
	p := setupClaudeProject(t)
	var buf bytes.Buffer
	if code := generateAll(p, claudeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d:\n%s", code, buf.String())
	}

	buf.Reset()
	if code := check(p, claudeRuntime, &buf); code != 0 {
		t.Fatalf("check con todo generado = %d, esperaba 0:\n%s", code, buf.String())
	}

	// Falta un archivo.
	if err := os.Remove(agentPath(p, claudeRuntime, "reviewer")); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if code := check(p, claudeRuntime, &buf); code != 1 {
		t.Fatalf("check con un archivo ausente = %d, esperaba 1:\n%s", code, buf.String())
	}

	// Contenido alterado.
	dest := agentPath(p, claudeRuntime, "implementer")
	writeTestFile(t, dest, readTestFile(t, dest)+"\nlínea manual\n")
	buf.Reset()
	if code := check(p, claudeRuntime, &buf); code != 1 {
		t.Fatalf("check con contenido alterado = %d, esperaba 1:\n%s", code, buf.String())
	}
}

func TestClaudeAdapterDelEsqueletoEsValido(t *testing.T) {
	for _, path := range []string{
		".rei/adapters/claude/tools.json",
		".rei/adapters/claude/agent.tmpl",
		".rei/adapters/claude/command.tmpl",
	} {
		if _, err := reiskel.Harness.ReadFile(path); err != nil {
			t.Errorf("el esqueleto no contiene %s: %v", path, err)
		}
	}
	data, err := reiskel.Harness.ReadFile(".rei/adapters/claude/tools.json")
	if err != nil {
		t.Fatal(err)
	}
	var tm toolMap
	if err := json.Unmarshal(data, &tm); err != nil {
		t.Fatalf("tools.json del esqueleto inválido: %v", err)
	}
	if err := validateToolMap(tm, claudeRuntime); err != nil {
		t.Errorf("tools.json del esqueleto no válido para el runtime Claude: %v", err)
	}
}

func TestClaudeGenerateCreaCommandYClaudeMD(t *testing.T) {
	p := setupClaudeProject(t)
	var buf bytes.Buffer
	if code := generateAll(p, claudeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d:\n%s", code, buf.String())
	}

	cmd := readTestFile(t, commandPath(p, claudeRuntime))
	if !strings.Contains(cmd, claudeMark) {
		t.Errorf("el comando de Claude no lleva la marca GENERATED:\n%s", cmd)
	}

	md := readTestFile(t, filepath.Join(p.Root, "CLAUDE.md"))
	if !strings.Contains(md, claudeMark) {
		t.Errorf("CLAUDE.md no lleva la marca GENERATED:\n%s", md)
	}
	if !strings.Contains(md, "@AGENTS.md") {
		t.Errorf("CLAUDE.md debe importar AGENTS.md (@AGENTS.md):\n%s", md)
	}
}

func TestClaudeCheckDetectaDerivaCommandYClaudeMD(t *testing.T) {
	p := setupClaudeProject(t)
	var buf bytes.Buffer
	if code := generateAll(p, claudeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d:\n%s", code, buf.String())
	}
	if code := check(p, claudeRuntime, &buf); code != 0 {
		t.Fatalf("check con todo generado = %d:\n%s", code, buf.String())
	}

	if err := os.Remove(filepath.Join(p.Root, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if code := check(p, claudeRuntime, &buf); code != 1 {
		t.Fatalf("check sin CLAUDE.md = %d, esperaba 1:\n%s", code, buf.String())
	}
}

func TestClaudeNoSobrescribeCommandYClaudeMDSinMarca(t *testing.T) {
	p := setupClaudeProject(t)
	custom := "contenido manual del usuario\n"
	cmdDest := commandPath(p, claudeRuntime)
	mdDest := filepath.Join(p.Root, "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(cmdDest), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, cmdDest, custom)
	writeTestFile(t, mdDest, custom)

	var buf bytes.Buffer
	if code := generateAll(p, claudeRuntime, &buf); code != 0 {
		t.Fatalf("generateAll = %d, esperaba 0:\n%s", code, buf.String())
	}
	if got := readTestFile(t, cmdDest); got != custom {
		t.Errorf("el comando sin marca fue sobrescrito:\n%s", got)
	}
	if got := readTestFile(t, mdDest); got != custom {
		t.Errorf("CLAUDE.md sin marca fue sobrescrito:\n%s", got)
	}
	if !strings.Contains(buf.String(), "[WARN]") {
		t.Errorf("debe avisar de los archivos sin marca:\n%s", buf.String())
	}
}
