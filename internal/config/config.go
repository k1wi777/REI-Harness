package config

import (
	"bytes"
	"encoding/json"
	"os"
)

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

// Config es el contenido de .rei/config.json.
type Config struct {
	Checks []Check `json:"checks"`
	Viewer Viewer  `json:"viewer"`
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

func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
