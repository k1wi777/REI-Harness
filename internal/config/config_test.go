package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadDefaultFalseSinClave comprueba que la ausencia de `viewer` no falla y
// deja AutoOpenOnReady en false (R34).
func TestLoadDefaultFalseSinClave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"checks":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Viewer.AutoOpenOnReady {
		t.Fatal("autoOpenOnReady debería ser false por defecto")
	}
}

func TestLoadSinArchivo(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "no-existe.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Viewer.AutoOpenOnReady {
		t.Fatal("autoOpenOnReady debería ser false sin archivo")
	}
}

// TestSetAutoOpenOnReadyPreservaClaves verifica que solo cambia la clave
// viewer.autoOpenOnReady y conserva checks y otras claves (R35).
func TestSetAutoOpenOnReadyPreservaClaves(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	original := `{
  "checks": [
    {"id": "V1", "description": "test", "command": ["go", "test", "./..."]}
  ],
  "otra": "valor",
  "viewer": {"autoOpenOnReady": false}
}`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := SetAutoOpenOnReady(path, true); err != nil {
		t.Fatalf("SetAutoOpenOnReady: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("JSON inválido tras escribir: %v", err)
	}
	if raw["otra"] != "valor" {
		t.Errorf("se perdió la clave 'otra': %v", raw["otra"])
	}
	viewer, ok := raw["viewer"].(map[string]any)
	if !ok || viewer["autoOpenOnReady"] != true {
		t.Fatalf("viewer.autoOpenOnReady no quedó en true: %v", raw["viewer"])
	}
	checks, ok := raw["checks"].([]any)
	if !ok || len(checks) != 1 {
		t.Fatalf("se perdieron los checks: %v", raw["checks"])
	}

	// Y el roundtrip tipado lo confirma.
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Viewer.AutoOpenOnReady || len(cfg.Checks) != 1 {
		t.Fatalf("config recargada inesperada: %+v", cfg)
	}

	// Volver a desactivar solo cambia esa clave.
	if err := SetAutoOpenOnReady(path, false); err != nil {
		t.Fatal(err)
	}
	cfg, err = Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Viewer.AutoOpenOnReady {
		t.Fatal("autoOpenOnReady debería volver a false")
	}
	if len(cfg.Checks) != 1 {
		t.Fatalf("los checks deben persistir: %+v", cfg.Checks)
	}
}

// TestSetAutoOpenOnReadyCreaArchivo cubre el caso de un proyecto sin config.
func TestSetAutoOpenOnReadyCreaArchivo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := SetAutoOpenOnReady(path, true); err != nil {
		t.Fatalf("SetAutoOpenOnReady: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Viewer.AutoOpenOnReady {
		t.Fatal("autoOpenOnReady debería ser true")
	}
}
