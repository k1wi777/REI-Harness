# Implementación — Modo directo: cambios sin Work Item autorizados

## Resumen

Implementación documental completa de T1–T11: el modo directo queda definido como
vía fuera del workflow, con triaje, negativa sin override, confirmación
explícita, ausencia de registro, verificación proporcional y entrega solo con
orden. Sin código Go ni cambios de CLI. El rework completó R17 en `README.md` y
acotó la formulación absoluta de `workflow.md`.

## Archivos modificados

- `.rei/docs/harness/workflow.md` — nueva sección «Modo directo (fuera del
  workflow)»; bullet en Reglas generales; ajuste del Objetivo.
- `.rei/agents/leader.md` — `description`, Identidad, Objetivo, paso de triaje en
  Arranque, subprotocolo «Modo directo autorizado (sin Work Item)» y Reglas duras.
- `AGENTS.md` — subsección §5, excepción en §6 y bullet en §9; y
  `skeleton/AGENTS.md` — misma subsección §5 y bullet §9, neutrales.
- `.rei/docs/usage.md` — «Atajo: cambios directos sin Work Item» en §3 y FAQ.
- `README.md` — nota en «Cómo funciona §3» y párrafo en «Work Items».
- `.opencode/agents/leader.md`, `.claude/agents/leader.md` — regenerados.

## Cambios

- **Norma canónica:** definición (no tipo ni estado), disparador explícito,
  elegibilidad cerrada (documentación, erratas, formato, comentarios, trivial sin
  comportamiento), negativa (comportamiento, seguridad, datos, dependencias, API
  pública, arquitectura, multiarchivo o algo que necesite spec/revisión) con
  oferta de Task/Feature y sin override.
- **Autorización:** confirmación explícita incluso en casos elegibles, aviso de
  «sin planificación, rama ni registro»; ante duda, aviso y confirmación; defecto
  negativo. Se avisa si hay un Work Item `in_progress` por árbol/diff compartido.
- **Sin registro:** nada de `rei new`/`rei session`/`rei status set`/`rei commit
  set`/`rei github` ni `current.md`, `meta.json`, `history.md` o `direct.md`.
- **Verificación y entrega:** `rei check` + `go test`/`make build` según el cambio,
  comandos y resultado reportados; sin Reviewer salvo petición; sin commit, push,
  PR ni merge sin orden.
- **Coherencia:** misma formulación de elegibilidad/negativa en los seis
  documentos; se acotaron frases absolutas («todo trabajo sigue el mismo flujo»,
  «sin implementar ni planificar tú mismo», «el código nunca se implementa»).

## Corrección de revisión (rework)

- **R17 en `README.md`:** el párrafo de «Work Items» nombra el disparador
  («cambio directo» o «sin Work Item»), conserva elegibilidad y negativa y añade
  que **sin una orden explícita del usuario no hay commit, push, PR ni merge**.
- **Observación no bloqueante:** en `workflow.md`, «la única vía ajena a los Work
  Items» pasa a «una vía ajena», para no chocar con la personalización (que
  tampoco crea Work Item). No había más formulaciones absolutas.
- T7 y T8 reabiertas en `tasks.md` y remarcadas tras verificar.

## Verificación

| Comando | Resultado |
|---------|-----------|
| `gofmt -l .` | sin salida (exit 0) |
| `go vet ./...` | OK |
| `go test ./...` | OK (todos los paquetes) |
| `make build` | OK (`bin/rei`, versión `dev`) |
| `./bin/rei check --quiet` | 0 |
| `./bin/rei init opencode --check` | 0 |
| `./bin/rei init claude --check` | 0 |
| `./bin/rei doctor` | 0, «Harness íntegro» |
| `./bin/rei init status` | 0, documentación personalizada |
| `./bin/rei validate <id>` | 0, `Resultado: OK` |

Diff (`git diff --name-only`): `AGENTS.md`, `README.md`, `skeleton/AGENTS.md`,
`.rei/docs/harness/workflow.md`, `.rei/docs/usage.md`,
`.rei/agents/leader.md`, `.opencode/agents/leader.md`,
`.claude/agents/leader.md`, `.rei/progress/current.md` y, dentro del Work Item,
`tasks.md` y `meta.json` (marcas y estado). Sin tocar `cmd/`, `internal/`,
`embed.go`, `.rei/adapters/` ni el resto de `skeleton/`.

## Observaciones

- `skeleton/AGENTS.md` conserva los marcadores `<!-- REI:PENDIENTE -->` y no cita
  comandos ni rutas del repositorio fuente.
- No se creó registro para el modo directo (es una política, no una sesión): este
  `impl.md` documenta el Work Item que la implementa, no un uso del modo.
- Sin commit, push, PR ni merge; el árbol queda listo para revisión.
