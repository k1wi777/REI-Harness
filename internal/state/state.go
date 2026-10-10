package state

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

const emptyWorkItem = "_ninguno_"

// ErrTransitionNotAllowed indica que la transición no está permitida por workflow.
var ErrTransitionNotAllowed = errors.New("transición no permitida")

// Session es el estado vivo reflejado en .rei/progress/current.md.
type Session struct {
	WorkItem string
	Type     string
	State    string
	Agent    string
}

// Active indica si hay un Work Item en curso.
func (s Session) Active() bool {
	return s.WorkItem != "" && s.WorkItem != emptyWorkItem
}

// ReadSession parsea .rei/progress/current.md.
func ReadSession(p *paths.Project) (Session, error) {
	data, err := os.ReadFile(p.CurrentFile())
	if err != nil {
		return Session{}, err
	}
	content := string(data)
	return Session{
		WorkItem: markdownField(content, "Work Item"),
		Type:     markdownField(content, "Tipo"),
		State:    markdownField(content, "Estado"),
		Agent:    markdownField(content, "Agente activo"),
	}, nil
}

// Transition valida y aplica una transición de estado sobre meta.json,
// reutilizando la misma validación que `rei status set`. Si el Work Item es la
// sesión activa, alinea el campo Estado de current.md. Devuelve el estado
// previo y si fue necesario forzar la transición. Si la transición no está
// permitida y force es false, devuelve ErrTransitionNotAllowed sin escribir.
func Transition(p *paths.Project, id, to string, force bool) (from string, forced bool, err error) {
	if !meta.IsValidStatus(to) {
		return "", false, fmt.Errorf("estado inválido: %q", to)
	}
	path := p.MetaFile(id)
	m, err := meta.Load(path)
	if err != nil {
		return "", false, err
	}
	from = m.Status
	if from == to {
		return from, false, nil
	}
	if !meta.TransitionAllowed(from, to) {
		if !force {
			return from, false, ErrTransitionNotAllowed
		}
		forced = true
	}
	if _, err := meta.SetStatus(path, to); err != nil {
		return from, forced, err
	}
	if err := alignSessionState(p, id, to); err != nil {
		return from, forced, err
	}
	return from, forced, nil
}

// alignSessionState reescribe solo la línea `- **Estado:**` de current.md
// cuando el Work Item es la sesión activa, preservando el resto del archivo.
func alignSessionState(p *paths.Project, id, status string) error {
	s, err := ReadSession(p)
	if err != nil || s.WorkItem != id {
		return nil
	}
	data, err := os.ReadFile(p.CurrentFile())
	if err != nil {
		return err
	}
	updated := replaceMarkdownField(string(data), "Estado", status)
	if updated == string(data) {
		return nil
	}
	return os.WriteFile(p.CurrentFile(), []byte(updated), 0o644)
}

// replaceMarkdownField sustituye el valor de la línea `- **<label>:**` por
// value, conservando el resto del contenido.
func replaceMarkdownField(content, label, value string) string {
	prefix := "- **" + label + ":**"
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = prefix + " " + value
			break
		}
	}
	return strings.Join(lines, "\n")
}

// ItemStatus resume el estado de un Work Item.
type ItemStatus struct {
	ID     string
	Status string
}

// ListWorkItems lee todos los meta.json y devuelve su estado (ordenado por id).
func ListWorkItems(p *paths.Project) ([]ItemStatus, error) {
	files, err := p.MetaFiles()
	if err != nil {
		return nil, err
	}
	items := make([]ItemStatus, 0, len(files))
	for _, f := range files {
		id := idFromMetaPath(f)
		status := "desconocido"
		if m, err := meta.Load(f); err == nil {
			if m.ID != "" {
				id = m.ID
			}
			if m.Status != "" {
				status = m.Status
			}
		}
		items = append(items, ItemStatus{ID: id, Status: status})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func markdownField(content, label string) string {
	prefix := "- **" + label + ":**"
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func idFromMetaPath(path string) string {
	path = strings.TrimSuffix(path, "/meta.json")
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}
