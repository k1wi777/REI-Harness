# Implementación — Separar skeleton del proyecto vivo y verificación rápida/lenta

## Resumen

Se separó el material desplegable del material personalizado: `skeleton/` es la
copia neutral que embarca `rei init`, mientras `AGENTS.md`, `.rei/config.json` y
`.rei/docs/project/*` vivos permanecen intactos. El arnés compartido se embebe en
un segundo FS. Se dejó `V1` como checkpoint rápido y `V2`/`V3`/`V4` como lentos.

## Archivos modificados

- `skeleton/AGENTS.md`, `skeleton/.rei/config.json`,
  `skeleton/.rei/docs/project/{README,architecture,conventions,verification}.md`
  — nuevos; documentos desde `HEAD`, config neutral (`checks: []`, sin
  `integrations`, `autoOpenOnReady: true`).
- `embed.go` — dos `embed.FS`: `Skeleton` (`all:skeleton`) y `Harness`
  (`all:.rei/docs/harness`, `all:.rei/docs/usage.md`, `all:.rei/agents`,
  `all:.rei/adapters`, `all:.rei/templates`). Sin lógica.
- `internal/initwizard/embedded.go` — nuevo `walkEmbedded`: `skeleton/` → raíz
  (sin prefijo) y `Harness` → ruta canónica.
- `internal/initwizard/{installer,update}.go` — usan `walkEmbedded`; manifiesto y
  mensajes con rutas destino; se conserva no-sobrescritura y regla marker-aware.
- `embed_test.go` — asserts repartidos entre FS; neutralidad, marcadores, config
  sin checks/integración y exclusión de vivos/estado.
- `internal/initwizard/*_test.go` — `embeddedFile` resuelve `skeleton/<rel>` o
  `Harness`; nuevo test de instalación limpia neutral; no-sobrescritura ampliada
  a docs de proyecto.
- `internal/adapter/{adapter,claude}_test.go` — leen roles/adaptadores de
  `Harness`.
- `.rei/config.json` — solo `V1`; `integrations.github` y `viewer` vivos
  conservados.
- `.rei/docs/project/verification.md` — `V2` a checkpoints lentos; ejemplo solo
  con `V1`; nota `V2/V3/V4`.
- `AGENTS.md` (§4), `README.md`, `.rei/docs/project/architecture.md` — reflejan
  `skeleton/` y los dos `embed.FS`.

## Verificación

- `V1` `go vet ./...` → OK. `./bin/rei check --quiet` → exit 0 (solo V1,
  ~0.2 s).
- `V2` `go test -count=1 ./...` → OK, ~2.8 s (incluye los 4 asserts antes
  fallidos: `TestInitMencionaPersonalize`, `TestInstalacionLimpiaIntegra`,
  `TestUpdateConservaPersonalizacionYModificados`).
- `V3` `make build` → OK; `V4` `make cross` → OK (linux/darwin/windows ×
  amd64/arm64).
- `./bin/rei doctor` → sin fallos; `./bin/rei init status` → personalizada
  (exit 0); `./bin/rei init opencode --check` y `claude --check` → exit 0;
  `./bin/rei validate <id>` → OK.
- Smoke en `t.TempDir()`: `rei init` deja 4 documentos con `REI:PENDIENTE`,
  config sin checks/integración, `init status` exit 1, `rei check`/`doctor`
  exit 0, `/personalize` mencionado y `--update` idempotente.
- `rei init --update` en el repo: hashes de `AGENTS.md`, `.rei/config.json`,
  `.rei/docs/project/*` idénticos antes/después (también en la 2.ª ejecución).
- `go.mod`/`go.sum` sin cambios (sin dependencias nuevas).

## Rework (review `changes_requested`)

- `skeleton/.rei/config.json`: `viewer.autoOpenOnReady` restaurado a `true`,
  conservando el default aprobado en `2026-10-13_15-51__viewer-lifecycle`.
- Planificación alineada con `true`: R4, config neutral/justificación de
  `design.md`, assert de `embed_test` y T1.
- Asserts invertidos en `TestSkeletonConfigNeutral`/`TestInstalacionLimpiaNeutral`;
  la lista negativa de `embed_test` incluye `.rei/docs/project/README.md`.
- Reejecución: `go vet ./...`, `go test ./...`, `make build`, `make cross`,
  `rei check --quiet`, `init {opencode,claude} --check`, `doctor`, `init status`
  y `validate` → todos OK.
- Smoke en proyecto limpio: config
  `{"checks": [], "viewer": {"autoOpenOnReady": true}}`, sin `integrations`, 4
  documentos con `REI:PENDIENTE`, `init status`=1, `check`/`doctor`=0 y
  `/personalize` anunciado.
- `rei init --update` en el repo conserva los hashes de `AGENTS.md`,
  `.rei/config.json` y `.rei/docs/project/*`.

## Observaciones

- `skeleton/` solo contiene archivos desplegables (sin README propio ni
  `.rei/specs`/`.rei/progress`).
- `internal/adapter` no cambió: reutiliza `InstallSkeleton` y hereda el mapeo.
- Cadencia: `rei check` = V1; V2/V3/V4 antes de cerrar/revisar/release, con
  evidencia en `impl.md`/`review.md`.
