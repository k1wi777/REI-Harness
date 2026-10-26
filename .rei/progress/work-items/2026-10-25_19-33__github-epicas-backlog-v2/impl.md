# Implementación — GitHub Issues v2: épicas, backlog y asignación

## Resumen

Ejecutadas T1–T12 en `feat/2-github-epicas-backlog-v2` según `design.md`. No hay
código Go nuevo: v2 vive en el playbook, el contrato del Leader, el índice y los
agentes nativos regenerados. La prueba de campo se ejecutó con la autorización
registrada en `current.md` y dejó las issues de prueba cerradas.

## Archivos modificados

- `.rei/docs/harness/integrations/github.md` — playbook v2 completo (13
  secciones), conservando el contenido v1.
- `.rei/agents/leader.md` — «Integración con GitHub (opcional)» con disparadores,
  conflictos, asignación, cierre y prueba de campo.
- `.rei/docs/harness/README.md` — fila del playbook a v2.
- `.opencode/agents/leader.md`, `.claude/agents/leader.md` — regenerados.
- `.rei/install-manifest.json` — regenerado (ignorado por git).
- `.rei/specs/<id>/tasks.md` — T1–T12 marcadas.
- `.rei/progress/current.md` y este `impl.md`.

## Cambios

- **Playbook:** alcance v1+v2; marcadores `epic`/`backlog`/`children`/`spec`;
  tabla de labels con colores; `links.json` con objeto `epic` o `null`; nombre
  canónico y rename/fork; épicas (plantilla, entrevista, borrador, aprobación,
  idempotencia REST canónica, creación parcial, épica cerrada); hijas de backlog;
  adopción/promoción y publicación idempotente de la spec; asignación con
  `gh issue edit`; cierre bajo orden; errores y recordatorios v2.
- **Leader:** reglas no negociables de backlog, adopción, asignación, cierre,
  conflictos y prueba de campo; remite al playbook sin añadir herramientas.
- Sin cambios en `internal/**`, `cmd/**`, `go.mod`/`go.sum`, tests, `RequiredFiles`,
  `embed_test.go`, `config.json` ni contratos de otros roles.

## Verificación

`gofmt -l internal cmd` sin salida; `make vet`, `make test`, `make build`,
`./bin/rei check --quiet`, `./bin/rei init opencode --check`,
`./bin/rei init claude --check` y `./bin/rei validate <id>` en verde.

## Prueba de campo (T12)

Repositorio `k1wi777/REI-Harness`; `canonical=k1wi777/REI-Harness`; `me=k1wi777`.

- **Creación:** épica #3 <https://github.com/k1wi777/REI-Harness/issues/3>
  (`rei:epic:prueba-v2`), hijas #4
  <https://github.com/k1wi777/REI-Harness/issues/4> y #5
  <https://github.com/k1wi777/REI-Harness/issues/5> (`rei:backlog:prueba-v2/hija-a|b`).
  Labels `rei`+`rei:epic` / `rei`+`rei:backlog`; task list con `#4`/`#5`.
- **Idempotencia:** nueva búsqueda REST por marcador devolvió exactamente 1
  resultado por issue (#3, #4, #5); sin duplicados.
- **Asignación:** `--add-assignee k1wi777` y `--remove-assignee` en #4 verificados
  (`["k1wi777"]` → `[]`); login `rei-v2-no-existe-xyz` → `users/...` 404 y
  `gh issue edit` falló con «not found» sin cambiar assignees; lote
  `gh issue edit 4 5 --add-assignee k1wi777` verificado en ambas y retirado.
- **Cierre con hija abierta:** #4/#5 en `OPEN` ⇒ rechazado; no se ejecutó
  `gh issue close` y la épica siguió `OPEN`.
- **Adopción en worktree:** `git worktree add /tmp/rei-v2-field main`;
  `rei new 2026-10-25_20-10__prueba-v2-hija-a feature`; marcador de #4 sustituido
  por `rei:work-item:2026-10-25_20-10__prueba-v2-hija-a` (0 marcadores de
  backlog en forma de comentario); labels `rei`+`rei:feature` sin `rei:backlog`;
  `links.json` con `epic {slug: prueba-v2, issue: 3, url}`; estado `pending`.
- **Spec idempotente:** bloque `rei:spec:<id>` publicado dos veces; tras la
  segunda, 1 delimitador de apertura y 1 de cierre, contenido actualizado en sitio
  y el antiguo ausente.
- **Cierre bajo orden:** #4 y #5 cerradas y verificadas `CLOSED`; task list de #3
  reconciliada a `[x]`; #3 cerrada y verificada `CLOSED`.
- **Épica cerrada:** intento de añadir `hija-c` rechazado por estado `CLOSED`;
  0 issues creadas con ese marcador.
- Worktree eliminado con `git worktree remove --force`; sin commits. Issue #2
  intacta (`OPEN`, `rei`+`rei:feature`, marcador del Work Item).

## Observaciones

- La REST devuelve `state` en minúsculas (`open`); el playbook normaliza con
  `ascii_upcase` para comparar `OPEN`/`CLOSED` (ajuste de precisión sobre el
  comando de `design.md` §7.2, sin cambio de comportamiento).
- `.rei/install-manifest.json` está en `.gitignore`; se regeneró pero no aparece
  en `git status`.
- No se crearon PRs, merges ni commits; `main` no se tocó.
