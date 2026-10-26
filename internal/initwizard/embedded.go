package initwizard

import (
	"fmt"
	"io/fs"
	"strings"

	reiskel "github.com/k1wi777/my-harness-SDD"
)

// skeletonRoot es el directorio raíz de reiskel.Skeleton; su prefijo no forma
// parte del proyecto destino y se elimina al calcular la ruta de destino.
const (
	skeletonRoot   = "skeleton"
	skeletonPrefix = skeletonRoot + "/"
)

// walkEmbedded recorre las dos raíces embebidas y entrega a fn, por cada
// archivo, la ruta destino relativa a la raíz del proyecto y su contenido.
//
//   - reiskel.Skeleton aporta el esqueleto neutral en la raíz del proyecto
//     (AGENTS.md, .rei/config.json y .rei/docs/project/**) sin el prefijo
//     skeleton/.
//   - reiskel.Harness aporta el material del arnés con su ruta canónica
//     (.rei/docs/harness/**, .rei/docs/usage.md, .rei/agents/**,
//     .rei/adapters/** y .rei/templates/**).
//
// Los directorios se omiten: los crea cada escritor con MkdirAll. El recorrido
// es determinista (cada FS en orden léxico, Skeleton antes que Harness).
func walkEmbedded(fn func(dest string, data []byte) error) error {
	err := fs.WalkDir(reiskel.Skeleton, skeletonRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := reiskel.Skeleton.ReadFile(path)
		if err != nil {
			return fmt.Errorf("leer %s del esqueleto: %w", path, err)
		}
		dest := strings.TrimPrefix(path, skeletonPrefix)
		if dest == path || dest == "" {
			return fmt.Errorf("ruta de esqueleto inesperada: %s", path)
		}
		return fn(dest, data)
	})
	if err != nil {
		return err
	}

	return fs.WalkDir(reiskel.Harness, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." || d.IsDir() {
			return nil
		}
		data, err := reiskel.Harness.ReadFile(path)
		if err != nil {
			return fmt.Errorf("leer %s del harness: %w", path, err)
		}
		return fn(path, data)
	})
}
