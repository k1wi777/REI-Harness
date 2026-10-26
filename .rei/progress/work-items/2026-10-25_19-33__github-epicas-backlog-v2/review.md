# Revisión — GitHub Issues v2: épicas, backlog y asignación

- **Work Item:** `2026-10-25_19-33__github-epicas-backlog-v2` (feature)
- **Resultado:** `done` (aprobado)
- **Fecha:** 2026-10-25

## Verificaciones

- **Paquete de revisión:** `rei review-diff --full` (exit 0, base `a7e00b3`).
  Solo cambian el playbook, el contrato del Leader, el índice, los agentes
  nativos y `current.md`; los nuevos de `.rei/specs/` y `.rei/progress/` se
  leyeron directamente.
- **Trazabilidad:** T1–T12 completadas; el diff cubre R1–R64. Cada requisito
  tiene su sección en el playbook o en el contrato del Leader.
- **Playbook v2 (R1–R57, R59):** alcance v1+v2 y regla de integración
  desactivada; marcadores `epic`/`backlog`/`children`/`spec`; tabla de labels;
  `links.json` con `epic` objeto o `null`; nombre canónico, rename y fork;
  épicas y backlog con entrevista, borrador, aprobación explícita e
  idempotencia REST paginada sin PRs; creación parcial y reintento; hijos de
  épica abierta y rechazo con épica cerrada; adopción, conflicto por marcador,
  promoción con sustitución de marcador/labels/`links.json`; publicación
  idempotente de la spec; asignación/desasignación verificada y sin
  automatismos; cierre solo bajo orden con estado real de las hijas; tabla de
  errores, aislamiento y recordatorios v1+v2.
- **Contrato del Leader (R58):** disparadores de backlog, adopción, asignación
  y cierre; conflictos; prohibición de autocierre/autoasignación; autorización
  de la prueba de campo; remite al playbook sin añadir herramientas. Las reglas
  v1 se conservan.
- **Sin código nuevo (R3, R61):** `git diff a7e00b3` no toca `internal/**`,
  `cmd/**`, `go.mod`/`go.sum` ni `.rei/config.json`; `links.json` de este Work
  Item mantiene `epic: null`.
- **Agentes nativos (R60):** `.opencode/agents/leader.md` y
  `.claude/agents/leader.md` regenerados; `init opencode --check` y
  `init claude --check` en verde.
- **Prueba de campo (R62–R64), comprobada en vivo con `gh`:**
  - #3 cerrada, `rei:epic:prueba-v2`, labels `rei`+`rei:epic`, task list
    `[x] #4/#5`.
  - #4 cerrada, marcador `rei:work-item:2026-10-25_20-10__prueba-v2-hija-a`,
    labels `rei`+`rei:feature`, un único bloque `rei:spec`; #5 cerrada con
    marcador/labels de backlog. Sin marcadores de backlog duplicados.
  - Búsqueda REST por marcador: exactamente 1 resultado por issue, sin
    duplicados; sin Work Item de prueba residual ni worktree.
  - #2 intacta: `OPEN`, `rei`+`rei:feature`, marcador de este Work Item, sin PR.
  - Sin PRs, merges ni commits; `main` sigue en `a7e00b3`.
- **Verificación local:** `gofmt -l internal cmd` sin salida, `go vet ./...`,
  `go test ./...`, build y las comprobaciones del CLI en verde (Go 1.27 en
  contenedor, montaje de solo lectura); `rei check --quiet` exit 0;
  `rei validate` OK. No hay checkpoints declarados en `.rei/config.json`.

## Observaciones

- La normalización `ascii_upcase` del estado REST en el cierre está
  documentada en `impl.md`: corrige la comparación `open`/`OPEN` de `design.md`
  §7.2 sin cambiar el comportamiento exigido (R49–R52).
- La prueba de campo deja #3/#4/#5 cerradas como evidencia permanente, según
  `design.md` §10.2.

## Acciones requeridas

Ninguna.
