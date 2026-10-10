// Package browser abre una URL en el navegador por defecto del sistema.
// Es una utilidad best-effort, reutilizada por `rei update` y `rei view`.
package browser

import (
	"os/exec"
	"runtime"
)

// openCommand es la costura de test: construye el comando según el SO. En
// producción apunta a command; los tests pueden sustituirla para verificar el
// comando elegido sin lanzar procesos.
var openCommand = command

// command construye el comando que abre url en el navegador por defecto según
// goos. No lo ejecuta.
func command(goos, url string) *exec.Cmd {
	switch goos {
	case "windows":
		return exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		return exec.Command("open", url)
	default:
		return exec.Command("xdg-open", url)
	}
}

// Open intenta abrir url en el navegador por defecto, según el SO. Es
// best-effort: devuelve el error para que el llamante lo ignore o lo reporte;
// la URL siempre se imprime como alternativa.
func Open(url string) error {
	return openCommand(runtime.GOOS, url).Start()
}
