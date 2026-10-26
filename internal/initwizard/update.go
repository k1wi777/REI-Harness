package initwizard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1wi777/REI-Harness/internal/paths"
)

// isPersonalizationDoc indica si una ruta relativa pertenece a la lista
// canónica de documentos de personalización (Docs): AGENTS.md y
// .rei/docs/project/{architecture,conventions,verification}.md.
func isPersonalizationDoc(path string) bool {
	for _, d := range Docs {
		if d.Path == path {
			return true
		}
	}
	return false
}

// Update refresca el esqueleto embebido en un proyecto ya inicializado sin
// pisar la personalización del usuario ni su estado (.rei/specs/** y
// .rei/progress/** nunca se recorren: el walk solo visita las raíces
// embebidas del esqueleto y del harness).
//
// Política (ver plan.md):
//   - No existe -> se crea.
//   - Igual al embebido -> sin cambios.
//   - Documento de personalización que difiere y aún contiene el marcador -> se
//     actualiza; sin marcador -> se conserva salvo --force.
//   - Archivo de harness que difiere: se actualiza solo si el manifiesto
//     confirma que sigue intacto desde la instalación; si no, se conserva salvo
//     --force.
//
// Devuelve 0 si completa (aunque conserve archivos) y 1 si falla una lectura o
// escritura del esqueleto.
func Update(p *paths.Project, force bool, out io.Writer) int {
	fmt.Fprintln(out, "== rei init --update ==")
	fmt.Fprintln(out)

	baseline, err := loadManifest(p)
	if err != nil {
		fmt.Fprintf(out, "[FAIL]  No se pudo leer %s: %v\n", ManifestPath, err)
		return 1
	}

	exit := 0
	created, updated, unchanged, skipped := 0, 0, 0, 0
	next := map[string]string{}

	walkErr := walkEmbedded(func(dest string, embedded []byte) error {
		full := filepath.Join(p.Root, filepath.FromSlash(dest))

		current, readErr := os.ReadFile(full)
		switch {
		case errors.Is(readErr, fs.ErrNotExist):
			if writeSkeletonFile(full, embedded, dest, out) != 0 {
				exit = 1
				return nil
			}
			fmt.Fprintf(out, "[NEW]   %s\n", dest)
			created++
			next[dest] = hashBytes(embedded)

		case readErr != nil:
			fmt.Fprintf(out, "[FAIL]  No se pudo leer %s: %v\n", dest, readErr)
			exit = 1
			return nil

		case bytes.Equal(current, embedded):
			fmt.Fprintf(out, "[OK]    %s\n", dest)
			unchanged++
			next[dest] = hashBytes(embedded)

		default:
			if shouldUpdate(dest, current, baseline, force) {
				if writeSkeletonFile(full, embedded, dest, out) != 0 {
					exit = 1
					return nil
				}
				fmt.Fprintf(out, "[UPD]   %s\n", dest)
				updated++
				next[dest] = hashBytes(embedded)
			} else {
				fmt.Fprintf(out, "[SKIP]  %s (conservado; usa --force para sobrescribir)\n", dest)
				skipped++
			}
		}
		return nil
	})
	if walkErr != nil {
		fmt.Fprintf(out, "[FAIL]  No se pudo recorrer el esqueleto: %v\n", walkErr)
		exit = 1
	}

	if err := saveManifest(p, next); err != nil {
		fmt.Fprintf(out, "[WARN]  No se pudo guardar el manifiesto %s: %v\n", ManifestPath, err)
	}

	fmt.Fprintf(out, "%d nuevo(s), %d actualizado(s), %d sin cambios, %d omitido(s).\n",
		created, updated, unchanged, skipped)
	return exit
}

// shouldUpdate decide si un archivo del esqueleto que difiere del embebido debe
// sobrescribirse. Con --force siempre; si no, aplica la regla marker-aware para
// los documentos de personalización y el manifiesto para el resto del harness.
func shouldUpdate(path string, current []byte, baseline map[string]string, force bool) bool {
	if force {
		return true
	}
	if isPersonalizationDoc(path) {
		return strings.Contains(string(current), marker)
	}
	h, ok := baseline[path]
	return ok && h == hashBytes(current)
}

// writeSkeletonFile crea los directorios necesarios y escribe contenido en dest
// con permisos 0644. Devuelve 1 si falla (y reporta [FAIL]).
func writeSkeletonFile(dest string, content []byte, rel string, out io.Writer) int {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		fmt.Fprintf(out, "[FAIL]  No se pudo crear el directorio de %s: %v\n", rel, err)
		return 1
	}
	if err := os.WriteFile(dest, content, 0o644); err != nil {
		fmt.Fprintf(out, "[FAIL]  No se pudo escribir %s: %v\n", rel, err)
		return 1
	}
	return 0
}
