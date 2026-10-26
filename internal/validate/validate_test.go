package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

const testID = "2026-10-01_10-00__demo"

func writeMeta(t *testing.T, p *paths.Project, m *meta.Meta) {
	t.Helper()
	if err := os.MkdirAll(p.SpecDir(m.ID), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.WorkItemDir(m.ID), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := m.Save(p.MetaFile(m.ID)); err != nil {
		t.Fatal(err)
	}
}

func TestFeatureReadyMissingDocs(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeMeta(t, p, &meta.Meta{
		ID: testID, Title: "t", Description: "d",
		Type: "feature", Status: meta.StatusReady, CreatedAt: "2026-10-01T10:00:00-05:00",
	})
	issues, err := WorkItem(p, testID)
	if err != nil {
		t.Fatal(err)
	}
	if countLevel(issues, LevelFail) == 0 {
		t.Fatal("esperaba FAIL por documentos de planificación faltantes")
	}
}

func TestTaskDoneOK(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeMeta(t, p, &meta.Meta{
		ID: testID, Title: "t", Description: "d",
		Type: "task", Status: meta.StatusDone, CreatedAt: "2026-10-01T10:00:00-05:00",
	})
	mustWrite(t, filepath.Join(p.SpecDir(testID), "plan.md"), "- [x] 1\n")
	mustWrite(t, filepath.Join(p.WorkItemDir(testID), "impl.md"), "impl\n")
	mustWrite(t, filepath.Join(p.WorkItemDir(testID), "review.md"), "review\n")

	issues, err := WorkItem(p, testID)
	if err != nil {
		t.Fatal(err)
	}
	if n := countLevel(issues, LevelFail); n != 0 {
		t.Fatalf("esperaba 0 FAIL, got %d: %+v", n, issues)
	}
}

func TestEmptyDescriptionFails(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeMeta(t, p, &meta.Meta{
		ID: testID, Title: "t", Type: "task", Status: meta.StatusPending,
		CreatedAt: "2026-10-01T10:00:00-05:00",
	})
	issues, _ := WorkItem(p, testID)
	if countLevel(issues, LevelFail) == 0 {
		t.Fatal("esperaba FAIL por description vacía")
	}
}

func TestReportTooLargeWarns(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeMeta(t, p, &meta.Meta{
		ID: testID, Title: "t", Description: "d",
		Type: "task", Status: meta.StatusReview, CreatedAt: "2026-10-01T10:00:00-05:00",
	})
	mustWrite(t, filepath.Join(p.SpecDir(testID), "plan.md"), "- [x] 1\n")
	big := strings.Repeat("palabra ", maxReportWords+1)
	mustWrite(t, filepath.Join(p.WorkItemDir(testID), "impl.md"), big)

	issues, err := WorkItem(p, testID)
	if err != nil {
		t.Fatal(err)
	}
	if n := countLevel(issues, LevelFail); n != 0 {
		t.Fatalf("esperaba 0 FAIL, got %d: %+v", n, issues)
	}
	if countLevel(issues, LevelWarn) == 0 {
		t.Fatalf("esperaba WARN por impl.md demasiado grande: %+v", issues)
	}
}

// TestReportTooLargeDoneNoWarn comprueba que el tope blando solo se aplica a
// Work Items activos: un `done` con un reporte grande no debe emitir WARN.
func TestReportTooLargeDoneNoWarn(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeMeta(t, p, &meta.Meta{
		ID: testID, Title: "t", Description: "d",
		Type: "task", Status: meta.StatusDone, CreatedAt: "2026-10-01T10:00:00-05:00",
	})
	mustWrite(t, filepath.Join(p.SpecDir(testID), "plan.md"), "- [x] 1\n")
	big := strings.Repeat("palabra ", maxReportWords+1)
	mustWrite(t, filepath.Join(p.WorkItemDir(testID), "impl.md"), big)
	mustWrite(t, filepath.Join(p.WorkItemDir(testID), "review.md"), "ok\n")

	issues, err := WorkItem(p, testID)
	if err != nil {
		t.Fatal(err)
	}
	for _, is := range issues {
		if is.Level == LevelWarn && strings.Contains(is.Message, "tope blando") {
			t.Fatalf("done no debe medir el tope blando: %+v", issues)
		}
	}
}

// TestLinksInvalidosWarn comprueba que un links.json roto o incompleto emite
// WARN, nunca FAIL, y que el archivo ausente no genera hallazgo (R16).
func TestLinksInvalidosWarn(t *testing.T) {
	cases := []struct {
		name, content string
		warn          bool
	}{
		{"json roto", `{"provider":`, true},
		{"forma inválida", `{"provider":"github","repo":"owner","issue":0,"url":"nope"}`, true},
		{"válido", `{"provider":"github","repo":"o/n","issue":1,"url":"https://github.com/o/n/issues/1"}`, false},
		{"ausente", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &paths.Project{Root: t.TempDir()}
			writeMeta(t, p, &meta.Meta{
				ID: testID, Title: "t", Description: "d",
				Type: "task", Status: meta.StatusPending, CreatedAt: "2026-10-01T10:00:00-05:00",
			})
			if tc.content != "" {
				mustWrite(t, filepath.Join(p.SpecDir(testID), "links.json"), tc.content)
			}
			issues, err := WorkItem(p, testID)
			if err != nil {
				t.Fatal(err)
			}
			if n := countLevel(issues, LevelFail); n != 0 {
				t.Fatalf("links.json nunca debe producir FAIL: %+v", issues)
			}
			got := countWarnContaining(issues, "links.json")
			if tc.warn && got == 0 {
				t.Fatalf("esperaba WARN de links.json: %+v", issues)
			}
			if !tc.warn && got != 0 {
				t.Fatalf("no esperaba WARN de links.json: %+v", issues)
			}
			if tc.warn && !strings.Contains(strings.Join(issueMessages(issues), " "), "recuperable") {
				t.Fatalf("el WARN debe indicar que es recuperable: %+v", issues)
			}
		})
	}
}

func issueMessages(issues []Issue) []string {
	msgs := make([]string, 0, len(issues))
	for _, is := range issues {
		msgs = append(msgs, is.Message)
	}
	return msgs
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// featureDocs escribe requirements.md, design.md y tasks.md con los tamaños dados.
func featureDocs(t *testing.T, p *paths.Project, requirements, designLines, tasks int) {
	t.Helper()
	var req strings.Builder
	for i := 1; i <= requirements; i++ {
		fmt.Fprintf(&req, "- **R%d.** El sistema DEBE cumplir el requisito %d.\n", i, i)
	}
	mustWrite(t, filepath.Join(p.SpecDir(testID), "requirements.md"), req.String())

	design := strings.Repeat("decisión de diseño\n", designLines)
	mustWrite(t, filepath.Join(p.SpecDir(testID), "design.md"), design)

	var taskDoc strings.Builder
	for i := 1; i <= tasks; i++ {
		fmt.Fprintf(&taskDoc, "- [ ] T%d — implementar la tarea %d.\n", i, i)
	}
	mustWrite(t, filepath.Join(p.SpecDir(testID), "tasks.md"), taskDoc.String())
}

func countWarnContaining(issues []Issue, sub string) int {
	n := 0
	for _, is := range issues {
		if is.Level == LevelWarn && strings.Contains(is.Message, sub) {
			n++
		}
	}
	return n
}

// TestPlanningBudgetWithinLimit comprueba que el presupuesto no avisa en el límite.
func TestPlanningBudgetWithinLimit(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeMeta(t, p, &meta.Meta{
		ID: testID, Title: "t", Description: "d",
		Type: "feature", Status: meta.StatusReady, CreatedAt: "2026-10-01T10:00:00-05:00",
	})
	featureDocs(t, p, maxFeatureRequirements, maxDesignLines, maxTasks)

	issues, err := WorkItem(p, testID)
	if err != nil {
		t.Fatal(err)
	}
	if n := countLevel(issues, LevelFail); n != 0 {
		t.Fatalf("esperaba 0 FAIL, got %d: %+v", n, issues)
	}
	if n := countWarnContaining(issues, "presupuesto"); n != 0 {
		t.Fatalf("esperaba 0 WARN de presupuesto, got %d: %+v", n, issues)
	}
}

// TestPlanningBudgetExceededWarns comprueba los tres avisos al superar el máximo.
func TestPlanningBudgetExceededWarns(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeMeta(t, p, &meta.Meta{
		ID: testID, Title: "t", Description: "d",
		Type: "feature", Status: meta.StatusReady, CreatedAt: "2026-10-01T10:00:00-05:00",
	})
	featureDocs(t, p, maxFeatureRequirements+1, maxDesignLines+1, maxTasks+1)

	issues, err := WorkItem(p, testID)
	if err != nil {
		t.Fatal(err)
	}
	if n := countLevel(issues, LevelFail); n != 0 {
		t.Fatalf("esperaba 0 FAIL, got %d: %+v", n, issues)
	}
	if n := countWarnContaining(issues, "presupuesto"); n != 3 {
		t.Fatalf("esperaba 3 WARN de presupuesto, got %d: %+v", n, issues)
	}
}

// TestTaskPlanBudgetWarns comprueba el aviso por un plan.md demasiado largo.
func TestTaskPlanBudgetWarns(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeMeta(t, p, &meta.Meta{
		ID: testID, Title: "t", Description: "d",
		Type: "task", Status: meta.StatusReady, CreatedAt: "2026-10-01T10:00:00-05:00",
	})
	plan := strings.Repeat("paso\n", maxPlanLines+1)
	mustWrite(t, filepath.Join(p.SpecDir(testID), "plan.md"), plan)

	issues, err := WorkItem(p, testID)
	if err != nil {
		t.Fatal(err)
	}
	if n := countLevel(issues, LevelFail); n != 0 {
		t.Fatalf("esperaba 0 FAIL, got %d: %+v", n, issues)
	}
	if n := countWarnContaining(issues, "presupuesto"); n != 1 {
		t.Fatalf("esperaba 1 WARN de presupuesto, got %d: %+v", n, issues)
	}
}

// TestPlanningBudgetDoneNoWarn fija el filtro de estados: `done` no se mide.
func TestPlanningBudgetDoneNoWarn(t *testing.T) {
	p := &paths.Project{Root: t.TempDir()}
	writeMeta(t, p, &meta.Meta{
		ID: testID, Title: "t", Description: "d",
		Type: "feature", Status: meta.StatusDone, CreatedAt: "2026-10-01T10:00:00-05:00",
	})
	featureDocs(t, p, maxFeatureRequirements+1, maxDesignLines+1, maxTasks+1)
	mustWrite(t, filepath.Join(p.WorkItemDir(testID), "impl.md"), "ok\n")
	mustWrite(t, filepath.Join(p.WorkItemDir(testID), "review.md"), "ok\n")

	issues, err := WorkItem(p, testID)
	if err != nil {
		t.Fatal(err)
	}
	if n := countWarnContaining(issues, "presupuesto"); n != 0 {
		t.Fatalf("done no debe medir el presupuesto: %+v", issues)
	}
}

func countLevel(issues []Issue, level Level) int {
	n := 0
	for _, is := range issues {
		if is.Level == level {
			n++
		}
	}
	return n
}
