package template

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

// Empty es el valor canónico para un campo vacío: "_—_".
const Empty = "_\u2014_"

// ErrArchiveNotAllowed indica que la sesión activa no puede archivarse porque el
// Work Item no está en un estado terminal o current.md no coincide con
// meta.json. Los llamadores pueden comprobarlo con errors.Is.
var ErrArchiveNotAllowed = errors.New("no se puede archivar la sesión")

// RenderFile sustituye cada {{CLAVE}} de la plantilla por su valor y escribe el
// destino. Las claves no presentes se dejan intactas.
func RenderFile(templatePath, targetPath string, vars map[string]string) error {
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}
	content := string(data)
	for key, value := range vars {
		content = strings.ReplaceAll(content, "{{"+key+"}}", value)
	}
	return os.WriteFile(targetPath, []byte(content), 0o644)
}

// ResetCurrent restablece current.md a su estado inicial.
func ResetCurrent(p *paths.Project) error {
	if err := os.MkdirAll(p.ProgressDir(), 0o755); err != nil {
		return err
	}
	return RenderFile(p.TemplatesDir()+"/current.md", p.CurrentFile(), map[string]string{
		"WORK_ITEM": "_ninguno_",
		"TYPE":      Empty,
		"STATE":     Empty,
		"START":     Empty,
		"AGENT":     Empty,
		"PLAN":      Empty,
		"LOG":       Empty,
		"NEXT":      Empty,
	})
}

// StartSession inicia la sesión activa para un Work Item nuevo. Antes de tocar
// el disco comprueba que el Work Item existe, que el tipo coincide con meta.json
// y que su estado es pending; cualquier rechazo incluye causa y siguiente paso y
// deja current.md intacto.
func StartSession(p *paths.Project, id, typ string) error {
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("el Work Item %q no existe. Siguiente paso: créalo con `rei new %s <feature|task>` y vuelve a iniciar la sesión", id, id)
		}
		return fmt.Errorf("no se pudo leer .rei/specs/%s/meta.json: %v. Siguiente paso: corrige el meta.json o crea un Work Item nuevo con `rei new`", id, err)
	}
	if m.Type != typ {
		return fmt.Errorf("el Work Item %s es de tipo %q y la sesión pide %q. Siguiente paso: inicia la sesión con el tipo real (`rei session start %s %s`)", id, m.Type, typ, id, m.Type)
	}
	if m.Status != meta.StatusPending {
		return fmt.Errorf("el Work Item %s está en %q y solo se inicia una sesión nueva en %q. %s", id, m.Status, meta.StatusPending, startSessionNextStep(id, m.Status))
	}
	if err := os.MkdirAll(p.ProgressDir(), 0o755); err != nil {
		return err
	}
	return RenderFile(p.TemplatesDir()+"/current.md", p.CurrentFile(), map[string]string{
		"WORK_ITEM": id,
		"TYPE":      typ,
		"STATE":     "pending",
		"START":     time.Now().Format(time.RFC3339),
		"AGENT":     "spec_author",
		"PLAN":      Empty,
		"LOG":       Empty,
		"NEXT":      "Planificación pendiente.",
	})
}

// startSessionNextStep sugiere qué hacer cuando el Work Item ya no está en
// pending, para acompañar siempre el rechazo con una salida.
func startSessionNextStep(id, status string) string {
	switch status {
	case meta.StatusReady:
		return "Siguiente paso: la planificación ya está lista; apruébala o solicita cambios (Caso C), no reinicies la sesión."
	case meta.StatusInProgress, meta.StatusReview, meta.StatusChangesRequested:
		return "Siguiente paso: retoma la sesión/etapa existente con `rei item show " + id + "`; no crees una sesión nueva."
	case meta.StatusDone, meta.StatusBlocked:
		return "Siguiente paso: el Work Item está cerrado; crea uno nuevo con `rei new`."
	default:
		return "Siguiente paso: revisa el estado del Work Item con `rei item show " + id + "`."
	}
}

// ArchiveSession mueve el resumen de current.md a history.md y restablece
// current.md. Devuelve el id archivado, o "" si no había sesión activa.
//
// Solo archiva cuando el Work Item está en un estado terminal (done o blocked)
// y el campo Estado de current.md coincide con meta.json. En caso contrario
// devuelve ErrArchiveNotAllowed sin escribir nada, para que history.md refleje
// siempre el veredicto final y no un estado intermedio.
func ArchiveSession(p *paths.Project) (string, error) {
	data, err := os.ReadFile(p.CurrentFile())
	if err != nil {
		return "", nil
	}
	content := string(data)
	workItem := field(content, "Work Item")
	if workItem == "" || workItem == "_ninguno_" {
		return "", nil
	}

	m, err := meta.Load(p.MetaFile(workItem))
	if err != nil {
		return "", fmt.Errorf("%w: no se pudo leer .rei/specs/%s/meta.json: %v. Corrige el Work Item o descarta la sesión con 'rei session reset'", ErrArchiveNotAllowed, workItem, err)
	}
	if m.Status != meta.StatusDone && m.Status != meta.StatusBlocked {
		return "", fmt.Errorf("%w: el Work Item %s está en '%s' y solo se archiva en '%s' o '%s'. Registra el veredicto en current.md y ejecuta 'rei status set %s %s'", ErrArchiveNotAllowed, workItem, m.Status, meta.StatusDone, meta.StatusBlocked, workItem, meta.StatusDone)
	}
	if state := field(content, "Estado"); state != m.Status {
		return "", fmt.Errorf("%w: current.md indica Estado '%s' pero meta.json está en '%s'. Actualiza el Estado de current.md a '%s' o corrige meta.json con 'rei status set %s <estado>' (ambos deben coincidir)", ErrArchiveNotAllowed, state, m.Status, m.Status, workItem)
	}

	history, err := os.OpenFile(p.HistoryFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer history.Close()

	header := fmt.Sprintf("\n## %s \u2014 %s\n\n", time.Now().Format("2006-01-02 15:04"), workItem)
	if _, err := history.WriteString(header + stripSessionHeader(content) + "\n"); err != nil {
		return "", err
	}
	if err := ResetCurrent(p); err != nil {
		return "", err
	}
	return workItem, nil
}

func field(content, label string) string {
	prefix := "- **" + label + ":**"
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

// stripSessionHeader elimina el título y la introducción de la plantilla de
// current.md, para que la entrada de history.md empiece en los campos.
func stripSessionHeader(content string) string {
	lines := strings.Split(content, "\n")
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "# Sesión actual") {
		i++
	}
	for i < len(lines) {
		t := strings.TrimSpace(lines[i])
		if t == "" || strings.HasPrefix(t, ">") {
			i++
			continue
		}
		break
	}
	return strings.Join(lines[i:], "\n")
}
