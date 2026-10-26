package view

import (
	"html"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/k1wi777/REI-Harness/internal/gitx"
	"github.com/k1wi777/REI-Harness/internal/paths"
)

// applyChanges calcula, para los documentos del Work Item, el cambio respecto
// del commit base de revisión (last_review_commit; en su defecto base_commit) y
// lo vuelca en el modelo. Degrada sin git (GitAvailable=false) o sin base
// (HasBase=false) sin romper la página (R1–R8).
func applyChanges(p *paths.Project, id string, vm *viewModel) {
	if !gitx.IsRepo(p.Root) {
		return
	}
	vm.GitAvailable = true
	base, origin, ok, err := gitx.ReviewBase(p, id)
	if err != nil || !ok {
		return
	}
	vm.HasBase = true
	vm.BaseSHA = base
	vm.BaseOrigin = origin

	rels := make([]string, 0, len(documentos))
	for _, name := range documentos {
		rels = append(rels, docRel(p, id, name))
	}
	changes, err := gitx.FileChanges(p, base, rels)
	if err != nil {
		return
	}
	for i := range vm.Docs {
		c, ok := changes[docRel(p, id, vm.Docs[i].Name)]
		if !ok {
			continue
		}
		vm.Docs[i].Changed = c.Changed
		vm.Docs[i].New = c.New
		vm.Docs[i].Added = c.Added
		vm.Docs[i].Removed = c.Removed
		if c.Changed {
			vm.Docs[i].Diff = renderDiff(c.Diff)
		}
	}
}

// docRel devuelve la ruta de un documento relativa a la raíz del proyecto, tal
// como la reporta git (separadores "/").
func docRel(p *paths.Project, id, name string) string {
	rel, err := filepath.Rel(p.Root, docPath(p, id, name))
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

// renderDiff compone el diff unificado a HTML seguro: cada línea se escapa y se
// clasifica con una clase propia. El contenido de los documentos nunca se
// interpreta como HTML (R2, R21).
func renderDiff(diff string) template.HTML {
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimRight(diff, "\n"), "\n") {
		var cls string
		switch {
		case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"),
			strings.HasPrefix(line, "diff "), strings.HasPrefix(line, "index "),
			strings.HasPrefix(line, "new file"), strings.HasPrefix(line, "deleted file"),
			strings.HasPrefix(line, "old mode"), strings.HasPrefix(line, "new mode"),
			strings.HasPrefix(line, "similarity"), strings.HasPrefix(line, "rename "):
			cls = "dl-meta"
		case strings.HasPrefix(line, "@@"):
			cls = "dl-hunk"
		case strings.HasPrefix(line, "+"):
			cls = "dl-add"
		case strings.HasPrefix(line, "-"):
			cls = "dl-del"
		default:
			cls = "dl-ctx"
		}
		b.WriteString(`<span class="` + cls + `">`)
		b.WriteString(html.EscapeString(line))
		b.WriteString("</span>\n")
	}
	return template.HTML(b.String())
}
