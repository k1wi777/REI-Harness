# Implementación — GitHub: idempotencia por API y aviso de remoto

## Resumen

Task re-scopeada ejecutada según `plan.md`. La idempotencia del playbook pasa del
search API (`gh issue list --search`, que no sigue renames y falla en silencio) a
un listado REST paginado filtrado por marcador, con resolución explícita del
`nameWithOwner` canónico al inicio de cada acción. `rei github status` avisa
cuando el remoto no coincide con `integrations.github.repo` y ya no concluye
«Sin acciones pendientes» en ese caso. Sin red en el CLI, sin flags nuevos y sin
cambios de contrato de rol.

## Archivos modificados

- `.rei/docs/harness/integrations/github.md` — nombre canónico, idempotencia REST,
  rename/fork, aviso de `status`, tabla de errores y limitación PR fork→upstream.
- `internal/github/github.go` — `mismatch`, `[WARN]` y `case` de siguiente paso.
- `internal/github/github_test.go` — `TestStatusRemoto` ampliado (4 casos).
- `.rei/specs/2026-10-25_19-01__github-repo-canonico-idempotencia/plan.md` —
  pasos marcados.

## Cambios

### Playbook (`.rei/docs/harness/integrations/github.md`)

- Sección 2: `status` documentado como offline y con aviso de remoto ≠ config.
- Sección 3: subsección «Nombre canónico del repositorio»
  (`gh repo view <repo> --json nameWithOwner --jq .nameWithOwner`, obligatorio al
  inicio de cada acción, uso en `--repo` y rutas REST; si falla, causa y siguiente
  paso, nunca el nombre viejo) y «Rename y fork» (el REST sigue el 301; el search
  API no; el fork resuelve al fork y nunca se cambia solo a `upstream`).
- Sección 4.1: paso 0 canónico, paso 1 `gh api
  "repos/$canonical/issues?state=all&per_page=100" --paginate` con filtro
  `pull_request == null` + `contains(marker)`, labels/issue/view con
  `--repo "$canonical"`. Eliminado `gh issue list --search`.
- Secciones 5, 6 y 7: recordatorio de que `<repo>` es el canónico resuelto.
- Sección 7: limitación v1 PR desde fork (`--repo <upstream> --head
  <fork-owner>:<branch>`), caso manual sin flags en `rei`.
- Sección 8: filas «Remoto ≠ nombre canónico», «Fallo al resolver el canónico» e
  «Issues deshabilitados» (404/410 → activar Issues o configurar otro repo).

### CLI (`internal/github/github.go`)

- `mismatch := gh != nil && gh.Repo != "" && remoteErr == nil &&
  !strings.EqualFold(gh.Repo, remote.String())`.
- Tras la línea del remoto: `[WARN]` + resolución canónica.
- `case mismatch:` insertado antes del `default` del «Siguiente paso»; la
  prioridad existente (`no configurada`, `desactivada`, `gh`, auth, remoto) se
  conserva. Sin llamadas nuevas al `Runner`, sin escritura y con exit 0.

### Tests (`internal/github/github_test.go`)

`TestStatusRemoto` cubre: coincide (sin aviso, «Sin acciones pendientes.»), no
coincide (aviso + siguiente paso con `gh repo view`), no coincide desactivada
(aviso + `rei github enable`) y `repo` vacío (sin comparación ni aviso). En todos
se comprueba exit 0, `config.json` intacto y que no se usan comandos no
simulados (`gh api`/`gh repo view` habrían caído en «comando no simulado»).

## Verificación

- `gofmt -l .` → sin salida.
- `go vet ./...` / `make vet` → OK (exit 0).
- `go test ./internal/github/` → OK; `go test ./...` / `make test` → OK.
- `make build` → OK (bin/rei).
- `rei check --quiet` → exit 0.
- `rei init opencode --check` y `rei init claude --check` → OK, sin regenerar.
- `grep -rn "issue list\|--search" .rei/agents/` → sin resultados.

## Observaciones

- No hubo contradicción entre `plan.md` y `meta.json`; el re-scope (sin flags
  `--repo/--base/--strategy`) se respetó.
- No se tocaron contratos de rol ni documentos históricos; la issue #1 no se
  modificó.
- Sin dependencias nuevas ni llamadas de red en `rei github status`.
- Nota de entorno: `go` no estaba en `PATH`; se usó
  `/home/jose/.local/go-toolchain/go/bin` para `make test|vet|build`.
- Sin commit, push ni merge.
