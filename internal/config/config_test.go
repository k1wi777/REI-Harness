package config

import (
	"encoding/json"
	"errors"
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

func TestLoadIntegrationsAusenteYDesactivada(t *testing.T) {
	dir := t.TempDir()
	ausente := filepath.Join(dir, "ausente.json")
	if err := os.WriteFile(ausente, []byte(`{"checks":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(ausente)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Integrations.GitHub != nil {
		t.Fatalf("GitHub debería ser nil sin la clave: %+v", cfg.Integrations.GitHub)
	}

	desactivada := filepath.Join(dir, "desactivada.json")
	if err := os.WriteFile(desactivada, []byte(`{"integrations":{"github":{"enabled":false}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err = Load(desactivada)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Integrations.GitHub == nil || cfg.Integrations.GitHub.Enabled {
		t.Fatalf("GitHub debería existir y estar desactivada: %+v", cfg.Integrations.GitHub)
	}
}

func TestGitHubEffectiveDefaults(t *testing.T) {
	var nilGH *GitHub
	if got := nilGH.EffectiveBase(); got != "main" {
		t.Errorf("EffectiveBase(nil) = %q", got)
	}
	if got := nilGH.EffectiveMergeStrategy(); got != "squash" {
		t.Errorf("EffectiveMergeStrategy(nil) = %q", got)
	}
	empty := &GitHub{}
	if got := empty.EffectiveBase(); got != "main" {
		t.Errorf("EffectiveBase(vacío) = %q", got)
	}
	if got := empty.EffectiveMergeStrategy(); got != "squash" {
		t.Errorf("EffectiveMergeStrategy(vacío) = %q", got)
	}
	custom := &GitHub{Base: "develop", MergeStrategy: "merge"}
	if got := custom.EffectiveBase(); got != "develop" {
		t.Errorf("EffectiveBase = %q", got)
	}
	if got := custom.EffectiveMergeStrategy(); got != "merge" {
		t.Errorf("EffectiveMergeStrategy = %q", got)
	}
}

func TestSetGitHubPreservaClaves(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	original := `{
  "checks": [{"id": "V1", "description": "test", "command": ["go", "test", "./..."]}],
  "viewer": {"autoOpenOnReady": true},
  "integrations": {"otra": {"x": 1}},
  "desconocida": "valor"
}`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	gh := GitHub{Enabled: true, Repo: "owner/name", Base: "main", MergeStrategy: "squash"}
	if err := SetGitHub(path, gh); err != nil {
		t.Fatalf("SetGitHub: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if raw["desconocida"] != "valor" {
		t.Errorf("se perdió la clave desconocida: %v", raw["desconocida"])
	}
	if len(raw["checks"].([]any)) != 1 {
		t.Errorf("se perdieron los checks: %v", raw["checks"])
	}
	viewer := raw["viewer"].(map[string]any)
	if viewer["autoOpenOnReady"] != true {
		t.Errorf("se perdió viewer: %v", viewer)
	}
	integrations := raw["integrations"].(map[string]any)
	if _, ok := integrations["otra"]; !ok {
		t.Errorf("se perdió otra integración: %v", integrations)
	}
	got := integrations["github"].(map[string]any)
	for k, want := range map[string]any{
		"enabled": true, "repo": "owner/name", "base": "main", "mergeStrategy": "squash",
	} {
		if got[k] != want {
			t.Errorf("github.%s = %v, want %v", k, got[k], want)
		}
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Integrations.GitHub == nil || cfg.Integrations.GitHub.Repo != "owner/name" {
		t.Fatalf("roundtrip inesperado: %+v", cfg.Integrations.GitHub)
	}
}

func TestSetGitHubIdempotente(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	original := `{"checks":[{"id":"V1","description":"x","command":["true"]}],"viewer":{"autoOpenOnReady":false}}`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	gh := GitHub{Enabled: true, Repo: "o/n", Base: "develop", MergeStrategy: "squash"}
	if err := SetGitHub(path, gh); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := SetGitHub(path, gh); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatalf("SetGitHub no es idempotente:\n%s\n---\n%s", first, second)
	}
}

func TestSetGitHubEnabledConservaValores(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	original := `{
  "checks": [{"id": "V1", "description": "x", "command": ["true"]}],
  "viewer": {"autoOpenOnReady": true},
  "integrations": {"github": {"enabled": true, "repo": "o/n", "base": "develop", "mergeStrategy": "merge"}}
}`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetGitHubEnabled(path, false); err != nil {
		t.Fatalf("SetGitHubEnabled: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	gh := cfg.Integrations.GitHub
	if gh == nil {
		t.Fatal("GitHub no debería ser nil")
	}
	if gh.Enabled {
		t.Error("enabled debería ser false")
	}
	if gh.Repo != "o/n" || gh.Base != "develop" || gh.MergeStrategy != "merge" {
		t.Errorf("se perdieron valores: %+v", gh)
	}
	if len(cfg.Checks) != 1 || !cfg.Viewer.AutoOpenOnReady {
		t.Errorf("se perdieron claves: %+v", cfg)
	}
}

func TestSetGitHubEnabledNoConfigurado(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	original := `{"checks":[],"viewer":{"autoOpenOnReady":true}}`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	err := SetGitHubEnabled(path, false)
	if !errors.Is(err, ErrGitHubNotConfigured) {
		t.Fatalf("esperaba ErrGitHubNotConfigured, got %v", err)
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != original {
		t.Fatalf("el archivo no debe cambiar:\n%s", data)
	}
}

func TestSetGitHubJSONInvalidoIntacto(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	original := `{"checks": [`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetGitHub(path, GitHub{Enabled: true, Repo: "o/n"}); err == nil {
		t.Fatal("esperaba error con JSON inválido")
	}
	if err := SetGitHubEnabled(path, true); err == nil {
		t.Fatal("esperaba error con JSON inválido")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatalf("el archivo inválido no debe modificarse: %q", data)
	}
}
