package update

import "testing"

func TestAssetName(t *testing.T) {
	cases := []struct {
		version, goos, goarch, want string
	}{
		{"0.2.0", "linux", "amd64", "rei_0.2.0_linux_amd64.tar.gz"},
		{"0.2.0", "darwin", "arm64", "rei_0.2.0_darwin_arm64.tar.gz"},
		{"0.2.0", "windows", "amd64", "rei_0.2.0_windows_amd64.zip"},
	}
	for _, c := range cases {
		if got := assetName(c.version, c.goos, c.goarch); got != c.want {
			t.Errorf("assetName(%q,%q,%q) = %q, want %q", c.version, c.goos, c.goarch, got, c.want)
		}
	}
}

func TestFindAssetYChecksums(t *testing.T) {
	rel := &githubRelease{
		TagName: "v0.2.0",
		Assets: []githubAsset{
			{Name: "rei_0.2.0_linux_amd64.tar.gz", URL: "u1"},
			{Name: "checksums.txt", URL: "u2"},
		},
	}
	if a, ok := findAsset(rel, "rei_0.2.0_linux_amd64.tar.gz"); !ok || a.URL != "u1" {
		t.Fatalf("findAsset no encontró el asset esperado: %+v", a)
	}
	if _, ok := findAsset(rel, "no-existe"); ok {
		t.Fatal("findAsset no debe encontrar un asset inexistente")
	}
	if c, ok := findChecksums(rel); !ok || c.URL != "u2" {
		t.Fatalf("findChecksums no encontró checksums.txt: %+v", c)
	}
}

func TestFindChecksumsAusente(t *testing.T) {
	rel := &githubRelease{TagName: "v0.2.0", Assets: []githubAsset{{Name: "otro"}}}
	if _, ok := findChecksums(rel); ok {
		t.Fatal("findChecksums no debe encontrar nada")
	}
}

// TestRepositorioCanonico fija el repositorio de auto-actualización y detecta
// una regresión al nombre anterior (R13).
func TestRepositorioCanonico(t *testing.T) {
	if repoPath != "k1wi777/REI-Harness" {
		t.Fatalf("repoPath = %q, want k1wi777/REI-Harness", repoPath)
	}
	if releasesPage != "https://github.com/k1wi777/REI-Harness/releases" {
		t.Fatalf("releasesPage = %q", releasesPage)
	}
}
