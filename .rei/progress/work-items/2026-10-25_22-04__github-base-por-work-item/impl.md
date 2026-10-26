# Implementación — GitHub: base por Work Item en la puerta

## Resumen

Se sustituyó la base global por una **base por Work Item** capturada en la puerta
«issue + rama». Cambio solo documental/contractual: playbook, contrato del Leader,
agentes nativos regenerados e índice. Sin código Go, dependencias ni flags.

Corrección del usuario incorporada: la puerta es **solo para Features**; las
Tasks no generan issue ni rama por defecto (solo opt-in explícito).

## Archivos modificados

- `.rei/docs/harness/integrations/github.md` — §1, §2, §3.3, §4–§4.3, §5.2, §5.4,
  §11, §12, §13.
- `.rei/agents/leader.md` — puerta (solo features, base y tipo), PR/merge y Caso C.
- `.opencode/agents/leader.md`, `.claude/agents/leader.md` — regenerados con
  `rei init opencode` / `rei init claude` (no editados a mano).
- `.rei/docs/harness/README.md` — descripción del playbook con «base por Work Item».

## Cambios

- §1/§2: `integrations.github.base` pasa a **fallback**; la base efectiva no vive
  en `.rei/config.json` y `rei github status` sigue offline. Puerta solo features.
- §3.3: campo opcional `base {branch, sha}` (sha corto de 12), corrupto = ausente,
  orden de recuperación `links.json` → `Base:` → fallback → preguntar.
- §4: captura de `HEAD`/sha/default, clasificación del tipo (default, integración,
  apilada), pregunta única con base, confirmación obligatoria fuera de default y
  rama desde la base sin exigir `HEAD == base`. §4.2 añade `Base: <rama> @ <sha>`.
- §5.2/§5.4: la base y su tipo se resuelven y guardan en `links.json`/`current.md`;
  nunca se reescribe el cuerpo de issues existentes.
- §11: verificación de la base en `origin`, `--base` del Work Item, cierre por
  tipo (default → `Closes`/autocierre; integración → cierre manual verificado;
  apilada → `Refs` y cierre pendiente).
- §12/§13: 7 filas de error nuevas y 3 recordatorios.

## Verificación

- `./bin/rei check --quiet` → exit 0.
- `./bin/rei init opencode --check` y `./bin/rei init claude --check` → exit 0.
- Greps: sin `HEAD == <base>` ni «detente y pregunta» por rama; todo `gh pr create`
  con `--base`; plantilla con `Base: <rama> @ <sha>`; `integrations.github.base`
  solo como fallback; sin `ls-remote ... dev` para decidir el cierre.
- `git diff --stat`: solo los archivos del plan; nada en `internal/**`/`cmd/**`.
- **Pendiente por entorno:** no hay toolchain Go (`go`, `gofmt`, `make vet`,
  `make build` no disponibles), así que esos comandos no se pudieron ejecutar.

## Prueba de campo (autorizada)

`dev` creado desde `main` en `origin` (`1d3ae0b`) antes de los tests. Work Items
efímeros en worktrees temporales, commits vacíos.

- **Test A** (base `main`, tipo default): issue **#7** con `Base: main @
  1d3ae0b89e81`, `links.json` `base {branch: main, sha: 1d3ae0b89e81}`, rama
  `feat/7-field-test-a` desde `main`, PR **#8** a `main` con `Closes #7` →
  MERGED; issue **CLOSED** automáticamente (autocierre aunque `dev` exista).
- **Test B** (base `dev`, tipo integración): issue **#9** con `Base: dev @
  1d3ae0b89e81`, `links.json` `base {branch: dev, sha: 1d3ae0b89e81}`, rama
  `feat/9-field-test-b` desde `dev`, PR **#10** a `dev` sin `Closes` → MERGED;
  issue seguía **OPEN** y se cerró con `gh issue close 9 --reason completed`
  (verificado **CLOSED**). Después `dev → main` (fast-forward `23c659b`).
- Sin efectos secundarios: `feat/*` ⇒ aviso de PR apilada (declinar no crea nada);
  base solo local ⇒ mensaje de §12 sin crear PR.
- Limpieza: worktrees fuera, ramas de prueba borradas (local y remoto), `dev`
  queda en `origin` (`23c659b`), `main` local/remoto en `23c659b`, sin archivos de
  prueba. Issues y PRs de prueba quedan cerradas/mergeadas como evidencia.

## Observaciones

- No se creó issue ni rama para esta Task (solo Features pasan la puerta).
- Sin commit de esta Task ni PR: los hará el Leader tras la revisión.
