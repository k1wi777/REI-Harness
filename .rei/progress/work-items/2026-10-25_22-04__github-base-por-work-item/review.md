# Revisión — GitHub: base por Work Item en la puerta

## Resultado

**Aprobado (`done`).** La implementación cumple el plan y la prueba de campo
está verificada en vivo.

## Verificaciones

- **Plan ↔ diff.** `rei review-diff` (base `1d3ae0b`) coincide con el plan:
  playbook, contrato del Leader, agentes nativos regenerados y `README.md`. Sin
  archivos ajenos a la tabla del plan.
- **Puerta solo `feature`.** Explícito en §1, §4 («No preguntes la puerta a una
  `task` «por si acaso»») y en el contrato del Leader (Caso C + viñeta de puerta:
  tasks sin issue/rama salvo opt-in). Corrección obligatoria cumplida.
- **Base por Work Item.** §4 captura `HEAD`/`--short=12`/default, clasifica tipo
  (default, integración, apilada, ambiguo), una sola pregunta con base y
  confirmación obligatoria fuera de default o en `feat/*`. §4.1 escribe
  `links.json` con `base {branch, sha}` y crea la rama desde la base sin exigir
  `HEAD == base`. §4.2 añade `Base: <rama> @ <sha>` solo en issues nuevas. §3.3
  documenta el campo (aditivo v1, corrupto = ausente, discrepancia ⇒ preguntar).
- **Cierre por tipo.** §11 y contrato: default → `Closes`/autocierre;
  integración no default → sin `Closes` + cierre manual verificado; apilada →
  `Refs`/cierre pendiente; ambiguo → confirmar. Sin reglas que decidan por
  `git ls-remote --heads origin dev`; `integrations.github.base` solo como
  fallback (grep: 11 menciones, todas de fallback).
- **Base en `origin` antes del PR.** §11 y contrato: `git ls-remote --heads
  origin "$base"` obligatorio; si falta, detenerse y pedir `git push origin
  <base>`. Fila correspondiente en §12.
- **Adopción.** §5.2 paso 6 y §5.4 paso 5: orden `links.json.base` → `Base:` de
  la issue → preguntar → fallback; guardar en `links.json`/`current.md`; nunca
  reescribir el cuerpo de issues existentes.
- **Alcance.** Sin cambios en `internal/**`, `cmd/**`, `go.mod`, `Makefile` ni
  `.rei/config.json`; `.rei/install-manifest.json` ignorado y sin tocar.
  `.opencode/agents/leader.md` y `.claude/agents/leader.md` regenerados
  (`init --check` OK). `README.md` con «base por Work Item».
- **Checkpoints.** `.rei/config.json` no define checks rápidos (no hay `V1...`
  que ejecutar). `./bin/rei check --quiet` exit 0; `rei init opencode --check` y
  `rei init claude --check` exit 0; `rei validate <id>` OK. Limitación del
  entorno: no hay toolchain Go (`go`, `gofmt`, `make vet`, `make build`
  ausentes), así que `go test`/`vet`/build no se ejecutaron; se mitiga con la
  ausencia total de cambios en código Go.
- **Prueba de campo (comprobada en vivo).** `gh`: issue **#7 CLOSED**, PR **#8**
  base `main` MERGED con `Closes #7`; issue **#9** cerrada 6 s después del merge
  del PR **#10** base `dev` (sin `Closes`), luego `dev → main` (`23c659b`).
  `git ls-remote --heads origin` solo `main`/`dev` en `23c659b`; `git worktree
  list` solo el repo principal; `git diff --stat 1d3ae0b..HEAD` vacío; `b997720`,
  `d237202` son commits vacíos y `23c659b` un merge sin archivos.

## Observaciones (no bloqueantes)

- El snippet de §11 usa `test -n "$(git ls-remote ...)" || { echo ...; }`, que
  solo avisa; el texto sí exige detener el PR. Para que el snippet sea
  ejecutable convendría añadir `exit 1`.
- `.rei/docs/project/verification.md` sigue siendo plantilla (`REI:PENDIENTE`);
  es previo y ajeno a esta Task.

## Acciones requeridas

Ninguna.
