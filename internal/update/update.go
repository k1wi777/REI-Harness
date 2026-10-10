package update

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/k1wi777/my-harness-SDD/internal/browser"
)

// Options configura una ejecución de `rei update`. Los campos de costura con
// valor cero se rellenan con los valores de producción dentro de Run; así los
// tests pueden inyectar httptest, un ejecutable temporal o un GOOS simulado.
type Options struct {
	CurrentVersion string
	Check          bool
	Stdout         io.Writer

	Client      *http.Client
	Executable  func() (string, error)
	GOOS        string
	GOARCH      string
	OpenBrowser func(url string) error
	APIBase     string
}

// Run ejecuta `rei update [--check]` y devuelve el código de salida (0/1).
func Run(opts Options) int {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Client == nil {
		opts.Client = &http.Client{Timeout: 30 * time.Second}
	}
	if opts.Executable == nil {
		opts.Executable = os.Executable
	}
	if opts.GOOS == "" {
		opts.GOOS = runtime.GOOS
	}
	if opts.GOARCH == "" {
		opts.GOARCH = runtime.GOARCH
	}
	if opts.OpenBrowser == nil {
		opts.OpenBrowser = browser.Open
	}
	if opts.APIBase == "" {
		opts.APIBase = defaultAPIBase
	}

	out := opts.Stdout
	dev := isDev(opts.CurrentVersion)

	rel, err := fetchLatestRelease(opts.Client, opts.APIBase)
	if err != nil {
		// Sin releases publicadas no es un fallo: nada que actualizar.
		if errors.Is(err, errNoReleases) {
			fmt.Fprintln(out, "No hay releases publicadas todavía. Nada que actualizar.")
			return 0
		}
		fmt.Fprintf(out, "ERROR: no se pudo consultar la última release: %v\n", err)
		return 1
	}
	latest := strings.TrimPrefix(strings.TrimSpace(rel.TagName), "v")

	if dev {
		fmt.Fprintln(out, "AVISO: la versión actual es 'dev' (compilación local), no una release publicada.")
	}

	// R4: sin actualización disponible.
	if !dev {
		if cur, ok := parseSemver(opts.CurrentVersion); ok {
			if lat, ok := parseSemver(rel.TagName); ok && compareSemver(lat, cur) <= 0 {
				fmt.Fprintf(out, "Ya estás en la última versión (%s).\n", strings.TrimSpace(opts.CurrentVersion))
				return 0
			}
		}
	}

	// R9: modo --check, solo informa.
	if opts.Check {
		fmt.Fprintf(out, "Última release disponible: %s (actual: %s).\n", rel.TagName, opts.CurrentVersion)
		fmt.Fprintln(out, "Ejecuta `rei update` para instalarla.")
		return 0
	}

	if dev {
		fmt.Fprintf(out, "Instalando la última release: %s\n", rel.TagName)
	} else {
		fmt.Fprintf(out, "Actualizando %s -> %s\n", opts.CurrentVersion, rel.TagName)
	}

	// R8: Windows no puede reemplazar el .exe en ejecución.
	if opts.GOOS == "windows" {
		fmt.Fprintln(out, "En Windows no es posible reemplazar el ejecutable en ejecución.")
		fmt.Fprintf(out, "Descarga la nueva versión desde: %s\n", releasesPage)
		if err := opts.OpenBrowser(releasesPage); err != nil {
			fmt.Fprintf(out, "(No se pudo abrir el navegador automáticamente: %v)\n", err)
		}
		return 0
	}

	// R5: selección del asset y de checksums.
	name := assetName(latest, opts.GOOS, opts.GOARCH)
	asset, ok := findAsset(rel, name)
	if !ok {
		fmt.Fprintf(out, "AVISO: la release %s no publica el asset %s.\n", rel.TagName, name)
		fmt.Fprintf(out, "Descarga la nueva versión manualmente desde: %s\n", releasesPage)
		return 1
	}
	cs, ok := findChecksums(rel)
	if !ok {
		fmt.Fprintf(out, "AVISO: la release %s no incluye checksums.txt.\n", rel.TagName)
		fmt.Fprintf(out, "Descarga la nueva versión manualmente desde: %s\n", releasesPage)
		return 1
	}

	tmpDir, err := os.MkdirTemp("", "rei-update-*")
	if err != nil {
		fmt.Fprintf(out, "ERROR: no se pudo crear el directorio temporal: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)

	// R6/R12: descarga.
	archivePath := filepath.Join(tmpDir, name)
	if err := downloadToFile(opts.Client, asset.URL, archivePath); err != nil {
		fmt.Fprintf(out, "ERROR: no se pudo descargar la actualización: %v\n", err)
		return 1
	}
	checksums, err := downloadToMemory(opts.Client, cs.URL)
	if err != nil {
		fmt.Fprintf(out, "ERROR: no se pudo descargar checksums.txt: %v\n", err)
		return 1
	}
	// R11: integridad antes de tocar el binario.
	if err := verifySHA256(archivePath, checksums, name); err != nil {
		fmt.Fprintf(out, "ERROR: %v; no se modificó el binario.\n", err)
		return 1
	}

	binPath := filepath.Join(tmpDir, "rei")
	if opts.GOOS == "windows" {
		binPath += ".exe"
	}
	if err := extractBinary(archivePath, binPath, opts.GOOS); err != nil {
		fmt.Fprintf(out, "ERROR: no se pudo extraer el binario: %v\n", err)
		return 1
	}

	dest, err := resolveExecutable(opts.Executable)
	if err != nil {
		fmt.Fprintf(out, "ERROR: no se pudo localizar el binario en ejecución: %v\n", err)
		return 1
	}
	// R7/R12/R13: sustitución atómica con backup y restauración.
	if err := replaceBinary(binPath, dest); err != nil {
		fmt.Fprintf(out, "ERROR: %v\n", err)
		return 1
	}
	fmt.Fprintf(out, "Actualizado a %s.\n", rel.TagName)
	return 0
}
