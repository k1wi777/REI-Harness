# Revisión — Separar skeleton del proyecto vivo y verificación rápida/lenta

## Resultado

**Aprobado (`done`).** Las cuatro acciones requeridas en la revisión anterior
están aplicadas y verificadas; no se detectan regresiones.

## Verificaciones

- **Correcciones de la revisión previa.**
  - `skeleton/.rei/config.json` vuelve a `viewer.autoOpenOnReady: true`,
    conservando el default aprobado en `2026-10-13_15-51__viewer-lifecycle`.
  - R4, `design.md` («Contenido neutral del esqueleto») y T1 alineados con
    `true`.
  - Asserts invertidos en `TestSkeletonConfigNeutral` y
    `TestInstalacionLimpiaNeutral`; ambos esperan `true`.
  - `.rei/docs/project/README.md` incluido en la lista negativa de
    `TestEmbedExcluyeArchivosVivosDelProyecto`.
- **Esqueleto neutral.** `skeleton/` replica `HEAD` sin cambios (`AGENTS.md` 2
  marcadores; architecture/conventions/verification 1; README 0). Config con
  `checks: []`, sin `integrations` y visor activo; sin `.rei/specs` ni
  `.rei/progress`. Docs vivos conservados (0 marcadores).
- **Embed.** `embed.go` declara `Skeleton` (`all:skeleton`) y `Harness`
  (harness, `usage.md`, agents, adapters, templates) sin lógica; ningún FS
  embebe `AGENTS.md`, `.rei/config.json` ni `.rei/docs/project/*` vivos.
- **initwizard.** `walkEmbedded` mapea `skeleton/`→raíz y `Harness`→ruta
  canónica; `installer`/`update` usan rutas destino, conservan no-sobrescritura
  y regla marker-aware; manifiesto con claves destino. `--update` idempotente.
- **Tests (V2).** `go test -count=1 ./...` OK; en verde
  `TestInitMencionaPersonalize`, `TestInstalacionLimpiaIntegra` y
  `TestUpdateConservaPersonalizacionYModificados`, más neutralidad, instalación
  limpia y preservación.
- **Verificación.** `go vet ./...` (V1) OK; `.rei/config.json` vivo solo V1;
  `./bin/rei check` ejecuta únicamente V1 (0); `verification.md` clasifica
  V2/V3/V4 como lentos. `make build` (V3) y `make cross` (V4) OK.
- **CLI.** `rei doctor`, `rei init status` (personalizada, 0),
  `init opencode --check`, `init claude --check` y `rei validate` OK.
- **Smoke en limpio.** `rei init` deja `{"checks": [], "viewer":
  {"autoOpenOnReady": true}}` sin `integrations`, 4 documentos con
  `REI:PENDIENTE`, `/personalize` anunciado, `init status`=1, `check`/`doctor`=0
  y `--update` idempotente.
- **Repo vivo.** `rei init --update` conserva los hashes de `AGENTS.md`,
  `.rei/config.json` y `.rei/docs/project/*`, sin `UPD/NEW`. `.rei/specs`/
  `.rei/progress` intactos; `go.mod`/`go.sum` sin cambios (sin dependencias).

## Observaciones

- La cadencia V1 (rápido) / V2–V4 (lentos) queda documentada y coherente con
  `rei check`.
- El default del visor vuelve a ser coherente con `initializer.md` (recomienda
  activo) y `leader.md` §7 (abre el visor si es `true`).

## Acciones requeridas

Ninguna.
