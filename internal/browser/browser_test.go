package browser

import (
	"reflect"
	"testing"
)

// TestCommand verifica el comando elegido por GOOS sin lanzar procesos. Es la
// costura que evita abrir navegadores reales durante los tests.
func TestCommand(t *testing.T) {
	cases := []struct {
		goos string
		url  string
		want []string
	}{
		{"windows", "http://127.0.0.1:1/", []string{"cmd", "/c", "start", "", "http://127.0.0.1:1/"}},
		{"darwin", "http://127.0.0.1:1/", []string{"open", "http://127.0.0.1:1/"}},
		{"linux", "http://127.0.0.1:1/", []string{"xdg-open", "http://127.0.0.1:1/"}},
		{"freebsd", "http://127.0.0.1:1/", []string{"xdg-open", "http://127.0.0.1:1/"}},
	}
	for _, c := range cases {
		got := command(c.goos, c.url).Args
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("command(%q) args = %v, want %v", c.goos, got, c.want)
		}
	}
}
