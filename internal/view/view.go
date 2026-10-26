// Package view implementa `rei view`: sirve en el navegador una vista de solo
// lectura de un Work Item, con render seguro de Markdown, recarga en vivo por
// SSE y ciclo de vida del servidor local. Nunca muta documentos ni estado.
package view

import (
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/k1wi777/my-harness-SDD/internal/browser"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

// Mode selecciona el modo de ejecución de `rei view`.
type Mode string

const (
	// ModeForeground sirve en primer plano hasta Ctrl+C (por defecto).
	ModeForeground Mode = "foreground"
	// ModeOpen arranca/reutiliza el servidor de fondo y abre el navegador.
	ModeOpen Mode = "open"
	// ModeStop detiene el servidor de fondo del proyecto.
	ModeStop Mode = "stop"
	// ModeEnable activa la apertura automática del visor.
	ModeEnable Mode = "enable"
	// ModeDisable desactiva la apertura automática del visor.
	ModeDisable Mode = "disable"
	// ModeServe es el proceso interno de fondo (`--serve`).
	ModeServe Mode = "serve"
)

// Options configura una ejecución de `rei view`.
type Options struct {
	Mode        Mode
	ID          string
	Stdout      io.Writer
	OpenBrowser func(url string) error
}

// Run ejecuta el modo solicitado y devuelve el código de salida (0/1).
func Run(p *paths.Project, opts Options) int {
	out := opts.Stdout
	if out == nil {
		out = os.Stdout
	}
	open := opts.OpenBrowser
	if open == nil {
		open = browser.Open
	}

	switch opts.Mode {
	case ModeEnable:
		if err := setAutoOpen(p, true); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		fmt.Fprintln(out, "Apertura automática del visor activada.")
		return 0
	case ModeDisable:
		if err := setAutoOpen(p, false); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		fmt.Fprintln(out, "Apertura automática del visor desactivada.")
		return 0
	case ModeStop:
		if err := stopViewer(p, out); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	case ModeOpen:
		if err := openViewer(p, opts.ID, open, out); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		return 0
	case ModeServe:
		return runServe(p, opts.ID)
	default:
		return runForeground(p, opts.ID, out)
	}
}

// runForeground sirve en el proceso actual y bloquea hasta Ctrl+C.
func runForeground(p *paths.Project, id string, out io.Writer) int {
	s, err := newServer(p, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	ln, err := s.Listen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	url := s.URL(portOf(ln))
	fmt.Fprintf(out, "Visor disponible en %s\n", url)
	fmt.Fprintln(out, "Pulsa Ctrl+C para detenerlo.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	go func() {
		<-sig
		s.requestClose()
	}()

	if err := s.Serve(ln); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	return 0
}
