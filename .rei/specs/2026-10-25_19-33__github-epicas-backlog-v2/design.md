# Design — GitHub Issues v2: épicas, backlog y asignación

> Work Item: `2026-10-25_19-33__github-epicas-backlog-v2` (feature).
>
> Este documento describe **cómo** se implementa la Feature. Las referencias
> `R<n>` apuntan a `requirements.md`.

---

# 1. Estrategia

v2 es una extensión de la **capa de agente** de la integración, no del CLI.

1. **Sin código Go nuevo.** El CLI sigue limitado a `rei github
   enable|disable|status`; no crea ni lee issues, labels, assignees, épicas ni
   comentarios. Toda acción sobre GitHub la ejecuta el agente con `gh` y `git`
   siguiendo el playbook (R3, R61).
2. **El playbook es el contrato operativo.** Se amplía
   `.rei/docs/harness/integrations/github.md` con el modelo de épica/hija, la
   creación del backlog, la adopción/promoción, la asignación, el cierre y los
   errores nuevos.
3. **El contrato del Leader incorpora los disparadores.** El Leader decide
   cuándo crear un backlog, adoptar una hija, asignar o cerrar una épica, y
   aplica las reglas no negociables. Los comandos exactos viven en el playbook.
4. **La fuente de verdad no cambia.** El marcador identifica la issue; la label
   es una pista; `links.json` es un índice local recuperable. GitHub manda sobre
   personas, estados y épicas; `.rei` manda sobre la ejecución del Work Item
   (R5, R12, R13).
5. **Sin sincronización en vivo.** La task list de la épica se actualiza solo al
   crear/reanudar el backlog, al añadir una hija o bajo una orden explícita de
   reconciliación/cierre. Nada se cierra ni se asigna automáticamente (R48,
   R52).

**Decisión de alcance:** la prueba de campo de este Work Item es la única
excepción a «las acciones GitHub las ejecuta el Leader». La ejecuta el
Implementer como parte de la verificación del Work Item, con la autorización
explícita del usuario registrada previamente por el Leader. Así la evidencia
queda en `impl.md` y el Reviewer puede verificarla sin modificar el trabajo de
otro rol. Las operaciones normales de backlog y adopción siguen siendo del
Leader (R64).

**Fuera de alcance:** Projects, milestones, sub-issues nativas, PRs desde
fork→upstream, sincronización bidireccional, Work Items `pending` locales para
el backlog, autoasignación por carga, time tracking, otros proveedores y flags
nuevos en `rei`.

**Contexto de dogfooding:** este Work Item usará además el flujo v1 completo
(issue, rama, comentarios de implementación/revisión, PR y merge bajo orden).
Ese flujo no es un requisito de v2, pero la implementación no debe impedirlo ni
cambiar su comportamiento.

**Restricciones globales:** se mantiene un solo Work Item `in_progress`; el
versionado de `.rei` es ortogonal y no se decide aquí; no hay sincronización en
vivo, autocierre ni autoasignación.

---

# 2. Archivos afectados

## Modificados

| Archivo | Cambio |
|---------|--------|
| `.rei/docs/harness/integrations/github.md` | Reescritura/extensión a v2: alcance, vocabulario, épicas, backlog, adopción/promoción, asignación, cierre, errores, recordatorios y plantillas. |
| `.rei/agents/leader.md` | Sección «Integración con GitHub (opcional)» ampliada con los disparadores v2 y las reglas de conflicto/automatismo. |
| `.rei/docs/harness/README.md` | Fila del playbook actualizada a v2. |
| `.opencode/agents/leader.md` | Regenerado por `rei init opencode`. |
| `.claude/agents/leader.md` | Regenerado por `rei init claude`. |
| `.rei/install-manifest.json` | Regenerado por los comandos anteriores. |
| `.rei/progress/work-items/2026-10-25_19-33__github-epicas-backlog-v2/impl.md` | Evidencia de implementación y de la prueba de campo. |

## No se modifican

- Código Go (`internal/**`, `cmd/**`), `go.mod`/`go.sum`, tests Go.
- `.rei/check/RequiredFiles`: el playbook ya está declarado.
- `embed_test.go`: el playbook ya está verificado como embebido.
- `.rei/config.json`: el repositorio ya tiene `integrations.github.enabled:
  true`; no se cambia el valor por defecto del esqueleto.
- Contratos de `spec_author`, `implementer`, `reviewer` e `initializer`: no
  intervienen en las acciones GitHub de v2.
- Plantillas de `current.md`, `history.md` y `meta.json`.
- `.rei/specs/**` y `.rei/progress/**` salvo este Work Item.

---

# 3. Vocabulario y modelo

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
- La búsqueda de idempotencia usa el marcador exacto, no la label.
- Un marcador de identidad debe aparecer **una sola vez** en el cuerpo. Si
  aparece más de una vez o hay más de una issue con el mismo marcador, se
  detiene la acción y se reporta la anomalía (R18, R54).

## 3.2 Labels

| Label | Color | Descripción |
|-------|-------|-------------|
| `rei` | `1D76DB` | Gestionado por REI |
| `rei:epic` | `6F42C1` | Épica: planificación de módulo (no Work Item) |
| `rei:backlog` | `FBCA04` | Hija de backlog pendiente de adopción |
| `rei:feature` | `0E8A16` | Work Item de tipo feature |
| `rei:blocked` | `B60205` | Bloqueado |

Se crean con `gh label create --force` antes de usarlas. La epic lleva `rei` y
`rei:epic`; la hija lleva `rei` y `rei:backlog`; al adoptar, la hija conserva
`rei`, recibe `rei:feature` y pierde `rei:backlog`.

## 3.3 `links.json`

```json
{
  "provider": "github",
  "repo": "owner/name",
  "issue": 123,
  "url": "https://github.com/owner/name/issues/123",
  "epic": {
    "slug": "modulo-x",
    "issue": 100,
    "url": "https://github.com/owner/name/issues/100"
  }
}
```

Sin épica, `epic` es `null` (compatible con v1). El objeto `epic` evita la
ambigüedad entre slug y número y permite recuperar la épica sin releer el
cuerpo. Sigue siendo un índice local: si falta o está corrupto, se recupera por
el marcador del cuerpo de la issue (R13, R36).

## 3.4 Plantilla de épica

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

La épica no lleva EARS, `design.md` ni `tasks.md`. El bloque `children` puede
crearse vacío y rellenarse después de crear las hijas.

## 3.5 Plantilla de hija

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

No incluye requisitos EARS, design, tasks ni plan de archivos. Al adoptarse, el
cuerpo se conserva y solo se sustituye el marcador y se añade la sección de
spec.

## 3.6 Sección de spec publicada

Se añade al final del cuerpo de una hija adoptada:

```md
<!-- rei:spec:2026-10-25_19-33__ejemplo -->

## Plan (REI)

- Work Item: `.rei/specs/<id>/`
- Requisitos: <resumen R1..Rn>
- Tareas: <resumen>
- Referencias: `.rei/specs/<id>/requirements.md`, `design.md`, `tasks.md`

<details>
<summary>design.md completo</summary>

<contenido íntegro de design.md>

</details>
<!-- /rei:spec:2026-10-25_19-33__ejemplo -->
```

Se publica solo para Work Items adoptados de backlog (`links.json.epic !=
null`), después de que `spec_author` deje la planificación en `ready`. En una
replanificación se reemplaza el bloque entre delimitadores; nunca se duplica.

---

# 4. Creación del backlog

## 4.1 Entrevista y borrador

Disparadores: «crea el backlog de X», «planifica el módulo X», «añade una hija
a la épica #N». Con la integración desactivada no se lee ni se crea nada; se
ofrece `rei github enable` (R25).

La entrevista cubre, preguntando solo lo que falte:

1. alcance del módulo;
2. no-objetivos;
3. features hijas (título y resultado esperado);
4. orden y dependencias;
5. responsables (solo si el usuario los menciona; no se inventan).

Con ese material se construye un borrador en memoria con:

- slug de épica e hijas (kebab-case, estable);
- título, labels y marcador de cada issue;
- cuerpo completo de la épica y de cada hija;
- orden y dependencias.

El borrador se muestra **completo** en el chat (tabla de resumen + cuerpo de
cada issue) y se pide una aprobación explícita. Sin «aprobado» (o equivalente
inequívoco) no se ejecuta ningún `gh issue create`. Si hay cambios, se corrige
el borrador y se vuelve a mostrar (R14–R17).

## 4.2 Idempotencia canónica

Al inicio de la acción se resuelve el nombre canónico:

```bash
canonical=$(gh repo view <repo> --json nameWithOwner --jq .nameWithOwner)
```

Si falla, se detiene con causa y siguiente paso; no se usa el nombre viejo ni el
valor de la config sin resolver (R19).

La búsqueda de un marcador es siempre REST, paginada y sobre issues (no PRs):

```bash
gh api "repos/$canonical/issues?state=all&per_page=100" --paginate \
  --jq '.[] | select(.pull_request == null)
        | select((.body // "") | contains("<!-- rei:epic:modulo-x -->"))
        | [.number, .state, .html_url] | @tsv'
```

- 1 resultado ⇒ se reutiliza e informa número/URL.
- 0 resultados ⇒ se crea.
- \>1 resultados ⇒ anomalía: se detiene y se pide corregir los duplicados.
- Nunca se usa `gh issue list --search` (el search no sigue redirects de rename
  y puede devolver vacío con exit 0; v1 ya lo descartó).

## 4.3 Orden de creación y creación parcial

1. Resolver canónico.
2. Crear/asegurar labels.
3. Buscar la épica por marcador; crearla si falta (con el bloque `children`
   vacío) y verificarla.
4. Para cada hija, en orden: buscar por marcador; crearla si falta; verificarla.
5. Reescribir el bloque `children` de la épica con los números y títulos
   reales. Si la épica es nueva, se rellena con las hijas del borrador; si se
   reutiliza, se añaden solo las entradas que falten y se preservan las
   existentes. Verificar el cuerpo.
6. Reportar creadas, reutilizadas y pendientes.

Ante un fallo de una hija se detiene la creación de las siguientes, se
actualiza la task list con lo que sí existe y se reporta lo pendiente. El
reintento vuelve a buscar por marcador: reutiliza la épica y las hijas ya
creadas y crea solo las que falten. No se borra nada (R21, R22).

La creación de la épica usa un cuerpo con el bloque `children` vacío; por eso
no hace falta inventar números. Si el bloque se corrompe, se detiene y se pide
corregirlo antes de reescribirlo.

## 4.4 Añadir una hija a una épica existente

Mismo flujo de borrador, aprobación e idempotencia. Antes de crear se verifica
que la épica existe, tiene marcador de épica y está abierta. Si está cerrada, se
detiene y se ofrece reabrirla o elegir otra épica. Al terminar, se añade la
nueva entrada al bloque `children` preservando las existentes (o se crea el
bloque si no existe); si el bloque está corrupto, se detiene y se pide
corregirlo antes de reescribirlo (R23).

---

# 5. Adopción y promoción

## 5.1 Lectura y conflictos

Con la integración activa, «trabajemos en la #N» ejecuta:

```bash
gh issue view <n> --repo "$canonical" \
  --json number,title,body,state,labels,assignees,comments,url
```

Orden de comprobaciones:

1. **Épica** (`rei:epic`) ⇒ no se adopta; se explica y se ofrecen sus hijas.
2. **Marcador de Work Item** (`rei:work-item:<id>`) ⇒ no se crea otro Work
   Item ni se sobrescribe el marcador. Si el Work Item existe en este clon, se
   reutiliza y se informa; si no existe (otro clon), se informa del conflicto y
   se detiene.
3. **Cerrada** (la issue o su épica) ⇒ se pregunta antes de continuar.
4. **Asignada a otra persona** ⇒ se pregunta antes de adoptar o reasignar.
5. **Marcador de backlog** ⇒ se promueve.
6. **Sin marcador** ⇒ se aplica la adopción genérica de v1.

La label `rei:feature` sin marcador no prueba que la issue ya esté adoptada; la
identidad es el marcador (R26–R30, R38).

## 5.2 Promoción

1. Preguntar lo que falte (objetivo, alcance, fuera de alcance, criterios,
   restricciones), usando cuerpo y comentarios como contexto.
2. Resolver la épica por su marcador `<!-- rei:epic:<epic-slug> -->`. Si no
   existe o hay duplicados, no se inicia la adopción.
3. Crear el Work Item con `rei new <id> feature` y rellenar `title` y
   `description` con el acuerdo, la referencia a la épica y la URL.
4. Sustituir la línea del marcador de backlog por
   `<!-- rei:work-item:<id> -->`, conservando el resto del cuerpo:

   ```bash
   gh issue view <n> --repo "$canonical" --json body --jq .body \
     > /tmp/rei-issue-<n>-body.md
   # sustituir la línea exacta del marcador de backlog por el de Work Item
   gh issue edit <n> --repo "$canonical" --body-file /tmp/rei-issue-<n>-body.md
   ```

   Verificar que el marcador de backlog ya no aparece y que el de Work Item
   aparece una sola vez.
5. Ajustar labels: `--add-label rei --add-label rei:feature --remove-label
   rei:backlog`; verificar.
6. Escribir `links.json` con el objeto `epic` (R32–R36).
7. Si la hija está sin asignar, ofrecer asignarla a quien la trabaja; nunca
   asignar sin confirmación (R37).
8. Continuar con el workflow normal: `spec_author`.

Si la adopción queda parcial (Work Item creado pero marcador, labels o
`links.json` pendientes), se informa de lo hecho y lo pendiente y el reintento
reutiliza el Work Item existente y completa solo los pasos que falten; nunca se
crea otro Work Item (R56).

## 5.3 Publicación de la spec

Se ejecuta cuando `spec_author` termina (Caso A) y cuando una replanificación
vuelve a `ready` (Caso B), solo si `links.json.epic != null`.

1. Leer `requirements.md` y `design.md`.
2. Construir el bloque de §3.6.
3. Leer el cuerpo actual de la issue.
4. Si existen los delimitadores `rei:spec:<id>`, reemplazar el bloque; si no,
   añadirlo al final.
5. Editar con `gh issue edit --body-file` y verificar que el delimitador de
   apertura aparece una sola vez.

Si falla, se reporta y se reintenta; no cambia el estado del Work Item (R39–
R41).

---

# 6. Asignación

Disparadores: «asigna la #N a @login», «asígnanos la #12 y la #13», «quítame de
la #N». Sin orden o confirmación no se ejecuta nada (R42).

1. Resolver `@me`:

   ```bash
   me=$(gh api user --jq .login)
   ```

2. Validar cada login:

   ```bash
   gh api "users/<login>" --jq .login
   ```

   Un 404 ⇒ «el login no existe» + siguiente paso. La asignabilidad real la
   confirma GitHub al editar; si rechaza por permisos o por no ser asignable, se
   reporta sin afirmar éxito.

3. Verificar que cada issue existe y es una issue (no un PR).
4. Ejecutar:

   ```bash
   # una issue
   gh issue edit <n> --repo "$canonical" --add-assignee <login>[,<login>]
   # varias issues con el mismo conjunto de personas
   gh issue edit <n1> <n2> --repo "$canonical" --add-assignee <login1>,<login2>
   # desasignar
   gh issue edit <n> --repo "$canonical" --remove-assignee <login>
   ```

5. Verificar:

   ```bash
   gh issue view <n> --repo "$canonical" \
     --json assignees --jq '[.assignees[].login]'
   ```

   Solo se informa de éxito si el login aparece (asignar) o no aparece
   (desasignar). Si ya estaba en el estado pedido, se informa «sin cambios»
   (R43–R46).

El lote solo se usa cuando el mismo conjunto de personas aplica a todas las
issues; si no, se ejecuta una llamada por issue. Ante un fallo parcial se
reportan las issues afectadas y las que quedaron sin cambiar.

---

# 7. Progreso y cierre de la épica

## 7.1 Task list

Formato: `- [ ] #N <título>`. Es la lista de hijas de la épica y el índice que
usa el cierre. Las casillas son informativas: el estado real se consulta
siempre a GitHub (R47).

Se actualiza solo:

- al crear o reanudar un backlog parcial;
- al añadir una hija;
- bajo una orden explícita de reconciliación o cierre.

No hay sincronización en vivo ni comentarios por cambios de estado de las
hijas (R48).

## 7.2 Cierre

Disparador: «cierra la épica #N». Solo aplica a issues con marcador de épica.

1. Leer la épica y verificar que está abierta y que su marcador no es ambiguo.
2. Extraer del bloque `children` (o de la task list si el bloque no existe) los
   números `#N`. Si no hay ninguno o no se pueden interpretar, detenerse.
3. Consultar el estado real de cada hija:

   ```bash
   gh api "repos/$canonical/issues/<n>" \
     --jq '[.number, .state, (.pull_request != null)] | @tsv'
   ```

   - `OPEN` ⇒ no cerrar; listar las hijas abiertas.
   - 404 ⇒ no cerrar; reportar la hija inexistente.
   - `pull_request != null` ⇒ no cerrar; reportar que la entrada no es una
     issue.
4. Si todas están cerradas, reconciliar la task list a `[x]` y verificar el
   cuerpo. Si la escritura falla, no cerrar.
5. Cerrar la épica:

   ```bash
   gh issue close <n> --repo "$canonical" --reason completed
   gh issue view <n> --repo "$canonical" --json state --jq .state  # CLOSED
   ```

6. Si el cierre falla, reportar lo hecho y lo pendiente; la task list ya
   reconciliada es recuperable.

Nunca se cierra la épica al cerrarse la última hija, nunca se cierra una hija
desde este protocolo y nunca se usa la casilla `[x]` como prueba de estado
(R49–R52).

---

# 8. Errores y reintentos

Se conserva la tabla de v1 (configuración, `gh`, repo, issue/rama/PR, permisos,
branch protection, red y rate limit) y se añaden estas filas:

| Situación | Detección | Mensaje (causa) | Siguiente paso |
|-----------|-----------|-----------------|----------------|
| Épica cerrada al añadir hija | estado `CLOSED` | «La épica #N está cerrada» | Reabrirla o elegir otra épica |
| Hija ya adoptada | marcador `rei:work-item:` | «La issue #N ya pertenece al Work Item `<id>`» | Trabajar en ese Work Item o elegir otra hija |
| Épica/hija inexistente | 404 en `gh issue view`/REST | «La issue #N no existe en `<repo>`» | Verificar número y repositorio |
| Marcador duplicado | >1 issue o >1 marcador | «Hay más de una issue con el marcador `<marcador>`» | Corregir los duplicados antes de continuar |
| Assignee inexistente | 404 en `users/<login>` | «El login `<login>` no existe» | Verificar el login |
| Assignee no asignable / sin permisos | 403/422 de GitHub | «No se pudo asignar `<login>` a #N» | Revisar acceso del usuario y `gh auth status` |
| Task list vacía o ilegible | parseo del bloque | «No se pudo determinar las hijas de la épica #N» | Corregir la task list |
| Hijas abiertas al cerrar | estado `OPEN` | «La épica #N tiene hijas abiertas: ...» | Cerrar las hijas primero |
| Creación parcial | contador creadas/pendientes | «Backlog parcial: creadas ..., pendientes ...» | Reintentar; los marcadores evitan duplicados |
| Fallo al publicar la spec | error de `gh issue edit` | «No se pudo publicar la spec en #N» | Reintentar sin duplicar la sección |

Reglas transversales (v1, reforzadas):

- Todo mensaje lleva causa y siguiente paso; nunca éxito sin verificar (R53).
- Un fallo de GitHub no cambia `meta.json`, `current.md` ni la etapa (R56).
- Los reintentos son idempotentes por marcador/sección; no se duplican issues,
  labels ni bloques de spec.
- Ante rate limit o red, no se reintenta en bucle; se muestra el reset si `gh`
  lo proporciona (R55).
- Si una operación de lote falla parcialmente, se indica issue por issue qué
  quedó hecho.

---

# 9. Contratos y documentación

## 9.1 Playbook

`.rei/docs/harness/integrations/github.md` pasa a v2 con esta estructura:

1. Alcance v1 + v2.
2. Configuración y diagnóstico (sin cambios).
3. Vocabulario: marcadores, labels, `links.json`, nombre canónico.
4. Puerta única al aprobar (v1, sin cambios).
5. Adopción y promoción (v1 genérica + hija de backlog + spec).
6. Épicas (modelo, plantilla, creación, idempotencia, parcial).
7. Backlog (hijas, entrevista, borrador, creación, añadir hija).
8. Asignación.
9. Progreso y cierre de épica.
10. Comentarios automáticos (v1).
11. PR y merge (v1).
12. Errores y reintentos (tabla extendida).
13. Recordatorios en `current.md`.

Las secciones v1 se conservan en contenido; solo se reordenan y se referencian
desde v2. Se actualiza el título y el alcance.

## 9.2 Leader

En `.rei/agents/leader.md`, dentro de «Integración con GitHub (opcional)»:

- **Arranque:** si el usuario pide un backlog, crear épica/hijas o asignar,
  aplicar el playbook antes de `rei new`.
- **Backlog:** entrevista, borrador completo, aprobación explícita, idempotencia
  canónica, creación parcial reintentable; nunca crear sin aprobación.
- **Adopción:** leer issue y comentarios; no adoptar épicas; no adoptar una
  issue con marcador de otro Work Item; ofrecer asignación si está sin asignar;
  publicar la spec tras `spec_author`.
- **Asignación:** solo bajo orden o confirmación; validar login; verificar
  resultado; GitHub es la fuente de verdad.
- **Cierre:** solo bajo orden, verificando hijas cerradas; nunca autocierre.
- **Conflictos y errores:** causa + siguiente paso; no bloquean REI; reintento
  sin duplicar.
- **Prueba de campo:** si la planificación del Work Item incluye una prueba de
  campo, registrar la autorización explícita del usuario en `current.md` antes
  de lanzar al Implementer y permitir que sea el Implementer quien la ejecute;
  las operaciones normales de backlog y adopción siguen siendo del Leader.
- **Recordatorios:** añadir las líneas de backlog/asignación/cierre a
  «Próximo paso».

No se añaden herramientas ni se duplica la tabla de errores; se remite al
playbook.

## 9.3 Índice y agentes nativos

- `.rei/docs/harness/README.md`: fila del playbook a v2.
- `check.RequiredFiles` y `embed_test.go` no cambian porque el archivo ya está
  declarado y embebido.
- Regenerar `.opencode/agents/*.md`, `.claude/agents/*.md` y
  `.rei/install-manifest.json` con `rei init opencode` y `rei init claude`;
  verificar con `--check`.

---

# 10. Verificación

## 10.1 Verificación local

- `gofmt -l internal cmd` sin salida.
- `make vet`.
- `make test`.
- `make build`.
- `rei check --quiet`.
- `./bin/rei init opencode --check`.
- `./bin/rei init claude --check`.
- `rei validate 2026-10-25_19-33__github-epicas-backlog-v2`.

No se añaden tests Go: no hay cambio de comportamiento del CLI. La verificación
funcional es la prueba de campo.

## 10.2 Prueba de campo

Se ejecuta en este repositorio, con la integración activa y la autorización
explícita del usuario registrada por el Leader en `current.md` antes de lanzar
al Implementer (R64). La ejecuta el Implementer como parte de T12; las
operaciones normales de backlog/adopción siguen siendo del Leader.

Issues de prueba (títulos inequívocos, se dejan cerradas como evidencia):

- `[prueba v2] Épica: módulo de prueba`
- `[prueba v2] Hija A: ...`
- `[prueba v2] Hija B: ...`

Pasos:

1. Crear la épica y las dos hijas siguiendo §4; verificar marcadores, labels,
   task list y URLs.
2. Repetir la creación por marcador; verificar que no se crean duplicados y que
   se reutilizan los mismos números.
3. Asignar la Hija A a `@me`; verificar; desasignar; verificar. Probar un login
   inexistente y comprobar que el error es claro y no cambia nada. Probar el
   lote con las dos hijas y el mismo assignee.
4. Intentar cerrar la épica con una hija abierta; verificar que se rechaza y
   que la épica sigue abierta.
5. En un worktree temporal (`git worktree add /tmp/rei-v2-field <base>`),
   adoptar una hija: `rei new`, marcador, labels y `links.json`; publicar una
   spec de prueba (fixture mínima) y repetir la publicación para verificar la
   idempotencia. El Work Item de prueba queda `pending` (no se inicia sesión),
   vive solo en el worktree y se elimina con él, sin commits.
6. Cerrar las dos hijas bajo la orden de prueba; cerrar la épica bajo orden;
   verificar `CLOSED`. Intentar añadir una hija a la épica cerrada y comprobar
   que se rechaza.
7. Registrar en `impl.md` números, URLs, comandos y resultados. La hija
   adoptada en el worktree conserva el marcador `rei:work-item:<id-de-prueba>`
   como evidencia; el Work Item de prueba no se conserva fuera del worktree.
   Las issues de prueba quedan cerradas; si alguna no puede cerrarse, se
   documenta el motivo.

Si la prueba falla, el Work Item vuelve a `changes_requested`/`in_progress`
según corresponda; no se envía a revisión con la prueba pendiente.

---

# 11. Alternativas descartadas

| Alternativa | Motivo del descarte |
|-------------|---------------------|
| Crear Work Items `pending` locales para el backlog | Duplica la fuente de verdad, no es portable y contradice el modelo acordado. |
| Implementar épicas/backlog/asignación en Go | Añade API/acoplamiento, exige flags y rompe «el CLI no llama a GitHub». |
| Usar labels como identidad | Las labels se editan a mano y no son únicas; el marcador es la identidad. |
| Usar `gh issue list --search` para idempotencia | No sigue redirects de rename y puede devolver vacío con exit 0; v1 ya lo descartó. |
| Guardar assignees o épicas en `.rei` | GitHub es la fuente de verdad de personas y estados; `.rei` solo conserva el índice local. |
| Autocerrar la épica al cerrar la última hija | El cierre es una decisión humana; se exige orden explícita y verificación. |
| Sincronizar la task list en cada cambio de estado | Genera ruido y escrituras no ordenadas; solo se reconcilia al crear o bajo orden. |
| Publicar la spec como comentario | El cuerpo es el artefacto portable; un comentario se pierde en la lectura de la issue. Se usa una sección delimitada. |
| Publicar la spec sobrescribiendo todo el cuerpo | Destruye el contexto humano de la hija; se reemplaza solo el bloque delimitado. |
| Asignar automáticamente al adoptante | Viola «sin autoasignación»; se ofrece y se confirma. |
| Cerrar hijas desde el protocolo de épica | Cada hija sigue su propio flujo (PR/merge o cierre explícito). |
| Probar la adopción en el repo principal | Dejaría Work Items de prueba en `.rei/specs`; se usa un worktree temporal. |
