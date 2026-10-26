# GitHub — núcleo v1 (issue + rama, adopción, comentarios y PR)

> **Archivo de dominio del playbook de GitHub.** Cárgalo solo cuando la tabla
> «Qué leer según la acción» de [`../github.md`](../github.md) lo indique para la
> acción en curso; `github.md` es el índice de orientación y configuración.
>
> Referencias: [índice](../github.md) · [backlog](backlog.md) ·
> [plantillas](templates.md) · [errores](errors.md)

---

# 1. Alcance v1

Incluido en v1:

- configuración y diagnóstico (`rei github enable|disable|status`);
- vínculo issue ↔ Work Item con marcador y `links.json`;
- puerta única al aprobar la planificación de una **feature** (issue + rama; las
  tasks no la ejecutan por defecto);
- comentarios automáticos de implementación y revisión;
- PR y merge bajo orden explícita;
- adopción de issues existentes;
- labels `rei`, `rei:feature`, `rei:blocked`;
- manejo de errores con causa y siguiente paso.

---

# 2. `links.json`

`.rei/specs/<id>/links.json` es un **índice local recuperable**, nunca la fuente
de verdad:

```json
{
  "provider": "github",
  "repo": "owner/name",
  "issue": 123,
  "url": "https://github.com/owner/name/issues/123",
  "base": {
    "branch": "dev",
    "sha": "0a1b2c3d4e5f"
  },
  "epic": {
    "slug": "modulo-x",
    "issue": 100,
    "url": "https://github.com/owner/name/issues/100"
  }
}
```

Sin épica, `epic` es `null` (compatible con v1). El objeto `epic` evita la
ambigüedad entre slug y número y permite recuperar la épica sin releer el
cuerpo. Si el archivo falta, está corrupto o se desincroniza, **no bloquea** el
Work Item: se recupera la issue por su marcador y se reescribe.

`base` (opcional, aditivo y compatible con v1) registra la base del Work Item
capturada en la puerta:

- `branch` es la rama de origen; `sha` es `git rev-parse --short=12 HEAD` en el
  momento de la puerta, el mismo valor que el cuerpo de la issue.
- **Ausente** ⇒ Work Item v1: la base se resuelve al actuar (§5 y §6) con el
  orden `links.json.base` → línea `Base:` de la issue →
  `integrations.github.base` (o `main` si está vacío) → preguntar al usuario.
  Nunca se infiere de `origin/HEAD` ni de la rama actual.
- **Corrupto o incompleto** (falta `branch`, `sha` inválido, JSON roto) ⇒ se
  trata como ausente, se informa y se recupera; **nunca** bloquea el Work Item.
- Si `links.json.base` y la línea `Base:` de la issue existen y **discrepan**, se
  pregunta cuál vale; no se elige en silencio.

La base así resuelta se guarda en `links.json` y en `current.md`, pero **no** se
reescribe el cuerpo de issues ya creadas para añadirles la línea `Base:`.

---

# 3. Nombre canónico del repositorio

En este documento, `<repo>` significa siempre el **nombre canónico vigente**
`owner/name`. Se resuelve **al inicio de cada acción GitHub** (no una vez por
sesión: un rename a mitad de sesión invalidaría el valor cacheado):

```bash
canonical=$(gh repo view <repo> --json nameWithOwner --jq .nameWithOwner)
```

Ese valor se usa en **todos** los comandos de la acción: `--repo <canonical>` en
`gh label create`, `gh issue view|create|edit|close|comment` y
`gh pr create|view|merge`, y en las rutas REST `repos/<canonical>/issues/...`.

- `<repo>` de partida es `integrations.github.repo`; si está vacío, el
  `owner/name` detectado en `origin`.
- Si la resolución falla (sin red, sin auth, sin permisos, repo inexistente), se
  informa de **causa y siguiente paso** y se detiene la acción. **Nunca** se
  continúa con el nombre viejo ni con el valor de la config sin resolver.

---

# 4. Rename y fork

GitHub **redirige** el nombre viejo al nuevo (301), por lo que el listado REST
funciona con ambos nombres. El **search API no sigue ese redirect**: indexa con
retraso y, ante un nombre viejo, devuelve lista vacía con exit 0 — un fallo
silencioso que lleva directo al duplicado. Por eso la idempotencia **nunca** se
apoya en `gh issue list --search`.

En un **fork**, `gh repo view` resuelve el nombre consultado al fork, no al
upstream. La integración opera siempre sobre el repositorio configurado (o, si
`repo` está vacío, sobre `origin`); **nunca** cambia sola a `upstream`. Para
operar sobre el upstream hay que ponerlo explícitamente en
`integrations.github.repo`.

---

# 5. Puerta única al aprobar la planificación de una feature (v1 + base por Work Item)

**Disparador:** `meta.json` está en `ready`, el usuario aprueba,
`integrations.github.enabled` es `true` **y** `type` es `feature`.

**Solo features.** La puerta «issue + rama» **no** se ejecuta para Work Items de
tipo `task`: una task se trabaja en la rama base sin crear issue ni rama, y no se
pregunta nada. Solo con un **opt-in explícito** del usuario (pide issue y rama
para la task) se aplica esta sección completa. No preguntes la puerta a una
`task` «por si acaso».

Antes de `rei status set <id> in_progress`:

1. Lee `.rei/config.json` y comprueba `enabled`.
2. Comprueba `type == feature` (o el opt-in explícito de una task).
3. Si ya existe `.rei/specs/<id>/links.json`, no preguntes issue ni rama; si le
   falta `base`, resuélvela (§2) y añádela sin repreguntar.
4. Si `current.md` ya registra la decisión de este Work Item (base, sha y tipo),
   no vuelvas a preguntar.
5. Captura la base **antes** de formular la pregunta única:

```bash
current=$(git rev-parse --abbrev-ref HEAD)   # "HEAD" ⇒ detached
sha=$(git rev-parse --short=12 HEAD)
default=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD | sed 's#^origin/##')
```

Clasificación de la base y su **tipo** (el tipo decide el cierre en §8):

| Situación | Base elegida | Acción |
|-----------|--------------|--------|
| `HEAD` = default | `HEAD` | Sin aviso; la pregunta única muestra la base (tipo **default**). |
| `HEAD` ≠ default, no es `feat/*` | `HEAD` | Mostrar «base = `<base>` (no es la default `<default>`)» y **confirmar en la misma pregunta** que es la rama de integración (tipo **integración**, p. ej. `dev`). |
| `HEAD` = `feat/*` | `HEAD` | Aviso de PR apilada: el PR irá contra `<base>`, no contra la default; exigir confirmación explícita (tipo **apilada**). |
| default no resoluble | — | Preguntar cuál es la default y la rama de integración; no asumir `main` ni deducir el tipo en silencio. |
| detached (`HEAD`) | `integrations.github.base` o `main` | Ofrecer el fallback; si no hay respuesta, usarlo y decirlo en la Bitácora. Nunca adivinar la rama original. |
| `links.json.base` ya existe | la registrada | Mostrarla; si `HEAD` difiere, preguntar si se conserva o se actualiza. Nunca cambiarla en silencio. |

Si `links.json.base` y la línea `Base:` de la issue discrepan, pregunta; nunca
elijas en silencio ([`errors.md §1`](errors.md)).

6. Formula **una sola** pregunta, con la base incorporada:

   > «¿Creo la issue y la rama `feat/<n>-<slug>` desde `<base>` (@`<sha>`)?»
   > (el `<n>` será el número de la issue y `<slug>` el slug del Work Item).

   La confirmación explícita es **obligatoria** si `<base>` no es la default o si
   es una rama `feat/*` (aviso de PR apilada). Si el usuario no confirma, no se
   crea nada.

Nunca repitas la pregunta en la misma sesión. Registra en `current.md`
(Bitácora): base, sha, **tipo de base** (default, integración o apilada) y, si
aplica, el fallback usado por detached.

## 5.1 Camino «sí»

Orden exacto. Sustituye `<id>`, `<repo>`, `<título>`, `<base>`, `<sha>` y `<n>` por
sus valores.

```bash
# 0. Nombre canónico (obligatorio al inicio de la acción)
canonical=$(gh repo view <repo> --json nameWithOwner --jq .nameWithOwner)

# 1. ¿Existe ya la issue? (fuente de verdad: el marcador, vía REST)
gh api "repos/$canonical/issues?state=all&per_page=100" --paginate \
  --jq '.[] | select(.pull_request == null)
        | select((.body // "") | contains("<!-- rei:work-item:<id> -->"))
        | [.number, .state, .html_url] | @tsv'

# 2. Labels idempotentes (solo si falta la issue)
gh label create rei --color 1D76DB --description "Gestionado por REI" --force \
  --repo "$canonical"
gh label create 'rei:feature' --color 0E8A16 \
  --description "Work Item de tipo feature" --force --repo "$canonical"
```

El endpoint de issues incluye también PRs: se descartan los que traen
`pull_request`. El marcador se busca en `.body` y el resultado es autoritativo
aunque el remoto sea un nombre viejo (el REST sigue el redirect del rename).
`--paginate` evita perder issues antiguas.

Si el paso 1 devuelve una issue, **reutilízala**: informa de número y URL y no
crees otra. Si devuelve más de una, detente y reporta los duplicados. Si no
existe, escribe el cuerpo con la plantilla de [`templates.md §1`](templates.md)
en un archivo temporal y crea (pasos 3-4). El cuerpo sigue §1 y **no** incluye el
`design.md` completo por defecto; si `.rei` no se versiona o el usuario pide
trazabilidad, publica el design como comentario idempotente
`<!-- rei:comment:design:<id> -->` ([`templates.md §7`](templates.md)).

```bash
# 3. Crear la issue (solo si no existe)
gh issue create --repo "$canonical" \
  --title "<título del Work Item>" \
  --body-file /tmp/rei-issue-<id>.md \
  --label rei --label rei:feature

# 4. Verificar número/URL/estado antes de afirmar éxito
gh issue view <n> --repo "$canonical" --json number,url,state,body
```

5. Escribe `.rei/specs/<id>/links.json` con `base {branch, sha}` **antes** de
   crear la rama, para no perder el vínculo si la rama falla. Si ya existía,
   actualízalo conservando el resto de campos.
6. Crea la rama **desde la base**, sin exigir que `HEAD` sea la base:

```bash
git rev-parse --abbrev-ref HEAD        # rama actual (o "HEAD" si detached)
git rev-parse --verify refs/heads/<base>
git switch -c feat/<n>-<slug> <base>                    # base local
# solo si la base no está en local:
git fetch origin <base>
git switch --no-track -c feat/<n>-<slug> origin/<base>  # sin upstream heredado
```

Si la base no existe ni en local ni en `origin`, **no** crees la rama: informa
con causa y siguiente paso ([`errors.md §1`](errors.md)). Nunca uses `-f` ni
recrees una rama existente; si `feat/<n>-<slug>` ya existe, informa y ofrece
reutilizarla o elegir otra.
7. Reporta issue, base, sha, rama y `links.json` («rama creada desde `<base>`
   (@`<sha>`)» aunque `HEAD` estuviera en otra rama).

Si la issue se creó pero falla la rama o `links.json`, informa de **qué quedó
hecho y qué está pendiente**, conserva la issue y permite reintentar sin
duplicarla.

## 5.2 Camino «no»

- No se crea issue ni rama.
- Se trabaja en `<base>` si es una rama local; si `HEAD` está detached, en el
  estado actual.
- Se registra en `current.md` (Bitácora):
  «GitHub: issue/rama declinadas; se trabaja en `<base>`».
- No se vuelve a preguntar.

---

# 6. Adopción

## 6.1 Lectura y conflictos

Con la integración activa, «trabajemos en la #N» ejecuta:

```bash
gh issue view <n> --repo "$canonical" \
  --json number,title,body,state,labels,assignees,comments,url
```

Orden de comprobaciones:

1. **Épica** (`rei:epic`) ⇒ no se adopta; se explica que las épicas no son Work
   Items y se ofrecen sus hijas.
2. **Marcador de Work Item** (`rei:work-item:<id>`) ⇒ no se crea otro Work Item
   ni se sobrescribe el marcador. Si el Work Item existe en este clon, se
   reutiliza y se informa; si no existe (otro clon u otro Work Item), se informa
   del conflicto y se detiene.
3. **Cerrada** (la issue o la épica a la que pertenece) ⇒ se pregunta antes de
   continuar.
4. **Asignada a otra persona** ⇒ se pregunta antes de adoptar o reasignar.
5. **Marcador de backlog** ⇒ se promueve ([`backlog.md §2`](backlog.md)).
6. **Sin marcador** ⇒ se aplica la adopción genérica de v1 (§6.2).

La label `rei:feature` sin marcador **no** prueba que la issue ya esté adoptada;
la identidad es el marcador.

## 6.2 Adopción genérica (v1, sin marcador)

1. Comprobar que la integración está activa. Si no, no leer ni modificar issues:
   ofrecer `rei github enable`.
2. Leer la issue con la orden de §6.1.
3. Si está cerrada, preguntar antes de continuar.
4. Sin marcador: preguntar lo que falte (objetivo, alcance, criterios, fuera de
   alcance, restricciones), crear el Work Item con `rei new`, rellenar
   `title`/`description` con el acuerdo y añadir el marcador **sin destruir el
   contenido**:

   ```bash
   gh issue view <n> --repo "$canonical" --json body --jq .body \
     > /tmp/rei-issue-<n>-body.md
   # antepone <!-- rei:work-item:<id> --> al archivo
   gh issue edit <n> --repo "$canonical" --body-file /tmp/rei-issue-<n>-body.md
   gh label create rei --color 1D76DB --description "Gestionado por REI" --force \
     --repo "$canonical"
   gh label create 'rei:feature' --color 0E8A16 \
     --description "Work Item de tipo feature" --force --repo "$canonical"
   gh issue edit <n> --repo "$canonical" \
     --add-label rei --add-label 'rei:feature'
   ```

5. Resolver la base del Work Item antes de escribir `links.json` o crear la rama:
   `links.json.base` → línea `Base: <rama> @ <sha>` de la issue → preguntar →
   `integrations.github.base` (fallback; si está vacío, `main`). Si el usuario no
   responde, aplica el fallback y decláralo; **nunca** inventes la rama de origen.
   Si el tipo de base no se deduce (no es la default ni `feat/*`), confírmalo con
   el usuario antes de crear la rama o el PR; nunca asumas integración. No
   reescribas la issue para añadir `Base:`; guarda la base en `links.json` y
   `current.md`. Si la base ya no existe, deriva a
   [`errors.md §1`](errors.md).
6. Escribir `.rei/specs/<id>/links.json` con `base {branch, sha}` y `epic: null`
   (o el objeto `epic` si la issue declara pertenecer a una épica y esta se
   resuelve sin ambigüedad).
7. Continuar con el workflow normal (`spec_author`).

---

# 7. Comentarios automáticos (v1)

Disparadores: el Leader acaba de recibir al Implementer (comentario `impl`) y al
Reviewer (comentario `review`). Se publican **sin preguntar**.

Búsqueda idempotente del comentario (REST, para obtener el `id` numérico):

```bash
gh api "repos/$canonical/issues/<n>/comments" --paginate \
  --jq '.[] | select(.body | contains("<!-- rei:comment:impl:<id> -->")) | .id'
```

- Si devuelve un `id`: actualiza en sitio.

  ```bash
  gh api --method PATCH "repos/$canonical/issues/comments/<comment-id>" \
    -F body=@/tmp/rei-comment-impl-<id>.md
  ```

- Si no devuelve nada: crea.

  ```bash
  gh issue comment <n> --repo "$canonical" \
    --body-file /tmp/rei-comment-impl-<id>.md
  ```

- Verifica releyendo el comentario o comprobando que el marcador aparece **una
  sola vez**. Nunca dupliques.

El contenido se escribe con las plantillas de [`templates.md §5`](templates.md)
(implementación) y [`templates.md §6`](templates.md) (revisión); el procedimiento
es el mismo cambiando `<!-- rei:comment:impl:<id> -->` por
`<!-- rei:comment:review:<id> -->`.

Reglas:

- Si una revisión se repite, **actualiza** el mismo comentario de review.
- **Nunca** publiques comentarios por cambios de estado del Work Item: los
  únicos comentarios automáticos son `impl` y `review`.
- Si el Work Item pasa a `blocked`, aplica `rei:blocked`; al desbloquearlo,
  retírala. Sin comentarios:

  ```bash
  gh label create 'rei:blocked' --color B60205 --description "Bloqueado" --force
  gh issue edit <n> --repo "$canonical" --add-label 'rei:blocked'
  gh issue edit <n> --repo "$canonical" --remove-label 'rei:blocked'
  ```

Un fallo al publicar o actualizar un comentario se reporta; **no** cambia el
estado de REI y se puede reintentar sin duplicar.

---

# 8. Pull request y merge (v1 + base por Work Item)

Solo **tras la revisión** y bajo **orden explícita** del usuario.

La base del PR es la **base del Work Item**, resuelta en la puerta y registrada
(§2/§5), no `integrations.github.base` si el Work Item ya tiene base. Antes de
crear el PR, verifica que la base está publicada en `origin`; si no lo está, se
detiene el PR y se pide publicarla (bajo orden explícita), nunca se cae en
silencio a la default:

```bash
# 0. Base del Work Item + verificación en origin (obligatoria antes del PR)
base=<base del Work Item>       # links.json.base → Base: de la issue → fallback confirmado
test -n "$(git ls-remote --heads origin "$base")" \
  || { echo "La base '$base' no existe en origin: git push origin '$base'"; }
```

```bash
# La orden de PR autoriza commitear SOLO los archivos del Work Item (nunca
# `git add -A` ciego) y publicar SOLO esta rama.
git add <archivos del Work Item>
git commit -m "<título>"
git push -u origin feat/<n>-<slug>

gh pr create --repo "$canonical" --base "$base" --head feat/<n>-<slug> \
  --title "<título>" --body-file /tmp/rei-pr-<id>.md

gh pr view <pr> --repo "$canonical" --json number,url,state,mergeable
```

El cuerpo y el cierre dependen del **tipo de base confirmado en la puerta**
(§5), nunca de la mera existencia de `dev`:

- **default** (`<base>` == default): cuerpo con `Closes #<n>`; la issue se
  autocierra al mergear.
- **integración** (`<base>` no default, p. ej. `dev`, confirmada): **sin
  `Closes`**; tras el merge se verifica y se cierra manualmente.
- **apilada** (`<base>` es `feat/*` o confirmada como tal): cuerpo con
  `Refs #<n>`, sin cierre; «cierre pendiente al llegar a `<integración>`».

**Limitación v1 — PR desde un fork:** la integración opera sobre **un único
repositorio** (el canónico configurado). Si el destino es el upstream pero la
rama vive en el fork, el flujo anterior no sirve: es un caso **manual**, con
`gh pr create --repo <upstream> --base "$base" --head <fork-owner>:<branch> ...`.
No hay flags `--repo`/`--base` en `rei`: la decisión y el comando los ejecuta el
Leader bajo orden explícita.

Merge solo con orden explícita:

```bash
gh pr merge <pr> --repo "$canonical" --squash --delete-branch \
  --subject "<título> (#<n>)"
```

- **Nunca** `--admin` ni auto-merge.
- Si branch protection, checks requeridos o permisos bloquean el merge: informa
  de la causa y del siguiente paso, deja el PR abierto y **no** reintentes en
  bucle ni fuerces el merge.
- Después del merge, verifica **siempre** la issue:

  ```bash
  gh issue view <n> --repo "$canonical" --json state,url
  ```

  - tipo **default**: si sigue abierta, ciérrala bajo la misma orden y verifica
    `CLOSED`; si ya la cerró `Closes #N`, informa de ello.
  - tipo **integración**: si sigue abierta, ciérrala manualmente y verifica:

    ```bash
    gh issue close <n> --repo "$canonical" --reason completed
    gh issue view <n> --repo "$canonical" --json state     # debe ser CLOSED
    ```

  - tipo **apilada**: **no** la cierres; reporta «cierre pendiente al llegar a
    `<integración>`».

  Si la verificación o el cierre falla, informa de qué quedó pendiente.

Cualquier otro commit o push distinto de la rama del Work Item requiere una
orden propia.
