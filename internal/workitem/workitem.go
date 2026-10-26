package workitem

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/k1wi777/REI-Harness/internal/paths"
	"github.com/k1wi777/REI-Harness/internal/template"
)

var idPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}-[0-9]{2}__[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidID indica si el id cumple el formato YYYY-MM-DD_HH-mm__slug.
func ValidID(id string) bool {
	return idPattern.MatchString(id)
}

// New crea la estructura de un Work Item: .rei/specs/<id>/meta.json y
// .rei/progress/work-items/<id>/.
func New(p *paths.Project, id, typ, title string) error {
	switch typ {
	case "feature", "task":
	default:
		return fmt.Errorf("type debe ser 'feature' o 'task' (recibido: '%s')", typ)
	}
	if !ValidID(id) {
		return fmt.Errorf("id inválido '%s'. Formato: YYYY-MM-DD_HH-mm__slug-en-kebab-case", id)
	}
	if _, err := os.Stat(p.SpecDir(id)); err == nil {
		return fmt.Errorf("ya existe un Work Item con id '%s'", id)
	}
	if _, err := os.Stat(p.WorkItemDir(id)); err == nil {
		return fmt.Errorf("ya existe un Work Item con id '%s'", id)
	}

	if err := os.MkdirAll(p.SpecDir(id), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(p.WorkItemDir(id), 0o755); err != nil {
		return err
	}

	vars := map[string]string{
		"ID":                 id,
		"TITLE":              jsonEscape(title),
		"DESCRIPTION":        "",
		"TYPE":               typ,
		"STATUS":             "pending",
		"CREATED_AT":         createdAt(id),
		"BASE_COMMIT":        "",
		"LAST_REVIEW_COMMIT": "",
	}
	return template.RenderFile(p.TemplatesDir()+"/meta.json", p.MetaFile(id), vars)
}

// createdAt deriva "YYYY-MM-DDTHH:mm:00<tz>" del prefijo del id.
func createdAt(id string) string {
	prefix := id
	if i := strings.Index(id, "__"); i >= 0 {
		prefix = id[:i]
	}
	i := strings.Index(prefix, "_")
	if i < 0 {
		return ""
	}
	datePart := prefix[:i]
	timePart := strings.Replace(prefix[i+1:], "-", ":", 1)
	return fmt.Sprintf("%sT%s:00%s", datePart, timePart, time.Now().Format("-07:00"))
}

func jsonEscape(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return s
	}
	return string(b[1 : len(b)-1])
}
