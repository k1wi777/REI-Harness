package initwizard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

// ResolveRootFrom determina el proyecto destino partiendo de dir: el ancestro
// más cercano que contenga .rei/ o, si no existe, dir mismo.
func ResolveRootFrom(dir string) (*paths.Project, error) {
	p, err := paths.FindFrom(dir)
	if err == nil {
		return p, nil
	}
	if errors.Is(err, paths.ErrNotFound) {
		return &paths.Project{Root: dir}, nil
	}
	return nil, err
}

// ResolveRoot aplica ResolveRootFrom sobre el directorio de trabajo actual.
func ResolveRoot() (*paths.Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return ResolveRootFrom(cwd)
}

// InstallSkeleton copia el esqueleto embebido en p.Root sin sobrescribir los
// archivos que ya existan. Devuelve 0 si lo completa y 1 si no pudo crear algún
// archivo o directorio.
//
// La omisión se decide por existencia (no comparando contenido): es lo que
// garantiza la idempotencia y que `rei init` nunca destruya personalizaciones
// del usuario (AGENTS.md, .rei/config.json, ...).
func InstallSkeleton(p *paths.Project, out io.Writer) int {
	exit := 0
	created, skipped := 0, 0
	baseline := map[string]string{}

	walkErr := walkEmbedded(func(dest string, embedded []byte) error {
		full := filepath.Join(p.Root, filepath.FromSlash(dest))

		if current, statErr := os.ReadFile(full); statErr == nil {
			if bytes.Equal(current, embedded) {
				// Sin cambios respecto al embebido: queda como baseline.
				baseline[dest] = hashBytes(embedded)
			}
			if dest == "AGENTS.md" {
				fmt.Fprintf(out, "[WARN]  %s ya existe; se conserva sin cambios.\n", dest)
			}
			fmt.Fprintf(out, "[SKIP]  %s (ya existe)\n", dest)
			skipped++
			return nil
		} else if !errors.Is(statErr, fs.ErrNotExist) {
			fmt.Fprintf(out, "[FAIL]  No se pudo comprobar %s: %v\n", dest, statErr)
			exit = 1
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			fmt.Fprintf(out, "[FAIL]  No se pudo crear el directorio de %s: %v\n", dest, err)
			exit = 1
			return nil
		}
		if err := os.WriteFile(full, embedded, 0o644); err != nil {
			fmt.Fprintf(out, "[FAIL]  No se pudo crear %s: %v\n", dest, err)
			exit = 1
			return nil
		}
		// Recién creado a partir del embebido: idéntico por construcción.
		baseline[dest] = hashBytes(embedded)
		fmt.Fprintf(out, "[OK]    %s\n", dest)
		created++
		return nil
	})
	if walkErr != nil {
		fmt.Fprintf(out, "[FAIL]  No se pudo recorrer el esqueleto: %v\n", walkErr)
		exit = 1
	}

	if err := saveManifest(p, baseline); err != nil {
		fmt.Fprintf(out, "[WARN]  No se pudo guardar el manifiesto %s: %v\n", ManifestPath, err)
	}

	fmt.Fprintf(out, "%d archivo(s) creado(s), %d omitido(s).\n", created, skipped)
	return exit
}
