// Package show presenta la ficha de un Work Item para la CLI (rei item show).
// Es de solo lectura: no escribe archivos ni ejecuta comandos.
package show

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1wi777/my-harness-SDD/internal/links"
	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
	"github.com/k1wi777/my-harness-SDD/internal/state"
	"github.com/k1wi777/my-harness-SDD/internal/validate"
)

// documentos es el orden fijo de la ficha. Se listan siempre los siete
// (existan o no) para que la ficha sea comparable entre Work Items.
var documentos = []string{
	"plan.md",
	"requirements.md",
	"design.md",
	"tasks.md",
	"impl.md",
	"review.md",
	"spec.md",
}

// WorkItem imprime la ficha de un Work Item y devuelve el código de salida.
// Es de solo lectura: no escribe archivos ni ejecuta comandos.
func WorkItem(p *paths.Project, id string, out io.Writer) int {
	metaPath := p.MetaFile(id)
	if !fileExists(metaPath) {
		fmt.Fprintf(os.Stderr, "ERROR: el Work Item %q no existe\n", id)
		return 1
	}
	m, err := meta.Load(metaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	fmt.Fprintf(out, "Work Item: %s\n", m.ID)
	fmt.Fprintf(out, "  %-19s %s\n", "title:", m.Title)
	fmt.Fprintf(out, "  %-19s %s\n", "type:", m.Type)
	fmt.Fprintf(out, "  %-19s %s\n", "status:", m.Status)
	fmt.Fprintf(out, "  %-19s %s\n", "created_at:", m.CreatedAt)
	fmt.Fprintf(out, "  %-19s %s\n", "base_commit:", orEmpty(m.BaseCommit))
	fmt.Fprintf(out, "  %-19s %s\n", "last_review_commit:", orEmpty(m.LastReviewCommit))

	fmt.Fprintln(out)
	writeLinks(out, p, id)

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Documentos:")
	for _, name := range documentos {
		fmt.Fprintf(out, "  %-19s %s\n", name, present(docPath(p, id, name)))
	}

	fmt.Fprintln(out)
	s, err := state.ReadSession(p)
	switch {
	case err != nil:
		fmt.Fprintln(out, "Sesión: no se pudo leer current.md")
	case s.Active() && s.WorkItem == id:
		fmt.Fprintf(out, "Sesión: activa (estado: %s, agente: %s)\n", s.State, s.Agent)
	default:
		fmt.Fprintln(out, "Sesión: no es la sesión activa")
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Validación:")
	issues, err := validate.WorkItem(p, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	fails := 0
	for _, is := range issues {
		fmt.Fprintf(out, "  [%s] %s\n", is.Level, is.Message)
		if is.Level == validate.LevelFail {
			fails++
		}
	}
	if len(issues) == 0 {
		fmt.Fprintln(out, "  (sin hallazgos)")
	}
	if fails > 0 {
		fmt.Fprintln(out, "Resultado: FALLO")
		return 1
	}
	fmt.Fprintln(out, "Resultado: OK")
	return 0
}

// writeLinks imprime la sección Enlaces a partir de links.json. Un archivo
// ausente, ilegible o inválido se informa como nota y no altera el código de
// salida de la ficha.
func writeLinks(out io.Writer, p *paths.Project, id string) {
	l, err := links.Read(filepath.Join(p.SpecDir(id), "links.json"))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		fmt.Fprintln(out, "Enlaces: (sin links.json)")
	case err != nil:
		fmt.Fprintf(out, "Enlaces: nota: links.json inválido (%v); se ignora\n", err)
	default:
		if problems := links.Problems(l); len(problems) > 0 {
			fmt.Fprintf(out, "Enlaces: nota: links.json inválido (%s); se ignora\n", strings.Join(problems, "; "))
			return
		}
		fmt.Fprintln(out, "Enlaces:")
		fmt.Fprintf(out, "  %-19s #%d\n", "issue:", l.Issue)
		fmt.Fprintf(out, "  %-19s %s\n", "url:", l.URL)
		base := "(sin registrar)"
		if l.Base != nil {
			base = l.Base.Branch + " @ " + l.Base.SHA
		}
		fmt.Fprintf(out, "  %-19s %s\n", "base:", base)
		if l.Epic != nil {
			fmt.Fprintf(out, "  %-19s %s (#%d)\n", "epic:", l.Epic.Slug, l.Epic.Issue)
			fmt.Fprintf(out, "  %-19s %s\n", "epic url:", l.Epic.URL)
		} else {
			fmt.Fprintf(out, "  %-19s %s\n", "epic:", "(sin épica)")
		}
	}
}

// docPath resuelve la ubicación de cada documento: los de planificación viven
// en .rei/specs/<id>/ y el resto en .rei/progress/work-items/<id>/.
func docPath(p *paths.Project, id, name string) string {
	switch name {
	case "plan.md", "requirements.md", "design.md", "tasks.md":
		return filepath.Join(p.SpecDir(id), name)
	default:
		return filepath.Join(p.WorkItemDir(id), name)
	}
}

func present(path string) string {
	if fileExists(path) {
		return "sí"
	}
	return "no"
}

func orEmpty(s string) string {
	if s == "" {
		return "(vacío)"
	}
	return s
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}
