package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/my-harness-SDD/internal/config"
	"github.com/k1wi777/my-harness-SDD/internal/meta"
)

func TestDescriptorRoundTripYPermisos(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	d := &descriptor{PID: 1234, Port: 9999, Token: "abc", WorkItem: id, StartedAt: "2026-01-01T00:00:00Z"}
	if err := writeDescriptor(p, d); err != nil {
		t.Fatalf("writeDescriptor: %v", err)
	}

	fi, err := os.Stat(descriptorPath(p))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("permisos = %v, want 0600", fi.Mode().Perm())
	}

	got, err := readDescriptor(p)
	if err != nil {
		t.Fatalf("readDescriptor: %v", err)
	}
	if *got != *d {
		t.Errorf("roundtrip = %+v, want %+v", got, d)
	}
}

func TestLiveDescriptorOBSOLETO(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	// PID fuera de rango: seguro que no vive.
	if err := writeDescriptor(p, &descriptor{PID: 1 << 30, WorkItem: id}); err != nil {
		t.Fatal(err)
	}
	if d := liveDescriptor(p); d != nil {
		t.Errorf("un pid muerto no debe considerarse vivo: %+v", d)
	}
	if _, err := os.Stat(descriptorPath(p)); !os.IsNotExist(err) {
		t.Errorf("el descriptor obsoleto debe eliminarse")
	}
}

func TestLiveDescriptorVivo(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	if err := writeDescriptor(p, &descriptor{PID: os.Getpid(), Port: 1, Token: "t", WorkItem: id}); err != nil {
		t.Fatal(err)
	}
	if d := liveDescriptor(p); d == nil {
		t.Fatal("el pid actual debe considerarse vivo")
	}
}

func TestStopViewerSinDescriptor(t *testing.T) {
	p, _ := newViewFixture(t, meta.StatusReady)
	var out strings.Builder
	if err := stopViewer(p, &out); err != nil {
		t.Fatalf("stopViewer: %v", err)
	}
	if !strings.Contains(out.String(), "No hay servidor") {
		t.Errorf("salida inesperada: %q", out.String())
	}
}

func TestStopViewerEliminaDescriptor(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	if err := writeDescriptor(p, &descriptor{PID: 1 << 30, WorkItem: id}); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := stopViewer(p, &out); err != nil {
		t.Fatalf("stopViewer: %v", err)
	}
	if _, err := os.Stat(descriptorPath(p)); !os.IsNotExist(err) {
		t.Errorf("stopViewer debe eliminar el descriptor")
	}
}

func TestOpenViewerReutiliza(t *testing.T) {
	p, id := newViewFixture(t, meta.StatusReady)
	if err := writeDescriptor(p, &descriptor{PID: os.Getpid(), Port: 4242, Token: "tok", WorkItem: id}); err != nil {
		t.Fatal(err)
	}
	var opened string
	var out strings.Builder
	err := openViewer(p, id, func(u string) error { opened = u; return nil }, &out)
	if err != nil {
		t.Fatalf("openViewer: %v", err)
	}
	if !strings.Contains(opened, "127.0.0.1:4242") || !strings.Contains(opened, "tok") {
		t.Errorf("URL abierta inesperada: %q", opened)
	}
	if !strings.Contains(out.String(), "ya activo") {
		t.Errorf("debe informar de la reutilización: %q", out.String())
	}
}

func TestSetAutoOpen(t *testing.T) {
	p, _ := newViewFixture(t, meta.StatusReady)
	if err := setAutoOpen(p, true); err != nil {
		t.Fatalf("setAutoOpen: %v", err)
	}
	cfg, err := config.Load(filepath.Join(p.ReiDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Viewer.AutoOpenOnReady {
		t.Error("autoOpenOnReady debería ser true")
	}
}
