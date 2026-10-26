package initwizard

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1wi777/REI-Harness/internal/paths"
)

// ManifestPath es la ruta, relativa a la raíz del proyecto, del manifiesto de
// hashes que registra los archivos del esqueleto que siguen idénticos al
// embebido (es decir, los que el usuario no ha modificado). Permite a
// `rei init --update` distinguir «sin tocar» de «modificado a mano» sin
// depender de marcadores. Vive fuera de specs/ y progress/, y no se embebe.
const ManifestPath = ".rei/install-manifest.json"

// manifest es el formato serializado del manifiesto: ruta relativa -> SHA-256
// hexadecimal del contenido.
type manifest struct {
	Files map[string]string `json:"files"`
}

// hashBytes devuelve el SHA-256 hexadecimal del contenido dado.
func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// manifestFile devuelve la ruta absoluta del manifiesto en el proyecto.
func manifestFile(p *paths.Project) string {
	return filepath.Join(p.Root, filepath.FromSlash(ManifestPath))
}

// loadManifest lee el manifiesto del proyecto. Devuelve nil sin error si no
// existe o si es inválido (no es motivo para abortar una actualización, solo
// implica perder el baseline). Devuelve error únicamente ante fallos de lectura
// distintos de «no existe».
func loadManifest(p *paths.Project) (map[string]string, error) {
	data, err := os.ReadFile(manifestFile(p))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, nil
	}
	if m.Files == nil {
		return nil, nil
	}
	return m.Files, nil
}

// saveManifest escribe el manifiesto con las entradas dadas. Las claves se
// serializan ordenadas (encoding/json ordena las claves de un map) e
// indentadas. Crea el directorio si falta y escribe con permisos 0644.
func saveManifest(p *paths.Project, entries map[string]string) error {
	if entries == nil {
		entries = map[string]string{}
	}
	// Normaliza las rutas a barras para que el manifiesto sea estable entre
	// plataformas.
	norm := make(map[string]string, len(entries))
	for k, v := range entries {
		norm[strings.ReplaceAll(k, "\\", "/")] = v
	}
	data, err := json.MarshalIndent(manifest{Files: norm}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dest := manifestFile(p)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o644)
}
