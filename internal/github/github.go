// Package github implementa la capa determinista de la integración opcional
// con GitHub Issues: detecta el repositorio y la rama base desde git y lee y
// escribe integrations.github en .rei/config.json.
//
// NUNCA crea issues, comentarios, ramas, PR ni merges: esas acciones son
// responsabilidad del agente, según el playbook
// .rei/docs/harness/integrations/github.md. Tampoco contacta con la API de
// GitHub: solo consulta `git` y, como diagnóstico, `gh auth status`.
package github

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/k1wi777/my-harness-SDD/internal/config"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

// Runner abstrae los comandos externos para poder testear sin `git`, `gh` ni
// red.
type Runner interface {
	LookPath(file string) (string, error)
	Run(dir, name string, args ...string) (stdout, stderr string, err error)
}

// ExecRunner implementa Runner con os/exec.
type ExecRunner struct{}

// LookPath busca un ejecutable en PATH.
func (ExecRunner) LookPath(file string) (string, error) { return exec.LookPath(file) }

// Run ejecuta un comando en dir y captura stdout y stderr por separado.
func (ExecRunner) Run(dir, name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// Remote es un repositorio de GitHub detectado.
type Remote struct {
	Host  string
	Owner string
	Name  string
}

// String devuelve owner/name.
func (r Remote) String() string { return r.Owner + "/" + r.Name }

var (
	scpLike    = regexp.MustCompile(`^[A-Za-z0-9._-]+@([^:]+):(.+)$`)
	identRe    = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	baseRef    = regexp.MustCompile(`^origin/(.+)$`)
	originArgs = []string{"remote", "get-url", "origin"}
)

// ParseRemote interpreta un remoto SSH (git@github.com:owner/name.git), URL
// (https://github.com/owner/name.git, ssh://git@github.com/owner/name) y sus
// variantes con `.git` y barra final. Exige host github.com y exactamente
// owner/name. Los errores no exponen credenciales.
func ParseRemote(raw string) (Remote, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Remote{}, errors.New("el remoto está vacío")
	}

	var host, path string
	if strings.Contains(trimmed, "://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return Remote{}, fmt.Errorf("remoto inválido %q: %v", sanitizeRemote(trimmed), err)
		}
		host = u.Hostname()
		path = u.Path
	} else {
		m := scpLike.FindStringSubmatch(trimmed)
		if m == nil {
			return Remote{}, fmt.Errorf("formato de remoto no reconocido: %q", sanitizeRemote(trimmed))
		}
		host = m[1]
		path = m[2]
	}

	path = strings.Trim(path, "/")
	path = strings.TrimSuffix(path, ".git")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return Remote{}, fmt.Errorf("el remoto %q no tiene forma owner/name", sanitizeRemote(trimmed))
	}
	owner, name := parts[0], parts[1]
	if !validIdent(owner) || !validIdent(name) {
		return Remote{}, fmt.Errorf("el remoto %q tiene owner o nombre inválidos", sanitizeRemote(trimmed))
	}
	if !strings.EqualFold(host, "github.com") {
		return Remote{}, fmt.Errorf("el remoto %q usa el host %q; v1 solo soporta github.com", sanitizeRemote(trimmed), host)
	}
	return Remote{Host: "github.com", Owner: owner, Name: name}, nil
}

func validIdent(s string) bool {
	return s != "." && s != ".." && identRe.MatchString(s)
}

// sanitizeRemote descarta la userinfo de una URL para no filtrar credenciales
// en los mensajes de error.
func sanitizeRemote(raw string) string {
	if !strings.Contains(raw, "://") {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.User = nil
	return u.String()
}

// originCause ejecuta `git remote get-url origin` y parsea el resultado sin
// adornar el error (lo usa Status para informar de la causa).
func originCause(dir string, r Runner) (Remote, error) {
	stdout, stderr, err := r.Run(dir, "git", originArgs...)
	if err != nil {
		cause := firstLine(stderr)
		if cause == "" {
			cause = err.Error()
		}
		return Remote{}, fmt.Errorf("`git remote get-url origin` falló (%s)", cause)
	}
	return ParseRemote(strings.TrimSpace(stdout))
}

// DetectOrigin ejecuta `git remote get-url origin` y parsea el resultado.
// Devuelve un error con causa y siguiente paso si no puede detectarlo.
func DetectOrigin(dir string, r Runner) (Remote, error) {
	remote, err := originCause(dir, r)
	if err != nil {
		return Remote{}, fmt.Errorf("no se pudo detectar el repositorio: %v. Siguiente paso: revisa `git remote -v` o configura integrations.github.repo a mano", err)
	}
	return remote, nil
}

// DetectBase determina la rama por defecto de origin con
// `git symbolic-ref --quiet --short refs/remotes/origin/HEAD`. Si no puede,
// devuelve "main" y fallback=true para que el llamante lo informe.
func DetectBase(dir string, r Runner) (base string, fallback bool) {
	stdout, _, err := r.Run(dir, "git", "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")
	if err != nil {
		return "main", true
	}
	m := baseRef.FindStringSubmatch(strings.TrimSpace(stdout))
	if m == nil || strings.TrimSpace(m[1]) == "" {
		return "main", true
	}
	return strings.TrimSpace(m[1]), false
}

// Enable detecta el repositorio y la rama base y escribe integrations.github,
// preservando el resto de la configuración. Los avisos de `gh` no bloquean.
func Enable(p *paths.Project, r Runner, out io.Writer) int {
	path := filepath.Join(p.ReiDir(), "config.json")
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(out, "[FAIL]  .rei/config.json inválido: %v\n", err)
		fmt.Fprintln(out, "        Siguiente paso: corrige el JSON de .rei/config.json; no se ha modificado.")
		return 1
	}

	remote, err := DetectOrigin(p.Root, r)
	if err != nil {
		fmt.Fprintf(out, "[FAIL]  %v\n", err)
		return 1
	}

	prev := cfg.Integrations.GitHub
	base := ""
	strategy := ""
	if prev != nil {
		base = prev.Base
		strategy = prev.MergeStrategy
	}
	fallback := false
	if base == "" {
		base, fallback = DetectBase(p.Root, r)
	}
	if strategy == "" {
		strategy = "squash"
	}

	gh := config.GitHub{Enabled: true, Repo: remote.String(), Base: base, MergeStrategy: strategy}
	if err := config.SetGitHub(path, gh); err != nil {
		fmt.Fprintf(out, "[FAIL]  No se pudo escribir .rei/config.json: %v\n", err)
		fmt.Fprintln(out, "        Siguiente paso: revisa los permisos de .rei/config.json; el archivo no se ha modificado.")
		return 1
	}

	fmt.Fprintln(out, "== rei github ==")
	fmt.Fprintln(out, "[OK]    Integración GitHub activada.")
	fmt.Fprintf(out, "        repo: %s\n", gh.Repo)
	fmt.Fprintf(out, "        base: %s\n", gh.Base)
	fmt.Fprintf(out, "        mergeStrategy: %s\n", gh.MergeStrategy)
	if fallback {
		fmt.Fprintln(out, "[WARN]  No se pudo detectar la rama por defecto de `origin`; se usa `main`.")
		fmt.Fprintln(out, "        Siguiente paso: si no es correcta, edita `integrations.github.base` en .rei/config.json.")
	}
	probeGH(p, r, out)
	fmt.Fprintln(out, "[OK]    Configuración escrita en .rei/config.json.")
	return 0
}

// Disable apaga la integración conservando repo, base, mergeStrategy y el
// resto de claves. No contacta con GitHub ni toca links.json.
func Disable(p *paths.Project, out io.Writer) int {
	path := filepath.Join(p.ReiDir(), "config.json")
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(out, "[FAIL]  .rei/config.json inválido: %v\n", err)
		fmt.Fprintln(out, "        Siguiente paso: corrige el JSON de .rei/config.json; no se ha modificado.")
		return 1
	}
	gh := cfg.Integrations.GitHub
	if gh == nil || !gh.Enabled {
		fmt.Fprintln(out, "== rei github ==")
		fmt.Fprintln(out, "[OK]    GitHub ya estaba desactivado; no hay nada que cambiar.")
		return 0
	}
	if err := config.SetGitHubEnabled(path, false); err != nil {
		fmt.Fprintf(out, "[FAIL]  No se pudo escribir .rei/config.json: %v\n", err)
		fmt.Fprintln(out, "        Siguiente paso: revisa los permisos de .rei/config.json; el archivo no se ha modificado.")
		return 1
	}
	fmt.Fprintln(out, "== rei github ==")
	fmt.Fprintln(out, "[OK]    Integración GitHub desactivada.")
	fmt.Fprintln(out, "        Se conservan repo, base, mergeStrategy y los enlaces locales (links.json).")
	return 0
}

// Status informa del estado sin escribir nada. Devuelve 0 en cualquier estado
// legible y 1 solo si el proyecto o el JSON no son legibles.
func Status(p *paths.Project, r Runner, out io.Writer) int {
	path := filepath.Join(p.ReiDir(), "config.json")
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(out, "[FAIL]  .rei/config.json inválido: %v\n", err)
		fmt.Fprintln(out, "        Siguiente paso: corrige el JSON de .rei/config.json; no se ha modificado.")
		return 1
	}

	gh := cfg.Integrations.GitHub
	fmt.Fprintln(out, "== rei github ==")
	switch {
	case gh == nil:
		fmt.Fprintln(out, "[INFO]  Estado: no configurada.")
	case gh.Enabled:
		fmt.Fprintln(out, "[OK]    Estado: activa.")
	default:
		fmt.Fprintln(out, "[INFO]  Estado: desactivada.")
	}
	if gh != nil {
		fmt.Fprintf(out, "        repo: %s\n", orNone(gh.Repo))
		fmt.Fprintf(out, "        base: %s\n", gh.EffectiveBase())
		fmt.Fprintf(out, "        mergeStrategy: %s\n", gh.EffectiveMergeStrategy())
	}

	remote, remoteErr := originCause(p.Root, r)
	switch {
	case remoteErr != nil:
		fmt.Fprintf(out, "        remoto origin: no detectado (%v).\n", remoteErr)
	case gh == nil || gh.Repo == "":
		fmt.Fprintf(out, "        remoto origin: %s\n", remote.String())
	default:
		relation := "no coincide"
		if strings.EqualFold(gh.Repo, remote.String()) {
			relation = "coincide"
		}
		fmt.Fprintf(out, "        remoto origin: %s (%s con integrations.github.repo)\n", remote.String(), relation)
	}

	ghOK, ghAuth := ghDiagnosis(p, r)
	switch {
	case !ghOK:
		fmt.Fprintln(out, "        gh: no instalado.")
	case !ghAuth:
		fmt.Fprintln(out, "        gh: instalado, sin autenticar.")
	default:
		fmt.Fprintln(out, "        gh: instalado y autenticado.")
	}

	switch {
	case gh == nil:
		fmt.Fprintln(out, "Siguiente paso: `rei github enable` para activar la integración.")
	case !gh.Enabled:
		fmt.Fprintln(out, "Siguiente paso: `rei github enable` si quieres volver a activarla.")
	case !ghOK:
		fmt.Fprintln(out, "Siguiente paso: instala GitHub CLI (https://cli.github.com/) para crear issues y comentarios.")
	case !ghAuth:
		fmt.Fprintln(out, "Siguiente paso: `gh auth login` para autenticar GitHub CLI.")
	case remoteErr != nil:
		fmt.Fprintln(out, "Siguiente paso: revisa `git remote -v` y vuelve a ejecutar `rei github enable`.")
	default:
		fmt.Fprintln(out, "Sin acciones pendientes.")
	}
	return 0
}

// probeGH emite avisos no bloqueantes sobre la disponibilidad y autenticación
// de `gh`.
func probeGH(p *paths.Project, r Runner, out io.Writer) {
	if _, err := r.LookPath("gh"); err != nil {
		fmt.Fprintln(out, "[WARN]  `gh` no está instalado. Siguiente paso: instala GitHub CLI (https://cli.github.com/).")
		return
	}
	if _, stderr, err := r.Run(p.Root, "gh", "auth", "status"); err != nil {
		cause := firstLine(stderr)
		if cause != "" {
			fmt.Fprintf(out, "[WARN]  `gh` no está autenticado (%s). Siguiente paso: `gh auth login`.\n", cause)
		} else {
			fmt.Fprintln(out, "[WARN]  `gh` no está autenticado. Siguiente paso: `gh auth login`.")
		}
	}
}

// ghDiagnosis informa si `gh` está instalado y autenticado.
func ghDiagnosis(p *paths.Project, r Runner) (installed, authenticated bool) {
	if _, err := r.LookPath("gh"); err != nil {
		return false, false
	}
	if _, _, err := r.Run(p.Root, "gh", "auth", "status"); err != nil {
		return true, false
	}
	return true, true
}

func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}

func orNone(s string) string {
	if s == "" {
		return "(sin configurar)"
	}
	return s
}
