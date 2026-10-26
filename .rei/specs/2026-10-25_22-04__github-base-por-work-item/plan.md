# Plan — GitHub: base por Work Item en la puerta

## Objetivo

Sustituir la base global del flujo GitHub por una **base capturada en la puerta
«issue + rama» y registrada por Work Item**, usada después para crear
`feat/<n>-<slug>` y para el `--base` del PR. `integrations.github.base` deja de
ser la decisión del flujo y pasa a ser **fallback**. La política de cierre no
cambia: la issue se autocierra al mergear a la default, se cierra al llegar a
`dev` cuando esa es la rama de integración confirmada, y el agente lo verifica
en ambos casos.

Solo documentación, plantillas y esquema de `links.json`: **sin código Go**, sin
cambios en `internal/**` ni `cmd/**`, sin flags nuevos y sin dependencias.

## Decisiones

**D1 — Precedencia de la base.** Al actuar sobre un Work Item:

1. `links.json.base.branch` (registrada en la puerta);
2. línea `Base: <rama> @ <sha>` del cuerpo de la issue (recuperación o adopción);
3. `integrations.github.base` (fallback; si está vacío, `main`);
4. preguntar al usuario.

Nunca se infiere la base de `origin/HEAD` ni de la rama actual cuando la
precedencia no lo autoriza. Si (1) y (2) existen y discrepan, se pregunta; no se
elige en silencio.

**D2 — Esquema aditivo en `links.json`.** Campo opcional:

```json
"base": { "branch": "dev", "sha": "<sha corto de 12 hex>" }
```

- Ausente ⇒ Work Item v1: se aplica D1 desde el paso 2.
- `sha` = `git rev-parse --short=12 HEAD` en el momento de la puerta; el cuerpo
  de la issue usa el mismo valor.
- Corrupto o incompleto ⇒ se trata como ausente, se informa y se recupera; nunca
  bloquea el Work Item.

**D3 — Tipo de base y cierre.** La default es la rama de
`git symbolic-ref --quiet --short refs/remotes/origin/HEAD` (sin el prefijo
`origin/`); si no se puede resolver, `main`. El cierre depende del **tipo de base
de la PR**, confirmado en la puerta, nunca de la mera existencia de `dev`:

1. `base == default` ⇒ tipo **default**: `Closes #N`; autocierre al mergear.
2. `base != default` y confirmada como rama de integración (p. ej. `dev`) ⇒ tipo
   **integración**: **sin `Closes`**; tras el merge el agente verifica y cierra
   manualmente (política acordada: cerrar al llegar a `dev`).
3. `base` es una rama `feat/*` o el usuario confirma PR apilada ⇒ tipo
   **apilada**: `Refs #N`, sin cierre; «cierre pendiente» hasta que el trabajo
   llegue a la rama de integración y se verifique.
4. Tipo ambiguo (no es la default, no es `feat/*` y no hay confirmación) ⇒ se
   confirma en la misma pregunta de la puerta; nunca se infiere en silencio.

Un Work Item con base `main` autocierra aunque `dev` exista en `origin`.

**D4 — Una sola pregunta.** La base se muestra dentro de la pregunta única de la
puerta: «¿Creo la issue y la rama `feat/<n>-<slug>` desde `<base>` (@sha)?».
La confirmación explícita es obligatoria si `<base>` no es la default o si es una
rama `feat/*` (aviso de PR apilada). La decisión (rama + sha, tipo de base —
default, integración o apilada— y fallback) se registra en la Bitácora de
`current.md` para no repreguntar.

**D5 — Rama desde la base sin exigir `HEAD == base`.**

```bash
git rev-parse --abbrev-ref HEAD        # rama actual (o "HEAD" si detached)
git rev-parse --verify refs/heads/<base>
git switch -c feat/<n>-<slug> <base>                      # base local
git fetch origin <base>                                   # solo si no es local
git switch --no-track -c feat/<n>-<slug> origin/<base>     # sin upstream heredado
```

Si la base no existe ni en local ni en `origin`, no se crea la rama (ver §12).
Nunca `-f`, nunca recrear una rama existente.

**D6 — Base publicada antes del PR.** `git ls-remote --heads origin <base>` debe
devolver una línea antes de `gh pr create`. Si está vacío, se detiene el PR y se
pide `git push origin <base>` (bajo orden explícita). Nunca se cae en silencio a
la default.

**D7 — Adopción sin reescribir issues.** No se añade `Base:` al cuerpo de issues
ya creadas (fuera de alcance). La base recuperada se guarda en `links.json` y en
`current.md`.

## Archivos

| Archivo | Acción |
|---------|--------|
| `.rei/docs/harness/integrations/github.md` | Reescritura de §1, §2, §3.3, §4 (4.1–4.3), §5.2/§5.4, §11, §12 y §13. |
| `.rei/agents/leader.md` | Actualizar la viñeta de la puerta y la de PR/merge del contrato. |
| `.opencode/agents/leader.md`, `.claude/agents/leader.md` | Regenerar con el CLI (`rei init opencode`, `rei init claude`); no editar a mano. |
| `.rei/docs/harness/README.md` | Actualizar la descripción del playbook (una línea). |
| `.rei/install-manifest.json` | Ignorado por git; puede reescribirse al regenerar. No commitear ni mencionar como cambio. |

No se crea ni se modifica ningún `links.json` existente: el cambio es de esquema
y de la escritura futura.

## Cambios

### 1. Alcance y configuración (§1, §2)

- En «Incluido en v2», añadir: «base por Work Item en la puerta (captura,
  registro en `links.json`/issue y uso en rama y PR); `integrations.github.base`
  es solo fallback».
- §2: redefinir `base` como **fallback** («se usa solo si el Work Item no tiene
  base registrada y no se puede preguntar»). Mantener el JSON y los comandos.
- Aclarar que la base por Work Item **no** se escribe en `.rei/config.json` ni la
  modifica `rei github enable`, y que `rei github status` sigue siendo offline y
  no conoce la base efectiva de cada Work Item.

### 2. `links.json` (§3.3)

- Añadir `base` al ejemplo y las reglas de D2 (opcional, `sha` corto, corrupto =
  ausente, no bloquea, recuperable por la issue).

### 3. Puerta (§4, §4.1, §4.2, §4.3)

Reescribir §4 con la captura de base **antes** de la pregunta única:

```bash
git rev-parse --abbrev-ref HEAD      # "HEAD" ⇒ detached
git rev-parse --short=12 HEAD
default=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD | sed 's#^origin/##')
```

| Situación | Base elegida | Acción |
|-----------|--------------|--------|
| `HEAD` = default | `HEAD` | Sin aviso; la pregunta única muestra la base (tipo **default**). |
| `HEAD` ≠ default, no es `feat/*` | `HEAD` | Mostrar «base = `<base>` (no es la default `<default>`)» y **confirmar en la misma pregunta** que es la rama de integración (tipo **integración**, p. ej. `dev`). |
| `HEAD` = `feat/*` | `HEAD` | Aviso de PR apilada: el PR irá contra `<base>`, no contra la default; exigir confirmación explícita (tipo **apilada**). |
| default no resoluble | — | Preguntar cuál es la default y la rama de integración; no asumir `main` ni deducir el tipo en silencio. |
| detached (`HEAD`) | `integrations.github.base` o `main` | Ofrecer el fallback; si no hay respuesta, usarlo y decirlo en la Bitácora. Nunca adivinar la rama original. |
| `links.json.base` ya existe | la registrada | Mostrarla; si `HEAD` difiere, preguntar si se conserva o se actualiza. Nunca cambiarla en silencio. |

- Mantener la regla de **una sola** pregunta por sesión, ahora con la base
  incorporada.
- Registrar en `current.md` (Bitácora): base, sha y **tipo de base**
  (default, integración o apilada) y, si aplica, fallback por detached.
- Si existe `links.json` pero sin `base`, resolverla y añadirla sin repreguntar
  issue/rama.
- Camino «no» (§4.3): no se crea issue ni rama; se trabaja en `<base>` si es una
  rama local, o en el estado actual si `HEAD` está detached; registrar la
  decisión.

En §4.1 (camino «sí»):

- Escribir `links.json` con `base` **antes** de crear la rama.
- Crear la rama con D5 y reportar «rama creada desde `<base>` (@sha)» aunque
  `HEAD` estuviera en otra rama.
- Reportar issue, base, sha, rama y `links.json`.
- Mantener el reintento sin duplicar si algo falla a medias.

§4.2 (plantilla del cuerpo de la issue): añadir la línea `Base: <rama> @ <sha>`
justo después del marcador y de la línea en blanco, antes de
`## Contexto y decisión`. Solo para issues nuevas.

### 4. Adopción y promoción (§5.2, §5.4)

- Añadir el orden de resolución de D1 antes de crear rama o publicar la spec:
  `links.json.base` → línea `Base:` de la issue → preguntar → fallback.
- En adopción/promoción, guardar la base resuelta en `links.json` y en
  `current.md` (rama, sha y tipo); **no** reescribir el cuerpo de la issue para
  añadir `Base:`.
- Si la issue no declara base y el usuario no responde, aplicar el fallback y
  declararlo; nunca inventar la rama de origen.
- Si el tipo de base no se deduce (no es la default ni `feat/*`), confirmarlo con
  el usuario antes de crear la rama o el PR; nunca asumir que es integración.
- Si la base registrada ya no existe, derivar a §12.

### 5. PR y merge (§11)

- Antes de crear el PR, verificar D6; si falta, detenerse con causa y siguiente
  paso (`git push origin <base>`).
- Usar `gh pr create ... --base <base>` con la base **del Work Item**.
- Decidir el cuerpo y el cierre por el **tipo de base** confirmado en la puerta
  (D3), no por la existencia de `dev`:
  - **default** (`<base>` == default): `Closes #N`; autocierre al mergear.
  - **integración** (`<base>` no default, p. ej. `dev`): **sin `Closes`**; tras
    el merge, verificar y cerrar manualmente con `gh issue close <n> --reason
    completed`, comprobando `CLOSED`.
  - **apilada** (`feat/*` o confirmada): `Refs #N`; sin cierre, con «cierre
    pendiente al llegar a `<integración>`».
- Tras el merge: verificar siempre la issue. Si sigue abierta y el tipo es
  default o integración, cerrarla bajo la misma orden y verificar `CLOSED`; si
  es apilada, no cerrar y reportar el pendiente.
- Mantener squash, `--subject "<título> (#<n>)"`, `--delete-branch`, nunca
  `--admin` ni auto-merge.

### 6. Errores y recordatorios (§12, §13)

Añadir filas a la tabla de errores:

| Situación | Detección | Mensaje (causa) | Siguiente paso |
|-----------|-----------|-----------------|----------------|
| `HEAD` detached en la puerta | `git rev-parse --abbrev-ref HEAD` = `HEAD` | «`HEAD` está detached: no hay rama actual» | Indicar la base o aceptar el fallback `integrations.github.base` |
| Base no registrada al adoptar | sin `links.json.base` ni línea `Base:` | «El Work Item no tiene base registrada» | Preguntar la rama y guardarla en `links.json` |
| Base solo local | `git ls-remote --heads origin <base>` vacío y rama local existente | «La base `<base>` no existe en `origin`» | `git push origin <base>` y reintentar el PR |
| Base borrada o renombrada | no existe en local ni en `origin` | «La base `<base>` del Work Item ya no existe» | Confirmar el nombre nuevo y actualizar `links.json`, o recrear/publicar la rama |
| Base contradictoria | `links.json.base` ≠ línea `Base:` | «La base registrada difiere de la issue» | Confirmar cuál vale y corregir `links.json` |
| Tipo de base ambiguo | base ≠ default, no es `feat/*` y no hay confirmación | «No está claro si `<base>` es la rama de integración» | Confirmar el tipo en la puerta (default, integración o apilada); nunca asumirlo |
| PR apilada sin confirmar | base `feat/*` sin confirmación | «La PR iría contra `<base>`, no contra la default `<default>`» | Confirmar la PR apilada o elegir otra base |

Añadir a §13:

- `GitHub: pendiente publicar la base <base> en origin antes del PR.`
- `GitHub: pendiente actualizar la base del Work Item (<base> no existe).`
- `GitHub: pendiente cierre de la issue #<n> al llegar a <integración> (PR apilada).`

### 7. Contrato del Leader (`.rei/agents/leader.md`)

Actualizar solo la sección «Integración con GitHub (opcional)»:

- **Puerta:** detectar `HEAD` y sha; mostrar/confirmar la base (obligatorio si no
  es la default o es `feat/*`); registrar `base` en `links.json` y `Base:` en el
  cuerpo; crear la rama desde esa base sin exigir `HEAD == base`; fallback
  `integrations.github.base`; confirmar PR apilada.
- **PR y merge:** `--base` por Work Item; verificar la base en `origin`; cierre
  por tipo de base: default → `Closes` (autocierre), integración no default →
  cierre manual verificado, apilada → `Refs` y cierre pendiente.
- Mantener el resto del contrato y la concisión actual (sin duplicar el
  playbook).

### 8. Índice de documentación (`.rei/docs/harness/README.md`)

Actualizar la descripción del playbook para incluir «base por Work Item» además
de épicas, backlog, adopción, asignación y cierre.

## Restricciones

- Sin código Go nuevo: nada en `internal/**` ni `cmd/**`; sin flags `--base` en
  `rei`; sin dependencias.
- El CLI sigue sin llamar a GitHub; las acciones las ejecuta el Leader.
- No cambiar la política de cierre acordada ni reescribir issues existentes.
- No cambiar `.rei/config.json` (la base global sigue en `main` como fallback).
- Conservar v1/v2 tal cual donde este plan no lo modifique; integración opcional
  con el mismo comportamiento cuando está desactivada.
- Un solo Work Item `in_progress`.

## Verificación

- `rei check --quiet` exit 0.
- `go test ./...` en verde (no hay cambios de código; es comprobación de que no
  se tocó nada).
- `rei init opencode --check` y `rei init claude --check` exit 0.
- `git diff --stat`: solo los archivos de la tabla; nada bajo `internal/` ni
  `cmd/`; `.rei/install-manifest.json` no aparece (ignorado).
- Greps de consistencia:
  - no queda ninguna regla que exija `HEAD == <base>` ni «detente y pregunta»
    por rama distinta;
  - todo `gh pr create` usa `--base <base>` del Work Item;
  - la plantilla de §4.2 contiene `Base: <rama> @ <sha>`;
  - §3.3 documenta `base` y §12 tiene las filas nuevas;
  - `integrations.github.base` solo aparece como fallback;
  - el cierre se decide por el tipo de base (default, integración o apilada), no
    por `git ls-remote --heads origin dev`.
- Prueba de campo (requiere autorización explícita del usuario registrada por el
  Leader en `current.md`, §13.1):
  1. con `dev` existiendo en `origin`, crear un Work Item pequeño **en `main`**:
     la puerta registra tipo default, el PR a `main` lleva `Closes #N` y la issue
     se autocierra al mergear;
  2. crear `dev` desde `main`, abrir otro Work Item pequeño en `dev` y recorrer
     puerta → issue con `Base: dev` + `links.json` con base → rama desde `dev` →
     PR a `dev` sin `Closes` → merge → cierre manual verificado → `dev → main`.
  No reutilizar `feat/2`. Dejar comandos, números, URLs y el tipo de base
  confirmado en `.rei/progress/work-items/<id>/impl.md`.
- Comprobaciones sin efectos secundarios: en una rama `feat/*`, la puerta muestra
  el aviso de PR apilada (declinar no crea nada); con una base local-only,
  verificar el mensaje de §12 antes de `gh pr create` sin llegar a crear el PR.

## Pasos

- [x] 1. Reescribir §1 y §2 del playbook (alcance v2 y `base` como fallback).
- [x] 2. Documentar el campo `base` en §3.3 y su recuperación.
- [x] 3. Reescribir §4/§4.1/§4.2/§4.3 con la captura, clasificación y
      confirmación de la base y de su **tipo**; añadir `Base:` a la plantilla y
      la creación de rama desde la base.
- [x] 4. Añadir la resolución de base y tipo en adopción y promoción (§5.2, §5.4).
- [x] 5. Actualizar §11 (verificación en `origin`, `--base` por Work Item,
      `Closes`/cierre según el tipo de base y verificación posterior).
- [x] 6. Añadir filas a §12 y recordatorios a §13.
- [x] 7. Actualizar la sección de GitHub del contrato del Leader.
- [x] 8. Regenerar los agentes nativos (`rei init opencode`, `rei init claude`) y
      comprobar con `--check`.
- [x] 9. Actualizar la línea del playbook en `.rei/docs/harness/README.md`.
- [x] 10. Ejecutar la verificación estática (`rei check`, `go test ./...`,
      greps, alcance del diff).
- [x] 11. Ejecutar la prueba de campo con autorización registrada y documentarla
      en `impl.md`; si no hay autorización, dejarla pendiente y reportarlo.
