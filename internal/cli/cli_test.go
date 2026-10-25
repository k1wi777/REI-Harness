package cli

import (
	"errors"
	"testing"

	"github.com/k1wi777/my-harness-SDD/internal/paths"
)

func TestProjectMessage(t *testing.T) {
	const want = "REI no está inicializado aquí; ejecuta `rei init` para inicializarlo."
	if got := projectMessage(paths.ErrNotFound); got != want {
		t.Fatalf("projectMessage(ErrNotFound) = %q, want %q", got, want)
	}
	if got := projectMessage(errors.New("boom")); got != "boom" {
		t.Fatalf("projectMessage(other) = %q, want %q", got, "boom")
	}
}

func TestCmdInitUpdateUsoIncorrecto(t *testing.T) {
	cases := [][]string{
		{"--update", "extra"},
		{"--update", "--bogus"},
		{"opencode", "--update"},
		{"--force", "--update", "extra"},
	}
	for _, args := range cases {
		if code := cmdInit(args); code != 2 {
			t.Errorf("cmdInit(%v) = %d, esperaba 2", args, code)
		}
	}
}

// TestCmdViewUsoIncorrecto cubre modos excluyentes, modos desconocidos y dos
// Work Items (R4, R5). El parseo falla antes de resolver el proyecto.
func TestCmdViewUsoIncorrecto(t *testing.T) {
	cases := [][]string{
		{"--open", "--stop"},
		{"--enable", "--disable"},
		{"--open", "--bogus"},
		{"--wat"},
		{"id1", "id2"},
	}
	for _, args := range cases {
		if code := cmdView(args); code != 2 {
			t.Errorf("cmdView(%v) = %d, esperaba 2", args, code)
		}
	}
}

// TestCmdGitHubUsoIncorrecto comprueba que el uso se valida antes de resolver
// el proyecto (R2): sin subcomando, desconocido o con argumentos sobrantes.
func TestCmdGitHubUsoIncorrecto(t *testing.T) {
	cases := [][]string{
		nil,
		{},
		{"bogus"},
		{"enable", "extra"},
		{"disable", "--force"},
		{"status", "id"},
	}
	for _, args := range cases {
		if code := cmdGitHub(args); code != 2 {
			t.Errorf("cmdGitHub(%v) = %d, esperaba 2", args, code)
		}
	}
}
