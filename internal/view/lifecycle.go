package view

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/k1wi777/my-harness-SDD/internal/config"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

// descriptorName es el registro operativo temporal del servidor de fondo. Vive
// mientras el servidor está activo y se elimina al cerrarlo; está gitignorado.
const descriptorName = ".view-server.json"

// descriptor describe el servidor de fondo vivo para un proyecto.
type descriptor struct {
	PID       int    `json:"pid"`
	Port      int    `json:"port"`
	Token     string `json:"token"`
	WorkItem  string `json:"workItem"`
	StartedAt string `json:"startedAt"`
}

func descriptorPath(p *paths.Project) string {
	return filepath.Join(p.ReiDir(), descriptorName)
}

func readDescriptor(p *paths.Project) (*descriptor, error) {
	data, err := os.ReadFile(descriptorPath(p))
	if err != nil {
		return nil, err
	}
	var d descriptor
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// writeDescriptor escribe el descriptor de forma atómica con permisos 0600.
func writeDescriptor(p *paths.Project, d *descriptor) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	target := descriptorPath(p)
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}

func removeDescriptor(p *paths.Project) error {
	err := os.Remove(descriptorPath(p))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// liveDescriptor devuelve el descriptor si su proceso sigue vivo; si el pid ya
// no vive, se considera obsoleto y se elimina.
func liveDescriptor(p *paths.Project) *descriptor {
	d, err := readDescriptor(p)
	if err != nil || d == nil {
		return nil
	}
	if !pidAlive(d.PID) {
		_ = removeDescriptor(p)
		return nil
	}
	return d
}

// setAutoOpen ajusta únicamente viewer.autoOpenOnReady en .rei/config.json.
func setAutoOpen(p *paths.Project, enabled bool) error {
	return config.SetAutoOpenOnReady(filepath.Join(p.ReiDir(), "config.json"), enabled)
}

// openViewer reutiliza el servidor vivo o lanza uno de fondo, y abre el
// navegador con su URL.
func openViewer(p *paths.Project, id string, open func(string) error, out io.Writer) error {
	if d := liveDescriptor(p); d != nil {
		if d.WorkItem != "" && d.WorkItem != id {
			// El servidor vivo sirve otro Work Item: se detiene y se arranca uno nuevo.
			if err := stopViewer(p, out); err != nil {
				return err
			}
		} else {
			url := viewerURL(d.Port, d.Token)
			fmt.Fprintf(out, "Visor ya activo: %s\n", url)
			return open(url)
		}
	}

	if err := spawnBackground(p, id); err != nil {
		return err
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if d := liveDescriptor(p); d != nil {
			url := viewerURL(d.Port, d.Token)
			fmt.Fprintf(out, "Visor iniciado: %s\n", url)
			return open(url)
		}
		if time.Now().After(deadline) {
			return errors.New("el servidor del visor no arrancó a tiempo")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func viewerURL(port int, token string) string {
	return fmt.Sprintf("http://127.0.0.1:%d/?token=%s", port, token)
}

// spawnBackground relanza el propio binario con `view <id> --serve` desacoplado
// del terminal (setsid/detach).
func spawnBackground(p *paths.Project, id string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "view", id, "--serve")
	cmd.Dir = p.Root
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	configureDetach(cmd)
	return cmd.Start()
}

// stopViewer termina el servidor de fondo y elimina su descriptor.
func stopViewer(p *paths.Project, out io.Writer) error {
	d, err := readDescriptor(p)
	if err != nil {
		fmt.Fprintln(out, "No hay servidor del visor en segundo plano.")
		return nil
	}
	if pidAlive(d.PID) {
		if err := terminateProcess(d.PID); err != nil {
			return err
		}
	}
	if err := removeDescriptor(p); err != nil {
		return err
	}
	fmt.Fprintln(out, "Servidor del visor detenido.")
	return nil
}

// runServe es el proceso de fondo (`view <id> --serve`): registra el
// descriptor, sirve y lo elimina al cerrarse.
func runServe(p *paths.Project, id string) int {
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
	s.background = true
	s.closeOnStatusChange = true

	d := &descriptor{
		PID:       os.Getpid(),
		Port:      portOf(ln),
		Token:     s.token,
		WorkItem:  id,
		StartedAt: time.Now().Format(time.RFC3339),
	}
	if err := writeDescriptor(p, d); err != nil {
		_ = ln.Close()
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	s.onClose = func() { _ = removeDescriptor(p) }
	defer func() { _ = removeDescriptor(p) }()

	if err := s.Serve(ln); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	return 0
}
