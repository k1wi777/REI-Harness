# Integración opcional con GitHub Issues (v1)

> **Playbook del agente.** GitHub es una **proyección** del Work Item, no su
> fuente de verdad. `.rei` manda sobre la ejecución; la issue manda sobre el
> problema y la discusión. Toda la integración es **opcional**: si está
> desactivada o no configurada, el flujo es idéntico al que no usa GitHub.
>
> La capa determinista es el comando `rei github`. Las acciones reales sobre
> GitHub las ejecuta el **Leader** con `gh` y `git`, siempre bajo las reglas de
> este documento: nunca commit, push, PR ni merge sin una orden explícita del
> usuario.

---

# 1. Alcance v1

Incluido:

- configuración y diagnóstico (`rei github enable|disable|status`);
- vínculo issue ↔ Work Item con marcador y `links.json`;
- puerta única al aprobar la planificación (issue + rama);
- comentarios automáticos de implementación y revisión;
- PR y merge bajo orden explícita;
- adopción de issues existentes;
- labels `rei`, `rei:feature`, `rei:blocked`;
- manejo de errores con causa y siguiente paso.

Fuera de alcance: épicas, backlog, asignación a personas, sub-issues,
sincronización bidireccional, proveedores distintos de GitHub, commits/push o
merge automáticos, `--admin`/auto-merge y decidir la política de versionado de
`.rei`. Las labels `rei:epic` y `rei:backlog` quedan **reservadas para v2** y no
se crean en v1.

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
- `base` vacío ⇒ `main`. `mergeStrategy` vacío ⇒ `squash`.

Comandos:

```bash
rei github enable    # detecta origin, escribe repo/base/mergeStrategy
rei github disable   # enabled: false; conserva el resto
rei github status    # solo lectura: estado, valores, remoto, gh y siguiente paso
```

`enable` detecta el repositorio con `git remote get-url origin` (SSH o HTTPS de
`github.com`) y la rama base con `git symbolic-ref --quiet --short
refs/remotes/origin/HEAD`; si no puede, usa `main` y lo informa. No consulta la
API de GitHub. Los avisos de `gh` ausente o sin autenticar **no bloquean**.

**Regla dura:** si `integrations.github` no existe o `enabled` es `false`, el
Leader **no menciona GitHub**, no ejecuta `gh` y trabaja sobre la rama base como
siempre (R20).

---

# 3. Vocabulario fijo (v1)

| Elemento | Valor |
|----------|-------|
| Marcador de issue | `<!-- rei:work-item:<id> -->` |
| Comentario de implementación | `<!-- rei:comment:impl:<id> -->` |
| Comentario de revisión | `<!-- rei:comment:review:<id> -->` |
| Labels v1 | `rei`, `rei:feature`, `rei:blocked` |
| Labels reservadas v2 | `rei:epic`, `rei:backlog` (no se crean en v1) |

El marcador es invisible y es la **fuente de verdad** del vínculo. La label es
solo una pista humana: **ningún paso depende de la label**. La idempotencia se
comprueba siempre por marcador.

`.rei/specs/<id>/links.json` es un índice local recuperable, nunca fuente de
verdad:

```json
{
  "provider": "github",
  "repo": "owner/name",
  "issue": 123,
  "url": "https://github.com/owner/name/issues/123",
  "epic": null
}
```

Si falta, está corrupto o se desincroniza, no bloquea el Work Item: se recupera
la issue buscando el marcador y se reescribe.

---

# 4. Puerta única al aprobar la planificación

**Disparador:** `meta.json` está en `ready`, el usuario aprueba y
`integrations.github.enabled` es `true`.

Antes de `rei status set <id> in_progress`:

1. Lee `.rei/config.json` y comprueba `enabled`.
2. Si ya existe `.rei/specs/<id>/links.json`, no preguntes.
3. Si `current.md` ya registra la decisión de este Work Item, no vuelvas a
   preguntar.
4. Formula **una sola** pregunta:

   > «¿Creo la issue y la rama `feat/<n>-<slug>`?» (el `<n>` será el número de
   > la issue y `<slug>` el slug del Work Item).

Nunca repitas la pregunta en la misma sesión.

## 4.1 Camino «sí»

Orden exacto. Sustituye `<id>`, `<repo>`, `<título>` y `<n>` por sus valores.

```bash
# 1. ¿Existe ya la issue? (idempotencia por marcador; NUNCA crear dos veces)
gh issue list --repo <repo> --state all --limit 10 \
  --search 'in:body "rei:work-item:<id>"' \
  --json number,url,state

# 2. Labels idempotentes (solo si falta la issue)
gh label create rei --color 1D76DB --description "Gestionado por REI" --force
gh label create 'rei:feature' --color 0E8A16 \
  --description "Work Item de tipo feature" --force
```

Si el paso 1 devuelve una issue, **reutilízala**: informa de número y URL y no
crees otra. Si no existe, escribe el cuerpo en un archivo temporal y crea:

```bash
# 3. Crear la issue (solo si no existe)
gh issue create --repo <repo> \
  --title "<título del Work Item>" \
  --body-file /tmp/rei-issue-<id>.md \
  --label rei --label rei:feature

# 4. Verificar número/URL/estado antes de afirmar éxito
gh issue view <n> --repo <repo> --json number,url,state,body
```

5. Escribe `.rei/specs/<id>/links.json` **antes** de crear la rama, para no
   perder el vínculo si la rama falla. Si ya existía, actualízalo.
6. Crea la rama solo desde la rama base:

```bash
git rev-parse --abbrev-ref HEAD     # debe ser <base>
git switch -c feat/<n>-<slug>
```

Si la rama actual no es `<base>`, **detente y pregunta**; nunca crees la rama
desde un punto arbitrario. Si `feat/<n>-<slug>` ya existe, no la fuerces ni la
recrees: informa de la causa y del siguiente paso (reutilizarla o elegir otra).
7. Reporta issue, rama y `links.json`.

Si la issue se creó pero falla la rama o `links.json`, informa de **qué quedó
hecho y qué está pendiente**, conserva la issue y permite reintentar sin
duplicarla.

## 4.2 Plantilla del cuerpo de la issue

El marcador va en la **primera línea**. El título es `meta.json.title`. El
cuerpo debe ser autosuficiente: alguien sin acceso al repo debe entender el
problema, la decisión y el alcance.

```md
<!-- rei:work-item:<id> -->

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
- Se trabaja sobre la rama base, como hoy.
- Se registra en `current.md` (Bitácora):
  «GitHub: issue/rama declinadas; se trabaja en `<base>`».
- No se vuelve a preguntar (R24).

---

# 5. Adopción de issues existentes

Cuando el usuario pida «trabajemos en la #N»:

1. Comprueba que la integración está activa. Si no, no leas ni modifiques
   issues: ofrece `rei github enable` como siguiente paso (R43).
2. Lee la issue:

   ```bash
   gh issue view <n> --repo <repo> --json number,title,body,state,labels,url
   ```

3. Si está **cerrada**, pregunta antes de continuar.
4. Busca `<!-- rei:work-item:<id> -->` en el cuerpo:
   - marcador de un Work Item existente ⇒ reutilízalo (no crees otro);
   - marcador de **otro** Work Item ⇒ detente, informa del conflicto y no
     sobrescribas el marcador (R42);
   - sin marcador ⇒ pregunta lo que falte (objetivo, alcance, criterios, fuera
     de alcance, restricciones), crea el Work Item con `rei new`, rellena
     `title`/`description` con el acuerdo y añade el marcador **sin destruir el
     contenido**:

   ```bash
   gh issue view <n> --repo <repo> --json body --jq .body > /tmp/rei-issue-<n>-body.md
   # antepone <!-- rei:work-item:<id> --> al archivo
   gh issue edit <n> --repo <repo> --body-file /tmp/rei-issue-<n>-body.md
   gh label create rei --force
   gh label create 'rei:feature' --force
   gh issue edit <n> --repo <repo> --add-label rei --add-label rei:feature
   ```

5. Escribe `.rei/specs/<id>/links.json`.
6. Continúa con el workflow normal (`spec_author`).

---

# 6. Comentarios automáticos

Disparadores: el Leader acaba de recibir al Implementer (comentario `impl`) y al
Reviewer (comentario `review`). Se publican **sin preguntar**.

Búsqueda idempotente del comentario (REST, para obtener el `id` numérico):

```bash
gh api repos/<repo>/issues/<n>/comments --paginate \
  --jq '.[] | select(.body | contains("<!-- rei:comment:impl:<id> -->")) | .id'
```

- Si devuelve un `id`: actualiza en sitio.

  ```bash
  gh api --method PATCH repos/<repo>/issues/comments/<comment-id> \
    -F body=@/tmp/rei-comment-impl-<id>.md
  ```

- Si no devuelve nada: crea.

  ```bash
  gh issue comment <n> --repo <repo> --body-file /tmp/rei-comment-impl-<id>.md
  ```

- Verifica releyendo el comentario o comprobando que el marcador aparece **una
  sola vez**. Nunca dupliques (R31).

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
- **Nunca** publiques comentarios por cambios de estado del Work Item (R32):
  los únicos comentarios automáticos son `impl` y `review`.
- Si el Work Item pasa a `blocked`, aplica `rei:blocked`; al desbloquearlo,
  retírala. Sin comentarios (R34):

  ```bash
  gh label create 'rei:blocked' --color B60205 --description "Bloqueado" --force
  gh issue edit <n> --repo <repo> --add-label 'rei:blocked'
  gh issue edit <n> --repo <repo> --remove-label 'rei:blocked'
  ```

Un fallo al publicar o actualizar un comentario se reporta; **no** cambia el
estado de REI y se puede reintentar sin duplicar (R33, R46).

---

# 7. Pull request y merge

Solo **tras la revisión** y bajo **orden explícita** del usuario (R35, R39).

```bash
# La orden de PR autoriza commitear SOLO los archivos del Work Item (nunca
# `git add -A` ciego) y publicar SOLO esta rama.
git add <archivos del Work Item>
git commit -m "<título>"
git push -u origin feat/<n>-<slug>

gh pr create --repo <repo> --base <base> --head feat/<n>-<slug> \
  --title "<título>" --body-file /tmp/rei-pr-<id>.md

gh pr view <pr> --repo <repo> --json number,url,state,mergeable
```

Cuerpo del PR: resumen, cambios, verificación y `Closes #<n>`.

Merge solo con orden explícita (R36):

```bash
gh pr merge <pr> --repo <repo> --squash --delete-branch \
  --subject "<título> (#<n>)"
```

- **Nunca** `--admin` ni auto-merge.
- Si branch protection, checks requeridos o permisos bloquean el merge:
  informa de la causa y del siguiente paso, deja el PR abierto y **no**
  reintentes en bucle ni fuerces el merge (R37).
- Después del merge, verifica la issue y ciérrala si sigue abierta (R38):

  ```bash
  gh issue view <n> --repo <repo> --json state,url
  gh issue close <n> --repo <repo> --reason completed   # solo si sigue OPEN
  gh issue view <n> --repo <repo> --json state          # verificar CLOSED
  ```

  Si ya estaba cerrada por `Closes #N`, informa de ello. Si la verificación o el
  cierre falla, informa de qué quedó pendiente.

Cualquier otro commit o push distinto de la rama del Work Item requiere una
orden propia.

---

# 8. Errores y reintentos

Todo fallo de `gh`, `git` o configuración se comunica con **causa** y **siguiente
paso**. Nunca se silencia ni se afirma un éxito sin verificarlo (R44, R47).

| Situación | Detección | Mensaje (causa) | Siguiente paso |
|-----------|-----------|-----------------|----------------|
| Sin `.rei` | `rei github` / Leader | «REI no está inicializado aquí» | `rei init` |
| Integración no configurada | `GitHub == nil` | «GitHub no está configurado» | `rei github enable` |
| Integración desactivada | `enabled: false` | «GitHub está desactivado» | `rei github enable` si se desea |
| Config incompleta | `repo` vacío o inválido | «Falta `integrations.github.repo`» | `rei github enable` o editar la config |
| `gh` ausente | `LookPath` falla | «GitHub CLI no está instalado» | Instalar `gh` |
| `gh` sin auth | `gh auth status` falla | «GitHub CLI no está autenticado» | `gh auth login` |
| Repo no detectado / ambiguo | `git remote get-url origin` falla o host inválido | «No se pudo detectar el repositorio» | Revisar `git remote -v` / configurar `repo` |
| Issue inexistente | `gh issue view` 404 | «La issue #N no existe en `<repo>`» | Verificar número/repo |
| Rama existente | `git switch -c` falla | «La rama ya existe» | Reutilizarla o elegir otra; nunca forzar |
| PR inexistente | `gh pr view` 404 | «No hay PR para esa rama» | Crear el PR solo si hay orden |
| Permisos insuficientes | 403/404 de API | «Sin permisos sobre `<repo>`» | Revisar acceso y `gh auth status` |
| Branch protection | merge rechazado | «El merge está bloqueado por protección de rama/checks» | Completar checks/reviews; nunca `--admin` |
| Red | error de conexión | «No se pudo contactar con GitHub» | Reintentar cuando haya red |
| Rate limit | mensaje de límite | «Límite de GitHub alcanzado» | Esperar al reset indicado; no reintentar en bucle |

Reglas transversales:

- **Aislamiento:** un fallo de GitHub **no** cambia `meta.json`, el estado de
  `current.md` ni la etapa del workflow. Se informa de lo hecho y de lo
  pendiente, y se reintenta sin duplicar issue ni comentario (R46).
- **Verificación:** comprueba número y URL de issue, `id` del comentario, PR,
  estado del merge y estado de la issue antes de informar de éxito.
- **Reintentos:** idempotentes por marcador. Ante rate limit, no reintentes en
  bucle y muestra el reset si `gh` lo proporciona (R48).

---

# 9. Recordatorios en `current.md`

Mientras haya sesión activa, el Leader lista en «Próximo paso» toda acción
GitHub pendiente y retira la línea cuando se completa (R54):

- `GitHub: pendiente crear rama feat/<n>-<slug> (issue #<n> ya creada).`
- `GitHub: pendiente comentario de implementación (reintentar).`
- `GitHub: pendiente comentario de revisión (reintentar).`
- `GitHub: pendiente PR de la rama feat/<n>-<slug>.`
- `GitHub: pendiente merge del PR #<pr>.`

Con la integración desactivada no se escribe ninguna línea. Tras `done`, el
Reviewer archiva la sesión; las acciones pendientes (comentario de review, PR o
merge) se reportan en la respuesta del Leader, sin recrear `current.md`.
