package view

import (
	"bytes"
	"embed"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1wi777/REI-Harness/internal/links"
	"github.com/k1wi777/REI-Harness/internal/meta"
	"github.com/k1wi777/REI-Harness/internal/paths"
	"github.com/k1wi777/REI-Harness/internal/state"
	"github.com/k1wi777/REI-Harness/internal/validate"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// viewAssets contiene la plantilla HTML y los estáticos propios del visor.
//
//go:embed assets/page.html assets/style.css assets/app.js
var viewAssets embed.FS

// documentos es el orden fijo de los documentos mostrados. Se listan siempre
// los siete (existan o no), igual que en `rei item show`.
var documentos = []string{
	"plan.md",
	"requirements.md",
	"design.md",
	"tasks.md",
	"impl.md",
	"review.md",
	"spec.md",
}

// sectionView es un encabezado del documento, para el índice de secciones (TOC).
type sectionView struct {
	Level int
	ID    string
	Text  string
}

// docView representa un documento en la página.
type docView struct {
	Name     string
	Present  bool
	HTML     template.HTML
	Pos      int // posición 1..N entre los documentos presentes (folio)
	Sections []sectionView

	// Cambios respecto del commit base de revisión (vacío si no hay base).
	Changed bool
	New     bool
	Added   int
	Removed int
	Diff    template.HTML
}

// issueView es un hallazgo de validación con el documento al que apunta.
type issueView struct {
	Level   string
	Message string
	Doc     string // documento relacionado, si el mensaje lo menciona
}

// linksView es la ficha de enlaces derivada de links.json. Solo la ficha
// válida expone campos; en otro caso Note explica el motivo y Problems detalla
// los incumplimientos, sin usar ningún campo del archivo (R6, R7).
type linksView struct {
	Valid    bool
	Note     string
	Problems []string

	Issue int
	URL   string

	HasBase bool
	Branch  string
	SHA     string

	HasEpic   bool
	EpicSlug  string
	EpicIssue int
	EpicURL   string
}

// viewModel es el conjunto de datos que la plantilla HTML necesita.
type viewModel struct {
	Meta          *meta.Meta
	Session       state.Session
	SessionActive bool
	Docs          []docView
	Issues        []issueView
	Links         linksView
	TotalPresent  int

	// Estado de la comparación contra el commit base de revisión.
	GitAvailable bool
	HasBase      bool
	BaseSHA      string
	BaseOrigin   string
}

// funcMap expone etiquetas en español a la plantilla.
var funcMap = template.FuncMap{
	"status": statusLabel,
	"level":  levelLabel,
	"short":  shortSHA,
}

// shortSHA devuelve los primeros 8 caracteres de un SHA (o el propio SHA).
func shortSHA(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// statusLabel traduce el estado crudo de meta.json a una etiqueta en español.
func statusLabel(s string) string {
	switch s {
	case "pending":
		return "pendiente"
	case "ready":
		return "en aprobación"
	case "in_progress":
		return "en curso"
	case "review":
		return "en revisión"
	case "done":
		return "terminado"
	case "blocked":
		return "bloqueado"
	case "changes_requested":
		return "cambios solicitados"
	default:
		return s
	}
}

// levelLabel traduce el nivel de un hallazgo de validación.
func levelLabel(l string) string {
	switch l {
	case "FAIL":
		return "Fallo"
	case "WARN":
		return "Aviso"
	case "OK":
		return "Correcto"
	case "INFO":
		return "Info"
	default:
		return l
	}
}

// mdRenderer renderiza Markdown en modo seguro por defecto: sin
// html.WithUnsafe, el HTML incrustado en los documentos no se interpreta. Con
// WithAutoHeadingID cada encabezado recibe un id, habilitando enlaces a
// secciones.
var mdRenderer = goldmark.New(goldmark.WithParserOptions(parser.WithAutoHeadingID()))

// pageTpl es la plantilla de la página; define "page" (documento completo) y
// "content" (fragmento para la recarga por SSE).
var pageTpl = template.Must(template.New("page.html").Funcs(funcMap).ParseFS(viewAssets, "assets/page.html"))

// renderDoc convierte Markdown a HTML con goldmark en modo seguro y extrae sus
// encabezados para el índice de secciones. Se parsea una sola vez: el HTML que
// produce es el mismo que `mdRenderer.Convert` (R9, R10).
func renderDoc(src []byte) (template.HTML, []sectionView, error) {
	doc := mdRenderer.Parser().Parse(text.NewReader(src))
	sections := collectSections(doc, src)

	var buf bytes.Buffer
	if err := mdRenderer.Renderer().Render(&buf, src, doc); err != nil {
		return "", nil, err
	}
	return template.HTML(buf.String()), sections, nil
}

// readLinks compone la FICHA de enlaces para el Work Item. No duplica las
// reglas de validación: se apoya en links.Read y links.Problems y solo consume
// el resultado (R6–R8).
func readLinks(p *paths.Project, id string) linksView {
	l, err := links.Read(linksPath(p, id))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return linksView{Note: "sin links.json"}
		}
		return linksView{Note: "links.json inválido", Problems: []string{err.Error()}}
	}
	if problems := links.Problems(l); len(problems) > 0 {
		return linksView{Note: "links.json inválido", Problems: problems}
	}
	vm := linksView{Valid: true, Issue: l.Issue, URL: l.URL}
	if l.Base != nil {
		vm.HasBase = true
		vm.Branch = l.Base.Branch
		vm.SHA = l.Base.SHA
	}
	if l.Epic != nil {
		vm.HasEpic = true
		vm.EpicSlug = l.Epic.Slug
		vm.EpicIssue = l.Epic.Issue
		vm.EpicURL = l.Epic.URL
	}
	return vm
}

// linksPath devuelve la ruta del índice local del Work Item.
func linksPath(p *paths.Project, id string) string {
	return filepath.Join(p.SpecDir(id), "links.json")
}

// buildModel compone los datos de la página para un Work Item.
func buildModel(p *paths.Project, id string) (*viewModel, error) {
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		return nil, err
	}
	vm := &viewModel{Meta: m}
	if s, err := state.ReadSession(p); err == nil {
		vm.Session = s
		vm.SessionActive = s.Active() && s.WorkItem == id
	}
	vm.Links = readLinks(p, id)
	if issues, err := validate.WorkItem(p, id); err == nil {
		for _, is := range issues {
			vm.Issues = append(vm.Issues, issueView{
				Level:   string(is.Level),
				Message: is.Message,
				Doc:     relatedDoc(is.Message),
			})
		}
	}
	total := 0
	for _, name := range documentos {
		dv := docView{Name: name}
		if data, err := os.ReadFile(docPath(p, id, name)); err == nil {
			if html, sections, rerr := renderDoc(data); rerr == nil {
				dv.Present = true
				dv.HTML = html
				dv.Sections = sections
				total++
				dv.Pos = total
			}
		}
		vm.Docs = append(vm.Docs, dv)
	}
	vm.TotalPresent = total
	applyChanges(p, id, vm)
	return vm, nil
}

// renderPage escribe la página HTML completa.
func renderPage(w io.Writer, vm *viewModel) error {
	return pageTpl.ExecuteTemplate(w, "page", vm)
}

// renderContent escribe solo el fragmento dinámico (para la recarga por SSE).
func renderContent(w io.Writer, vm *viewModel) error {
	return pageTpl.ExecuteTemplate(w, "content", vm)
}

// relatedDoc devuelve el documento del Work Item mencionado en un mensaje de
// validación (por su nombre de archivo), o "" si no menciona ninguno.
func relatedDoc(msg string) string {
	for _, name := range documentos {
		if strings.Contains(msg, name) {
			return name
		}
	}
	return ""
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
