package validate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/k1wi777/my-harness-SDD/internal/links"
	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
	"github.com/k1wi777/my-harness-SDD/internal/state"
)

// Level clasifica un hallazgo.
type Level string

const (
	LevelFail Level = "FAIL"
	LevelWarn Level = "WARN"
)

// Tope blando de los reportes impl.md/review.md. Superarlo produce un WARN,
// nunca un FAIL (ver .rei/docs/harness/progress.md).
const (
	maxReportLines = 80
	maxReportWords = 600
)

// Presupuesto de planificación (ver .rei/docs/harness/specs.md y task.md).
// Superar el máximo produce un WARN, nunca un FAIL.
const (
	maxFeatureRequirements = 30
	maxDesignLines         = 400
	maxTasks               = 15
	maxPlanLines           = 200
)

var (
	requirementLine = regexp.MustCompile(`(?m)^\s*(?:[-*]\s+)?(?:\*\*)?R\d+\b`)
	taskLine        = regexp.MustCompile(`(?m)^\s*-\s*\[[ xX]\]`)
)

// Issue es un hallazgo de la validación.
type Issue struct {
	Level   Level
	Message string
}

// WorkItem comprueba la consistencia interna de un Work Item.
func WorkItem(p *paths.Project, id string) ([]Issue, error) {
	var issues []Issue
	fail := func(format string, a ...any) {
		issues = append(issues, Issue{LevelFail, fmt.Sprintf(format, a...)})
	}
	warn := func(format string, a ...any) {
		issues = append(issues, Issue{LevelWarn, fmt.Sprintf(format, a...)})
	}

	metaPath := p.MetaFile(id)
	if !fileExists(metaPath) {
		fail("no existe %s", metaPath)
		return issues, nil
	}

	m, err := meta.Load(metaPath)
	if err != nil {
		fail("meta.json inválido: %v", err)
		return issues, nil
	}

	for _, kv := range []struct{ name, val string }{
		{"id", m.ID},
		{"title", m.Title},
		{"description", m.Description},
		{"type", m.Type},
		{"status", m.Status},
		{"created_at", m.CreatedAt},
	} {
		if strings.TrimSpace(kv.val) == "" {
			fail("campo vacío en meta.json: %s", kv.name)
		}
	}

	if m.ID != "" && m.ID != id {
		fail("id de meta.json ('%s') no coincide con la carpeta ('%s')", m.ID, id)
	}
	if m.Type != "" && !meta.IsValidType(m.Type) {
		fail("type inválido: '%s'", m.Type)
	}
	if m.Status != "" && !meta.IsValidStatus(m.Status) {
		fail("status inválido: '%s'", m.Status)
	}
	if m.CreatedAt != "" {
		if prefix := createdAtPrefix(id); prefix != "" && !strings.HasPrefix(m.CreatedAt, prefix) {
			warn("created_at ('%s') no coincide con el prefijo del id (%s...)", m.CreatedAt, prefix)
		}
	}

	// links.json es un índice opcional y recuperable: un archivo inválido se
	// informa como WARN y nunca bloquea el Work Item.
	if l, err := links.Read(filepath.Join(p.SpecDir(id), "links.json")); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			warn("links.json inválido (%v); es un índice recuperable (se reconstruye por el marcador) y no bloquea el Work Item", err)
		}
	} else if problems := links.Problems(l); len(problems) > 0 {
		warn("links.json inválido (%s); es un índice recuperable (se reconstruye por el marcador) y no bloquea el Work Item", strings.Join(problems, "; "))
	}

	if fi, err := os.Stat(p.WorkItemDir(id)); err != nil || !fi.IsDir() {
		warn("no existe %s", p.WorkItemDir(id))
	}

	if needsPlanning(m.Status) {
		switch m.Type {
		case "feature":
			for _, f := range []string{"requirements.md", "design.md", "tasks.md"} {
				if !fileExists(filepath.Join(p.SpecDir(id), f)) {
					fail("falta %s (feature en estado %s)", f, m.Status)
				}
			}
		case "task":
			if !fileExists(filepath.Join(p.SpecDir(id), "plan.md")) {
				fail("falta plan.md (task en estado %s)", m.Status)
			}
		}
	}

	switch m.Status {
	case meta.StatusReview, meta.StatusDone, meta.StatusChangesRequested:
		if !fileExists(filepath.Join(p.WorkItemDir(id), "impl.md")) {
			fail("falta impl.md (estado %s)", m.Status)
		}
	}
	if m.Status == meta.StatusDone && !fileExists(filepath.Join(p.WorkItemDir(id), "review.md")) {
		fail("falta review.md (estado done)")
	}

	// Aviso blando: la planificación no debe exceder su presupuesto.
	// Solo se evalúa en los estados de planificación activos.
	if budgetApplicable(m.Status) {
		switch m.Type {
		case "feature":
			specDir := p.SpecDir(id)
			if n, ok := countMatches(filepath.Join(specDir, "requirements.md"), requirementLine); ok && n > maxFeatureRequirements {
				warn("requirements.md excede el presupuesto (máximo %d; objetivo 10–20): %d requisitos. Justifica o divide.",
					maxFeatureRequirements, n)
			}
			if lines, ok := fileLines(filepath.Join(specDir, "design.md")); ok && lines > maxDesignLines {
				warn("design.md excede el presupuesto (máximo %d líneas; objetivo 150–300): %d líneas. Justifica o divide.",
					maxDesignLines, lines)
			}
			if n, ok := countMatches(filepath.Join(specDir, "tasks.md"), taskLine); ok && n > maxTasks {
				warn("tasks.md excede el presupuesto (máximo %d tareas; objetivo 8–12): %d tareas. Justifica o divide.",
					maxTasks, n)
			}
		case "task":
			if lines, ok := fileLines(filepath.Join(p.SpecDir(id), "plan.md")); ok && lines > maxPlanLines {
				warn("plan.md excede el presupuesto (máximo %d líneas; objetivo 150–200): %d líneas. Justifica o divide.",
					maxPlanLines, lines)
			}
		}
	}

	// Aviso blando: impl.md/review.md no deben exceder el tope de tamaño.
	// Solo se evalúa en Work Items activos; el histórico ya cerrado no se mide.
	if isActiveStatus(m.Status) {
		for _, f := range []string{"impl.md", "review.md"} {
			path := filepath.Join(p.WorkItemDir(id), f)
			if !fileExists(path) {
				continue
			}
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			lines, words := reportSize(string(data))
			if lines > maxReportLines || words > maxReportWords {
				warn("%s excede el tope blando (~%d líneas / ~%d palabras): %d líneas, %d palabras",
					f, maxReportLines, maxReportWords, lines, words)
			}
		}
	}

	if s, err := state.ReadSession(p); err == nil && s.WorkItem == id {
		if s.State != "" && s.State != m.Status {
			warn("current.md dice Estado='%s' pero meta.json dice status='%s'", s.State, m.Status)
		}
	}

	return issues, nil
}

func needsPlanning(status string) bool {
	switch status {
	case meta.StatusReady, meta.StatusInProgress, meta.StatusReview,
		meta.StatusChangesRequested, meta.StatusDone:
		return true
	}
	return false
}

// isActiveStatus indica si el Work Item sigue en curso y, por tanto, se le
// aplica el tope blando de tamaño de impl.md/review.md. El histórico cerrado
// (done/pending/ready/blocked) no se mide.
func isActiveStatus(status string) bool {
	switch status {
	case meta.StatusInProgress, meta.StatusReview, meta.StatusChangesRequested:
		return true
	}
	return false
}

// budgetApplicable indica si se aplica el presupuesto de planificación. Incluye
// `ready` (a diferencia de isActiveStatus) y excluye el histórico cerrado.
func budgetApplicable(status string) bool {
	switch status {
	case meta.StatusReady, meta.StatusInProgress, meta.StatusReview,
		meta.StatusChangesRequested:
		return true
	}
	return false
}

// countMatches cuenta las coincidencias de re en un archivo. Devuelve ok=false si
// el archivo no existe o no se puede leer, para omitir la comprobación.
func countMatches(path string, re *regexp.Regexp) (int, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	return len(re.FindAllIndex(data, -1)), true
}

// fileLines cuenta las líneas de un archivo con el mismo criterio que reportSize.
func fileLines(path string) (int, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	lines, _ := reportSize(string(data))
	return lines, true
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// reportSize devuelve (líneas, palabras) de un reporte. Las líneas se cuentan
// sobre el contenido sin el salto de línea final; las palabras, con Fields.
func reportSize(content string) (int, int) {
	lines := 0
	if trimmed := strings.TrimSuffix(content, "\n"); trimmed != "" {
		lines = len(strings.Split(trimmed, "\n"))
	}
	return lines, len(strings.Fields(content))
}

// createdAtPrefix deriva "YYYY-MM-DDTHH:mm" del id "YYYY-MM-DD_HH-mm__slug".
func createdAtPrefix(id string) string {
	prefix := id
	if i := strings.Index(id, "__"); i >= 0 {
		prefix = id[:i]
	}
	i := strings.Index(prefix, "_")
	if i < 0 {
		return ""
	}
	datePart := prefix[:i]
	timePart := prefix[i+1:]
	if datePart == "" || timePart == "" {
		return ""
	}
	return datePart + "T" + strings.Replace(timePart, "-", ":", 1)
}
