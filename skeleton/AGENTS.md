# AGENTS.md — REI Harness

> Este archivo es el **punto de entrada universal** para cualquier agente de IA que trabaje en este repositorio.
>
> Su objetivo es proporcionar únicamente el contexto esencial para orientarse dentro del proyecto y saber dónde encontrar información adicional. **No debe contener toda la documentación del proyecto.**
>Carga únicamente el contexto necesario para la tarea actual.
> Las responsabilidades específicas de cada rol se encuentran en `.rei/agents/`.
> Cada rol se define en `.rei/agents/<rol>.md` con frontmatter, un `## Contrato`
> autosuficiente y una `## Referencia` opcional; a los subagentes se les
> transmite únicamente el `## Contrato`. Consulta `.rei/docs/harness/adapters.md`.

**No ignores estas reglas.**

---

# 1. Antes de comenzar

Antes de realizar cualquier modificación en el proyecto:

1. Ejecuta `rei check --quiet`.
2. Si falla, DETENTE e informa el problema.
3. Carga únicamente la documentación necesaria para la tarea actual siguiendo el orden de la Sección 8.

No continúes hasta completar estos pasos.

---

# 2. Propósito del proyecto

> **Personaliza esta sección** al adaptar REI Harness a tu repositorio. Describe el producto concreto, no el arnés.

<!-- REI:PENDIENTE -->

*Resume aquí el objetivo del proyecto. Esta información guía las decisiones del agente y evita cambios que contradigan la visión del producto.*

## Qué construimos

<Una o dos frases que describan claramente el producto y el problema que resuelve.>

## Para quién

- <Usuario principal.>
- <Usuarios secundarios, si existen.>
- <Otros interesados, si aplica.>

## Principios

*Todas las decisiones deben respetar estos principios.*

- **<Principio>** — <Qué implica en la práctica.>
- **<Principio>** — <Qué implica en la práctica.>
- **<Principio>** — <Qué implica en la práctica.>

## Qué NO es

*Acota explícitamente el alcance del proyecto para evitar feature creep.*

- <Fuera del alcance.>
- <Otro límite importante.>

---

# 3. Stack técnico

> **Personaliza esta sección** con el stack y los comandos reales de tu proyecto.

<!-- REI:PENDIENTE -->

*Documenta únicamente la información técnica esencial para comprender el proyecto.*

## Tecnologías principales

- **Lenguaje:** <...>
- **Framework:** <...>
- **Base de datos:** <...>
- **Testing:** <...>
- **Despliegue:** <...>

## Comandos principales

*Comandos que cualquier agente utilizará con frecuencia.*

- `<comando>` — Desarrollo.
- `<comando>` — Tests.
- `<comando>` — Linter.
- `<comando>` — Build.

---

# 4. Estructura del repositorio

*Utiliza este mapa para orientarte dentro del proyecto.*

| Ruta | Propósito |
|------|-----------|
| `src/` | Código fuente del proyecto. |
| `tests/` | Suite de pruebas. |
| `.rei/specs/` | Work Items y planificación del REI Harness. |
| `.rei/progress/` | Estado actual (`current.md`), historial general (`history.md`) y detalle de Work Items. |
| `.rei/progress/work-items/` | Registros detallados de cada Work Item, organizados cronológicamente. |
| `.rei/docs/harness/` | Documentación del arnés (workflow, specs, progreso). |
| `.rei/docs/project/` | Documentación específica del proyecto. |
| `.rei/agents/` | Roles y comportamiento de los subagentes (frontmatter + `## Contrato` + `## Referencia`). |
| `.rei/adapters/` | Adaptadores de runtime (plantilla + mapa de herramientas); p. ej. `opencode/`. |
| `cmd/`, `internal/` | Código del CLI `rei` (Go). |
| `.rei/templates/` | Plantillas canónicas (`current.md`, `history.md`, `meta.json`). Fuente única de verdad. |

---

# 5. Workflow

Todo Work Item debe seguir el workflow definido por el proyecto.

Antes de comenzar cualquier trabajo consulta:

`.rei/docs/harness/workflow.md`

**No omitas ninguna etapa del workflow.**

---

# 6. Rol del agente principal

El agente que recibe directamente las solicitudes del usuario actúa siempre como **Leader**.

Antes de comenzar cualquier tarea DEBE consultar:

`.rei/agents/leader.md`

El Leader es el único responsable de:

- comprender la solicitud del usuario;
- seleccionar el workflow adecuado;
- coordinar a los demás subagentes.

Ningún otro agente debe asumir estas responsabilidades.

Al delegar en un subagente, el Leader transmite únicamente la sección `## Contrato`
del rol correspondiente, nunca el archivo completo (`## Referencia` es detalle
opcional y no se envía). Si el runtime dispone de un adaptador nativo
(`rei init opencode`), el subagente recibe ese mismo Contrato como system prompt.

Si el usuario pide personalizar el proyecto (por ejemplo, «personaliza mi
proyecto»), el Leader NO crea un Work Item: delega directamente en el rol
`initializer` (`.rei/agents/initializer.md`) transmitiéndole su `## Contrato` y
espera a que finalice. En un runtime con adaptador nativo, el comando
`/personalize` ejecuta este mismo camino.

---

# 7. Documentación

Consulta únicamente la documentación necesaria para la tarea actual.

## Documentación de REI Harness

Define cómo funciona el arnés. No requiere personalización por proyecto.

| Si necesitas... | Consulta... |
|-----------------|-------------|
| Workflow | `.rei/docs/harness/workflow.md` |
| Spec Driven Development | `.rei/docs/harness/specs.md` |
| Sistema de progreso | `.rei/docs/harness/progress.md` |
| Estructura de `meta.json` | `.rei/docs/harness/meta.md` |
| Roles y adaptadores | `.rei/docs/harness/adapters.md` |
| El comportamiento de un agente | `.rei/agents/<role>.md` |

## Documentación del proyecto

Describe las reglas de **este repositorio**. Debe personalizarse al implementar REI Harness.

| Si necesitas... | Consulta... | Quién lo lee por defecto |
|-----------------|-------------|--------------------------|
| Arquitectura | `.rei/docs/project/architecture.md` | Implementer, Spec Author (diseño) |
| Convenciones | `.rei/docs/project/conventions.md` | Implementer |
| Verificación | `.rei/docs/project/verification.md` | Reviewer |

## Work Items

| Si necesitas... | Consulta... |
|-----------------|-------------|
| Un Work Item concreto | `.rei/specs/<work-item-id>/` |
| El comportamiento de un agente | `.rei/agents/<role>.md` |
No cargues documentación que no aporte contexto a la tarea actual.

---

# 8. Carga de contexto

Prioriza siempre la carga de información en el siguiente orden:

1. `AGENTS.md`
2. `.rei/agents/<role>.md`
3. `docs/...`
4. `.rei/specs/<work-item-id>/`

Carga únicamente el contexto necesario para completar la tarea actual.

---

# 9. Principios generales

Estas reglas aplican a cualquier agente del repositorio.

- La documentación es la fuente de verdad.
- Trabaja sobre un único Work Item por sesión.
- No omitas etapas del workflow.
- Consulta la documentación antes de asumir comportamientos no especificados.
- Si encuentras documentación contradictoria, DETENTE y repórtala.
- Para conocer el estado de la sesión o de los Work Items, usa el CLI (`rei session`, `rei items status`); no escanees `meta.json` manualmente.
- No releas un archivo que ya esté en tu contexto; para localizar un dato concreto usa búsquedas en lugar de releer el archivo completo.
- En caso de conflicto entre documentos, prevalece el orden inverso de la Sección 8:
  `.rei/specs/<work-item-id>/` > `.rei/docs/...` > `.rei/agents/<role>.md` > `AGENTS.md`.
