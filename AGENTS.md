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

## Qué construimos

**REI Harness** es una herramienta de línea de comandos (CLI en Go, binario `rei`)
que implanta un arnés de **Spec Driven Development (SDD)** y orquestación
multiagente dentro de cualquier repositorio, con roles explícitos y una **puerta
de aprobación humana obligatoria**.

Resuelve el problema de que un agente implemente a ciegas: primero produce una
planificación revisable, el humano la aprueba y solo entonces se implementa. El
estado (planificación, implementación, revisión e historial) vive dentro del
propio proyecto destino.

## Para quién

- **Desarrolladores que trabajan con agentes de IA bajo SDD** y revisan/aprueban
  la planificación de un Work Item antes de implementarlo, a menudo alternando
  varios runtimes (OpenCode, Claude Code, Cursor, Codex).
- **Mantenedores del harness**, que desarrollan el CLI `rei`, sus contratos de
  rol y sus adaptadores.
- **Equipos que adoptan REI** en repositorios existentes mediante `rei init`.

## Principios

*Todas las decisiones deben respetar estos principios.*

- **Aprobación humana obligatoria y explícita** — El Work Item no pasa a
  implementación sin aprobación registrada en disco; nunca se asume por silencio.
- **La documentación es la fuente de verdad** — Las decisiones viven en los
  documentos del Work Item y en `.rei/docs/`, no en el chat.
- **Runtime-agnóstico** — El valor reside en el contrato de rol y en el CLI; los
  adaptadores son opcionales y no pueden convertirse en el mecanismo central.
- **Código propio y dependencias mínimas** — Se prefiere la biblioteca estándar y
  código propio antes que incorporar dependencias de terceros.
- **Baja fricción de revisión** — El material debe poder revisarse de verdad:
  legible, navegable y con evidencia clara.

## Qué NO es

*Acota explícitamente el alcance del proyecto para evitar feature creep.*

- **No es un generador de código ni un agente autónomo** — Orquesta el workflow;
  el código lo escriben los agentes bajo supervisión humana.
- **No es un framework de agentes atado a un runtime concreto** — No habrá hooks
  de un runtime como mecanismo central; los adaptadores son una capa opcional.
- **No es un servicio alojado ni un editor web** — `rei view` es un visor local
  de solo lectura sobre los documentos del Work Item.
- **No incorpora JavaScript ni CSS de terceros en el visor** — Los assets son
  propios y van embebidos en el binario.

---

# 3. Stack técnico

> **Personaliza esta sección** con el stack y los comandos reales de tu proyecto.

## Tecnologías principales

- **Lenguaje:** Go 1.27 (`go.mod`), sin `CGO`.
- **Framework:** CLI propio sobre la biblioteca estándar (`flag`/dispatch manual,
  `net/http` + `html/template` para `rei view`); sin framework web.
- **Base de datos:** ninguna. El estado vive en ficheros del repositorio:
  Markdown (`.rei/specs/`, `.rei/progress/`, `.rei/docs/`) y JSON
  (`.rei/config.json`, `meta.json`).
- **Testing:** `testing` de la biblioteca estándar (`go test ./...`), con tests
  junto al código en `internal/**`.
- **Despliegue:** binarios precompilados con GoReleaser (`.goreleaser.yaml`) al
  publicar un tag `v*`; también `make install` para instalación local.

## Comandos principales

*Comandos que cualquier agente utilizará con frecuencia.*

- `make build` — Desarrollo: compila el binario en `bin/rei` (acepta
  `VERSION=<v>` para sellar la versión).
- `make test` — Tests: ejecuta `go test ./...`.
- `make vet` — Linter: ejecuta `go vet ./...`; `make fmt` aplica `gofmt -w .`.
- `make build` / `make cross` — Build local y build multiplataforma en `dist/`
  (linux/darwin/windows, amd64/arm64).

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
| `.rei/docs/project/` | Documentación específica del proyecto (viva, personalizada). |
| `skeleton/` | Esqueleto neutral desplegable por `rei init` (`AGENTS.md`, `.rei/config.json` y `.rei/docs/project/**` con placeholders). Separado del proyecto vivo. |
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
