# Revisión — GitHub: dividir el playbook y cargar solo la sección necesaria

## Resultado

**Aprobado (`done`).** `integrations/github.md` es ahora un índice de enrutado de
200 líneas y el detalle operativo vive en cuatro archivos de dominio. La
migración es de solo organización: no cambian comandos, marcadores, labels, tipos
de base, errores ni flujo.

## Verificaciones

- **Índice (§1–§6, 200 líneas).** Alcance, configuración/`rei github`,
  vocabulario (marcadores, labels, `links.json`, canónico, rename/fork), puerta
  solo Features, tabla «Qué leer según la acción» y enlaces. Contiene **0** líneas
  de comando `gh`/`git`, ningún bloque `git switch` y ninguna plantilla; la única
  cerca es el bloque de `rei github enable|disable|status` (configuración, como
  planifica el plan).
- **Migración sin pérdida.** Comparación mecánica contra `f222933`: los **35
  bloques de código** del original aparecen íntegros y sin duplicados nuevos (35
  = 35; 48 líneas `gh`/`git` = 31 en `core.md` + 17 en `backlog.md`); la tabla de
  errores (32 filas), reintentos, recordatorios v1/v2 y prueba de campo son
  idénticas. Los párrafos solo cambian por reajuste de línea, renumeración de
  encabezados y reescritura de referencias `§N` → archivo correcto
  (`core.md §2/§3/§4/§5`, `backlog.md §2`, `templates.md §1–§6`, `errors.md §1/§3`).
- **Términos clave.** `rei:epic`, `rei:backlog`, `rei:work-item`, `Closes`,
  `Refs`, `--add-assignee`, `links.json`, `Base:`, `gh pr merge` y `rei:blocked`
  presentes en el dominio correcto; enlaces relativos resueltos
  (`../github.md`, `../README.md`, plantillas entre archivos).
- **Contrato del Leader.** `.rei/agents/leader.md` incluye la tabla acción →
  archivo exacto, prohíbe cargar el playbook completo o todos los archivos, y
  actualiza Arranque (`github/backlog.md`) y prueba de campo (`github/errors.md
  §3`). Reglas de negocio intactas. `initializer.md` no cita el playbook: sin
  cambios, como preveía el plan. `.opencode/` y `.claude/` regenerados (5 rutas
  cada uno) y `--check` en 0.
- **Registro y documentación.** Los cuatro archivos están en
  `internal/check.RequiredFiles` y en `embed_test.go`; el patrón `all:` ya los
  embebe. Actualizados `.rei/docs/harness/README.md`, `AGENTS.md` §7 (`task.md` +
  integración), `internal/cli/help.go` (conserva `integrations/github.md`),
  `README.md` y `.rei/docs/usage.md` §1.1.
- **Alcance.** Solo los archivos del plan; `skeleton/`, specs/progress
  históricos y `go.mod`/`go.sum` intactos; sin acciones GitHub ejecutadas; V4 no
  aplica (documentación sin código por plataforma).
- **Checkpoints.** V1 `go vet ./...` OK; V2 `go test -count=1 ./...` OK; V3
  `make build` OK. `rei check --quiet` 0; `rei init opencode --check` y
  `rei init claude --check` 0; `rei doctor` 0 (los 5 archivos GitHub `[OK]`);
  `rei init status` 0; `rei validate <id>` OK.

## Observaciones

- El índice conserva la semántica de resolución de base (`links.json.base` →
  `Base:` → fallback) como vocabulario; el procedimiento completo vive solo en
  `core.md §5` y las repeticiones internas de comandos de labels son las mismas
  que ya existían en el original.
- `rei github status` sigue describiendo `gh repo view` en el índice; es
  comportamiento del diagnóstico, no una acción operativa que invada `core.md`.

## Acciones requeridas

Ninguna.
