package meta

import (
	"bytes"
	"encoding/json"
	"os"
)

// Estados válidos de un Work Item (ver workflow.md).
const (
	StatusPending          = "pending"
	StatusReady            = "ready"
	StatusInProgress       = "in_progress"
	StatusReview           = "review"
	StatusDone             = "done"
	StatusBlocked          = "blocked"
	StatusChangesRequested = "changes_requested"
)

// Meta representa .rei/specs/<id>/meta.json.
type Meta struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Type             string `json:"type"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
	BaseCommit       string `json:"base_commit"`
	LastReviewCommit string `json:"last_review_commit"`
}

// Load lee y parsea un meta.json.
func Load(path string) (*Meta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Meta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Save escribe el meta.json preservando el orden de campos.
// No escapa HTML (<, >, &) y escribe de forma atómica.
func (m *Meta) Save(path string) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return err
	}
	return writeFileAtomic(path, buf.Bytes())
}

func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ValidStatuses contiene todos los estados admitidos.
var ValidStatuses = []string{
	StatusPending, StatusReady, StatusInProgress,
	StatusReview, StatusDone, StatusBlocked, StatusChangesRequested,
}

// ValidTypes contiene los tipos admitidos.
var ValidTypes = []string{"feature", "task"}

func IsValidStatus(s string) bool {
	for _, v := range ValidStatuses {
		if v == s {
			return true
		}
	}
	return false
}

func IsValidType(s string) bool {
	for _, v := range ValidTypes {
		if v == s {
			return true
		}
	}
	return false
}

// allowedTransitions describe las transiciones estándar de workflow.md.
// Cualquier estado salvo `done` puede pasar a blocked (ver TransitionAllowed);
// desde blocked se puede reanudar a ready (volver a planificación) o a
// in_progress (retomar la implementación). `done` es terminal: solo se sale con
// `--force` como transición no estándar.
var allowedTransitions = map[string][]string{
	StatusPending:          {StatusReady},
	StatusReady:            {StatusInProgress},
	StatusInProgress:       {StatusReview},
	StatusReview:           {StatusDone, StatusChangesRequested},
	StatusChangesRequested: {StatusInProgress},
	StatusBlocked:          {StatusReady, StatusInProgress},
}

// TransitionAllowed indica si from -> to es una transición estándar.
func TransitionAllowed(from, to string) bool {
	if to == StatusBlocked && from != StatusDone {
		return true
	}
	for _, s := range allowedTransitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

// SetStatus actualiza el status de un meta.json preservando el resto de campos.
// Devuelve el estado anterior.
func SetStatus(path, status string) (string, error) {
	m, err := Load(path)
	if err != nil {
		return "", err
	}
	previous := m.Status
	m.Status = status
	if err := m.Save(path); err != nil {
		return "", err
	}
	return previous, nil
}
