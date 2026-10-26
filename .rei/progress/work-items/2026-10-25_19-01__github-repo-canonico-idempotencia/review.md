# Revisión — GitHub: idempotencia por API y aviso de remoto

## Resultado

**Aprobado.** El Work Item cumple `plan.md`: idempotencia por listado REST del
marcador con nombre canónico explícito en el playbook, aviso de remoto no
canónico en `rei github status` y tests que lo cubren.

## Verificaciones

| Check | Resultado | Evidencia |
|---|---|---|
| Playbook sin `gh issue list --search` para idempotencia | OK | §4.1 usa `gh api "repos/$canonical/issues?state=all&per_page=100" --paginate` con `select(.pull_request == null)` + `contains(marker)`; única mención restante de `--search` es la explicación de §3 («Rename y fork») |
| Nombre canónico al inicio de cada acción | OK | §3 «Nombre canónico del repositorio» + paso 0 de §4.1; `--repo "$canonical"` en labels/issue/view; recordatorios en §5, §6 y §7 |
| Fallo de resolución explícito (causa + siguiente paso, nunca nombre viejo) | OK | §3 y fila «Fallo al resolver el canónico» de §8 |
| Rename/fork, `origin` vs `upstream`, Issues deshabilitados, PR fork→upstream | OK | §3 «Rename y fork» (301 vs search API, fork resuelve al fork, nunca cambia sola a `upstream`, `repo` vacío→`origin`) y §7 limitación v1 |
| `status` avisa y no dice «Sin acciones pendientes» | OK | `mismatch` en `Status`; `[WARN]` tras la línea del remoto y `case mismatch:` antes de `default`; smoke manual: remoto viejo → aviso + siguiente paso, exit 0, `config.json` intacta; remoto coincidente → «Sin acciones pendientes.» |
| `status` sigue offline | OK | Sin llamadas nuevas al `Runner` ni escritura; solo compara `gh.Repo` con `remote.String()` |
| Tests | OK | `go test ./internal/github/ -run TestStatusRemoto -v` → 4/4 PASS (coincide, no coincide, desactivada, `repo` vacío); exit 0 y config intacta |
| Sin flags, dependencias ni cambios de contrato | OK | Diff toca solo docs + `internal/github/{github.go,github_test.go}`; `go.mod` sin cambios; `.rei/agents/` y adaptadores sin diff |
| Adaptadores sincronizados | OK | `rei init opencode --check` y `rei init claude --check` → exit 0 |
| `make test` / `make vet` / `make build` | OK | exit 0 (Go 1.27.2, toolchain local) |
| `gofmt -l .` | OK | sin salida |
| `rei check --quiet` | OK | exit 0 (binario instalado y `./bin/rei` recién compilado) |

`rei validate <id>` → OK. `.rei/config.json` no declara checkpoints rápidos
(`checks: []`), así que no hay IDs `V*` que referenciar.

## Observaciones (no bloqueantes)

- `TestStatusRemoto` no afirma explícitamente la ausencia de «Sin acciones
  pendientes.» en el caso de discrepancia (el plan lo sugería); el `switch` es
  exclusivo y el smoke manual lo confirma: no aparece.
- Se añadió una tercera fila a la tabla de errores («Issues deshabilitados»),
  exigida por el alcance de la Task y coherente con el plan (que hablaba de
  «fila(s)»); no altera comportamiento.
- Nota de entorno: `go` no está en `PATH`; se usó
  `/home/jose/.local/go-toolchain/go/bin`. El `rei` del `PATH` es 0.2.0 y no
  incluye `github`, por lo que el smoke se hizo con `./bin/rei`.

## Acciones requeridas

Ninguna.
