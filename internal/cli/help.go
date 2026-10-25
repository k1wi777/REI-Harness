package cli

import (
	"fmt"
	"os"
	"strings"
)

// commandInfo describe un comando para la ayuda general y por comando.
type commandInfo struct {
	name   string
	usage  string
	list   string // forma corta para la tabla general; si vacío, se usa usage
	short  string
	detail []string
}

// commands es la tabla única que alimenta la ayuda general y la de cada
// comando (evita duplicar el texto entre ambas).
var commands = []commandInfo{
	{
		name:  "check",
		usage: "check [--quiet]",
		short: "Verifica el harness, inicializa la estructura y corre los checks",
		detail: []string{
			"--quiet, -q    omite los [OK] y el resumen.",
			"Además de los checks, crea la estructura y, si falta, inicializa git.",
			"Sale 0 si todo pasa; 1 si hay algún [FAIL].",
		},
	},
	{
		name:  "init",
		usage: "init [status|opencode|claude [--check]|--update [--force]]",
		list:  "init [status|opencode|claude|--update]",
		short: "Instala el esqueleto del harness, reporta la personalización y genera agentes nativos",
		detail: []string{
			"Sin argumentos: despliega el esqueleto embebido (AGENTS.md, .rei/docs/, .rei/agents/, .rei/adapters/, .rei/templates/ y .rei/config.json) sin sobrescribir archivos existentes, crea la estructura de estado (.rei/specs/, .rei/progress/work-items/, current.md, history.md) y, si falta y git está disponible, inicializa el repositorio.",
			"status: reporta qué documentos siguen pendientes, sin modificar archivos (0 si no queda ninguno; 1 si queda alguno).",
			"--update: refresca el esqueleto embebido en un proyecto YA inicializado; NO instala desde cero. Exige un proyecto REI (.rei/); si no lo está, avisa y sale 1. Nunca toca .rei/specs/** ni .rei/progress/**.",
			"--update es marker-aware para los documentos de personalización (AGENTS.md §2/§3 y .rei/docs/project/{architecture,conventions,verification}.md): se actualizan si aún contienen `<!-- REI:PENDIENTE -->`; si ya no lo contienen (personalizados), se conservan con aviso.",
			"--update para el resto del harness: se actualiza solo si el manifiesto .rei/install-manifest.json confirma que el archivo no se ha modificado desde la instalación; si se modificó o no hay entrada, se conserva con aviso.",
			"--update --force: sobrescribe cualquier archivo que difiera, incluidos los personalizados o modificados a mano; los archivos que difieren se reportan como [NEW], [UPD], [OK] o [SKIP].",
			"--update regenera los nativos solo si el proyecto ya tiene .opencode/ (OpenCode) y/o .claude/ (Claude Code); si no existen, no instala nativos.",
			"opencode: instala el esqueleto (incluido .rei/adapters/), asegura la estructura y genera los agentes nativos de OpenCode en .opencode/agents/<rol>.md a partir del `## Contrato` de cada rol. Genera además el comando `/personalize` en .opencode/commands/personalize.md. No sobrescribe archivos que no lleven la marca GENERATED.",
			"opencode --check: verifica que los agentes nativos y el comando `/personalize` existen y coinciden con la generación canónica, sin escribir (0 si todo coincide; 1 si falta alguno, no lleva la marca o difiere).",
			"claude: instala el esqueleto (incluido .rei/adapters/) y genera los agentes nativos de Claude Code en .claude/agents/<rol>.md con frontmatter name, description, tools y model opcional a partir del `## Contrato` de cada rol. Genera además el comando `/personalize` en .claude/commands/personalize.md y `CLAUDE.md` (con `@AGENTS.md`). No sobrescribe archivos que no lleven la marca GENERATED.",
			"claude --check: verifica que los agentes nativos, el comando `/personalize` y `CLAUDE.md` existen y coinciden con la generación canónica, sin escribir (0 si todo coincide; 1 si falta alguno, no lleva la marca o difiere).",
			"Runtimes sin subagentes nativos (Cursor, Codex): funcionan por fallback, sin configuración; el Leader transmite al subagente únicamente la sección `## Contrato` del rol.",
			"Códigos de salida: 0 si completa la instalación; 1 si no puede crear algún archivo del esqueleto o de la estructura, o si la generación/verificación de OpenCode o Claude falla; 2 uso incorrecto.",
			"Si AGENTS.md ya existe se avisa y se conserva sin cambios.",
			"La personalización guiada se inicia con el comando `/personalize` (OpenCode o Claude Code) o invocando el rol `initializer` (.rei/agents/initializer.md).",
		},
	},
	{
		name:   "doctor",
		usage:  "doctor",
		short:  "Diagnóstico de solo lectura del harness (no modifica nada)",
		detail: []string{"No escribe archivos ni ejecuta comandos que modifiquen el proyecto."},
	},
	{
		name:  "new",
		usage: "new <id> <feature|task> [title]",
		short: "Crea un Work Item nuevo",
		detail: []string{
			"Crea .rei/specs/<id>/meta.json y la carpeta de progreso.",
			"El estado inicial es 'pending'.",
		},
	},
	{
		name:  "session",
		usage: "session [status]",
		short: "Muestra o gestiona la sesión activa",
		detail: []string{
			"start <id> <feature|task>    inicia la sesión de un Work Item.",
			"archive                       archiva la sesión si el Work Item está en done/blocked y current.md es coherente; si no, falla.",
			"reset                         restablece current.md.",
		},
	},
	{
		name:   "items",
		usage:  "items status",
		short:  "Lista el estado de todos los Work Items",
		detail: []string{"Marca con '<< ACTIVO' el Work Item en in_progress."},
	},
	{
		name:  "status",
		usage: "status set <id> <status> [--force]",
		short: "Cambia el estado de un Work Item en meta.json",
		detail: []string{
			"Valida la transición según workflow.md.",
			"--force permite transiciones no estándar.",
		},
	},
	{
		name:   "commit",
		usage:  "commit set <id> <base_commit|last_review_commit>",
		list:   "commit set <id> <field>",
		short:  "Registra el commit actual de git en meta.json",
		detail: []string{"Guarda HEAD en el campo indicado del Work Item."},
	},
	{
		name:  "validate",
		usage: "validate [<id>]",
		short: "Comprueba la consistencia interna de un Work Item",
		detail: []string{
			"Sin <id>, usa la sesión activa.",
			"Sale 0 si no hay [FAIL]; 1 si hay al menos uno.",
		},
	},
	{
		name:   "review-diff",
		usage:  "review-diff <id> [--full]",
		short:  "Genera el paquete de revisión de un Work Item",
		detail: []string{"--full muestra el diff completo, no solo el resumen."},
	},
	{
		name:  "test",
		usage: "test",
		short: "Ejecuta los checks de .rei/config.json",
		detail: []string{
			"Ejecuta SOLO los checks declarados en .rei/config.json, sin efectos secundarios.",
			"NO es la suite de tests de REI; esos son `make test` (dev, requiere Go).",
			"Sin checks configurados informa y sale 0; sale 1 si algún check falla.",
		},
	},
	{
		name:  "item",
		usage: "item show <id>",
		short: "Muestra la ficha de un Work Item",
		detail: []string{
			"Ficha de solo lectura: metadatos, documentos, sesión y validación.",
			"Sale 1 si hay algún [FAIL] de validación.",
		},
	},
	{
		name:  "view",
		usage: "view [<id>] [--open|--stop|--enable|--disable]",
		short: "Abre el visor de especificaciones de un Work Item en el navegador",
		detail: []string{
			"Sin <id>, usa el Work Item de la sesión activa.",
			"Sin modo (primer plano): sirve el visor y bloquea hasta Ctrl+C.",
			"--open: arranca (o reutiliza) el servidor en segundo plano y abre el navegador sin bloquear la conversación.",
			"--stop: detiene el servidor de fondo del proyecto y elimina su registro temporal.",
			"--enable / --disable: activan o desactivan viewer.autoOpenOnReady en .rei/config.json.",
			"Los modos son mutuamente excluyentes; usar más de uno es uso incorrecto (2).",
			"Códigos de salida: 0 correcto; 1 sin sesión activa o error de E/S; 2 uso incorrecto.",
		},
	},
	{
		name:  "github",
		usage: "github <enable|disable|status>",
		short: "Configura la integración opcional con GitHub Issues",
		detail: []string{
			"enable: detecta el repositorio desde el remoto `origin` (`git remote get-url origin`) y la rama base desde `origin/HEAD`; escribe integrations.github en .rei/config.json preservando checks, viewer y el resto de claves. Si `gh` falta o no está autenticado, avisa sin bloquear (sale 0).",
			"disable: apaga la integración (`enabled: false`) conservando repo, base, mergeStrategy y los enlaces locales (links.json). No contacta con GitHub.",
			"status: reporte de solo lectura del estado, valores efectivos, coincidencia con el remoto, disponibilidad/autenticación de `gh` y siguiente paso.",
			"El comando nunca crea issues, comentarios, ramas, PR ni merges: esas acciones las ejecuta el agente según .rei/docs/harness/integrations/github.md.",
			"Códigos de salida: 0 correcto (también con avisos y en status), 1 error operativo o de configuración, 2 uso incorrecto.",
		},
	},
	{
		name:  "update",
		usage: "update [--check]",
		short: "Auto-actualiza el binario de REI desde GitHub Releases",
		detail: []string{
			"Consulta la última release publicada, la compara (semver) con la versión en ejecución y, si hay una más reciente, descarga el asset del SO/arch, verifica su SHA-256 contra checksums.txt y reemplaza el binario en ejecución (temp + rename con backup).",
			"--check: solo informa del resultado de la comprobación; no descarga ni escribe (en cualquier SO). Sale 0 si la consulta funcionó.",
			"En Windows no es posible reemplazar un .exe en ejecución: informa y abre la página de releases en el navegador, imprimiendo la URL como alternativa.",
			"Versión 'dev' (compilación local): avisa de que no es una release publicada y ofrece instalar la última disponible.",
			"Fallo de integridad (SHA-256): aborta sin modificar el binario. Fallo de red o de permisos: informa, sale 1 y no deja el binario degradado; si la sustitución falla tras iniciarse, restaura el original.",
			"Códigos de salida: 0 si está actualizado, al día, `--check` correcto o Windows guiado; 1 ante red/permisos/integridad/extracción; 2 uso incorrecto.",
		},
	},
	{
		name:   "version",
		usage:  "version",
		short:  "Muestra la versión",
		detail: []string{"También disponible como --version o -v."},
	},
	{
		name:   "help",
		usage:  "help [<comando>]",
		short:  "Muestra la ayuda general o de un comando",
		detail: []string{"Sin argumentos muestra la ayuda general."},
	},
}

// helpText construye la ayuda general desde la tabla.
func helpText() string {
	var b strings.Builder
	b.WriteString("rei — CLI de REI Harness\n")
	b.WriteString("\nUso:\n")
	b.WriteString("  rei <comando> [argumentos]\n")
	b.WriteString("\nComandos:\n")

	rows := make([][2]string, 0, len(commands))
	for _, c := range commands {
		left := c.list
		if left == "" {
			left = c.usage
		}
		rows = append(rows, [2]string{left, c.short})
	}
	b.WriteString(twoColumn(rows, "  ", 3))

	b.WriteString("\nUsa 'rei <comando> --help' para ver el detalle de un comando.\n")
	return b.String()
}

// twoColumn formatea filas de dos columnas alineando la segunda tras la más
// ancha de la primera, con un margen fijo entre ambas.
func twoColumn(rows [][2]string, indent string, gap int) string {
	width := 0
	for _, r := range rows {
		if len(r[0]) > width {
			width = len(r[0])
		}
	}
	var b strings.Builder
	for _, r := range rows {
		fmt.Fprintf(&b, "%s%-*s%s\n", indent, width+gap, r[0], r[1])
	}
	return b.String()
}

func printHelp() {
	fmt.Print(helpText())
}

// helpWidth es el ancho objetivo para ajustar las líneas de la ayuda por comando.
const helpWidth = 100

// wrap envuelve `text` en líneas de a lo sumo `width` columnas, prefijando cada
// línea con `indent` (la primera incluida). Las líneas de continuación de una
// viñeta pueden usar una sangría distinta.
func wrap(text string, width int, indent, hang string) string {
	words := strings.Fields(text)
	var b strings.Builder
	cur := indent
	b.WriteString(cur)
	col := len(cur)
	for i, w := range words {
		needSpace := i > 0 && col > len(cur)
		if needSpace && col+1+len(w) > width {
			b.WriteString("\n")
			cur = hang
			b.WriteString(cur)
			col = len(cur)
			needSpace = false
		}
		if needSpace {
			b.WriteString(" ")
			col++
		}
		b.WriteString(w)
		col += len(w)
	}
	b.WriteString("\n")
	return b.String()
}

// commandHelpText devuelve el texto de ayuda de un comando a partir de la
// tabla. No escribe a stdout, para poder testearlo.
func commandHelpText(name string) (string, bool) {
	for _, c := range commands {
		if c.name != name {
			continue
		}
		var b strings.Builder
		b.WriteString(wrap("rei "+c.name+" — "+c.short, helpWidth, "", ""))
		b.WriteString("\nUso:\n")
		b.WriteString(wrap(c.usage, helpWidth, "  ", "  "))
		b.WriteString("\nDetalle:\n")
		for _, d := range c.detail {
			// Las líneas con alineación manual (varios espacios seguidos) se
			// respetan tal cual; el resto se ajusta al ancho.
			if strings.Contains(d, "   ") {
				fmt.Fprintf(&b, "  %s\n", d)
				continue
			}
			b.WriteString(wrap("- "+d, helpWidth, "  ", "    "))
		}
		return b.String(), true
	}
	return "", false
}

// printCommandHelp imprime la ayuda de un comando. Si el comando no existe,
// informa por stderr, muestra la ayuda general y devuelve 2.
func printCommandHelp(name string) int {
	text, ok := commandHelpText(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "comando desconocido: %s\n\n", name)
		printHelp()
		return 2
	}
	fmt.Print(text)
	return 0
}
