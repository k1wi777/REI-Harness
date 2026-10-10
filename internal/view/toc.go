package view

import (
	"strings"

	"github.com/yuin/goldmark/ast"
)

// collectSections recorre el AST de goldmark y extrae los encabezados en orden
// de aparición, con su nivel, su id (asignado por `parser.WithAutoHeadingID`) y
// su texto llano (R9, R10).
func collectSections(doc ast.Node, src []byte) []sectionView {
	var sections []sectionView
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		id := ""
		if v, ok := h.AttributeString("id"); ok {
			switch t := v.(type) {
			case []byte:
				id = string(t)
			case string:
				id = t
			}
		}
		sections = append(sections, sectionView{Level: h.Level, ID: id, Text: headingText(h, src)})
		return ast.WalkContinue, nil
	})
	return sections
}

// headingText reúne el texto llano de un encabezado a partir de sus nodos.
func headingText(h *ast.Heading, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(h, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := n.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(src))
		case *ast.String:
			b.Write(t.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}
