# Implementación — GitHub: dividir el playbook y cargar solo la sección necesaria

## Resumen

`integrations/github.md` (1.170 líneas) pasa a ser un índice de enrutado de 200
líneas y el detalle operativo se migra sin cambios a cuatro archivos de dominio.
El contrato del Leader enruta por acción y prohíbe cargar todo el playbook. El
registro (`RequiredFiles`, embed, docs y ayuda) conoce los archivos nuevos.

## Archivos

- `.rei/docs/harness/integrations/github.md` — índice (200 líneas): cabecera,
  alcance resumido, configuración, vocabulario, puerta solo Features, tabla
  acción → archivo y enlaces.
- `.rei/docs/harness/integrations/github/core.md` — v1: alcance, `links.json`,
  nombre canónico, rename/fork, puerta, adopción v1, comentarios y PR/merge.
- `.rei/docs/harness/integrations/github/backlog.md` — v2: promoción,
  publicación de spec, épicas, hijas, asignación, task list y cierre.
- `.rei/docs/harness/integrations/github/templates.md` — plantillas de issue,
  spec, épica, hija y comentarios impl/review.
- `.rei/docs/harness/integrations/github/errors.md` — tabla de errores,
  reintentos, aislamiento, recordatorios, prueba de campo y cierre tras `done`.
- `.rei/agents/leader.md` — tabla acción → archivo, prohibición de cargar todo,
  Arranque → `backlog.md`, prueba de campo → `errors.md §3`.
- `internal/check/check.go`, `embed_test.go` — los cuatro archivos nuevos.
- `.rei/docs/harness/README.md`, `AGENTS.md` §7, `internal/cli/help.go`,
  `README.md`, `.rei/docs/usage.md` — registro y enrutado.
- `.opencode/agents/leader.md`, `.claude/agents/leader.md` — regenerados con el
  CLI.

## Mapa de migración

| Original | Destino |
|---|---|
| Cabecera; §1 alcance; §2 configuración; §3.1–3.2 marcadores/labels; §3.4 regla | `github.md` |
| §3.3 `links.json`; §3.4 comando; §3.5 rename/fork; §4 y 4.1/4.3; §5.1 y 5.4; §10 procedimiento; §11 | `core.md` |
| §1 v2; §5.2; §5.3; §6.1/6.3; §7; §8; §9 | `backlog.md` |
| §4.2; §5.3 bloque; §6.2; §7.1; §10 plantillas | `templates.md` |
| §12; §13; §13.1; cierre tras `done` | `errors.md` |

## Verificación

- V1 `go vet ./...` — OK.
- V2 `go test ./...` — OK (incluye embed de los cuatro archivos).
- V3 `make build` — OK (`bin/rei`).
- `rei check --quiet` — 0; los 5 archivos de GitHub aparecen [OK].
- `rei init opencode --check` / `rei init claude --check` — 0.
- `rei doctor` — 0; `rei init status` — 0; `rei validate <id>` — OK.
- No pérdida: la tabla de errores y las de marcadores/labels/base son idénticas
  al original; los bloques de código coinciden salvo el desindentado del bloque
  de spec en `templates.md`; recordatorios idénticos.
- Comandos clave presentes en el dominio (`rei:epic`, `rei:backlog`,
  `rei:work-item`, `Closes`, `Refs`, `--add-assignee`, `links.json`, `Base:`,
  `gh pr merge`, `rei:blocked`).
- El índice no contiene bloques `gh api`, `gh issue`, `gh pr`, `git switch` ni
  plantillas (solo reglas y enlaces).

## Observaciones

- V4 (`make cross`) no aplica: cambio de documentación y registro, sin código
  por plataforma.
- No se tocó `skeleton/`, ni specs/progress históricos, ni se ejecutaron acciones
  GitHub reales.
- `initializer.md` se auditó: no cita el playbook, no se editó.
- El manifiesto local (`.rei/install-manifest.json`, ignorado por git) quedó
  actualizado al regenerar los adaptadores.
