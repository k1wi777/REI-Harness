# Revisión — Visor: vivo tras aprobar y recomendar auto-open

## Resultado

**Aprobado (`done`).** La implementación cumple el plan aprobado sin desviaciones.

## Verificaciones

- **Paquete de revisión** (`rei review-diff 2026-10-13_15-51__viewer-lifecycle`,
  exit 0, base `915eab6`): solo los 5 archivos del plan, más `meta.json`/`plan.md`
  del propio Work Item. Sin cambios colaterales en `render.go`, `assets/`,
  `leader.md` ni `config.go`.
- **Cierre solo en `done`**: `watchStatus` (`server.go:147`) y la rama SSE de
  `handleEvents` (`server.go:313-314`) usan `m.Status == meta.StatusDone`; ambas
  exigen `background` (SSE) y arrancan/actúan solo con `closeOnDone`. `--stop`
  (`stopViewer`) y el idle timeout (`monitor`) intactos.
- **Renombrado**: `closeOnStatusChange` → `closeOnDone` en declaración, `Serve`,
  `handleEvents` y `runServe`; sin referencias residuales en código (solo
  menciones históricas en `plan.md`/`impl.md`/`current.md`). Comentarios de
  `Serve`, `watchStatus`, `handleEvents` y `runServe` actualizados.
- **Test del ciclo de vida**: `TestFondoSeCierraEnDone` (`server_test.go:214`)
  con `background`, `closeOnDone` y `pollInterval = 10ms`; verifica que
  `ready→in_progress` y `review` NO cierran y que `done` SÍ cierra y `Serve`
  retorna sin error. Con la condición antigua el test falla (cobertura efectiva).
- **Default y recomendación**: `.rei/config.json` → `viewer.autoOpenOnReady:
  true` (es el esqueleto embebido por `embed.go`, por lo que `rei init` despliega
  `true`). `initializer.md` § «Configuración del visor» conserva la pregunta,
  recomienda mantenerla activa, usa `true` si no hay preferencia y respeta un
  `false` explícito, con el mismo alcance de escritura.
- **Checkpoints rápidos**: `verification.md` sigue siendo plantilla y
  `.rei/config.json` declara `checks: []`, así que no existen IDs `V*`
  ejecutables; `rei check --quiet` termina con exit 0. No hay checkpoints lentos
  definidos.
- **Verificación ejecutada** (Go 1.27.3 temporal en `/tmp/goroot`, sin cambios en
  el repo): `gofmt -l .` sin salida, `go vet ./...` OK, `go test ./...` OK,
  `go test ./internal/view/ -count=1 -run TestFondoSeCierraEnDone` PASS,
  `make build` OK.
- **v1/Fase 2**: selector de documento, aprobación en dos pasos, TOC, diff,
  enlaces profundos, scroll y reconexión SSE no se tocan; sus tests siguen
  pasando.

## Observaciones

- Tras `done` la página pasa a «Reconectando…»: comportamiento esperado y
  documentado en el plan, no un defecto.
- El fallback compilado de `config.Load` sigue en `false` cuando falta la clave,
  según la decisión explícita del plan (coherente con `leader.md`, no tocado).
- Deriva **preexistente** (fuera de alcance): `.opencode/agents/initializer.md` y
  `.claude/agents/initializer.md` no incluyen la sección «Configuración del
  visor» (generados antes de que existiera). No es una desviación de este plan;
  se regeneran con `rei init opencode` / `rei init claude`.
- Evidencia del Implementer (`impl.md`) coincide con lo ejecutado en esta
  revisión.

## Acciones requeridas

Ninguna.
