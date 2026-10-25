package check

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/k1wi777/my-harness-SDD/internal/config"
	"github.com/k1wi777/my-harness-SDD/internal/gitx"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
	"github.com/k1wi777/my-harness-SDD/internal/template"
)

// RequiredFiles son los archivos críticos de REI Harness (rutas relativas a la raíz).
var RequiredFiles = []string{
	"AGENTS.md",
	".rei/docs/harness/workflow.md",
	".rei/docs/harness/specs.md",
	".rei/docs/harness/task.md",
	".rei/docs/harness/meta.md",
	".rei/docs/harness/progress.md",
	".rei/docs/harness/integrations/github.md",
	".rei/docs/project/architecture.md",
	".rei/docs/project/conventions.md",
	".rei/docs/project/verification.md",
	".rei/agents/leader.md",
	".rei/agents/spec_author.md",
	".rei/agents/implementer.md",
	".rei/agents/initializer.md",
	".rei/agents/reviewer.md",
	".rei/templates/current.md",
	".rei/templates/history.md",
	".rei/templates/meta.json",
}

// Run comprueba la integridad del harness, inicializa la estructura y ejecuta
// los checks definidos en .rei/config.json. Devuelve el código de salida.
func Run(p *paths.Project, quiet bool, out io.Writer) int {
	exit := 0
	ok := func(format string, a ...any) {
		if !quiet {
			fmt.Fprintf(out, "[OK]    %s\n", fmt.Sprintf(format, a...))
		}
	}
	fail := func(format string, a ...any) {
		fmt.Fprintf(out, "[FAIL]  %s\n", fmt.Sprintf(format, a...))
		exit = 1
	}

	// 1. Integridad del harness
	for _, f := range RequiredFiles {
		if fileExists(filepath.Join(p.Root, f)) {
			ok("%s", f)
		} else {
			fail("Falta %s", f)
		}
	}

	// 2-4. Estructura (specs, progreso, plantillas) y repositorio git
	if ensureStructure(p, out, quiet) != 0 {
		exit = 1
	}

	// 5. Checks del proyecto (.rei/config.json)
	cfg, err := config.Load(filepath.Join(p.ReiDir(), "config.json"))
	switch {
	case err != nil:
		fail(".rei/config.json inválido: %v", err)
	case len(cfg.Checks) == 0:
		if !quiet {
			fmt.Fprintf(out, "[INFO]  Sin checks configurados (.rei/config.json).\n")
		}
	default:
		for _, ch := range cfg.Checks {
			desc := ch.Description
			if desc == "" {
				desc = ch.ID
			}
			if Exec(p.Root, ch.Command, out) {
				ok("%s", desc)
			} else {
				fail("%s", desc)
			}
		}
	}

	// 6. Resumen
	if !quiet {
		fmt.Fprintln(out)
	}
	if exit == 0 {
		ok("REI Harness listo para trabajar.")
	} else {
		fail("REI Harness contiene errores.")
	}
	return exit
}

// EnsureStructure crea .rei/specs/, .rei/progress/work-items/, current.md e
// history.md si faltan, e inicializa git si no existe. Devuelve un código de
// salida (0 correcto; 1 si no pudo crear los archivos base). No ejecuta los
// checks de .rei/config.json.
//
// Es la pieza compartida entre `rei check` y `rei init`: reutilizarla evita
// duplicar el scaffold. El fallo al inicializar git es un aviso, no un error.
func EnsureStructure(p *paths.Project, out io.Writer) int {
	return ensureStructure(p, out, false)
}

// ensureStructure implementa EnsureStructure. El parámetro quiet permite a
// check.Run conservar su salida actual (sin [OK] con --quiet) sin que
// EnsureStructure tenga que conocer los flags del CLI.
func ensureStructure(p *paths.Project, out io.Writer, quiet bool) int {
	exit := 0
	ok := func(format string, a ...any) {
		if !quiet {
			fmt.Fprintf(out, "[OK]    %s\n", fmt.Sprintf(format, a...))
		}
	}
	warn := func(format string, a ...any) {
		fmt.Fprintf(out, "[WARN]  %s\n", fmt.Sprintf(format, a...))
	}
	fail := func(format string, a ...any) {
		fmt.Fprintf(out, "[FAIL]  %s\n", fmt.Sprintf(format, a...))
		exit = 1
	}

	// 1. Estructura
	if err := os.MkdirAll(p.SpecsDir(), 0o755); err != nil {
		fail("No se pudo crear .rei/specs/: %v", err)
	} else {
		ok(".rei/specs/")
	}
	if err := os.MkdirAll(p.WorkItemsDir(), 0o755); err != nil {
		fail("No se pudo crear .rei/progress/work-items/: %v", err)
	} else {
		ok(".rei/progress/work-items/")
	}

	// 2. current.md / history.md
	if !fileExists(p.CurrentFile()) {
		if err := template.ResetCurrent(p); err != nil {
			fail("No se pudo crear current.md: %v", err)
		} else {
			ok(".rei/progress/current.md creado desde plantilla")
		}
	} else {
		ok(".rei/progress/current.md")
	}
	if !fileExists(p.HistoryFile()) {
		data, err := os.ReadFile(filepath.Join(p.TemplatesDir(), "history.md"))
		if err != nil {
			fail("No se pudo leer la plantilla de history.md: %v", err)
		} else if err := os.WriteFile(p.HistoryFile(), data, 0o644); err != nil {
			fail("No se pudo crear history.md: %v", err)
		} else {
			ok(".rei/progress/history.md creado desde plantilla")
		}
	} else {
		ok(".rei/progress/history.md")
	}

	// 3. Repositorio git (opcional, mejora el review por diff)
	switch {
	case !hasGit():
		warn("git no disponible; REI funciona, pero el review por diff quedará deshabilitado.")
	case gitx.IsRepo(p.Root):
		ok("Repositorio git detectado.")
	default:
		if err := exec.Command("git", "-C", p.Root, "init").Run(); err != nil {
			warn("No se pudo inicializar git; el review por diff quedará deshabilitado.")
		} else {
			ok("Repositorio git inicializado.")
		}
	}

	return exit
}

func Exec(dir string, command []string, out io.Writer) bool {
	if len(command) == 0 {
		return false
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		if s := strings.TrimSpace(buf.String()); s != "" {
			for _, line := range strings.Split(s, "\n") {
				fmt.Fprintf(out, "    %s\n", line)
			}
		}
		return false
	}
	return true
}

// RunDeclared ejecuta solo los checks de .rei/config.json, sin efectos
// secundarios (no crea estructura, no inicializa git, no valida integridad).
// NO es la suite de tests de REI (esa es `make test`).
func RunDeclared(p *paths.Project, out io.Writer) int {
	cfg, err := config.Load(filepath.Join(p.ReiDir(), "config.json"))
	switch {
	case err != nil:
		fmt.Fprintf(out, "[FAIL]  .rei/config.json inválido: %v\n", err)
		return 1
	case len(cfg.Checks) == 0:
		fmt.Fprintln(out, "[INFO]  Sin checks configurados (.rei/config.json). Nada que ejecutar.")
		return 0
	}
	exit := 0
	for _, ch := range cfg.Checks {
		desc := ch.Description
		if desc == "" {
			desc = ch.ID
		}
		if Exec(p.Root, ch.Command, out) {
			fmt.Fprintf(out, "[OK]    %s\n", desc)
		} else {
			fmt.Fprintf(out, "[FAIL]  %s\n", desc)
			exit = 1
		}
	}
	return exit
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func hasGit() bool {
	_, err := exec.LookPath("git")
	return err == nil
}
