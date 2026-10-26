package adapter

import (
	"io/fs"
	"strings"
	"testing"

	reiskel "github.com/k1wi777/REI-Harness"
)

// canonicalRole construye un rol canónico válido para los tests.
func canonicalRole(name, mode, tools string) string {
	return "---\n" +
		"name: " + name + "\n" +
		"description: Rol " + name + " de prueba.\n" +
		"mode: " + mode + "\n" +
		"tools: " + tools + "\n" +
		"---\n\n" +
		"## Contrato\n\n" +
		"### Identidad\n\nSoy " + name + ".\n\n" +
		"### Objetivo\n\nCumplir el rol.\n\n" +
		"### Precondiciones\n\nNinguna.\n\n" +
		"### Protocolo\n\n1. Primer paso.\n\n" +
		"### Reglas duras\n\n- Una regla.\n\n" +
		"### Formato de salida\n\nok.\n\n" +
		"### Herramientas permitidas\n\n`read`.\n\n" +
		"### Documentos de referencia\n\nAGENTS.md.\n\n" +
		"## Referencia\n\nTexto de referencia opcional.\n"
}

func TestParseRoleValido(t *testing.T) {
	role, err := ParseRole([]byte(canonicalRole("implementer", "subagent", "[read, write, edit, search, shell]")))
	if err != nil {
		t.Fatalf("ParseRole: %v", err)
	}
	if role.Name != "implementer" {
		t.Errorf("Name = %q, esperaba implementer", role.Name)
	}
	if role.Mode != "subagent" {
		t.Errorf("Mode = %q, esperaba subagent", role.Mode)
	}
	if role.Description != "Rol implementer de prueba." {
		t.Errorf("Description = %q", role.Description)
	}
	if len(role.Tools) != 5 {
		t.Errorf("Tools = %v, esperaba 5 elementos", role.Tools)
	}
	if role.Model != "" {
		t.Errorf("Model = %q, esperaba vacío", role.Model)
	}
	if !strings.Contains(role.Contract, "### Identidad") || !strings.Contains(role.Contract, "### Protocolo") {
		t.Errorf("Contract incompleto:\n%s", role.Contract)
	}
	if strings.Contains(role.Contract, "## Referencia") || strings.Contains(role.Contract, "Texto de referencia") {
		t.Errorf("Contract no debe incluir la Referencia:\n%s", role.Contract)
	}
}

func TestParseRoleModelOpcional(t *testing.T) {
	data := strings.Replace(
		canonicalRole("leader", "primary", "[read, subagent]"),
		"mode: primary\n",
		"mode: primary\nmodel: anthropic/claude\n",
		1,
	)
	role, err := ParseRole([]byte(data))
	if err != nil {
		t.Fatalf("ParseRole: %v", err)
	}
	if role.Model != "anthropic/claude" {
		t.Errorf("Model = %q, esperaba anthropic/claude", role.Model)
	}
}

func TestParseRoleFaltaFrontmatter(t *testing.T) {
	if _, err := ParseRole([]byte("## Contrato\n\nx\n\n## Referencia\n\ny\n")); err == nil {
		t.Fatal("se esperaba error por falta de frontmatter")
	}
}

func TestParseRoleModeInvalido(t *testing.T) {
	if _, err := ParseRole([]byte(canonicalRole("x", "maestro", "[read]"))); err == nil {
		t.Fatal("se esperaba error por mode inválido")
	}
}

func TestParseRoleHerramientaDesconocida(t *testing.T) {
	if _, err := ParseRole([]byte(canonicalRole("x", "subagent", "[read, telepatia]"))); err == nil {
		t.Fatal("se esperaba error por herramienta desconocida")
	}
}

func TestParseRoleToolsNoEsLista(t *testing.T) {
	if _, err := ParseRole([]byte(canonicalRole("x", "subagent", "read, write"))); err == nil {
		t.Fatal("se esperaba error por tools no inline")
	}
}

func TestParseRoleFaltaContrato(t *testing.T) {
	data := "---\nname: x\ndescription: d\nmode: subagent\ntools: [read]\n---\n\n## Referencia\n\nsolo referencia\n"
	if _, err := ParseRole([]byte(data)); err == nil {
		t.Fatal("se esperaba error por falta de ## Contrato")
	}
}

func TestParseRoleFaltaReferencia(t *testing.T) {
	data := "---\nname: x\ndescription: d\nmode: subagent\ntools: [read]\n---\n\n## Contrato\n\nsolo contrato\n"
	if _, err := ParseRole([]byte(data)); err == nil {
		t.Fatal("se esperaba error por falta de ## Referencia")
	}
}

func TestParseRoleOrdenInvertido(t *testing.T) {
	data := "---\nname: x\ndescription: d\nmode: subagent\ntools: [read]\n---\n\n## Referencia\n\nref\n\n## Contrato\n\ncontrato\n"
	if _, err := ParseRole([]byte(data)); err == nil {
		t.Fatal("se esperaba error por orden invertido de secciones")
	}
}

func TestExtractContractExcluyeReferenciaYRecorta(t *testing.T) {
	body := "\n## Contrato\n\n  contenido útil  \n\n## Referencia\n\ndetalle opcional\n"
	got, err := ExtractContract(body)
	if err != nil {
		t.Fatalf("ExtractContract: %v", err)
	}
	if got != "contenido útil" {
		t.Errorf("ExtractContract = %q, esperaba %q", got, "contenido útil")
	}
	if strings.Contains(got, "Referencia") || strings.Contains(got, "detalle opcional") {
		t.Errorf("ExtractContract no debe incluir la Referencia: %q", got)
	}
}

func TestContratosDeLosRolesSonEstructurados(t *testing.T) {
	required := []string{
		"### Identidad",
		"### Objetivo",
		"### Precondiciones",
		"### Protocolo",
		"### Reglas duras",
		"### Formato de salida",
		"### Herramientas permitidas",
		"### Documentos de referencia",
	}
	for _, name := range Roles {
		data, err := fs.ReadFile(reiskel.Harness, ".rei/agents/"+name+".md")
		if err != nil {
			t.Fatalf("no se pudo leer .rei/agents/%s.md del esqueleto: %v", name, err)
		}
		role, err := ParseRole(data)
		if err != nil {
			t.Errorf("ParseRole(%s): %v", name, err)
			continue
		}
		for _, marker := range required {
			if !strings.Contains(role.Contract, marker) {
				t.Errorf("el Contrato de %s no contiene %q", name, marker)
			}
		}
		if strings.Contains(role.Contract, "Sin detalle adicional") {
			t.Errorf("el Contrato de %s no debe incluir el cuerpo de la Referencia", name)
		}
	}
}
