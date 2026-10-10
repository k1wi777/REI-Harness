package gitx

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

// ErrNoGit indica que no hay un repositorio git disponible.
var ErrNoGit = errors.New("sin git")

// ErrNoBase indica que no hay un commit base para calcular el diff.
var ErrNoBase = errors.New("sin base")

func gitOut(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), err
}

// IsRepo indica si dir está dentro de un repositorio git.
func IsRepo(dir string) bool {
	if _, err := exec.LookPath("git"); err != nil {
		return false
	}
	_, err := gitOut(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

// HeadSHA devuelve el SHA de HEAD, o "" si no hay commits.
func HeadSHA(dir string) string {
	out, err := gitOut(dir, "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// SetCommit registra el HEAD actual en meta.json.
// Devuelve un mensaje descriptivo; no es un error si no hay git/commits.
func SetCommit(p *paths.Project, id, field string) (string, error) {
	if field != "base_commit" && field != "last_review_commit" {
		return "", fmt.Errorf("campo inválido: %s", field)
	}
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		return "", err
	}
	if !IsRepo(p.Root) {
		return "Sin git: no se registra '" + field + "'.", nil
	}
	sha := HeadSHA(p.Root)
	if sha == "" {
		return "Sin commits todavía: no se registra '" + field + "'.", nil
	}
	if field == "base_commit" {
		m.BaseCommit = sha
	} else {
		m.LastReviewCommit = sha
	}
	if err := m.Save(p.MetaFile(id)); err != nil {
		return "", err
	}
	return field + "=" + sha, nil
}

// ReviewBase devuelve el commit base de revisión de un Work Item
// (last_review_commit; en su defecto base_commit) y su origen. ok=false si no
// hay ninguno o si el commit no es válido; err solo si no se puede leer
// meta.json o si no hay un repositorio git.
func ReviewBase(p *paths.Project, id string) (base, origin string, ok bool, err error) {
	m, err := meta.Load(p.MetaFile(id))
	if err != nil {
		return "", "", false, err
	}
	if !IsRepo(p.Root) {
		return "", "", false, ErrNoGit
	}
	origin = "base_commit"
	switch {
	case m.LastReviewCommit != "":
		base, origin = m.LastReviewCommit, "last_review_commit"
	case m.BaseCommit != "":
		base = m.BaseCommit
	}
	if base == "" {
		return "", "", false, nil
	}
	if _, err := gitOut(p.Root, "cat-file", "-e", base+"^{commit}"); err != nil {
		return "", "", false, nil
	}
	return base, origin, true, nil
}

// FileChange describe el cambio de un archivo respecto de un commit base.
type FileChange struct {
	Changed bool   // el contenido difiere del base
	New     bool   // no existía en el base (todo añadido)
	Added   int    // líneas añadidas
	Removed int    // líneas eliminadas
	Diff    string // diff unificado (--no-color); "" si no cambió
}

// FileChanges calcula el cambio de cada ruta (relativa a la raíz del proyecto)
// respecto de un commit base. No descarta el bookkeeping de sesión
// (.rei/progress/**): el visor muestra esos documentos y quiere sus cambios.
func FileChanges(p *paths.Project, base string, relPaths []string) (map[string]FileChange, error) {
	out := make(map[string]FileChange, len(relPaths))
	for _, rel := range relPaths {
		fc, err := fileChange(p, base, rel)
		if err != nil {
			return nil, err
		}
		out[rel] = fc
	}
	return out, nil
}

// fileChange resuelve el cambio de una única ruta respecto del base: los
// archivos sin rastrear presentes en disco son material añadido por completo;
// los rastreados se comparan con git.
func fileChange(p *paths.Project, base, rel string) (FileChange, error) {
	var fc FileChange
	if _, err := gitOut(p.Root, "ls-files", "--error-unmatch", "--", rel); err != nil {
		// No rastreado: si existe en disco, todo su contenido es añadido.
		if !fileExists(filepath.Join(p.Root, rel)) {
			return fc, nil
		}
		diff, _ := gitOut(p.Root, "diff", "--no-color", "--no-index", "--", "/dev/null", rel)
		fc.Changed = true
		fc.New = true
		fc.Added, fc.Removed = countDiff(diff)
		fc.Diff = diff
		return fc, nil
	}
	numstat, _ := gitOut(p.Root, "diff", "--numstat", "--no-color", base, "--", rel)
	if strings.TrimSpace(numstat) == "" {
		return fc, nil // sin cambios
	}
	fc.Changed = true
	fc.Added, fc.Removed = parseNumstat(numstat)
	diff, _ := gitOut(p.Root, "diff", "--no-color", base, "--", rel)
	fc.Diff = diff
	return fc, nil
}

// countDiff cuenta las líneas añadidas y eliminadas de un diff unificado.
func countDiff(diff string) (added, removed int) {
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"):
			continue
		case strings.HasPrefix(line, "+"):
			added++
		case strings.HasPrefix(line, "-"):
			removed++
		}
	}
	return added, removed
}

// parseNumstat suma las columnas añadida/eliminada de `git diff --numstat`
// (los archivos binarios reportan "-", que cuenta como cero).
func parseNumstat(out string) (added, removed int) {
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		added += atoiOrZero(fields[0])
		removed += atoiOrZero(fields[1])
	}
	return added, removed
}

func atoiOrZero(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// ReviewDiff genera el paquete de revisión de un Work Item.
func ReviewDiff(p *paths.Project, id string, full bool) (string, error) {
	base, origin, ok, err := ReviewBase(p, id)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrNoBase
	}

	tracked := filterPaths(gitLines(p.Root, "diff", "--name-only", base))
	untracked := filterPaths(gitLines(p.Root, "ls-files", "--others", "--exclude-standard"))

	var b strings.Builder
	fmt.Fprintf(&b, "== Paquete de revisión: %s ==\n", id)
	fmt.Fprintf(&b, "Base: %s (%s)\n\n", base, origin)

	b.WriteString("--- Resumen ---\n")
	if len(tracked) > 0 {
		args := append([]string{"diff", "--stat", base, "--"}, tracked...)
		stat, _ := gitOut(p.Root, args...)
		b.WriteString(stat)
	} else {
		b.WriteString("(sin cambios)\n")
	}

	if len(untracked) > 0 {
		b.WriteString("\n--- Nuevos sin rastrear ---\n")
		for _, u := range untracked {
			fmt.Fprintf(&b, "  %s\n", u)
		}
	}

	if full {
		b.WriteString("\n--- Diff completo ---\n")
		if len(tracked) > 0 {
			args := append([]string{"diff", base, "--"}, tracked...)
			diff, _ := gitOut(p.Root, args...)
			b.WriteString(diff)
		}
		for _, u := range untracked {
			if strings.HasPrefix(u, ".rei/progress/") {
				continue
			}
			fmt.Fprintf(&b, "\n--- nuevo: %s ---\n", u)
			diff, _ := gitOut(p.Root, "diff", "--no-index", "--", "/dev/null", u)
			b.WriteString(diff)
		}
	}

	return b.String(), nil
}

func gitLines(dir string, args ...string) []string {
	out, _ := gitOut(dir, args...)
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	sort.Strings(lines)
	return lines
}

// filterPaths descarta solo el bookkeeping de sesión (.rei/progress/**);
// conserva el resto: código del proyecto, AGENTS.md, la spec del Work Item,
// docs, agentes, adaptadores y plantillas del harness.
func filterPaths(paths []string) []string {
	var kept []string
	for _, p := range paths {
		if strings.HasPrefix(p, ".rei/progress/") {
			continue
		}
		kept = append(kept, p)
	}
	return kept
}
