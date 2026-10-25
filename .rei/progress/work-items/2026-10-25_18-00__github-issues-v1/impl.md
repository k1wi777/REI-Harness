# Implementation — Integración opcional con GitHub Issues (v1)

## Resumen

Implementadas T1–T12 según `design.md`, sin contradicciones con
`requirements.md`. La capa Go (config + `internal/github` + CLI) es
determinista y no contacta con la API de GitHub; el resto es playbook y
contratos de agente. La integración sigue desactivada por defecto: no se
escribió `integrations.github` en `.rei/config.json`.

## Archivos

Nuevos:

- `internal/github/github.go` — `Runner`/`ExecRunner`, `ParseRemote`,
  `DetectOrigin`, `DetectBase`, `Enable`, `Disable`, `Status`.
- `internal/github/github_test.go` — tests con `fakeRunner` (sin red ni `gh`).
- `.rei/docs/harness/integrations/github.md` — playbook completo v1.

Modificados:

- `internal/config/config.go` — modelo `GitHub`/`Integrations`,
  `EffectiveBase`/`EffectiveMergeStrategy`, `SetGitHub`,
  `SetGitHubEnabled`, `ErrGitHubNotConfigured`, escritura atómica.
- `internal/config/config_test.go` — presencia/ausencia, preservación,
  idempotencia, defaults y JSON inválido intacto.
- `internal/cli/cli.go` — dispatch `github` y `cmdGitHub` (uso = 2).
- `internal/cli/help.go` — entrada `github` (subcomandos, códigos y playbook).
- `internal/cli/cli_test.go`, `internal/cli/help_test.go`.
- `internal/check/check.go` — playbook en `RequiredFiles`.
- `embed_test.go` — el playbook debe estar embebido.
- `.rei/docs/harness/README.md` — fila del playbook.
- `.rei/agents/leader.md` — sección «Integración con GitHub (opcional)»,
  pasos en Caso C/E, adopción y reglas duras.
- `.rei/agents/initializer.md` — «Configuración de GitHub Issues».
- `.opencode/agents/{leader,initializer}.md`, `.claude/agents/{leader,initializer}.md`
  y `.rei/install-manifest.json` — regenerados con `rei init opencode|claude`.

## Cambios

- `enable`: detecta `origin` (SSH/HTTPS/`ssh://`), conserva `base` y
  `mergeStrategy` previos, usa `main`/`squash` por defecto (avisa del fallback)
  y sondea `gh` sin bloquear; errores con causa y siguiente paso, sin escribir.
- `disable`: solo apaga, conserva el resto y no contacta con GitHub.
- `status`: solo lectura; estado, valores efectivos, remoto, `gh` y siguiente
  paso; sale 0 en estados legibles y 1 solo con JSON ilegible.
- Playbook: puerta única, issue con marcador y `design.md` plegado, rama
  `feat/<n>-<slug>`, `links.json`, comentarios idempotentes, `rei:blocked`,
  PR/merge bajo orden, adopción, tabla de errores y recordatorios.

## Verificación

- `gofmt -l internal cmd` → sin salida.
- `make vet` → OK.
- `make test` → OK (todos los paquetes; nuevos tests de config/github/CLI).
- `make build` → OK.
- `./bin/rei check --quiet` → 0.
- `./bin/rei init opencode --check` → 0; `./bin/rei init claude --check` → 0.
- Manual en proyectos temporales: sin `.rei` → 1; sin config → `status` 0 «no
  configurada»; `origin` SSH y HTTPS → `enable` escribe y preserva
  `checks`/`viewer`; `disable` conserva; JSON inválido → 1 sin sobrescribir;
  argumentos inválidos → 2.

## Observaciones

- Sin dependencias Go nuevas ni cambios en `.rei/config.json` del repositorio.
- `.rei/install-manifest.json` está en `.gitignore`; se regeneró en disco.
- No se detectaron contradicciones entre requisitos y diseño.
