package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1wi777/my-harness-SDD/internal/adapter"
	"github.com/k1wi777/my-harness-SDD/internal/check"
	"github.com/k1wi777/my-harness-SDD/internal/doctor"
	"github.com/k1wi777/my-harness-SDD/internal/github"
	"github.com/k1wi777/my-harness-SDD/internal/gitx"
	"github.com/k1wi777/my-harness-SDD/internal/initwizard"
	"github.com/k1wi777/my-harness-SDD/internal/links"
	"github.com/k1wi777/my-harness-SDD/internal/meta"
	"github.com/k1wi777/my-harness-SDD/internal/paths"
	"github.com/k1wi777/my-harness-SDD/internal/show"
	"github.com/k1wi777/my-harness-SDD/internal/state"
	"github.com/k1wi777/my-harness-SDD/internal/template"
	"github.com/k1wi777/my-harness-SDD/internal/update"
	"github.com/k1wi777/my-harness-SDD/internal/validate"
	"github.com/k1wi777/my-harness-SDD/internal/view"
	"github.com/k1wi777/my-harness-SDD/internal/workitem"
)

// version se inyecta en build time con:
//
//	-ldflags "-X github.com/k1wi777/my-harness-SDD/internal/cli.version=vX.Y.Z"
//
// Por defecto identifica una compilación local sin versión publicada.
var version = "dev"

// Run despacha un comando y devuelve el código de salida.
func Run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 0
	}
	cmd := args[0]
	rest := args[1:]
	for _, a := range rest {
		if a == "--help" || a == "-h" {
			return printCommandHelp(cmd)
		}
	}
	switch cmd {
	case "version", "--version", "-v":
		fmt.Printf("rei %s\n", version)
		return 0
	case "help", "--help", "-h":
		if len(rest) > 0 {
			return printCommandHelp(rest[0])
		}
		printHelp()
		return 0
	case "check":
		return cmdCheck(rest)
	case "init":
		return cmdInit(rest)
	case "doctor":
		return cmdDoctor(rest)
	case "new":
		return cmdNew(rest)
	case "session":
		return cmdSession(rest)
	case "items":
		return cmdItems(rest)
	case "status":
		return cmdStatus(rest)
	case "commit":
		return cmdCommit(rest)
	case "validate":
		return cmdValidate(rest)
	case "review-diff":
		return cmdReviewDiff(rest)
	case "test":
		return cmdTest(rest)
	case "item":
		return cmdItem(rest)
	case "view":
		return cmdView(rest)
	case "github":
		return cmdGitHub(rest)
	case "update":
		return cmdUpdate(rest)
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %s\n\n", cmd)
		printHelp()
		return 2
	}
}

func project() (*paths.Project, int) {
	p, err := paths.Find()
	if err != nil {
		fmt.Fprintln(os.Stderr, projectMessage(err))
		return nil, 1
	}
	return p, 0
}

// projectMessage traduce un error de resolución de proyecto a un mensaje útil.
func projectMessage(err error) string {
	if errors.Is(err, paths.ErrNotFound) {
		return "REI no está inicializado aquí; ejecuta `rei init` para inicializarlo."
	}
	return err.Error()
}

func cmdCheck(args []string) int {
	quiet := false
	for _, a := range args {
		if a == "--quiet" || a == "-q" {
			quiet = true
		}
	}
	p, code := project()
	if p == nil {
		return code
	}
	return check.Run(p, quiet, os.Stdout)
}

// cmdInit implementa `rei init [status|opencode|claude [--check]|--update
// [--force]]`. En modo instalador (sin argumentos) despliega el esqueleto
// embebido en el proyecto destino —resuelto con initwizard.ResolveRoot, sin
// exigir un .rei/ previo— y reporta la personalización. `status` es un reporte
// de solo lectura que exige un proyecto REI existente (project()). `opencode` y
// `claude` instalan el esqueleto (incluido .rei/adapters/) y generan los
// agentes nativos del runtime; `<runtime> --check` verifica sin escribir.
// `--update [--force]` refresca el esqueleto de un proyecto ya inicializado
// (exige project()) sin pisar la personalización ni el estado. Cualquier otra
// combinación es uso incorrecto (2, R30).
func cmdInit(args []string) int {
	if hasArg(args, "--update") {
		return cmdInitUpdate(args)
	}
	switch {
	case len(args) == 0:
		p, err := initwizard.ResolveRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return initwizard.Init(p, os.Stdout)
	case len(args) == 1 && args[0] == "status":
		p, code := project()
		if p == nil {
			return code
		}
		return initwizard.Status(p, os.Stdout)
	case len(args) == 1 && args[0] == "opencode":
		p, err := initwizard.ResolveRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return adapter.Install(p, os.Stdout)
	case len(args) == 2 && args[0] == "opencode" && args[1] == "--check":
		p, err := initwizard.ResolveRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return adapter.Check(p, os.Stdout)
	case len(args) == 1 && args[0] == "claude":
		p, err := initwizard.ResolveRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return adapter.InstallClaude(p, os.Stdout)
	case len(args) == 2 && args[0] == "claude" && args[1] == "--check":
		p, err := initwizard.ResolveRoot()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return adapter.CheckClaude(p, os.Stdout)
	default:
		fmt.Fprintln(os.Stderr, "uso: rei init [status|opencode|claude [--check]|--update [--force]]")
		return 2
	}
}

// cmdInitUpdate ejecuta `rei init --update [--force]`. Exige un proyecto REI
// (project()); refresca el esqueleto con initwizard.Update y, si el proyecto ya
// tiene los directorios nativos, regenera los adaptadores de OpenCode y/o
// Claude Code. Devuelve 2 ante cualquier argumento distinto de --update/--force.
func cmdInitUpdate(args []string) int {
	force := false
	for _, a := range args {
		switch a {
		case "--update":
		case "--force":
			force = true
		default:
			fmt.Fprintln(os.Stderr, "uso: rei init --update [--force]")
			return 2
		}
	}
	p, code := project()
	if p == nil {
		return code
	}
	rc := initwizard.Update(p, force, os.Stdout)
	if isDir(filepath.Join(p.Root, ".opencode")) {
		if adapter.Install(p, os.Stdout) != 0 {
			rc = 1
		}
	}
	if isDir(filepath.Join(p.Root, ".claude")) {
		if adapter.InstallClaude(p, os.Stdout) != 0 {
			rc = 1
		}
	}
	return rc
}

// hasArg indica si args contiene la bandera exacta flag.
func hasArg(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

// isDir informa si path existe y es un directorio.
func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// cmdTest ejecuta solo los checks declarados en .rei/config.json, sin efectos
// secundarios. NO es la suite de tests de REI (esa es `make test`).
func cmdTest(args []string) int {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "uso: rei test")
		return 2
	}
	p, code := project()
	if p == nil {
		return code
	}
	return check.RunDeclared(p, os.Stdout)
}

// cmdItem muestra la ficha de un Work Item (solo lectura).
func cmdItem(args []string) int {
	if len(args) != 2 || args[0] != "show" {
		fmt.Fprintln(os.Stderr, "uso: rei item show <id>")
		return 2
	}
	p, code := project()
	if p == nil {
		return code
	}
	return show.WorkItem(p, args[1], os.Stdout)
}

// cmdView implementa `rei view [<id>] [--open|--stop|--enable|--disable]`.
// Resuelve el Work Item (id explícito o sesión activa) y delega en view.Run.
// Los modos son mutuamente excluyentes; un modo desconocido o una combinación
// incompatible es uso incorrecto (2, R5).
func cmdView(args []string) int {
	const usage = "uso: rei view [<id>] [--open|--stop|--enable|--disable]"

	id := ""
	mode := view.ModeForeground
	modes := 0
	for _, a := range args {
		switch a {
		case "--open":
			mode, modes = view.ModeOpen, modes+1
		case "--stop":
			mode, modes = view.ModeStop, modes+1
		case "--enable":
			mode, modes = view.ModeEnable, modes+1
		case "--disable":
			mode, modes = view.ModeDisable, modes+1
		case "--serve":
			mode, modes = view.ModeServe, modes+1
		default:
			if strings.HasPrefix(a, "-") {
				fmt.Fprintln(os.Stderr, usage)
				return 2
			}
			if id != "" {
				fmt.Fprintln(os.Stderr, usage)
				return 2
			}
			id = a
		}
	}
	if modes > 1 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}

	p, code := project()
	if p == nil {
		return code
	}

	if mode == view.ModeForeground || mode == view.ModeOpen || mode == view.ModeServe {
		if id == "" {
			s, err := state.ReadSession(p)
			if err != nil || !s.Active() {
				fmt.Fprintln(os.Stderr, "No hay sesión activa; indica un Work Item: rei view <id>")
				return 1
			}
			id = s.WorkItem
		}
	}

	return view.Run(p, view.Options{Mode: mode, ID: id, Stdout: os.Stdout})
}

// cmdGitHub implementa `rei github <enable|disable|status>`. Valida el uso (2)
// antes de resolver el proyecto; sin `.rei/` devuelve 1. Las acciones sobre
// GitHub (issues, ramas, PR) son del agente, nunca de este comando.
func cmdGitHub(args []string) int {
	const usage = "uso: rei github <enable|disable|status> [--repo owner/name] [--base <rama>] [--strategy squash|merge|rebase]"
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	switch args[0] {
	case "enable":
		opts, err := parseGitHubEnable(args[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			fmt.Fprintln(os.Stderr, usage)
			return 2
		}
		p, code := project()
		if p == nil {
			return code
		}
		return github.Enable(p, github.ExecRunner{}, opts, os.Stdout)
	case "disable", "status":
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, usage)
			return 2
		}
		p, code := project()
		if p == nil {
			return code
		}
		if args[0] == "disable" {
			return github.Disable(p, os.Stdout)
		}
		return github.Status(p, github.ExecRunner{}, os.Stdout)
	default:
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
}

// parseGitHubEnable parsea y valida los flags de `rei github enable`. Los flags
// desconocidos, repetidos o sin valor son uso incorrecto. Un valor vacío (por
// ejemplo `$VAR` sin definir) también se rechaza aquí, antes de resolver el
// proyecto, para que no llegue a `Enable` como si el flag no se hubiera pasado.
func parseGitHubEnable(args []string) (github.Options, error) {
	var opts github.Options
	seen := map[string]bool{}
	for i := 0; i < len(args); i++ {
		flag := args[i]
		switch flag {
		case "--repo", "--base", "--strategy":
			if seen[flag] {
				return opts, fmt.Errorf("flag repetido: %s", flag)
			}
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return opts, fmt.Errorf("falta el valor de %s", flag)
			}
			i++
			value := args[i]
			seen[flag] = true
			switch flag {
			case "--repo":
				opts.Repo, opts.RepoSet = value, true
			case "--base":
				opts.Base, opts.BaseSet = value, true
			case "--strategy":
				opts.Strategy, opts.StrategySet = value, true
			}
		default:
			return opts, fmt.Errorf("argumento desconocido: %s", flag)
		}
	}
	if err := github.ValidateOptions(opts); err != nil {
		return opts, err
	}
	return opts, nil
}

// cmdUpdate implementa `rei update [--check]`. Solo acepta `--check` como
// argumento (cualquier otro es uso incorrecto, 2) y delega en update.Run, que
// resuelve la descarga/instalación y los códigos 0/1.
func cmdUpdate(args []string) int {
	check := false
	for _, a := range args {
		switch a {
		case "--check":
			check = true
		default:
			fmt.Fprintln(os.Stderr, "uso: rei update [--check]")
			return 2
		}
	}
	return update.Run(update.Options{
		CurrentVersion: version,
		Check:          check,
		Stdout:         os.Stdout,
	})
}

func cmdDoctor(args []string) int {
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "uso: rei doctor")
		return 2
	}
	p, code := project()
	if p == nil {
		return code
	}
	return doctor.Run(p, os.Stdout)
}

func cmdNew(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: rei new <id> <feature|task> [title]")
		return 2
	}
	id, typ := args[0], args[1]
	title := ""
	if len(args) > 2 {
		title = args[2]
	}
	p, code := project()
	if p == nil {
		return code
	}
	if err := workitem.New(p, id, typ, title); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	fmt.Printf("Work Item creado: .rei/specs/%s/ (type: %s, status: pending)\n", id, typ)
	fmt.Printf("Completa 'description' en .rei/specs/%s/meta.json.\n", id)
	return 0
}

func cmdSession(args []string) int {
	if len(args) == 0 || args[0] == "status" {
		return cmdSessionStatus()
	}
	p, code := project()
	if p == nil {
		return code
	}
	switch args[0] {
	case "start":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "uso: rei session start <id> <feature|task>")
			return 2
		}
		if err := template.StartSession(p, args[1], args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		fmt.Printf("Sesión iniciada: %s (%s) -> pending (agente: spec_author)\n", args[1], args[2])
		return 0
	case "archive":
		id, err := template.ArchiveSession(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		if id == "" {
			fmt.Println("No hay sesión activa que archivar.")
			return 0
		}
		fmt.Printf("Sesión '%s' archivada en history.md. current.md restablecido.\n", id)
		return 0
	case "reset":
		if err := template.ResetCurrent(p); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			return 1
		}
		fmt.Println("current.md restablecido.")
		return 0
	default:
		fmt.Fprintf(os.Stderr, "subcomando desconocido: session %s\n", args[0])
		return 2
	}
}

func cmdSessionStatus() int {
	p, code := project()
	if p == nil {
		return code
	}
	s, err := state.ReadSession(p)
	if err != nil || !s.Active() {
		fmt.Println("Sin sesión activa.")
		return 0
	}
	fmt.Printf("Sesión activa: %s (estado: %s, agente: %s)\n", s.WorkItem, orQ(s.State), orQ(s.Agent))
	return 0
}

func cmdItems(args []string) int {
	if len(args) == 0 || args[0] != "status" {
		fmt.Fprintln(os.Stderr, "uso: rei items status")
		return 2
	}
	p, code := project()
	if p == nil {
		return code
	}
	items, err := state.ListWorkItems(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(items) == 0 {
		fmt.Println("No hay Work Items registrados.")
		return 0
	}
	active, err := state.InProgressItems(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("Work Items:")
	width := 0
	for _, it := range items {
		if len(it.ID) > width {
			width = len(it.ID)
		}
	}
	for _, it := range items {
		marker := ""
		if it.Status == meta.StatusInProgress {
			marker = "   << ACTIVO"
		}
		fmt.Printf("  - %-*s [%s] (%s)%s\n", width, it.ID, it.Status, itemDetails(p, it), marker)
	}
	if len(active) > 1 {
		fmt.Fprintf(os.Stderr, "ERROR: hay %d Work Items en in_progress (%s); solo puede haber uno.\n",
			len(active), strings.Join(active, ", "))
		return 1
	}
	return 0
}

// itemDetails compone el detalle de un Work Item para `rei items status`:
// siempre el tipo y, si links.json es válido, la issue y la base. Un índice
// inválido se indica como nota sin impedir el listado.
func itemDetails(p *paths.Project, it state.ItemStatus) string {
	details := []string{"type: " + it.Type}
	l, err := links.Read(filepath.Join(p.SpecDir(it.ID), "links.json"))
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		details = append(details, "links inválido")
	default:
		if len(links.Problems(l)) > 0 {
			details = append(details, "links inválido")
			break
		}
		if l.Issue > 0 {
			details = append(details, fmt.Sprintf("issue: #%d", l.Issue))
		}
		if l.Base != nil && l.Base.Branch != "" && l.Base.SHA != "" {
			details = append(details, fmt.Sprintf("base: %s@%s", l.Base.Branch, l.Base.SHA))
		}
	}
	return strings.Join(details, ", ")
}

func cmdStatus(args []string) int {
	if len(args) == 0 || args[0] != "set" {
		fmt.Fprintln(os.Stderr, "uso: rei status set <id> <status> [--force]")
		return 2
	}
	force := false
	var positional []string
	for _, a := range args[1:] {
		switch {
		case a == "--force":
			force = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintln(os.Stderr, "uso: rei status set <id> <status> [--force]")
			return 2
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) < 2 {
		fmt.Fprintln(os.Stderr, "uso: rei status set <id> <status> [--force]")
		return 2
	}
	id, status := positional[0], positional[1]
	if !meta.IsValidStatus(status) {
		fmt.Fprintf(os.Stderr, "ERROR: estado inválido: %q\n", status)
		return 1
	}
	p, code := project()
	if p == nil {
		return code
	}
	path := p.MetaFile(id)
	if _, err := os.Stat(path); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: el Work Item %q no existe\n", id)
		return 1
	}
	previous, forced, err := state.Transition(p, id, status, force)
	switch {
	case errors.Is(err, state.ErrAnotherInProgress):
		activos, _ := state.InProgressItems(p)
		otro := "otro Work Item"
		if len(activos) > 0 {
			otro = strings.Join(activos, ", ")
		}
		fmt.Fprintf(os.Stderr, "ERROR: no se puede pasar %s a in_progress: %s ya está en curso; ciérralo (review/done), bloquéalo o revísalo antes de continuar.\n", id, otro)
		return 1
	case errors.Is(err, state.ErrTransitionNotAllowed):
		fmt.Fprintf(os.Stderr, "ERROR: transición no permitida: %s -> %s (ver workflow.md; usa --force para forzar)\n", previous, status)
		return 1
	case err != nil:
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	if previous == status {
		fmt.Printf("Work Item %s: ya está en '%s'.\n", id, status)
		return 0
	}
	if forced {
		fmt.Printf("Work Item %s: %s -> %s (forzado).\n", id, previous, status)
	} else {
		fmt.Printf("Work Item %s: %s -> %s.\n", id, previous, status)
	}
	return 0
}

func cmdCommit(args []string) int {
	if len(args) < 3 || args[0] != "set" {
		fmt.Fprintln(os.Stderr, "uso: rei commit set <id> <base_commit|last_review_commit>")
		return 2
	}
	p, code := project()
	if p == nil {
		return code
	}
	msg, err := gitx.SetCommit(p, args[1], args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	fmt.Println(msg)
	return 0
}

func cmdValidate(args []string) int {
	id := ""
	if len(args) > 0 {
		id = args[0]
	}
	p, code := project()
	if p == nil {
		return code
	}
	if id == "" {
		s, err := state.ReadSession(p)
		if err != nil || !s.Active() {
			fmt.Println("Sin sesión activa.")
			return 0
		}
		id = s.WorkItem
	}
	issues, err := validate.WorkItem(p, id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("== Validando Work Item: %s ==\n", id)
	fails := 0
	for _, is := range issues {
		fmt.Printf("%-8s%s\n", "["+string(is.Level)+"]", is.Message)
		if is.Level == validate.LevelFail {
			fails++
		}
	}
	fmt.Println()
	if fails > 0 {
		fmt.Println("Resultado: FALLO")
		return 1
	}
	fmt.Println("Resultado: OK")
	return 0
}

func cmdReviewDiff(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "uso: rei review-diff <work-item-id> [--full]")
		return 2
	}
	id := args[0]
	full := false
	for _, a := range args[1:] {
		if a == "--full" {
			full = true
		}
	}
	p, code := project()
	if p == nil {
		return code
	}
	out, err := gitx.ReviewDiff(p, id, full)
	switch {
	case errors.Is(err, gitx.ErrNoGit):
		fmt.Println("SIN_GIT")
		return 2
	case errors.Is(err, gitx.ErrNoBase):
		fmt.Println("SIN_BASE")
		return 2
	case err != nil:
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Print(out)
	return 0
}

func orQ(s string) string {
	if s == "" {
		return "?"
	}
	return s
}
