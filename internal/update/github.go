package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// errNoReleases señala que la API no tiene ninguna release publicada (HTTP
// 404). No es un fallo: el comando lo trata como "nada que actualizar".
var errNoReleases = errors.New("no hay releases publicadas")

const (
	// defaultAPIBase es la raíz de la API de GitHub.
	defaultAPIBase = "https://api.github.com"
	// releasesPage es la página de releases, usada como fallback informativo.
	releasesPage = "https://github.com/k1wi777/REI-Harness/releases"
	// repoPath es la ruta del repositorio canónico dentro de la API de GitHub.
	repoPath = "k1wi777/REI-Harness"
	// userAgent es exigido por la API de GitHub.
	userAgent = "rei-update"
)

// githubAsset describe un asset de una release.
type githubAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// githubRelease modela lo mínimo necesario de una release de GitHub.
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

// fetchLatestRelease consulta la última release publicada.
func fetchLatestRelease(client *http.Client, apiBase string) (*githubRelease, error) {
	url := strings.TrimRight(apiBase, "/") + "/repos/" + repoPath + "/releases/latest"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errNoReleases
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("respuesta inesperada de GitHub: %s", resp.Status)
	}
	var rel githubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&rel); err != nil {
		return nil, fmt.Errorf("respuesta JSON inválida de GitHub: %w", err)
	}
	if strings.TrimSpace(rel.TagName) == "" {
		return nil, fmt.Errorf("la release no incluye tag_name")
	}
	return &rel, nil
}

// assetName construye el nombre del asset con el esquema de goreleaser.
func assetName(version, goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("rei_%s_%s_%s%s", version, goos, goarch, ext)
}

// findAsset busca un asset por nombre exacto.
func findAsset(rel *githubRelease, name string) (githubAsset, bool) {
	for _, a := range rel.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return githubAsset{}, false
}

// findChecksums localiza el archivo de checksums de la release.
func findChecksums(rel *githubRelease) (githubAsset, bool) {
	return findAsset(rel, "checksums.txt")
}
