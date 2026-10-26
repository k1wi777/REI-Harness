# GitHub — backlog v2 (épicas, hijas, promoción y cierre)

> **Archivo de dominio del playbook de GitHub.** Cárgalo solo cuando la tabla
> «Qué leer según la acción» de [`../github.md`](../github.md) lo indique para la
> acción en curso; `github.md` es el índice de orientación y configuración.
>
> Referencias: [índice](../github.md) · [núcleo v1](core.md) ·
> [plantillas](templates.md) · [errores](errors.md)

---

# 1. Alcance v2

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

---

# 2. Promoción de una hija de backlog

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
   registrada ya no existe, deriva a [`errors.md §1`](errors.md).
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

---

# 3. Publicación de la spec

Se ejecuta cuando `spec_author` termina la planificación de un Work Item
adoptado de backlog (Caso A) y cuando una replanificación vuelve a `ready` (Caso
B), solo si `links.json.epic != null`.

1. Leer `requirements.md` y las decisiones con peso de `design.md`.
2. Construir el bloque de [`templates.md §2`](templates.md), **sin** el design
   completo.
3. Leer el cuerpo actual de la issue.
4. Si existen los delimitadores `rei:spec:<id>`, **reemplazar el bloque en
   sitio**; si no, añadirlo al final. Nunca se duplica la sección.
5. Editar con `gh issue edit --body-file` y verificar que el delimitador de
   apertura aparece **una sola vez**.
6. Solo si `.rei` no se versiona o el usuario pide trazabilidad completa,
   publicar o actualizar el design como comentario idempotente
   `<!-- rei:comment:design:<id> -->` ([`templates.md §7`](templates.md)),
   reutilizando el marcador y sin duplicarlo. Por defecto **no** se publica.

Si falla, se reporta la causa y el siguiente paso, **no** cambia el estado del
Work Item y se puede reintentar sin duplicar la sección.

---

# 4. Épicas

## 4.1 Modelo

Una **épica** es una issue de GitHub que planifica un módulo. **No es un Work
Item**: no tiene `requirements.md`, `design.md`, `tasks.md` ni requisitos EARS.
Lleva:

- marcador `<!-- rei:epic:<epic-slug> -->` en la **primera línea** del cuerpo;
- labels `rei` y `rei:epic`.

El **slug** es kebab-case, estable y único en el repositorio. Cambiar el título
no cambia el slug ni el marcador.

El cuerpo se construye con la plantilla de [`templates.md §3`](templates.md). El
bloque `children` puede crearse **vacío** y rellenarse después de crear las
hijas. El plan completo es opcional y, si existe, va plegado en `<details>`. Si
el bloque se corrompe, se detiene y se pide corregirlo antes de reescribirlo.

## 4.2 Creación del backlog de un módulo

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

Al inicio de la acción se resuelve el nombre canónico
([`core.md §3`](core.md)). La búsqueda de un marcador es **siempre REST,
paginada y sobre issues (no PRs)**:

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

# 5. Backlog de hijas

## 5.1 Modelo

Una **hija** es una issue completa de GitHub, **no un Work Item** hasta que se
adopta. Lleva:

- marcador `<!-- rei:backlog:<epic-slug>/<item-slug> -->` en la **primera
  línea** del cuerpo;
- labels `rei` y `rei:backlog`.

El slug de la hija es kebab-case, estable y **único dentro de la épica**. La
identidad de la hija es su marcador, no su título.

El cuerpo se construye con la plantilla de [`templates.md §4`](templates.md).

**No incluye** requisitos EARS, `design.md`, `tasks.md` ni plan de archivos. Al
adoptarse, el cuerpo se conserva y solo se sustituye el marcador y se añade la
sección de spec.

## 5.2 Creación y ampliación

La creación de hijas usa la misma entrevista, borrador, aprobación explícita,
idempotencia canónica y orden de §4.2. Cada issue creada o reutilizada se
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

# 6. Asignación

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

# 7. Progreso y cierre de la épica

## 7.1 Task list

Formato: `- [ ] #N <título>`. Es la lista de hijas de la épica y el índice que
usa el cierre. Las **casillas son informativas**: el estado real se consulta
siempre a GitHub.

Se actualiza solo:

- al crear o reanudar un backlog parcial;
- al añadir una hija;
- bajo una orden explícita de reconciliación o cierre.

No hay sincronización en vivo ni comentarios por cambios de estado de las hijas.

## 7.2 Cierre

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
