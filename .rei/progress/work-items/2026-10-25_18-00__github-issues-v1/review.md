# Review — Integración opcional con GitHub Issues (v1)

## Resultado

**Aprobado (`done`).** T1–T12 implementadas; requisitos R1–R56 cubiertos;
sin desviaciones respecto a `requirements.md`/`design.md`/`tasks.md`.

## Verificaciones ejecutadas

| Check | Resultado | Evidencia |
|-------|-----------|-----------|
| `gofmt -l internal cmd` | pasa | sin salida |
| `make vet` | pasa | `go vet ./...` OK |
| `make test -count=1` | pasa | todos los paquetes OK (incluye `internal/github`, `internal/config`, `internal/cli`) |
| `make build` | pasa | `bin/rei` compilado |
| `./bin/rei check --quiet` | pasa | exit 0 |
| `./bin/rei init opencode --check` | pasa | agentes OK |
| `./bin/rei init claude --check` | pasa | agentes + `CLAUDE.md` OK |
| `./bin/rei validate <id>` | pasa | OK |

Prueba manual en proyectos temporales (binario real):

- sin `.rei` ⇒ `github status` = `1` con causa/siguiente paso (R3).
- sin subcomando / desconocido / extra ⇒ `2` (R2, R49).
- sin config ⇒ `status` = `0` «no configurada»; `disable` = `0` «ya estaba
  desactivada» (R10, R11).
- `origin` SSH `git@github.com:acme/widget.git` y HTTPS
  `https://github.com/acme/widget.git` ⇒ `enable` = `0`, escribe
  `enabled/repo/base/mergeStrategy` y **preserva** `checks`, `viewer` y claves
  desconocidas (R5, R6, R13).
- `gh` ausente ⇒ aviso + `0` y config escrita (R9).
- `disable` conserva `repo/base/mergeStrategy` y el resto; no invoca `gh`
  (R10).
- remoto gitlab y ausencia de `origin` ⇒ `1` sin escribir (R7).
- JSON inválido ⇒ `enable|disable|status` = `1`, archivo intacto (R12).
- segunda ejecución de `enable` conserva `base: develop` y
  `mergeStrategy: merge` (R8, idempotencia R5).

Integración desactivada por defecto: `.rei/config.json` **no** contiene
`integrations.github` y no se modificó respecto a `base_commit` (R20, R43); el
contrato del Leader y el playbook solo actúan con `enabled: true`.

## Revisión de alcance y contratos

- Playbook `.rei/docs/harness/integrations/github.md` (431 líneas) cubre:
  principio de proyección, config/comandos, vocabulario fijo (marcadores y
  labels v1; `rei:epic`/`rei:backlog` reservadas), puerta única con pregunta
  exacta e idempotencia por marcador, plantilla autosuficiente con `design.md`
  plegado, camino «no», adopción, comentarios `impl`/`review` idempotentes por
  REST, `rei:blocked` sin comentarios, PR/merge solo bajo orden, cierre y
  verificación de la issue, tabla de errores (sin `.rei`, desactivada,
  incompleta, `gh` ausente/sin auth, repo/remoto, issue/rama/PR, permisos,
  branch protection, red, rate limit), aislamiento del estado de REI y
  recordatorios en `current.md` (R21–R55).
- Leader (`.rei/agents/leader.md` y nativos) incorpora puerta única en Caso C,
  comentarios tras Implementer/Reviewer, adopción al arrancar, PR/merge bajo
  orden, errores no bloqueantes y recordatorios (R52, R54).
- Initializer incorpora la pregunta de configuración sin editar JSON a mano y
  sin bloquear por `gh` (R15–R19, R52).
- Playbook declarado en `RequiredFiles`, `README.md` y `embed_test.go`
  (R51). Sin dependencias Go nuevas (`go.mod` intacto) ni llamadas HTTP a la
  API de GitHub; el paquete solo ejecuta `git` y `gh auth status` como
  diagnóstico (R14).
- Tests nuevos sin red: `fakeRunner` para `internal/github` (R56).

## Observaciones (no bloqueantes)

- `.rei/docs/project/verification.md` sigue siendo la plantilla del arnés y
  `.rei/config.json` no declara checks, por lo que no existen checkpoints `V1`
  concretos; se re-ejecutaron los comandos de T12 como evidencia.
- `status` clasifica la autenticación como instalado/sin autenticar/no
  instalado; no emite el estado «no comprobable» que `design.md` §5.4 menciona
  de forma orientativa. R11 se cumple.
- Con `enabled: true` y `repo` vacío (solo por edición manual), `status` no
  propone `rei github enable` como siguiente paso. Caso límite no exigido por
  R11.

## Acciones requeridas

Ninguna.
