# Integración opcional con GitHub Issues (v2)

> **Playbook del agente.** GitHub es una **proyección** del trabajo, no su
> fuente de verdad. `.rei` manda sobre la ejecución del Work Item; la issue manda
> sobre el problema, el alcance y la discusión. Toda la integración es
> **opcional**: si está desactivada o no configurada, el flujo es idéntico al que
> no usa GitHub.
>
> La capa determinista es el comando `rei github`. Las acciones reales sobre
> GitHub las ejecuta el **Leader** con `gh` y `git`, siempre bajo las reglas de
> este documento: nunca commit, push, PR ni merge sin una orden explícita del
> usuario. La única excepción es la prueba de campo de un Work Item, que puede
> ejecutar el Implementer con la autorización explícita del usuario registrada
> por el Leader (ver §13).
>
> **El CLI nunca llama a GitHub.** No crea, lee ni modifica issues, labels,
> assignees, comentarios, PR ni épicas. No hay subcomandos, flags ni dependencias
> Go nuevas por esta integración.

---

# 1. Alcance v1 + v2

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

Incluido en v2 (construye sobre v1 sin alterarlo):

- **épicas**: issue que planifica un módulo, no un Work Item;
- **backlog de hijas**: issues completas pendientes de adopción;
- **adopción y promoción** de una hija a Work Item, con publicación de la spec;
- **base por Work Item en la puerta**: captura, registro en `links.json` y en la
  issue, y uso para la rama y el PR; `integrations.github.base` es solo
  *fallback*;
- **asignación y desasignación** de personas con `gh issue edit`;
- **progreso y cierre** de la épica bajo orden explícita;
- tabla de errores, aislamiento y reintentos ampliados.

Fuera de alcance: Projects, milestones, sub-issues nativas, PRs desde
fork→upstream, sincronización bidireccional, Work Items `pending` locales como
sustituto del backlog, autoasignación por carga, time tracking, proveedores
distintos de GitHub, commits/push o merge automáticos, `--admin`/auto-merge y
flags nuevos en `rei`.

**Alcance por tipo de Work Item.** La puerta «issue + rama» es **solo para
Work Items de tipo `feature`**. Una `task` **no** genera issue ni rama por
defecto: se trabaja en la rama base sin tocar GitHub. Solo si el usuario pide
explícitamente una issue y una rama para una `task` (opt-in) se aplica el mismo
flujo, incluida la captura de la base de §4.

**Regla dura:** si `integrations.github` no existe o `enabled` es `false`, el
Leader **no menciona GitHub**, no ejecuta `gh`, no lee ni crea issues y trabaja
sobre la rama base como siempre.

---

# 2. Configuración y diagnóstico

La configuración vive en `.rei/config.json`, bajo `integrations.github`:

```json
{
  "integrations": {
    "github": {
      "enabled": true,
      "repo": "owner/name",
      "base": "main",
      "mergeStrategy": "squash"
    }
  }
}
```

- **Clave ausente** ⇒ integración **desactivada**. El Leader no menciona GitHub
  ni ejecuta `gh`.
- `enabled: false` ⇒ integración **desactivada**; se conservan `repo`, `base`,
  `mergeStrategy` y `links.json`.
- `base` es solo el **fallback** de la base por Work Item: se usa únicamente si
  el Work Item no tiene base registrada (`links.json.base` ni la línea `Base:`
  de la issue) y no se puede preguntar al usuario. La base efectiva es **por
  Work Item** y se decide en la puerta (§4). Si `base` está vacío ⇒ `main`.
  `mergeStrategy` vacío ⇒ `squash`.

La base por Work Item **no** se escribe en `.rei/config.json` ni la modifica
`rei github enable`; vive en `links.json` y en el cuerpo de la issue. `rei
github status` sigue siendo offline y **no** conoce la base efectiva de cada
Work Item.

Comandos:

```bash
rei github enable    # detecta origin, escribe repo/base/mergeStrategy
rei github disable   # enabled: false; conserva el resto
rei github status    # solo lectura: estado, valores, remoto, gh y siguiente paso
```

`enable` detecta el repositorio con `git remote get-url origin` (SSH o HTTPS de
`github.com`) y la rama base (*fallback*) con `git symbolic-ref --quiet --short
refs/remotes/origin/HEAD`; si no puede, usa `main` y lo informa. Ese valor no
decide la base de cada Work Item (§4). No consulta la API de GitHub. Los avisos
de `gh` ausente o sin autenticar **no bloquean**.

`status` es **offline y de solo lectura**: no llama a `gh` ni a la API de GitHub.
Además de estado, valores, remoto y disponibilidad de `gh`, **avisa si el remoto
de `origin` no coincide con `integrations.github.repo`** (rename, clon viejo o
fork) y propone el siguiente paso: resolver el nombre canónico con `gh repo view`
y ponerlo en la config, o actualizar `origin`. Con esa discrepancia **no**
concluye «Sin acciones pendientes».

---

# 3. Vocabulario

## 3.1 Marcadores

| Elemento | Valor |
|----------|-------|
| Issue de Work Item | `<!-- rei:work-item:<id> -->` |
| Épica | `<!-- rei:epic:<epic-slug> -->` |
| Hija de backlog | `<!-- rei:backlog:<epic-slug>/<item-slug> -->` |
| Comentario de implementación | `<!-- rei:comment:impl:<id> -->` |
| Comentario de revisión | `<!-- rei:comment:review:<id> -->` |
| Delimitador de la task list de la épica | `<!-- rei:children -->` … `<!-- /rei:children -->` |
| Delimitador de la spec publicada | `<!-- rei:spec:<id> -->` … `<!-- /rei:spec:<id> -->` |

Reglas:

- Los marcadores de identidad (`work-item`, `epic`, `backlog`) van en la
  **primera línea** del cuerpo.
- Los delimitadores `children` y `spec` son **auxiliares**: permiten reemplazar
  un bloque sin destruir el resto del cuerpo. No identifican issues.
- **El marcador identifica; la label es solo una pista humana.** Ningún paso
  depende de la label. La ausencia de marcador **no** se suple con labels: si
  hay duda, se pregunta.
- La búsqueda de idempotencia usa el **marcador exacto**, nunca la label ni el
  título.
- Un marcador de identidad debe aparecer **una sola vez** en el cuerpo. Si
  aparece más de una vez **o hay más de una issue con el mismo marcador**, se
  detiene la acción y se reporta la anomalía. Nunca se crea ni se sobrescribe
  nada en ese estado.
- Cambiar el título de una issue **no** cambia su slug ni su marcador.

## 3.2 Labels

| Label | Color | Descripción |
|-------|-------|-------------|
| `rei` | `1D76DB` | Gestionado por REI |
| `rei:epic` | `6F42C1` | Épica: planificación de módulo (no Work Item) |
| `rei:backlog` | `FBCA04` | Hija de backlog pendiente de adopción |
| `rei:feature` | `0E8A16` | Work Item de tipo feature |
| `rei:blocked` | `B60205` | Bloqueado |

Se aseguran con `gh label create --force` antes de usarlas (idempotente: crea o
actualiza color y descripción). Composición:

- **épica**: `rei` + `rei:epic`;
- **hija de backlog**: `rei` + `rei:backlog`;
- **hija adoptada**: conserva `rei`, recibe `rei:feature` y pierde
  `rei:backlog`;
- **Work Item normal**: `rei` + `rei:feature` (v1);
- **bloqueado**: añade `rei:blocked` sin comentario (v1).

## 3.3 `links.json`

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
- **Ausente** ⇒ Work Item v1: la base se resuelve al actuar (§4 y §5) con el
  orden `links.json.base` → línea `Base:` de la issue →
  `integrations.github.base` (o `main` si está vacío) → preguntar al usuario.
  Nunca se infiere de `origin/HEAD` ni de la rama actual.
- **Corrupto o incompleto** (falta `branch`, `sha` inválido, JSON roto) ⇒ se
  trata como ausente, se informa y se recupera; **nunca** bloquea el Work Item.
- Si `links.json.base` y la línea `Base:` de la issue existen y **discrepan**, se
  pregunta cuál vale; no se elige en silencio.

La base así resuelta se guarda en `links.json` y en `current.md`, pero **no** se
reescribe el cuerpo de issues ya creadas para añadirles la línea `Base:`.

## 3.4 Nombre canónico del repositorio

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

## 3.5 Rename y fork

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

# 4. Puerta única al aprobar la planificación de una feature (v1 + base por Work Item)

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
   falta `base`, resuélvela (§3.3) y añádela sin repreguntar.
4. Si `current.md` ya registra la decisión de este Work Item (base, sha y tipo),
   no vuelvas a preguntar.
5. Captura la base **antes** de formular la pregunta única:

```bash
current=$(git rev-parse --abbrev-ref HEAD)   # "HEAD" ⇒ detached
sha=$(git rev-parse --short=12 HEAD)
default=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD | sed 's#^origin/##')
```

Clasificación de la base y su **tipo** (el tipo decide el cierre en §11):

| Situación | Base elegida | Acción |
|-----------|--------------|--------|
| `HEAD` = default | `HEAD` | Sin aviso; la pregunta única muestra la base (tipo **default**). |
| `HEAD` ≠ default, no es `feat/*` | `HEAD` | Mostrar «base = `<base>` (no es la default `<default>`)» y **confirmar en la misma pregunta** que es la rama de integración (tipo **integración**, p. ej. `dev`). |
| `HEAD` = `feat/*` | `HEAD` | Aviso de PR apilada: el PR irá contra `<base>`, no contra la default; exigir confirmación explícita (tipo **apilada**). |
| default no resoluble | — | Preguntar cuál es la default y la rama de integración; no asumir `main` ni deducir el tipo en silencio. |
| detached (`HEAD`) | `integrations.github.base` o `main` | Ofrecer el fallback; si no hay respuesta, usarlo y decirlo en la Bitácora. Nunca adivinar la rama original. |
| `links.json.base` ya existe | la registrada | Mostrarla; si `HEAD` difiere, preguntar si se conserva o se actualiza. Nunca cambiarla en silencio. |

Si `links.json.base` y la línea `Base:` de la issue discrepan, pregunta; nunca
elijas en silencio (§12).

6. Formula **una sola** pregunta, con la base incorporada:

   > «¿Creo la issue y la rama `feat/<n>-<slug>` desde `<base>` (@`<sha>`)?»
   > (el `<n>` será el número de la issue y `<slug>` el slug del Work Item).

   La confirmación explícita es **obligatoria** si `<base>` no es la default o si
   es una rama `feat/*` (aviso de PR apilada). Si el usuario no confirma, no se
   crea nada.

Nunca repitas la pregunta en la misma sesión. Registra en `current.md`
(Bitácora): base, sha, **tipo de base** (default, integración o apilada) y, si
aplica, el fallback usado por detached.

## 4.1 Camino «sí»

Orden exacto. Sustituye `<id>`, `<repo>`, `<título>`, `<base>`, `<sha>` y `<n>`
por sus valores.

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
existe, escribe el cuerpo en un archivo temporal y crea:

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
con causa y siguiente paso (§12). Nunca uses `-f` ni recrees una rama existente;
si `feat/<n>-<slug>` ya existe, informa y ofrece reutilizarla o elegir otra.
7. Reporta issue, base, sha, rama y `links.json` («rama creada desde `<base>`
   (@`<sha>`)» aunque `HEAD` estuviera en otra rama).

Si la issue se creó pero falla la rama o `links.json`, informa de **qué quedó
hecho y qué está pendiente**, conserva la issue y permite reintentar sin
duplicarla.

## 4.2 Plantilla del cuerpo de la issue

El marcador va en la **primera línea**. El título es `meta.json.title`. El
cuerpo debe ser autosuficiente: alguien sin acceso al repo debe entender el
problema, la decisión y el alcance. La línea `Base:` (`<rama>` es la base elegida
en la puerta y `<sha>` su sha corto) solo se escribe en issues **nuevas**; nunca
se añade a issues ya creadas.

```md
<!-- rei:work-item:<id> -->

Base: <rama> @ <sha>

## Contexto y decisión

<description de meta.json: problema, decisión y acuerdo>

## Requisitos (resumen)

- **R1** — ...
- **R2** — ...

## Decisiones de diseño relevantes

- ...

## Fuera de alcance

- ...

## Referencias

- Work Item: `.rei/specs/<id>/`
- Requisitos: `.rei/specs/<id>/requirements.md`
- Tareas: `.rei/specs/<id>/tasks.md`

<details>
<summary>design.md completo</summary>

<contenido íntegro de design.md>

</details>
```

## 4.3 Camino «no»

- No se crea issue ni rama.
- Se trabaja en `<base>` si es una rama local; si `HEAD` está detached, en el
  estado actual.
- Se registra en `current.md` (Bitácora):
  «GitHub: issue/rama declinadas; se trabaja en `<base>`».
- No se vuelve a preguntar.

---

# 5. Adopción y promoción

## 5.1 Lectura y conflictos

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
5. **Marcador de backlog** ⇒ se promueve (§5.2).
6. **Sin marcador** ⇒ se aplica la adopción genérica de v1 (§5.4).

La label `rei:feature` sin marcador **no** prueba que la issue ya esté adoptada;
la identidad es el marcador.

## 5.2 Promoción de una hija de backlog

1. Preguntar solo lo que falte para definir el Work Item (objetivo, alcance,
   fuera de alcance, criterios de aceptación y restricciones), usando el cuerpo y
   los comentarios como contexto. **No inventar** información.
2. Resolver la épica por su marcador `<!-- rei:epic:<epic-slug> -->` antes de
   crear el Work Item. Si no existe o hay duplicados, no se inicia la adopción y
   se informa del conflicto.
3. Crear el Work Item con `rei new <id> feature` y rellenar `title` y
   `description` con el acuerdo alcanzado, la referencia a la épica y la URL de
   la issue.
4. Sustituir la línea del marcador de backlog por
   `<!-- rei:work-item:<id> -->`, conservando el resto del cuerpo:

   ```bash
   gh issue view <n> --repo "$canonical" --json body --jq .body \
     > /tmp/rei-issue-<n>-body.md
   # sustituir la línea exacta del marcador de backlog por el de Work Item
   gh issue edit <n> --repo "$canonical" --body-file /tmp/rei-issue-<n>-body.md
   ```

   Verificar que el marcador de backlog ya no aparece y que el de Work Item
   aparece **una sola vez**.
5. Ajustar labels y verificar:

   ```bash
   gh label create 'rei:feature' --color 0E8A16 \
     --description "Work Item de tipo feature" --force --repo "$canonical"
   gh issue edit <n> --repo "$canonical" \
     --add-label rei --add-label 'rei:feature' --remove-label 'rei:backlog'
   gh issue view <n> --repo "$canonical" --json labels --jq '[.labels[].name]'
   ```

   Debe conservar `rei`, tener `rei:feature` y no tener `rei:backlog`.
6. Resolver la base del Work Item **antes** de escribir `links.json` o crear la
   rama, en este orden: `links.json.base` → línea `Base: <rama> @ <sha>` del
   cuerpo de la issue → preguntar al usuario → `integrations.github.base`
   (fallback; si está vacío, `main`). Si la issue no declara base y el usuario no
   responde, aplica el fallback y decláralo; **nunca** inventes la rama de origen.
   Si `links.json.base` y la línea `Base:` discrepan, pregunta. Si el **tipo** de
   base no se deduce (no es la default ni `feat/*`), confírmalo con el usuario
   antes de crear la rama o el PR; nunca asumas que es integración. Si la base
   registrada ya no existe, deriva a §12.
7. Escribir `links.json` con `provider`, `repo`, `issue`, `url`,
   `base {branch, sha}` y el objeto `epic` resuelto desde el marcador de la épica
   (slug, número y URL). Registrar base, sha y tipo en `current.md`. **No**
   reescribir el cuerpo de la issue para añadir `Base:`. Si `links.json` faltaba
   o estaba corrupto, se reescribe (no bloquea).
8. Si la hija está **sin asignar**, ofrecer asignarla a quien la trabaja y **no
   asignarla sin confirmación**. Si está asignada a otra persona, preguntar
   antes.
9. Continuar con el workflow normal: `spec_author`.

Si la adopción queda parcial (Work Item creado pero marcador, labels o
`links.json` pendientes), se informa de lo hecho y lo pendiente; el reintento
reutiliza el Work Item existente y completa solo los pasos que falten. **Nunca**
se crea otro Work Item.

## 5.3 Publicación de la spec

Se ejecuta cuando `spec_author` termina la planificación de un Work Item
adoptado de backlog (Caso A) y cuando una replanificación vuelve a `ready` (Caso
B), solo si `links.json.epic != null`.

1. Leer `requirements.md` y `design.md`.
2. Construir el bloque:

   ```md
   <!-- rei:spec:<id> -->

   ## Plan (REI)

   - Work Item: `.rei/specs/<id>/`
   - Requisitos: <resumen R1..Rn>
   - Tareas: <resumen>
   - Referencias: `.rei/specs/<id>/requirements.md`, `design.md`, `tasks.md`

   <details>
   <summary>design.md completo</summary>

   <contenido íntegro de design.md>

   </details>
   <!-- /rei:spec:<id> -->
   ```

3. Leer el cuerpo actual de la issue.
4. Si existen los delimitadores `rei:spec:<id>`, **reemplazar el bloque en
   sitio**; si no, añadirlo al final. Nunca se duplica la sección.
5. Editar con `gh issue edit --body-file` y verificar que el delimitador de
   apertura aparece **una sola vez**.

Si falla, se reporta la causa y el siguiente paso, **no** cambia el estado del
Work Item y se puede reintentar sin duplicar la sección.

## 5.4 Adopción genérica (v1, sin marcador)

1. Comprobar que la integración está activa. Si no, no leer ni modificar issues:
   ofrecer `rei github enable`.
2. Leer la issue con la orden de §5.1.
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
   `current.md`. Si la base ya no existe, deriva a §12.
6. Escribir `.rei/specs/<id>/links.json` con `base {branch, sha}` y `epic: null`
   (o el objeto `epic` si la issue declara pertenecer a una épica y esta se
   resuelve sin ambigüedad).
7. Continuar con el workflow normal (`spec_author`).

---

# 6. Épicas

## 6.1 Modelo

Una **épica** es una issue de GitHub que planifica un módulo. **No es un Work
Item**: no tiene `requirements.md`, `design.md`, `tasks.md` ni requisitos EARS.
Lleva:

- marcador `<!-- rei:epic:<epic-slug> -->` en la **primera línea** del cuerpo;
- labels `rei` y `rei:epic`.

El **slug** es kebab-case, estable y único en el repositorio. Cambiar el título
no cambia el slug ni el marcador.

## 6.2 Plantilla

```md
<!-- rei:epic:modulo-x -->

## Contexto del módulo

<problema y situación actual>

## Objetivo

<qué debe conseguirse>

## No-objetivos

- ...

## Hijas

<!-- rei:children -->
- [ ] #12 <título de la hija>
- [ ] #13 <título de la hija>
<!-- /rei:children -->

## Orden y dependencias

1. #12 → #13: <motivo>

## Criterio de éxito

<condición verificable>

<details>
<summary>Plan completo (opcional)</summary>

<contenido opcional>

</details>
```

El bloque `children` puede crearse **vacío** y rellenarse después de crear las
hijas. El plan completo es opcional y, si existe, va plegado en `<details>`. Si
el bloque se corrompe, se detiene y se pide corregirlo antes de reescribirlo.

## 6.3 Creación del backlog de un módulo

**Disparadores:** «crea el backlog de X», «planifica el módulo X», «añade una
hija a la épica #N». Con la integración desactivada no se lee ni se crea nada y
se ofrece `rei github enable` como siguiente paso.

### Entrevista y borrador

Se pregunta **solo lo que falte**, sin inventar alcance, hijas, responsables ni
criterios:

1. alcance del módulo;
2. no-objetivos;
3. features hijas (título y resultado esperado);
4. orden y dependencias;
5. responsables (solo si el usuario los menciona; no se inventan).

Con ese material se construye un borrador en memoria con: slug de épica e hijas
(kebab-case, estable y único), título, labels, marcador, cuerpo completo, orden y
dependencias.

Antes de crear nada se muestra el **borrador completo** (tabla de resumen +
cuerpo de cada issue) y se pide una **aprobación explícita**. Sin «aprobado» (o
equivalente inequívoco) no se ejecuta ningún `gh issue create`. Si el usuario
pide cambios, se corrige el borrador y se repite la aprobación. **El silencio o
una respuesta ambigua nunca son aprobación.**

### Idempotencia canónica

Al inicio de la acción se resuelve el nombre canónico (§3.4). La búsqueda de un
marcador es **siempre REST, paginada y sobre issues (no PRs)**:

```bash
gh api "repos/$canonical/issues?state=all&per_page=100" --paginate \
  --jq '.[] | select(.pull_request == null)
        | select((.body // "") | contains("<!-- rei:epic:modulo-x -->"))
        | [.number, .state, .html_url] | @tsv'
```

- 1 resultado ⇒ se reutiliza e informa número/URL.
- 0 resultados ⇒ se crea.
- \>1 resultados ⇒ anomalía: se detiene y se pide corregir los duplicados.

Incluye issues cerradas y descarta pull requests. Nunca se usa
`gh issue list --search`. Para las hijas se usa el mismo filtro con su marcador
`<!-- rei:backlog:<epic-slug>/<item-slug> -->`.

### Orden de creación y creación parcial

1. Resolver el canónico.
2. Asegurar labels con `gh label create --force`:

   ```bash
   gh label create rei --color 1D76DB --description "Gestionado por REI" --force --repo "$canonical"
   gh label create 'rei:epic' --color 6F42C1 --description "Épica: planificación de módulo (no Work Item)" --force --repo "$canonical"
   gh label create 'rei:backlog' --color FBCA04 --description "Hija de backlog pendiente de adopción" --force --repo "$canonical"
   ```

3. Buscar la épica por marcador; crearla si falta (con el bloque `children`
   vacío) y verificarla (número, URL, estado, cuerpo y labels).
4. Para cada hija, en orden: buscar por marcador; crearla si falta; verificarla.
5. Reescribir el bloque `children` de la épica con los números y títulos reales.
   Si la épica es nueva, se rellena con las hijas del borrador; si se reutiliza,
   se añaden solo las entradas que falten y se preservan las existentes.
   Verificar el cuerpo.
6. Reportar creadas, reutilizadas y pendientes.

Ante un fallo de una hija se detiene la creación de las siguientes, se actualiza
la task list con lo que sí existe y se reporta lo pendiente. El **reintento**
vuelve a buscar por marcador: reutiliza la épica y las hijas ya creadas y crea
solo las que falten. No se borra nada.

Si no se puede resolver el canónico, listar las issues o buscar el marcador, se
detiene con causa y siguiente paso y no se crea nada con un nombre sin resolver.
Si las Issues están deshabilitadas en el repositorio, se informa de la causa y
del siguiente paso (activarlas o configurar otro repositorio) y no se afirma
éxito.

---

# 7. Backlog de hijas

## 7.1 Modelo y plantilla

Una **hija** es una issue completa de GitHub, **no un Work Item** hasta que se
adopta. Lleva:

- marcador `<!-- rei:backlog:<epic-slug>/<item-slug> -->` en la **primera
  línea** del cuerpo;
- labels `rei` y `rei:backlog`.

El slug de la hija es kebab-case, estable y **único dentro de la épica**. La
identidad de la hija es su marcador, no su título.

```md
<!-- rei:backlog:modulo-x/hija-a -->

## Contexto

<por qué existe esta hija>

## Alcance

- ...

## Fuera de alcance

- ...

## Criterios de aceptación (borrador)

- [ ] ...

## Épica

#100 — <título de la épica>

## Assignee

<@login | sin asignar>

## Notas y dependencias

- ...
```

**No incluye** requisitos EARS, `design.md`, `tasks.md` ni plan de archivos. Al
adoptarse, el cuerpo se conserva y solo se sustituye el marcador y se añade la
sección de spec.

## 7.2 Creación y ampliación

La creación de hijas usa la misma entrevista, borrador, aprobación explícita,
idempotencia canónica y orden de §6.3. Cada issue creada o reutilizada se
verifica (número, URL, estado, cuerpo y labels) antes de continuar.

Para **añadir una hija a una épica existente**:

1. Verificar que la épica existe, tiene marcador de épica y está **abierta**.
2. Aplicar el mismo flujo de borrador, aprobación e idempotencia.
3. Crear la hija si falta y verificarla.
4. Añadir su entrada al bloque `children` preservando las existentes (o crear el
   bloque si no existe) y verificar el cuerpo.

**Si la épica está cerrada, no se añade la hija**: se informa de la causa y del
siguiente paso (reabrirla o elegir otra épica). Si el bloque `children` está
corrupto, se detiene y se pide corregirlo antes de reescribirlo.

---

# 8. Asignación

Disparadores: «asigna la #N a @login», «asígnanos la #12 y la #13», «quítame de
la #N». GitHub es la **fuente de verdad de las personas**; REI no modela gente ni
guarda estado local de assignees.

**Sin orden explícita o confirmación de una oferta no se ejecuta nada.** Nunca
se autoasigna por carga, por adopción ni por creación de backlog.

1. Resolver `@me`:

   ```bash
   me=$(gh api user --jq .login)
   ```

2. Validar cada login:

   ```bash
   gh api "users/<login>" --jq .login
   ```

   Un 404 ⇒ «el login `<login>` no existe» + siguiente paso. La asignabilidad
   real la confirma GitHub al editar; si rechaza por permisos o por no ser
   asignable, se reporta sin afirmar éxito.

3. Verificar que cada issue existe y es una issue (no un PR).
4. Ejecutar:

   ```bash
   # una issue
   gh issue edit <n> --repo "$canonical" --add-assignee <login>[,<login>]
   # varias issues con el MISMO conjunto de personas (lote opcional)
   gh issue edit <n1> <n2> --repo "$canonical" --add-assignee <login1>,<login2>
   # desasignar
   gh issue edit <n> --repo "$canonical" --remove-assignee <login>
   ```

   El lote solo se usa cuando el mismo conjunto de personas aplica a todas las
   issues; si no, se ejecuta una llamada por issue.

5. Verificar los assignees reales **antes** de informar de éxito:

   ```bash
   gh issue view <n> --repo "$canonical" \
     --json assignees --jq '[.assignees[].login]'
   ```

   Solo se informa de éxito si el login aparece (asignar) o no aparece
   (desasignar). Si ya estaba en el estado pedido, se informa «sin cambios».

Ante un fallo total o parcial se informan las issues afectadas, las que quedaron
sin cambiar y el siguiente paso, sin ocultar el fallo.

---

# 9. Progreso y cierre de la épica

## 9.1 Task list

Formato: `- [ ] #N <título>`. Es la lista de hijas de la épica y el índice que
usa el cierre. Las **casillas son informativas**: el estado real se consulta
siempre a GitHub.

Se actualiza solo:

- al crear o reanudar un backlog parcial;
- al añadir una hija;
- bajo una orden explícita de reconciliación o cierre.

No hay sincronización en vivo ni comentarios por cambios de estado de las hijas.

## 9.2 Cierre

Disparador: «cierra la épica #N». Solo aplica a issues con **marcador de épica**.
La épica **nunca** se cierra automáticamente al cerrarse la última hija, y este
protocolo **nunca** cierra hijas: su cierre sigue su propio flujo.

1. Leer la épica y verificar que está abierta y que su marcador no es ambiguo.
2. Extraer del bloque `children` (o de la task list si el bloque no existe) los
   números `#N`. Si no hay ninguno o no se pueden interpretar, detenerse.
3. Consultar el estado real de cada hija:

   ```bash
   gh api "repos/$canonical/issues/<n>" \
     --jq '[.number, (.state | ascii_upcase), (.pull_request != null)] | @tsv'
   ```

   La REST devuelve `state` en minúsculas (`open`/`closed`); se normaliza a
   `OPEN`/`CLOSED` para comparar sin ambigüedad.

   - `OPEN` ⇒ no cerrar; listar las hijas abiertas.
   - 404 ⇒ no cerrar; reportar la hija inexistente.
   - `pull_request != null` ⇒ no cerrar; reportar que la entrada no es una issue.

4. Si todas están cerradas, reconciliar la task list a `[x]` y verificar el
   cuerpo. Si la escritura falla, **no cerrar**.
5. Cerrar la épica bajo orden:

   ```bash
   gh issue close <n> --repo "$canonical" --reason completed
   gh issue view <n> --repo "$canonical" --json state --jq .state  # CLOSED
   ```

6. Si el cierre falla, reportar lo hecho y lo pendiente; la task list ya
   reconciliada es recuperable.

Nunca se usa la casilla `[x]` como prueba de estado.

---

# 10. Comentarios automáticos (v1)

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

Plantilla de implementación:

```md
<!-- rei:comment:impl:<id> -->

## Implementación terminada

- Work Item: `<id>`
- Estado: `review`
- Resumen: ...
- Archivos principales: ...
- Verificación: ...
- Detalle: `.rei/progress/work-items/<id>/impl.md`
```

Plantilla de revisión (mismo procedimiento con
`<!-- rei:comment:review:<id> -->`):

```md
<!-- rei:comment:review:<id> -->

## Revisión: <done|changes_requested|blocked>

- Work Item: `<id>`
- Verificaciones: ...
- Observaciones: ...
- Detalle: `.rei/progress/work-items/<id>/review.md`
```

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

# 11. Pull request y merge (v1 + base por Work Item)

Solo **tras la revisión** y bajo **orden explícita** del usuario.

La base del PR es la **base del Work Item**, resuelta en la puerta y registrada
(§3.3/§4), no `integrations.github.base` si el Work Item ya tiene base. Antes de
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
(§4), nunca de la mera existencia de `dev`:

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

---

# 12. Errores y reintentos

Todo fallo de `gh`, `git` o configuración se comunica con **causa** y **siguiente
paso**. Nunca se silencia ni se afirma un éxito sin verificarlo.

| Situación | Detección | Mensaje (causa) | Siguiente paso |
|-----------|-----------|-----------------|----------------|
| Sin `.rei` | `rei github` / Leader | «REI no está inicializado aquí» | `rei init` |
| Integración no configurada | `GitHub == nil` | «GitHub no está configurado» | `rei github enable` |
| Integración desactivada | `enabled: false` | «GitHub está desactivado» | `rei github enable` si se desea |
| Config incompleta | `repo` vacío o inválido | «Falta `integrations.github.repo`» | `rei github enable` o editar la config |
| `gh` ausente | `LookPath` falla | «GitHub CLI no está instalado» | Instalar `gh` |
| `gh` sin auth | `gh auth status` falla | «GitHub CLI no está autenticado» | `gh auth login` |
| Repo no detectado / ambiguo | `git remote get-url origin` falla o host inválido | «No se pudo detectar el repositorio» | Revisar `git remote -v` / configurar `repo` |
| Remoto ≠ nombre canónico | `git remote get-url origin` frente a `integrations.github.repo` | «El remoto no coincide con `integrations.github.repo`» | Resolver el canónico con `gh repo view <repo> --json nameWithOwner --jq .nameWithOwner` y ponerlo en `integrations.github.repo`, o actualizar el remoto |
| Fallo al resolver el canónico | `gh repo view` falla (red/auth/permisos/404) | «No se pudo resolver el nombre canónico de `<repo>`» | Revisar `gh auth status` y la red; **no** continuar con el nombre viejo |
| Issues deshabilitados | 404/410 al crear o listar la issue | «Las issues están deshabilitadas en `<repo>`» | Activar Issues en Settings del repositorio (o del fork) o configurar otro repositorio |
| Épica cerrada al añadir hija | estado `CLOSED` | «La épica #N está cerrada» | Reabrirla o elegir otra épica |
| Épica/hija inexistente | 404 en `gh issue view`/REST | «La issue #N no existe en `<repo>`» | Verificar número y repositorio |
| Hija ya adoptada | marcador `rei:work-item:` | «La issue #N ya pertenece al Work Item `<id>`» | Trabajar en ese Work Item o elegir otra hija |
| Marcador duplicado | >1 issue o >1 marcador | «Hay más de una issue con el marcador `<marcador>`» | Corregir los duplicados antes de continuar |
| Assignee inexistente | 404 en `users/<login>` | «El login `<login>` no existe» | Verificar el login |
| Assignee no asignable / sin permisos | 403/422 de GitHub | «No se pudo asignar `<login>` a #N» | Revisar acceso del usuario y `gh auth status` |
| Task list vacía o ilegible | parseo del bloque | «No se pudo determinar las hijas de la épica #N» | Corregir la task list |
| Hijas abiertas al cerrar | estado `OPEN` | «La épica #N tiene hijas abiertas: ...» | Cerrar las hijas primero |
| Creación parcial | contador creadas/pendientes | «Backlog parcial: creadas ..., pendientes ...» | Reintentar; los marcadores evitan duplicados |
| Fallo al publicar la spec | error de `gh issue edit` | «No se pudo publicar la spec en #N» | Reintentar sin duplicar la sección |
| `HEAD` detached en la puerta | `git rev-parse --abbrev-ref HEAD` = `HEAD` | «`HEAD` está detached: no hay rama actual» | Indicar la base o aceptar el fallback `integrations.github.base` |
| Base no registrada al adoptar | sin `links.json.base` ni línea `Base:` | «El Work Item no tiene base registrada» | Preguntar la rama y guardarla en `links.json` |
| Base solo local | `git ls-remote --heads origin <base>` vacío y rama local existente | «La base `<base>` no existe en `origin`» | `git push origin <base>` y reintentar el PR |
| Base borrada o renombrada | no existe en local ni en `origin` | «La base `<base>` del Work Item ya no existe» | Confirmar el nombre nuevo y actualizar `links.json`, o recrear/publicar la rama |
| Base contradictoria | `links.json.base` ≠ línea `Base:` | «La base registrada difiere de la issue» | Confirmar cuál vale y corregir `links.json` |
| Tipo de base ambiguo | base ≠ default, no es `feat/*` y no hay confirmación | «No está claro si `<base>` es la rama de integración» | Confirmar el tipo en la puerta (default, integración o apilada); nunca asumirlo |
| PR apilada sin confirmar | base `feat/*` sin confirmación | «La PR iría contra `<base>`, no contra la default `<default>`» | Confirmar la PR apilada o elegir otra base |
| Rama existente | `git switch -c` falla | «La rama ya existe» | Reutilizarla o elegir otra; nunca forzar |
| PR inexistente | `gh pr view` 404 | «No hay PR para esa rama» | Crear el PR solo si hay orden |
| Permisos insuficientes | 403/404 de API | «Sin permisos sobre `<repo>`» | Revisar acceso y `gh auth status` |
| Branch protection | merge rechazado | «El merge está bloqueado por protección de rama/checks» | Completar checks/reviews; nunca `--admin` |
| Red | error de conexión | «No se pudo contactar con GitHub» | Reintentar cuando haya red |
| Rate limit | mensaje de límite | «Límite de GitHub alcanzado» | Esperar al reset indicado; no reintentar en bucle |

Reglas transversales:

- **Causa + siguiente paso** en todo mensaje; nunca éxito sin verificar número,
  URL, estado, labels, assignees o marcador, según corresponda.
- **Aislamiento:** un fallo de GitHub **no** cambia `meta.json`, el estado de
  `current.md` ni la etapa del workflow. Se informa de lo hecho y de lo pendiente.
- **Reintentos idempotentes:** por marcador y por delimitador de sección. No se
  duplican issues, labels, marcadores, Work Items, comentarios ni bloques de
  spec.
- **Rate limit / red:** no reintentar en bucle; mostrar el reset si `gh` lo
  proporciona.
- **Lote parcial:** indicar issue por issue qué quedó hecho y qué no.

---

# 13. Recordatorios en `current.md`

Mientras haya sesión activa, el Leader lista en «Próximo paso» toda acción
GitHub pendiente y retira la línea cuando se completa. Con la integración
desactivada no se escribe ninguna línea.

v1:

- `GitHub: pendiente crear rama feat/<n>-<slug> (issue #<n> ya creada).`
- `GitHub: pendiente comentario de implementación (reintentar).`
- `GitHub: pendiente comentario de revisión (reintentar).`
- `GitHub: pendiente PR de la rama feat/<n>-<slug>.`
- `GitHub: pendiente merge del PR #<pr>.`
- `GitHub: pendiente publicar la base <base> en origin antes del PR.`
- `GitHub: pendiente actualizar la base del Work Item (<base> no existe).`
- `GitHub: pendiente cierre de la issue #<n> al llegar a <integración> (PR apilada).`

v2:

- `GitHub: pendiente crear el backlog de <módulo> (borrador pendiente de aprobación).`
- `GitHub: pendiente crear épica #<n> / hijas #<n>, #<n> (creación parcial).`
- `GitHub: pendiente actualizar la task list de la épica #<n>.`
- `GitHub: pendiente adoptar la hija #<n> (marcador/labels/links.json).`
- `GitHub: pendiente publicar la spec en la hija #<n>.`
- `GitHub: pendiente asignar/desasignar la #<n> a @<login>.`
- `GitHub: pendiente cerrar la épica #<n> (hijas pendientes: #<n>, #<n>).`
- `GitHub: pendiente cerrar la épica #<n> bajo orden explícita.`

## 13.1 Prueba de campo

La prueba de campo de un Work Item es la **única excepción** a «las acciones
GitHub las ejecuta el Leader»: la ejecuta el Implementer como parte de la
verificación del Work Item, con la **autorización explícita del usuario**
registrada previamente por el Leader en `current.md` (Bitácora). Las operaciones
normales de backlog, adopción, asignación y cierre siguen siendo del Leader.

- El Leader registra la autorización y el alcance exacto de la prueba antes de
  lanzar al Implementer.
- El Implementer sigue este playbook y deja números, URLs, comandos y resultados
  en `.rei/progress/work-items/<id>/impl.md`.
- Si la planificación no incluye prueba de campo, el Implementer no ejecuta
  ninguna acción GitHub.
- La prueba nunca justifica escrituras no aprobadas ni cambia el flujo normal.

Tras `done`, el Reviewer archiva la sesión; las acciones pendientes (comentario
de review, PR o merge) se reportan en la respuesta del Leader, sin recrear
`current.md`.
