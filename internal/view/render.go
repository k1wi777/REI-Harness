package view

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
	"github.com/k1wi777/my-harness-SDD/internal/state"
	"github.com/k1wi777/my-harness-SDD/internal/validate"
	"github.com/yuin/goldmark"
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
}

// viewModel es el conjunto de datos que la plantilla HTML necesita.
type viewModel struct {
	Meta          *meta.Meta
	Session       state.Session
	SessionActive bool
	Docs          []docView
	Issues        []validate.Issue
	CanApprove    bool
}

// mdRenderer renderiza Markdown en modo seguro por defecto: sin
// html.WithUnsafe, el HTML incrustado en los documentos no se interpreta.
var mdRenderer = goldmark.New()

// pageTpl es la plantilla de la página; define "page" (documento completo) y
// "content" (fragmento para la recarga por SSE).
var pageTpl = template.Must(template.New("page.html").ParseFS(viewAssets, "assets/page.html"))

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
		vm.Issues = issues
	}
	for _, name := range documentos {
		dv := docView{Name: name}
		if data, err := os.ReadFile(docPath(p, id, name)); err == nil {
			if html, rerr := renderMarkdown(data); rerr == nil {
				dv.Present = true
				dv.HTML = html
			}
		}
		vm.Docs = append(vm.Docs, dv)
	}
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
