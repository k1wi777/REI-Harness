package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/k1wi777/my-harness-SDD/internal/check"
	"github.com/k1wi777/my-harness-SDD/internal/config"
	"github.com/k1wi777/my-harness-SDD/internal/gitx"
	"github.com/k1wi777/my-harness-SDD/internal/initwizard"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
	"github.com/k1wi777/my-harness-SDD/internal/state"
	"github.com/k1wi777/my-harness-SDD/internal/validate"
)

// Run ejecuta un diagnóstico de solo lectura del harness y escribe el reporte
// en out. Devuelve 0 si no hay ningún FAIL y 1 en caso contrario.
//
// Nunca crea archivos ni ejecuta comandos que modifiquen el proyecto (en
// particular, no inicializa git ni ejecuta los checks de .rei/config.json).
func Run(p *paths.Project, out io.Writer) int {
	failures := 0
	ok := func(format string, a ...any) {
		fmt.Fprintf(out, "[OK]    %s\n", fmt.Sprintf(format, a...))
	}
	warn := func(format string, a ...any) {
		fmt.Fprintf(out, "[WARN]  %s\n", fmt.Sprintf(format, a...))
	}
	info := func(format string, a ...any) {
		fmt.Fprintf(out, "[INFO]  %s\n", fmt.Sprintf(format, a...))
	}
	fail := func(format string, a ...any) {
		fmt.Fprintf(out, "[FAIL]  %s\n", fmt.Sprintf(format, a...))
		failures++
	}

	// 1. Integridad del harness (solo lectura)
	for _, f := range check.RequiredFiles {
		if fileExists(filepath.Join(p.Root, f)) {
			ok("%s", f)
		} else {
			fail("Falta %s", f)
		}
	}
	for _, d := range []string{".rei/specs/", ".rei/progress/work-items/"} {
		if dirExists(filepath.Join(p.Root, filepath.FromSlash(d))) {
			ok("%s", d)
		} else {
			fail("Falta %s", d)
		}
	}
	for _, f := range []string{".rei/progress/current.md", ".rei/progress/history.md"} {
		if fileExists(filepath.Join(p.Root, filepath.FromSlash(f))) {
			ok("%s", f)
		} else {
			fail("Falta %s", f)
		}
	}

	// 1b. Personalización del proyecto (WARN, no incrementa fallos)
	pending, err := initwizard.Pending(p)
	switch {
	case err != nil:
		warn("No se pudo comprobar la personalización: %v", err)
	case len(pending) == 0:
		ok("Documentación del proyecto personalizada.")
	default:
		for _, d := range pending {
			warn("Personalización pendiente: %s", d.Path)
		}
	}

	// 2. Work Items
	items, err := state.ListWorkItems(p)
	switch {
	case err != nil:
		fail("No se pudieron listar los Work Items: %v", err)
	case len(items) == 0:
		info("No hay Work Items.")
	default:
		for _, it := range items {
			issues, err := validate.WorkItem(p, it.ID)
			if err != nil {
				fail("%s: %v", it.ID, err)
				continue
			}
			if len(issues) == 0 {
				ok("%s", it.ID)
				continue
			}
			for _, is := range issues {
				if is.Level == validate.LevelFail {
					fail("%s: %s", it.ID, is.Message)
				} else {
					warn("%s: %s", it.ID, is.Message)
				}
			}
		}
	}

	// 2b. Invariante: solo un Work Item en in_progress
	active, err := state.InProgressItems(p)
	switch {
	case err != nil:
		fail("No se pudo comprobar el invariante de in_progress: %v", err)
	case len(active) > 1:
		fail("Hay %d Work Items en in_progress: %s. Solo puede haber uno.", len(active), strings.Join(active, ", "))
	}

	// 3. Sesión
	s, err := state.ReadSession(p)
	switch {
	case err != nil:
		warn("No se pudo leer la sesión: %v", err)
	case s.Active():
		ok("Sesión activa: %s (estado: %s, agente: %s)", s.WorkItem, orQ(s.State), orQ(s.Agent))
	default:
		ok("Sin sesión activa.")
	}

	// 4. Git (solo consulta)
	switch {
	case !hasGit():
		warn("git no disponible; el review por diff quedará deshabilitado.")
	case gitx.IsRepo(p.Root):
		if sha := gitx.HeadSHA(p.Root); sha == "" {
			ok("Repositorio git detectado. Sin commits todavía.")
		} else {
			ok("Repositorio git detectado. HEAD=%s", sha)
		}
	default:
		warn("Sin repositorio git; el review por diff quedará deshabilitado.")
	}

	// 5. Checks (.rei/config.json), sin ejecutarlos
	cfg, err := config.Load(filepath.Join(p.ReiDir(), "config.json"))
	if err != nil {
		fail("No se pudo leer .rei/config.json: %v", err)
	} else {
		info("%d checks configurados en .rei/config.json (no ejecutados).", len(cfg.Checks))
	}

	// 6. Resultado
	if failures == 0 {
		ok("Sin fallos. Harness íntegro.")
		return 0
	}
	fail("%d fallo(s) detectado(s).", failures)
	return 1
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func hasGit() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func orQ(s string) string {
	if s == "" {
		return "?"
	}
	return s
}
