package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	harnesscheck "github.com/k1wi777/REI-Harness/internal/check"
	"github.com/k1wi777/REI-Harness/internal/initwizard"
	"github.com/k1wi777/REI-Harness/internal/paths"
)

// runtime describe cómo adaptar el formato canónico de rol a un runtime nativo.
// Es el punto de extensión: para añadir un runtime solo hay que declarar su
// descriptor y registrar su plantilla y su mapa de herramientas en
// .rei/adapters/<name>/.
type runtime struct {
	// name es el identificador del runtime (nombre del directorio del
	// adaptador y del subcomando de `rei init`).
	name string
	// hostedDir es el directorio destino de los agentes nativos.
	hostedDir string
	// mark es la marca GENERATED que autoriza a sobrescribir un archivo.
	mark string
	// nativeKey es la clave del mapa de herramientas que contiene los destinos
	// nativos de ese runtime (p. ej. "permission" o "tools").
	nativeKey string
	// validTargets son los destinos nativos admitidos por el runtime.
	validTargets map[string]bool
	// buildData construye los datos que consume agent.tmpl a partir del rol y de
	// los destinos nativos ya resueltos y deduplicados.
	buildData func(role Role, targets []string) any
	// artifacts son archivos adicionales generados por el runtime (p. ej. el
	// comando `/personalize` o `CLAUDE.md`), además de los agentes de los roles.
	artifacts []artFile
}

// artFile describe un archivo generado por el adaptador que no depende de un
// rol: el comando `/personalize` o `CLAUDE.md`. rel es la ruta destino relativa
// a la raíz y build devuelve su contenido canónico.
type artFile struct {
	rel   string
	build func(p *paths.Project, rt runtime) (string, error)
}

// fileData son los datos mínimos que consumen las plantillas de archivos
// generados que solo necesitan la marca GENERATED.
type fileData struct {
	Mark string
}

// commandArtifact construye el artefacto del comando `/personalize`, que
// renderiza <tmpl> del adaptador.
func commandArtifact(rel, tmpl string) artFile {
	return artFile{
		rel: rel,
		build: func(p *paths.Project, rt runtime) (string, error) {
			return renderTemplate(p, rt, tmpl, fileData{Mark: rt.mark})
		},
	}
}

// toolEntry es una entrada de tools.json: una clave nativa por runtime
// (p. ej. "permission" o "tools") con la lista de destinos nativos.
type toolEntry map[string][]string

// toolMap es el contenido de tools.json: herramienta genérica -> entrada.
type toolMap map[string]toolEntry

// sortedKeys devuelve las claves de un toolMap ordenadas.
func sortedKeys(tm toolMap) []string {
	keys := make([]string, 0, len(tm))
	for key := range tm {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// readRole lee y parsea .rei/agents/<rol>.md.
func readRole(p *paths.Project, name string) (Role, error) {
	data, err := os.ReadFile(filepath.Join(p.ReiDir(), "agents", name+".md"))
	if err != nil {
		return Role{}, fmt.Errorf("no se pudo leer el rol %s: %w", name, err)
	}
	return ParseRole(data)
}

// adapterDir devuelve el directorio del adaptador de un runtime dentro del
// proyecto destino.
func adapterDir(p *paths.Project, rt runtime) string {
	return filepath.Join(p.ReiDir(), "adapters", rt.name)
}

// agentPath devuelve la ruta destino del agente nativo de un rol.
func agentPath(p *paths.Project, rt runtime, name string) string {
	return filepath.Join(p.Root, filepath.FromSlash(rt.hostedDir), name+".md")
}

// loadToolMap lee y valida .rei/adapters/<rt.name>/tools.json.
func loadToolMap(p *paths.Project, rt runtime) (toolMap, error) {
	data, err := os.ReadFile(filepath.Join(adapterDir(p, rt), "tools.json"))
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer tools.json del adaptador: %w", err)
	}
	var tm toolMap
	if err := json.Unmarshal(data, &tm); err != nil {
		return nil, fmt.Errorf("tools.json inválido: %w", err)
	}
	if err := validateToolMap(tm, rt); err != nil {
		return nil, err
	}
	return tm, nil
}

// validateToolMap exige que el mapa cubra todas las herramientas genéricas y
// que cada entrada declare al menos un destino nativo válido para el runtime.
func validateToolMap(tm toolMap, rt runtime) error {
	generics := make([]string, 0, len(genericTools))
	for tool := range genericTools {
		generics = append(generics, tool)
	}
	sort.Strings(generics)
	for _, tool := range generics {
		if _, ok := tm[tool]; !ok {
			return fmt.Errorf("tools.json no cubre la herramienta genérica %q", tool)
		}
	}
	for _, tool := range sortedKeys(tm) {
		targets := tm[tool][rt.nativeKey]
		if len(targets) == 0 {
			return fmt.Errorf("tools.json: %q no declara ninguna clave de %s", tool, rt.nativeKey)
		}
		for _, target := range targets {
			if !rt.validTargets[target] {
				return fmt.Errorf("tools.json: destino nativo desconocido %q para %q", target, tool)
			}
		}
	}
	return nil
}

// resolveTargets traduce las herramientas genéricas del rol a destinos nativos
// del runtime, en el orden de declaración de `tools` y deduplicando destinos
// (p. ej. write y edit colapsan en Write, Edit).
func resolveTargets(tools []string, tm toolMap, rt runtime) ([]string, error) {
	var out []string
	seen := map[string]bool{}
	for _, tool := range tools {
		entry, ok := tm[tool]
		if !ok {
			return nil, fmt.Errorf("herramienta genérica sin mapeo en tools.json: %q", tool)
		}
		for _, target := range entry[rt.nativeKey] {
			if seen[target] {
				continue
			}
			seen[target] = true
			out = append(out, target)
		}
	}
	return out, nil
}

// buildAgent construye el contenido nativo canónico de un rol.
func buildAgent(p *paths.Project, rt runtime, name string) (string, error) {
	role, err := readRole(p, name)
	if err != nil {
		return "", err
	}
	return renderAgent(p, rt, role)
}

// renderAgent construye el archivo nativo de un rol: frontmatter mapeado, marca
// GENERATED y el ## Contrato como cuerpo.
func renderAgent(p *paths.Project, rt runtime, role Role) (string, error) {
	tm, err := loadToolMap(p, rt)
	if err != nil {
		return "", err
	}
	targets, err := resolveTargets(role.Tools, tm, rt)
	if err != nil {
		return "", err
	}
	return renderTemplate(p, rt, "agent.tmpl", rt.buildData(role, targets))
}

// renderTemplate lee y aplica una plantilla del adaptador con los datos dados.
func renderTemplate(p *paths.Project, rt runtime, name string, data any) (string, error) {
	raw, err := os.ReadFile(filepath.Join(adapterDir(p, rt), name))
	if err != nil {
		return "", fmt.Errorf("no se pudo leer %s del adaptador: %w", name, err)
	}
	tmpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("%s inválido: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("no se pudo aplicar %s: %w", name, err)
	}
	return buf.String(), nil
}

// generateAll genera <rt.hostedDir>/<rol>.md para cada rol canónico y los
// artefactos del runtime (comando `/personalize` y `CLAUDE.md`). No sobrescribe
// archivos que no lleven la marca GENERATED: los reporta con [WARN] y los
// conserva. Devuelve 0 si completa (aunque omita algún archivo) y 1 si no pudo
// escribir un archivo generado.
func generateAll(p *paths.Project, rt runtime, out io.Writer) int {
	exit := 0
	for _, name := range Roles {
		rel := rt.hostedDir + "/" + name + ".md"
		content, err := buildAgent(p, rt, name)
		if err != nil {
			fmt.Fprintf(out, "[FAIL]  %s: %v\n", rel, err)
			exit = 1
			continue
		}
		if code := writeGenerated(agentPath(p, rt, name), rel, content, rt.mark, out); code != 0 {
			exit = 1
		}
	}
	for _, art := range rt.artifacts {
		content, err := art.build(p, rt)
		if err != nil {
			fmt.Fprintf(out, "[FAIL]  %s: %v\n", art.rel, err)
			exit = 1
			continue
		}
		dest := filepath.Join(p.Root, filepath.FromSlash(art.rel))
		if code := writeGenerated(dest, art.rel, content, rt.mark, out); code != 0 {
			exit = 1
		}
	}
	return exit
}

// writeGenerated escribe el contenido generado en dest aplicando las reglas de
// la marca GENERATED: un archivo sin la marca se conserva con [WARN]; uno igual
// se reporta [OK]; uno generado que difiere se reescribe con [UPD]. Devuelve 0
// si no hubo que reportar un fallo de escritura.
func writeGenerated(dest, rel, content, mark string, out io.Writer) int {
	existing, err := os.ReadFile(dest)
	switch {
	case err == nil:
		if !strings.Contains(string(existing), mark) {
			fmt.Fprintf(out, "[WARN]  %s no fue generado por rei; se conserva sin cambios. Elimínalo o regenera el archivo a mano.\n", rel)
			return 0
		}
		if string(existing) == content {
			fmt.Fprintf(out, "[OK]    %s\n", rel)
			return 0
		}
		if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
			fmt.Fprintf(out, "[FAIL]  No se pudo escribir %s: %v\n", rel, err)
			return 1
		}
		fmt.Fprintf(out, "[UPD]   %s\n", rel)
		return 0
	case os.IsNotExist(err):
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			fmt.Fprintf(out, "[FAIL]  No se pudo crear el directorio de %s: %v\n", rel, err)
			return 1
		}
		if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
			fmt.Fprintf(out, "[FAIL]  No se pudo escribir %s: %v\n", rel, err)
			return 1
		}
		fmt.Fprintf(out, "[OK]    %s\n", rel)
		return 0
	default:
		fmt.Fprintf(out, "[FAIL]  No se pudo comprobar %s: %v\n", rel, err)
		return 1
	}
}

// check comprueba, sin escribir, que cada archivo generado (agentes y
// artefactos) existe y coincide byte a byte con la generación canónica. Devuelve
// 0 si todo coincide y 1 si falta, no lleva la marca o difiere algún archivo.
func check(p *paths.Project, rt runtime, out io.Writer) int {
	exit := 0
	for _, name := range Roles {
		rel := rt.hostedDir + "/" + name + ".md"
		content, err := buildAgent(p, rt, name)
		if err != nil {
			fmt.Fprintf(out, "[FAIL]  %s: %v\n", rel, err)
			exit = 1
			continue
		}
		if code := checkGenerated(agentPath(p, rt, name), rel, content, rt.mark, out); code != 0 {
			exit = 1
		}
	}
	for _, art := range rt.artifacts {
		content, err := art.build(p, rt)
		if err != nil {
			fmt.Fprintf(out, "[FAIL]  %s: %v\n", art.rel, err)
			exit = 1
			continue
		}
		dest := filepath.Join(p.Root, filepath.FromSlash(art.rel))
		if code := checkGenerated(dest, art.rel, content, rt.mark, out); code != 0 {
			exit = 1
		}
	}
	return exit
}

// checkGenerated verifica, sin escribir, que dest existe, lleva la marca y
// coincide con el contenido canónico. Devuelve 0 si coincide y 1 en caso
// contrario.
func checkGenerated(dest, rel, content, mark string, out io.Writer) int {
	existing, err := os.ReadFile(dest)
	switch {
	case os.IsNotExist(err):
		fmt.Fprintf(out, "[FAIL]  Falta %s\n", rel)
		return 1
	case err != nil:
		fmt.Fprintf(out, "[FAIL]  No se pudo leer %s: %v\n", rel, err)
		return 1
	}
	if !strings.Contains(string(existing), mark) {
		fmt.Fprintf(out, "[FAIL]  %s existe pero no fue generado por rei (sin la marca GENERATED)\n", rel)
		return 1
	}
	if string(existing) != content {
		fmt.Fprintf(out, "[FAIL]  %s difiere de la generación canónica\n", rel)
		return 1
	}
	fmt.Fprintf(out, "[OK]    %s\n", rel)
	return 0
}

// install despliega el esqueleto (incluido .rei/adapters/**), asegura la
// estructura de estado y genera los agentes nativos. Reutiliza
// initwizard.InstallSkeleton y check.EnsureStructure para no duplicar la
// instalación. Devuelve 1 si alguna fase falla.
func install(p *paths.Project, rt runtime, out io.Writer) int {
	fmt.Fprintf(out, "== rei init %s ==\n", rt.name)
	fmt.Fprintln(out)

	installed := initwizard.InstallSkeleton(p, out)
	fmt.Fprintln(out)

	scaffold := harnesscheck.EnsureStructure(p, out)
	fmt.Fprintln(out)

	generated := generateAll(p, rt, out)
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Para personalizar el proyecto ejecuta el comando `/personalize` (runtime %s)\n", rt.name)
	fmt.Fprintf(out, "o invoca el rol `initializer` (%s).\n", initwizard.InitializerPath)

	if installed != 0 || scaffold != 0 || generated != 0 {
		return 1
	}
	return 0
}
