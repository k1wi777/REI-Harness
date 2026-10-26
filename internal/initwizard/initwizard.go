// Package initwizard implementa el nivel determinista de `rei init`: crea la
// estructura base del arnés, detecta los documentos de personalización
// pendientes (los que contienen el marcador canónico) y muestra el plan de
// pasos. No usa IA, red ni dependencias externas salvo git (opcional).
package initwizard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1wi777/REI-Harness/internal/check"
	"github.com/k1wi777/REI-Harness/internal/paths"
)

// marker es el marcador canónico que señala contenido pendiente de personalizar.
const marker = "<!-- REI:PENDIENTE -->"

// InitializerPath es el rol de IA que conduce la personalización guiada.
const InitializerPath = ".rei/agents/initializer.md"

// Doc describe un documento de personalización.
type Doc struct {
	ID    string // "agentes", "arquitectura", "convenciones", "verificacion"
	Path  string // ruta relativa a la raíz del proyecto
	Label string // descripción legible
	Steps []int  // pasos del plan asociados
}

// Docs es la lista canónica de documentos de personalización.
var Docs = []Doc{
	{
		ID:    "agentes",
		Path:  "AGENTS.md",
		Label: "Identidad, propósito y stack (AGENTS.md §2 y §3)",
		Steps: []int{1, 2},
	},
	{
		ID:    "arquitectura",
		Path:  ".rei/docs/project/architecture.md",
		Label: "Arquitectura del proyecto",
		Steps: []int{3},
	},
	{
		ID:    "convenciones",
		Path:  ".rei/docs/project/conventions.md",
		Label: "Convenciones de desarrollo",
		Steps: []int{4},
	},
	{
		ID:    "verificacion",
		Path:  ".rei/docs/project/verification.md",
		Label: "Verificación del proyecto",
		Steps: []int{5},
	},
}

// step es una entrada de la tabla canónica del plan de pasos.
type step struct {
	Num    int
	Title  string
	Target string
}

// steps es el plan de pasos ordenado que muestra `rei init`.
var steps = []step{
	{1, "Identidad y propósito", "AGENTS.md §2"},
	{2, "Stack y comandos", "AGENTS.md §3"},
	{3, "Arquitectura", ".rei/docs/project/architecture.md"},
	{4, "Convenciones", ".rei/docs/project/conventions.md"},
	{5, "Verificación", ".rei/docs/project/verification.md + .rei/config.json"},
	{6, "Cierre y validación", "rei init status / rei doctor"},
}

// Pending devuelve los documentos de personalización que contienen el marcador.
// Un documento inexistente se considera pendiente (no es un error fatal), para
// que el listado sea completo.
func Pending(p *paths.Project) ([]Doc, error) {
	var pending []Doc
	for _, d := range Docs {
		data, err := os.ReadFile(filepath.Join(p.Root, filepath.FromSlash(d.Path)))
		if err != nil {
			if os.IsNotExist(err) {
				pending = append(pending, d)
				continue
			}
			return nil, err
		}
		if strings.Count(string(data), marker) > 0 {
			pending = append(pending, d)
		}
	}
	return pending, nil
}

// Init despliega el esqueleto embebido, crea la estructura de estado, reporta
// los documentos pendientes y muestra el plan de pasos. Devuelve 0 en el camino
// correcto y 1 si falla el despliegue o la creación base.
func Init(p *paths.Project, out io.Writer) int {
	fmt.Fprintln(out, "== rei init ==")
	fmt.Fprintln(out)

	installed := InstallSkeleton(p, out)
	fmt.Fprintln(out)

	scaffold := check.EnsureStructure(p, out)
	fmt.Fprintln(out)

	pending, err := Pending(p)
	if err != nil {
		fmt.Fprintf(out, "[FAIL]  No se pudo detectar la personalización: %v\n", err)
		return 1
	}

	printStatus(out, pending)
	printPlan(out)
	fmt.Fprintln(out)

	if len(pending) > 0 {
		fmt.Fprintln(out, "Personalización pendiente. Inicia la entrevista con el comando")
		fmt.Fprintf(out, "`/personalize` (OpenCode o Claude Code) o invocando el rol `initializer`\n(%s). Consulta el detalle con `rei init status`.\n", InitializerPath)
	} else {
		fmt.Fprintln(out, "Documentación del proyecto personalizada. No hay pasos pendientes.")
	}

	if installed != 0 || scaffold != 0 {
		return 1
	}
	return 0
}

// Status reporta los documentos pendientes sin modificar ningún archivo.
// Devuelve 0 si no queda ninguno y 1 si queda al menos uno.
func Status(p *paths.Project, out io.Writer) int {
	pending, err := Pending(p)
	if err != nil {
		fmt.Fprintf(out, "[FAIL]  No se pudo detectar la personalización: %v\n", err)
		return 1
	}

	fmt.Fprintln(out, "== rei init status ==")
	fmt.Fprintln(out)

	printStatus(out, pending)
	fmt.Fprintln(out)

	if len(pending) == 0 {
		fmt.Fprintln(out, "Documentación del proyecto personalizada.")
		return 0
	}
	fmt.Fprintf(out, "%d documento(s) pendiente(s) de personalización.\n", len(pending))
	return 1
}

// printStatus imprime el estado de todos los documentos canónicos.
func printStatus(out io.Writer, pending []Doc) {
	fmt.Fprintln(out, "Documentos de personalización:")
	pendingIDs := make(map[string]bool, len(pending))
	for _, d := range pending {
		pendingIDs[d.ID] = true
	}
	stateW := len("[PENDIENTE]")
	pathW := 0
	for _, d := range Docs {
		if len(d.Path) > pathW {
			pathW = len(d.Path)
		}
	}
	for _, d := range Docs {
		state := "[COMPLETO]"
		if pendingIDs[d.ID] {
			state = "[PENDIENTE]"
		}
		fmt.Fprintf(out, "  %-*s %-*s %s\n", stateW, state, pathW+2, d.Path, d.Label)
	}
}

// printPlan imprime la tabla canónica del plan de pasos.
func printPlan(out io.Writer) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Plan de pasos:")
	titleW := 0
	for _, s := range steps {
		if len(s.Title) > titleW {
			titleW = len(s.Title)
		}
	}
	for _, s := range steps {
		fmt.Fprintf(out, "  %d. %-*s -> %s\n", s.Num, titleW, s.Title, s.Target)
	}
}
