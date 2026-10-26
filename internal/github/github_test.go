package github

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1wi777/my-harness-SDD/internal/config"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

type runResult struct {
	stdout, stderr string
	err            error
}

type fakeRunner struct {
	paths map[string]string
	runs  map[string]runResult
}

func (f fakeRunner) LookPath(file string) (string, error) {
	if p, ok := f.paths[file]; ok {
		return p, nil
	}
	return "", errors.New("no encontrado en PATH")
}

func (f fakeRunner) Run(dir, name string, args ...string) (string, string, error) {
	key := name
	if len(args) > 0 {
		key += " " + strings.Join(args, " ")
	}
	if r, ok := f.runs[key]; ok {
		return r.stdout, r.stderr, r.err
	}
	return "", "comando no simulado: " + key, errors.New("no simulado")
}

func gitOK() map[string]runResult {
	return map[string]runResult{
		"git remote get-url origin":                                 {stdout: "git@github.com:owner/name.git\n"},
		"git symbolic-ref --quiet --short refs/remotes/origin/HEAD": {stdout: "origin/main\n"},
	}
}

func ghOK() map[string]string {
	return map[string]string{"gh": "/usr/bin/gh"}
}

func newProject(t *testing.T, cfg string) *paths.Project {
	t.Helper()
	root := t.TempDir()
	p := &paths.Project{Root: root}
	if err := os.MkdirAll(p.ReiDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if cfg != "" {
		if err := os.WriteFile(filepath.Join(p.ReiDir(), "config.json"), []byte(cfg), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func configPath(p *paths.Project) string { return filepath.Join(p.ReiDir(), "config.json") }

func loadConfig(t *testing.T, p *paths.Project) *config.Config {
	t.Helper()
	cfg, err := config.Load(configPath(p))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg
}

func TestParseRemote(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want Remote
	}{
		{"ssh scp", "git@github.com:owner/name.git", Remote{"github.com", "owner", "name"}},
		{"ssh scp sin git", "git@github.com:owner/name", Remote{"github.com", "owner", "name"}},
		{"https con git", "https://github.com/owner/name.git", Remote{"github.com", "owner", "name"}},
		{"https sin git", "https://github.com/owner/name", Remote{"github.com", "owner", "name"}},
		{"https barra final", "https://github.com/owner/name/", Remote{"github.com", "owner", "name"}},
		{"ssh url", "ssh://git@github.com/owner/name.git", Remote{"github.com", "owner", "name"}},
		{"mayúsculas en host", "https://GitHub.com/owner/name", Remote{"github.com", "owner", "name"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRemote(tc.raw)
			if err != nil {
				t.Fatalf("ParseRemote(%q): %v", tc.raw, err)
			}
			if got != tc.want {
				t.Fatalf("ParseRemote(%q) = %+v, want %+v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseRemoteErrores(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantSub string
	}{
		{"vacío", "   ", "vacío"},
		{"host distinto", "git@gitlab.com:owner/name.git", "github.com"},
		{"tres segmentos", "https://github.com/owner/group/name", "owner/name"},
		{"sin nombre", "https://github.com/owner", "owner/name"},
		{"punto", "git@github.com:./name.git", "inválidos"},
		{"dos puntos", "git@github.com:../name.git", "inválidos"},
		{"sin arroba", "github.com:owner/name", "no reconocido"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseRemote(tc.raw)
			if err == nil {
				t.Fatalf("ParseRemote(%q) debía fallar", tc.raw)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error = %q, esperaba %q", err, tc.wantSub)
			}
		})
	}
}

func TestParseRemoteNoExponeCredenciales(t *testing.T) {
	_, err := ParseRemote("https://user:secreto@gitlab.com/owner/name")
	if err == nil {
		t.Fatal("debía fallar por host")
	}
	if strings.Contains(err.Error(), "secreto") {
		t.Fatalf("el error filtra credenciales: %v", err)
	}
}

func TestDetectOrigin(t *testing.T) {
	r := fakeRunner{runs: map[string]runResult{
		"git remote get-url origin": {stdout: "  https://github.com/o/n.git  \n"},
	}}
	got, err := DetectOrigin(t.TempDir(), r)
	if err != nil {
		t.Fatalf("DetectOrigin: %v", err)
	}
	if got.String() != "o/n" {
		t.Fatalf("DetectOrigin = %q", got.String())
	}
}

func TestDetectOriginError(t *testing.T) {
	r := fakeRunner{runs: map[string]runResult{
		"git remote get-url origin": {stderr: "error: No such remote 'origin'", err: errors.New("exit 2")},
	}}
	_, err := DetectOrigin(t.TempDir(), r)
	if err == nil {
		t.Fatal("esperaba error")
	}
	for _, want := range []string{"No such remote", "git remote -v"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, debe contener %q", err, want)
		}
	}
}

func TestDetectBase(t *testing.T) {
	cases := []struct {
		name     string
		res      runResult
		wantBase string
		wantFB   bool
	}{
		{"main", runResult{stdout: "origin/main\n"}, "main", false},
		{"develop", runResult{stdout: "origin/develop\n"}, "develop", false},
		{"error", runResult{err: errors.New("sin HEAD")}, "main", true},
		{"vacío", runResult{stdout: "\n"}, "main", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := fakeRunner{runs: map[string]runResult{
				"git symbolic-ref --quiet --short refs/remotes/origin/HEAD": tc.res,
			}}
			base, fb := DetectBase(t.TempDir(), r)
			if base != tc.wantBase || fb != tc.wantFB {
				t.Fatalf("DetectBase = (%q, %v), want (%q, %v)", base, fb, tc.wantBase, tc.wantFB)
			}
		})
	}
}

func TestEnableExito(t *testing.T) {
	p := newProject(t, `{"checks":[{"id":"V1","description":"x","command":["true"]}],"viewer":{"autoOpenOnReady":true}}`)
	r := fakeRunner{paths: ghOK(), runs: gitOK()}
	r.runs["gh auth status"] = runResult{}

	var out strings.Builder
	if code := Enable(p, r, Options{}, &out); code != 0 {
		t.Fatalf("Enable = %d:\n%s", code, out.String())
	}
	cfg := loadConfig(t, p)
	gh := cfg.Integrations.GitHub
	if gh == nil || !gh.Enabled || gh.Repo != "owner/name" || gh.Base != "main" || gh.MergeStrategy != "squash" {
		t.Fatalf("config GitHub inesperada: %+v", gh)
	}
	if len(cfg.Checks) != 1 || !cfg.Viewer.AutoOpenOnReady {
		t.Fatalf("se perdieron claves: %+v", cfg)
	}
	for _, want := range []string{"activada", "owner/name", "base: main", "mergeStrategy: squash", "Configuración escrita"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("salida no contiene %q:\n%s", want, out.String())
		}
	}
}

func TestEnableAvisoSinGh(t *testing.T) {
	p := newProject(t, `{"checks":[]}`)
	r := fakeRunner{runs: gitOK()}
	var out strings.Builder
	if code := Enable(p, r, Options{}, &out); code != 0 {
		t.Fatalf("Enable = %d, esperaba 0:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "no está instalado") || !strings.Contains(out.String(), "cli.github.com") {
		t.Fatalf("falta el aviso de gh:\n%s", out.String())
	}
	if loadConfig(t, p).Integrations.GitHub == nil {
		t.Fatal("la configuración debía escribirse pese al aviso")
	}
}

func TestEnableAvisoSinAuth(t *testing.T) {
	p := newProject(t, `{"checks":[]}`)
	r := fakeRunner{paths: ghOK(), runs: gitOK()}
	r.runs["gh auth status"] = runResult{stderr: "You are not logged into any GitHub hosts.", err: errors.New("exit 1")}
	var out strings.Builder
	if code := Enable(p, r, Options{}, &out); code != 0 {
		t.Fatalf("Enable = %d, esperaba 0:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "no está autenticado") || !strings.Contains(out.String(), "gh auth login") {
		t.Fatalf("falta el aviso de auth:\n%s", out.String())
	}
}

func TestEnableSinOriginNoEscribe(t *testing.T) {
	p := newProject(t, `{"checks":[]}`)
	r := fakeRunner{runs: map[string]runResult{
		"git remote get-url origin": {stderr: "No such remote", err: errors.New("exit 2")},
	}}
	var out strings.Builder
	if code := Enable(p, r, Options{}, &out); code != 1 {
		t.Fatalf("Enable = %d, esperaba 1", code)
	}
	if loadConfig(t, p).Integrations.GitHub != nil {
		t.Fatal("no debía escribir configuración")
	}
}

func TestEnableHostInvalidoNoEscribe(t *testing.T) {
	p := newProject(t, `{"checks":[]}`)
	r := fakeRunner{runs: map[string]runResult{
		"git remote get-url origin": {stdout: "git@gitlab.com:owner/name.git"},
	}}
	var out strings.Builder
	if code := Enable(p, r, Options{}, &out); code != 1 {
		t.Fatalf("Enable = %d, esperaba 1", code)
	}
	if loadConfig(t, p).Integrations.GitHub != nil {
		t.Fatal("no debía escribir configuración")
	}
	if !strings.Contains(out.String(), "github.com") {
		t.Fatalf("salida debe mencionar el host soportado:\n%s", out.String())
	}
}

func TestEnableConservaValoresPrevios(t *testing.T) {
	p := newProject(t, `{"integrations":{"github":{"enabled":false,"repo":"viejo/repo","base":"develop","mergeStrategy":"merge"}}}`)
	r := fakeRunner{paths: ghOK(), runs: gitOK()}
	var out strings.Builder
	if code := Enable(p, r, Options{}, &out); code != 0 {
		t.Fatalf("Enable = %d:\n%s", code, out.String())
	}
	gh := loadConfig(t, p).Integrations.GitHub
	if gh.Base != "develop" || gh.MergeStrategy != "merge" {
		t.Fatalf("debía conservar base/mergeStrategy: %+v", gh)
	}
	if gh.Repo != "viejo/repo" {
		t.Fatalf("con repo previo no debe detectar ni sobrescribir: %+v", gh)
	}
}

func TestValidateOptions(t *testing.T) {
	valid := []Options{
		{},
		{Repo: "owner/name"},
		{Base: "develop"},
		{Strategy: "squash"},
		{Repo: "a/b", Base: "main", Strategy: "rebase"},
	}
	for _, o := range valid {
		if err := ValidateOptions(o); err != nil {
			t.Errorf("ValidateOptions(%+v) = %v, want nil", o, err)
		}
	}
	invalid := []Options{
		{Repo: "owner"},
		{Repo: "https://github.com/owner/name"},
		{Repo: "owner/name.git"},
		{Base: " "},
		{Base: "feat con espacios"},
		{Strategy: "octopus"},
		{RepoSet: true},
		{BaseSet: true},
		{StrategySet: true},
	}
	for _, o := range invalid {
		if err := ValidateOptions(o); err == nil {
			t.Errorf("ValidateOptions(%+v) = nil, want error", o)
		}
	}
}

// TestEnableFlagsGananPrecedencia comprueba flag > previo y que no se consulta
// git cuando el flag ya resuelve cada valor.
func TestEnableFlagsGananPrecedencia(t *testing.T) {
	p := newProject(t, `{"integrations":{"github":{"enabled":true,"repo":"previo/repo","base":"prev-base","mergeStrategy":"merge"}}}`)
	r := fakeRunner{paths: ghOK(), runs: map[string]runResult{"gh auth status": {}}}
	var out strings.Builder
	code := Enable(p, r, Options{Repo: "flag/repo", Base: "flag-base", Strategy: "rebase"}, &out)
	if code != 0 {
		t.Fatalf("Enable = %d:\n%s", code, out.String())
	}
	gh := loadConfig(t, p).Integrations.GitHub
	if gh.Repo != "flag/repo" || gh.Base != "flag-base" || gh.MergeStrategy != "rebase" {
		t.Fatalf("los flags debían ganar: %+v", gh)
	}
	if !gh.Enabled {
		t.Fatal("enable debía activar la integración")
	}
}

// TestEnableFlagsSinGit comprueba que con todos los flags no hace falta ni
// `origin` ni `origin/HEAD`.
func TestEnableFlagsSinGit(t *testing.T) {
	p := newProject(t, `{"checks":[]}`)
	r := fakeRunner{paths: ghOK(), runs: map[string]runResult{"gh auth status": {}}}
	var out strings.Builder
	code := Enable(p, r, Options{Repo: "o/n", Base: "main", Strategy: "squash"}, &out)
	if code != 0 {
		t.Fatalf("Enable = %d, no debía necesitar git:\n%s", code, out.String())
	}
	gh := loadConfig(t, p).Integrations.GitHub
	if gh == nil || gh.Repo != "o/n" || gh.Base != "main" || gh.MergeStrategy != "squash" {
		t.Fatalf("config inesperada: %+v", gh)
	}
}

// TestEnableRepoPrevioSinGit comprueba que un repo previo evita la detección.
func TestEnableRepoPrevioSinGit(t *testing.T) {
	p := newProject(t, `{"integrations":{"github":{"enabled":false,"repo":"o/n"}}}`)
	r := fakeRunner{paths: ghOK(), runs: map[string]runResult{
		"git symbolic-ref --quiet --short refs/remotes/origin/HEAD": {stdout: "origin/main\n"},
		"gh auth status": {},
	}}
	var out strings.Builder
	if code := Enable(p, r, Options{}, &out); code != 0 {
		t.Fatalf("Enable = %d:\n%s", code, out.String())
	}
	gh := loadConfig(t, p).Integrations.GitHub
	if gh.Repo != "o/n" || gh.Base != "main" {
		t.Fatalf("config inesperada: %+v", gh)
	}
}

func TestEnableOpcionInvalidaNoEscribe(t *testing.T) {
	original := `{"checks":[]}`
	cases := []Options{
		{Strategy: "octopus"},
		{RepoSet: true},
		{BaseSet: true},
		{StrategySet: true},
	}
	for _, opts := range cases {
		p := newProject(t, original)
		r := fakeRunner{paths: ghOK(), runs: gitOK()}
		var out strings.Builder
		code := Enable(p, r, opts, &out)
		if code != 2 {
			t.Errorf("Enable(%+v) = %d, esperaba 2", opts, code)
		}
		if !strings.Contains(out.String(), "Uso: rei github enable") {
			t.Errorf("Enable(%+v) debía indicar el uso correcto:\n%s", opts, out.String())
		}
		data, err := os.ReadFile(configPath(p))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != original {
			t.Errorf("Enable(%+v) no debía escribir: %q", opts, data)
		}
	}
}

func TestEnablePreservaClavesAjenas(t *testing.T) {
	p := newProject(t, `{"checks":[],"viewer":{"autoOpenOnReady":true},"extra":{"x":1}}`)
	r := fakeRunner{paths: ghOK(), runs: gitOK()}
	r.runs["gh auth status"] = runResult{}
	var out strings.Builder
	if code := Enable(p, r, Options{}, &out); code != 0 {
		t.Fatalf("Enable = %d:\n%s", code, out.String())
	}
	data, err := os.ReadFile(configPath(p))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"extra"`, `"autoOpenOnReady": true`, `"checks"`, `"integrations"`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("config no conserva %q:\n%s", want, data)
		}
	}
}

func TestEnableBaseFallback(t *testing.T) {
	p := newProject(t, `{"checks":[]}`)
	r := fakeRunner{paths: ghOK(), runs: map[string]runResult{
		"git remote get-url origin":                                 {stdout: "https://github.com/o/n"},
		"git symbolic-ref --quiet --short refs/remotes/origin/HEAD": {err: errors.New("sin HEAD")},
		"gh auth status": {},
	}}
	var out strings.Builder
	if code := Enable(p, r, Options{}, &out); code != 0 {
		t.Fatalf("Enable = %d:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "se usa `main`") {
		t.Fatalf("debía informar del fallback:\n%s", out.String())
	}
	if gh := loadConfig(t, p).Integrations.GitHub; gh.Base != "main" {
		t.Fatalf("base = %q", gh.Base)
	}
}

func TestEnableJSONInvalidoIntacto(t *testing.T) {
	original := `{"checks": [`
	p := newProject(t, original)
	r := fakeRunner{paths: ghOK(), runs: gitOK()}
	var out strings.Builder
	if code := Enable(p, r, Options{}, &out); code != 1 {
		t.Fatalf("Enable = %d, esperaba 1", code)
	}
	data, err := os.ReadFile(configPath(p))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatalf("el JSON inválido no debe modificarse: %q", data)
	}
}

func TestDisableConservaValores(t *testing.T) {
	p := newProject(t, `{
  "checks": [{"id": "V1", "description": "x", "command": ["true"]}],
  "viewer": {"autoOpenOnReady": true},
  "integrations": {"github": {"enabled": true, "repo": "o/n", "base": "develop", "mergeStrategy": "merge"}}
}`)
	var out strings.Builder
	if code := Disable(p, &out); code != 0 {
		t.Fatalf("Disable = %d:\n%s", code, out.String())
	}
	cfg := loadConfig(t, p)
	gh := cfg.Integrations.GitHub
	if gh.Enabled {
		t.Fatal("enabled debía quedar false")
	}
	if gh.Repo != "o/n" || gh.Base != "develop" || gh.MergeStrategy != "merge" {
		t.Fatalf("se perdieron valores: %+v", gh)
	}
	if len(cfg.Checks) != 1 || !cfg.Viewer.AutoOpenOnReady {
		t.Fatalf("se perdieron claves: %+v", cfg)
	}
	if !strings.Contains(out.String(), "links.json") {
		t.Fatalf("debía informar de los enlaces:\n%s", out.String())
	}
}

func TestDisableYaDesactivada(t *testing.T) {
	original := `{"integrations":{"github":{"enabled":false,"repo":"o/n"}}}`
	p := newProject(t, original)
	var out strings.Builder
	if code := Disable(p, &out); code != 0 {
		t.Fatalf("Disable = %d, esperaba 0", code)
	}
	data, err := os.ReadFile(configPath(p))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatalf("no debía reescribir:\n%s", data)
	}
	if !strings.Contains(out.String(), "ya estaba desactivado") {
		t.Fatalf("salida inesperada:\n%s", out.String())
	}
}

func TestDisableSinConfig(t *testing.T) {
	p := newProject(t, "")
	var out strings.Builder
	if code := Disable(p, &out); code != 0 {
		t.Fatalf("Disable = %d, esperaba 0", code)
	}
	if _, err := os.Stat(configPath(p)); !os.IsNotExist(err) {
		t.Fatalf("no debía crear config.json (err=%v)", err)
	}
}

func TestDisableJSONInvalidoIntacto(t *testing.T) {
	original := `no es json`
	p := newProject(t, original)
	var out strings.Builder
	if code := Disable(p, &out); code != 1 {
		t.Fatalf("Disable = %d, esperaba 1", code)
	}
	data, err := os.ReadFile(configPath(p))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatalf("no debía tocar el archivo: %q", data)
	}
}

func TestStatusEstados(t *testing.T) {
	cases := []struct {
		name string
		cfg  string
		want string
	}{
		{"no configurada", `{"checks":[]}`, "no configurada"},
		{"desactivada", `{"integrations":{"github":{"enabled":false,"repo":"o/n"}}}`, "desactivada"},
		{"activa", `{"integrations":{"github":{"enabled":true,"repo":"owner/name"}}}`, "activa"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := newProject(t, tc.cfg)
			r := fakeRunner{paths: ghOK(), runs: gitOK()}
			r.runs["gh auth status"] = runResult{}
			before, err := os.ReadFile(configPath(p))
			if err != nil {
				t.Fatal(err)
			}
			var out strings.Builder
			if code := Status(p, r, &out); code != 0 {
				t.Fatalf("Status = %d:\n%s", code, out.String())
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Fatalf("salida no contiene %q:\n%s", tc.want, out.String())
			}
			after, err := os.ReadFile(configPath(p))
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("Status no debe escribir")
			}
		})
	}
}

func TestStatusRemoto(t *testing.T) {
	cases := []struct {
		name      string
		cfg       string
		wantAviso bool
		wantStep  string
	}{
		{
			name:      "coincide",
			cfg:       `{"integrations":{"github":{"enabled":true,"repo":"owner/name"}}}`,
			wantAviso: false,
			wantStep:  "Configuración completa; sin acciones de configuración pendientes.",
		},
		{
			name:      "no coincide",
			cfg:       `{"integrations":{"github":{"enabled":true,"repo":"otro/repo"}}}`,
			wantAviso: true,
			wantStep:  "Siguiente paso: pon el nombre canónico",
		},
		{
			name:      "no coincide con integración desactivada",
			cfg:       `{"integrations":{"github":{"enabled":false,"repo":"otro/repo"}}}`,
			wantAviso: true,
			wantStep:  "Siguiente paso: `rei github enable`",
		},
		{
			name:      "repo vacío no compara ni avisa",
			cfg:       `{"integrations":{"github":{"enabled":true}}}`,
			wantAviso: false,
			wantStep:  "Configuración completa; sin acciones de configuración pendientes.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := newProject(t, tc.cfg)
			// Sin comandos de red simulados: si Status invocase `gh api` o
			// `gh repo view`, el fakeRunner devolvería «comando no simulado».
			r := fakeRunner{paths: ghOK(), runs: gitOK()}
			r.runs["gh auth status"] = runResult{}
			before, err := os.ReadFile(configPath(p))
			if err != nil {
				t.Fatal(err)
			}
			var out strings.Builder
			if code := Status(p, r, &out); code != 0 {
				t.Fatalf("Status = %d, esperaba 0:\n%s", code, out.String())
			}
			got := out.String()
			if !strings.Contains(got, tc.wantStep) {
				t.Fatalf("salida no contiene %q:\n%s", tc.wantStep, got)
			}
			hasAviso := strings.Contains(got, "El remoto no es el nombre canónico")
			if hasAviso != tc.wantAviso {
				t.Fatalf("aviso = %v, esperaba %v:\n%s", hasAviso, tc.wantAviso, got)
			}
			if tc.wantAviso && !strings.Contains(got, "gh repo view <repo> --json nameWithOwner") {
				t.Fatalf("el aviso debe indicar la resolución canónica:\n%s", got)
			}
			if !tc.wantAviso && strings.Contains(got, "no coincide") {
				t.Fatalf("no debía comparar el remoto:\n%s", got)
			}
			after, err := os.ReadFile(configPath(p))
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("Status no debe escribir config.json")
			}
		})
	}
}

func TestStatusSinOriginYJSONInvalido(t *testing.T) {
	p := newProject(t, `{"integrations":{"github":{"enabled":true,"repo":"o/n"}}}`)
	r := fakeRunner{runs: map[string]runResult{
		"git remote get-url origin": {err: errors.New("sin origin")},
	}}
	var out strings.Builder
	if code := Status(p, r, &out); code != 0 {
		t.Fatalf("Status = %d, esperaba 0 sin origin:\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "no detectado") {
		t.Fatalf("debía informar del remoto no detectado:\n%s", out.String())
	}

	bad := newProject(t, `{`)
	out.Reset()
	if code := Status(bad, r, &out); code != 1 {
		t.Fatalf("Status con JSON inválido = %d, esperaba 1", code)
	}
}
