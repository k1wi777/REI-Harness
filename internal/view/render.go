package view

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
	"github.com/k1wi777/my-harness-SDD/internal/state"
	"github.com/k1wi777/my-harness-SDD/internal/validate"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
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

// docView representa un documento en la página.
type docView struct {
	Name    string
	Present bool
	HTML    template.HTML
	Pos     int // posición 1..N entre los documentos presentes (folio)
}

// issueView es un hallazgo de validación con el documento al que apunta.
type issueView struct {
	Level   string
	Message string
	Doc     string // documento relacionado, si el mensaje lo menciona
}

// viewModel es el conjunto de datos que la plantilla HTML necesita.
type viewModel struct {
	Meta          *meta.Meta
	Session       state.Session
	SessionActive bool
	Docs          []docView
	Issues        []issueView
	CanApprove    bool
	TotalPresent  int
}

// funcMap expone etiquetas en español a la plantilla.
var funcMap = template.FuncMap{
	"status": statusLabel,
	"level":  levelLabel,
}

// statusLabel traduce el estado crudo de meta.json a una etiqueta en español.
func statusLabel(s string) string {
	switch s {
	case "pending":
		return "pendiente"
	case "ready":
		return "listo para aprobar"
	case "in_progress":
		return "en curso"
	case "review":
		return "en revisión"
	case "done":
		return "completado"
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

// renderMarkdown convierte Markdown a HTML con goldmark en modo seguro.
func renderMarkdown(src []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := mdRenderer.Convert(src, &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// buildModel compone los datos de la página para un Work Item.
func buildModel(p *paths.Project, id string) (*viewModel, error) {
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		return nil, err
	}
	vm := &viewModel{
		Meta:       m,
		CanApprove: m.Status == meta.StatusReady,
	}
	if s, err := state.ReadSession(p); err == nil {
		vm.Session = s
		vm.SessionActive = s.Active() && s.WorkItem == id
	}
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
			if html, rerr := renderMarkdown(data); rerr == nil {
				dv.Present = true
				dv.HTML = html
				total++
				dv.Pos = total
			}
		}
		vm.Docs = append(vm.Docs, dv)
	}
	vm.TotalPresent = total
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
