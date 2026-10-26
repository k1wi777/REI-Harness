package meta

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.json")
	m := &Meta{
		ID:          "2026-10-01_10-00__demo",
		Title:       "Demo",
		Description: "desc",
		Type:        "task",
		Status:      StatusPending,
		CreatedAt:   "2026-10-01T10:00:00-05:00",
	}
	if err := m.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if *got != *m {
		t.Fatalf("got %+v, want %+v", got, m)
	}
}

func TestSaveDoesNotEscapeHTML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.json")
	m := &Meta{
		ID:          "x",
		Title:       "t",
		Description: "a < b > c & d",
		Type:        "task",
		Status:      StatusPending,
		CreatedAt:   "x",
	}
	if err := m.Save(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `\u003c`) || strings.Contains(string(data), `\u0026`) {
		t.Fatalf("Save escapó HTML: %s", data)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Description != m.Description {
		t.Fatalf("roundtrip description = %q, want %q", got.Description, m.Description)
	}
}

func TestIsValid(t *testing.T) {
	if !IsValidStatus(StatusInProgress) {
		t.Error("in_progress debería ser válido")
	}
	if IsValidStatus("nope") {
		t.Error("'nope' no debería ser válido")
	}
	if !IsValidType("feature") || IsValidType("x") {
		t.Error("validación de type incorrecta")
	}
}

func TestTransitionAllowed(t *testing.T) {
	allowed := []struct{ from, to string }{
		{StatusPending, StatusReady},
		{StatusReady, StatusInProgress},
		{StatusInProgress, StatusReview},
		{StatusReview, StatusDone},
		{StatusReview, StatusChangesRequested},
		{StatusChangesRequested, StatusInProgress},
		{StatusBlocked, StatusReady},
		{StatusBlocked, StatusInProgress},
	}
	for _, tc := range allowed {
		if !TransitionAllowed(tc.from, tc.to) {
			t.Errorf("%s -> %s debería estar permitida", tc.from, tc.to)
		}
	}

	rejected := []struct{ from, to string }{
		{StatusPending, StatusInProgress},
		{StatusDone, StatusInProgress},
		{StatusDone, StatusBlocked},
		{StatusBlocked, StatusDone},
		{StatusReady, StatusDone},
	}
	for _, tc := range rejected {
		if TransitionAllowed(tc.from, tc.to) {
			t.Errorf("%s -> %s NO debería estar permitida", tc.from, tc.to)
		}
	}

	for _, from := range ValidStatuses {
		if from == StatusDone {
			continue
		}
		if !TransitionAllowed(from, StatusBlocked) {
			t.Errorf("%s -> blocked debería estar permitida", from)
		}
	}
}

func TestSetStatusPreservesFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.json")
	m := &Meta{
		ID:               "2026-10-02_00-33__demo",
		Title:            "Demo",
		Description:      "desc",
		Type:             "task",
		Status:           StatusPending,
		CreatedAt:        "2026-10-02T00:33:00-05:00",
		BaseCommit:       "726a658661373a8caaec8134187061213739cb46",
		LastReviewCommit: "deadbeef",
	}
	if err := m.Save(path); err != nil {
		t.Fatal(err)
	}

	prev, err := SetStatus(path, StatusReady)
	if err != nil {
		t.Fatal(err)
	}
	if prev != StatusPending {
		t.Fatalf("estado anterior = %q, want %q", prev, StatusPending)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusReady {
		t.Fatalf("status = %q, want %q", got.Status, StatusReady)
	}
	if got.ID != m.ID || got.Title != m.Title || got.Description != m.Description ||
		got.Type != m.Type || got.CreatedAt != m.CreatedAt ||
		got.BaseCommit != m.BaseCommit || got.LastReviewCommit != m.LastReviewCommit {
		t.Fatalf("campos no preservados: got %+v, want %+v", got, m)
	}
}
