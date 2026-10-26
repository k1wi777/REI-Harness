# GitHub — errores, reintentos y recordatorios

> **Archivo de dominio del playbook de GitHub.** Cárgalo solo cuando la tabla
> «Qué leer según la acción» de [`../github.md`](../github.md) lo indique: ante un
> error, una ambigüedad o un reintento, para los recordatorios de `current.md` y
> para la prueba de campo.
>
> Referencias: [índice](../github.md) · [núcleo v1](core.md) ·
> [backlog](backlog.md) · [plantillas](templates.md)

---

# 1. Errores y reintentos

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

# 2. Recordatorios en `current.md`

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

---

# 3. Prueba de campo

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

---

# 4. Cierre tras `done`

Tras `done`, el Reviewer archiva la sesión; las acciones pendientes (comentario
de review, PR o merge) se reportan en la respuesta del Leader, sin recrear
`current.md`.
