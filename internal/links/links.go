// Package links lee y valida el índice local .rei/specs/<id>/links.json.
//
// El índice es opcional y recuperable: nunca es la fuente de verdad del Work
// Item. Un archivo ausente, ilegible o incompleto se informa como aviso y se
// ignora; nunca bloquea el workflow. El paquete solo usa la biblioteca estándar
// y no conoce GitHub, la CLI ni el estado del proyecto.
package links

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// Base registra la rama y el commit de origen de un Work Item.
type Base struct {
	Branch string `json:"branch"`
	SHA    string `json:"sha"`
}

// Epic vincula el Work Item con una épica.
type Epic struct {
	Slug  string `json:"slug"`
	Issue int    `json:"issue"`
	URL   string `json:"url"`
}

// Links es el contenido de links.json. Los campos extra se ignoran al decodificar
// para no romper la compatibilidad futura.
type Links struct {
	Provider string `json:"provider"`
	Repo     string `json:"repo"`
	Issue    int    `json:"issue"`
	URL      string `json:"url"`
	Base     *Base  `json:"base,omitempty"`
	Epic     *Epic  `json:"epic,omitempty"`
}

var (
	identRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	shaRe   = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)
)

// Read carga links.json. La ausencia del archivo se distingue con
// errors.Is(err, fs.ErrNotExist); un JSON mal formado devuelve error. Un archivo
// sintácticamente válido pero incompleto se devuelve sin error: sus
// incumplimientos se consultan con Problems.
func Read(path string) (Links, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Links{}, err
		}
		return Links{}, fmt.Errorf("no se pudo leer: %w", err)
	}
	var l Links
	if err := json.Unmarshal(data, &l); err != nil {
		return Links{}, fmt.Errorf("JSON inválido: %w", err)
	}
	return l, nil
}

// Problems devuelve los incumplimientos de forma de links.json. Una lista vacía
// indica que el índice es válido. No consulta la red ni resuelve el nombre
// canónico del repositorio.
func Problems(l Links) []string {
	var problems []string
	add := func(format string, a ...any) {
		problems = append(problems, fmt.Sprintf(format, a...))
	}

	if l.Provider == "" {
		add("provider vacío (se espera github)")
	} else if l.Provider != "github" {
		add("provider %q no soportado (v1 solo admite github)", l.Provider)
	}
	if !ValidRepo(l.Repo) {
		add("repo %q inválido (se espera owner/name)", l.Repo)
	}
	if l.Issue <= 0 {
		add("issue debe ser un entero mayor que cero")
	}
	if !absoluteURL(l.URL) {
		add("url %q inválida (se espera una URL http/https absoluta con host y ruta)", l.URL)
	}
	if l.Base != nil {
		if strings.TrimSpace(l.Base.Branch) == "" {
			add("base.branch vacío")
		}
		if !shaRe.MatchString(l.Base.SHA) {
			add("base.sha %q inválido (se esperan 7–40 caracteres hexadecimales)", l.Base.SHA)
		}
	}
	if l.Epic != nil {
		if strings.TrimSpace(l.Epic.Slug) == "" {
			add("epic.slug vacío")
		}
		if l.Epic.Issue <= 0 {
			add("epic.issue debe ser un entero mayor que cero")
		}
		if !absoluteURL(l.Epic.URL) {
			add("epic.url %q inválida (se espera una URL http/https absoluta con host y ruta)", l.Epic.URL)
		}
	}
	return problems
}

// ValidRepo indica si s tiene la forma owner/name usada por el remoto de GitHub:
// exactamente dos segmentos, sin host, sin `.git` y sin barra extra.
func ValidRepo(s string) bool {
	repo := strings.TrimSpace(s)
	if repo == "" || repo != s {
		return false
	}
	if strings.Contains(repo, "://") || strings.ContainsAny(repo, "@:") {
		return false
	}
	if strings.HasSuffix(repo, ".git") {
		return false
	}
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return false
	}
	for _, part := range parts {
		if part == "." || part == ".." || !identRe.MatchString(part) {
			return false
		}
	}
	return true
}

// absoluteURL indica si s es una URL http/https absoluta con host y ruta no
// vacíos. No exige que coincida con repo/issue: los renames de GitHub mantienen
// redirects legítimos.
func absoluteURL(s string) bool {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != "" && strings.Trim(u.Path, "/") != ""
}
