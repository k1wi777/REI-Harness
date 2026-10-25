package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
)

// ErrGitHubNotConfigured indica que integrations.github no existe en la
// configuración; no se crea implícitamente al cambiar solo `enabled`.
var ErrGitHubNotConfigured = errors.New("integrations.github no está configurado")

// Check es una verificación del proyecto (checkpoint rápido).
type Check struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Command     []string `json:"command"`
}

// Viewer agrupa la configuración del visor de especificaciones (rei view).
type Viewer struct {
	AutoOpenOnReady bool `json:"autoOpenOnReady"`
}

// GitHub es la configuración de la integración con GitHub Issues.
type GitHub struct {
	Enabled       bool   `json:"enabled"`
	Repo          string `json:"repo,omitempty"`
	Base          string `json:"base,omitempty"`
	MergeStrategy string `json:"mergeStrategy,omitempty"`
}

// EffectiveBase devuelve la rama base configurada o "main" si está vacía.
func (g *GitHub) EffectiveBase() string {
	if g == nil || g.Base == "" {
		return "main"
	}
	return g.Base
}

// EffectiveMergeStrategy devuelve la estrategia de merge configurada o
// "squash" si está vacía.
func (g *GitHub) EffectiveMergeStrategy() string {
	if g == nil || g.MergeStrategy == "" {
		return "squash"
	}
	return g.MergeStrategy
}

// Integrations agrupa integraciones externas opcionales.
type Integrations struct {
	GitHub *GitHub `json:"github,omitempty"`
}

// Config es el contenido de .rei/config.json.
type Config struct {
	Checks       []Check      `json:"checks"`
	Viewer       Viewer       `json:"viewer"`
	Integrations Integrations `json:"integrations"`
}

// Load lee .rei/config.json. Si no existe, devuelve una configuración vacía. Si
// la clave viewer (o autoOpenOnReady) falta, el valor cero es false y no falla.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// SetAutoOpenOnReady escribe viewer.autoOpenOnReady preservando el resto del
// archivo (checks y cualquier otra clave), de forma atómica e indentada.
func SetAutoOpenOnReady(path string, enabled bool) error {
	raw := map[string]any{}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
	case os.IsNotExist(err):
		// Se crea el archivo con la única clave escrita.
	default:
		return err
	}

	viewer, ok := raw["viewer"].(map[string]any)
	if !ok || viewer == nil {
		viewer = map[string]any{}
	}
	viewer["autoOpenOnReady"] = enabled
	raw["viewer"] = viewer

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(raw); err != nil {
		return err
	}
	return writeFileAtomic(path, buf.Bytes())
}

// SetGitHub escribe integrations.github con las cuatro claves de la
// integración, preservando checks, viewer, otras integraciones y cualquier otra
// clave del archivo. Es atómico: si falla, el archivo original queda intacto.
func SetGitHub(path string, gh GitHub) error {
	raw, err := readObject(path)
	if err != nil {
		return err
	}
	integrations := objectAt(raw, "integrations")
	integrations["github"] = map[string]any{
		"enabled":       gh.Enabled,
		"repo":          gh.Repo,
		"base":          gh.Base,
		"mergeStrategy": gh.MergeStrategy,
	}
	raw["integrations"] = integrations
	return writeObject(path, raw)
}

// SetGitHubEnabled cambia únicamente enabled. Devuelve ErrGitHubNotConfigured
// si la clave no existe: nunca la crea implícitamente.
func SetGitHubEnabled(path string, enabled bool) error {
	cfg, err := Load(path)
	if err != nil {
		return err
	}
	if cfg.Integrations.GitHub == nil {
		return ErrGitHubNotConfigured
	}
	raw, err := readObject(path)
	if err != nil {
		return err
	}
	integrations := objectAt(raw, "integrations")
	github := objectAt(integrations, "github")
	github["enabled"] = enabled
	integrations["github"] = github
	raw["integrations"] = integrations
	return writeObject(path, raw)
}

// readObject lee el archivo como un objeto JSON. Si no existe, devuelve un
// objeto vacío. Un JSON inválido devuelve error (nunca se sobrescribe).
func readObject(path string) (map[string]any, error) {
	raw := map[string]any{}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		if raw == nil {
			raw = map[string]any{}
		}
	case os.IsNotExist(err):
		// Se crea el archivo con la única clave escrita.
	default:
		return nil, err
	}
	return raw, nil
}

// objectAt devuelve el objeto hijo de raw[key], creándolo si falta o si no es
// un objeto.
func objectAt(raw map[string]any, key string) map[string]any {
	child, ok := raw[key].(map[string]any)
	if !ok || child == nil {
		child = map[string]any{}
	}
	return child
}

// writeObject serializa raw indentado (sin escapar HTML) y lo escribe de forma
// atómica.
func writeObject(path string, raw map[string]any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(raw); err != nil {
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
