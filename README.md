# REI Harness

REI Harness es una **herramienta de línea de comandos** (CLI en Go) para el desarrollo asistido por IA basado en **Spec Driven Development (SDD)** y **orquestación multiagente**.

Su objetivo no es generar código automáticamente, sino proporcionar una estructura donde distintos agentes colaboran de forma controlada, verificable y siempre bajo supervisión humana.

Con un solo comando (`rei init`), la herramienta **inicializa un arnés completo dentro de cualquier proyecto**: no hace falta clonar el repositorio ni copiar carpetas a mano. El estado (planificación, implementación, revisión e historial) vive dentro del propio proyecto.

---

# Instalación

El CLI `rei` se distribuye como **binario precompilado** en las
[releases de GitHub](https://github.com/k1wi777/REI-Harness/releases); no
necesita Go ni ninguna dependencia para ejecutarse.

1. Entra en la página de releases:
   <https://github.com/k1wi777/REI-Harness/releases>.
2. Descarga el archive correspondiente a tu sistema operativo y arquitectura:
   `rei_<versión>_<os>_<arch>.tar.gz` (en Windows, `rei_<versión>_windows_<arch>.zip`).
3. Descomprímelo y mueve el binario `rei` a un directorio de tu `PATH`
   (por ejemplo `/usr/local/bin` o `~/.local/bin`):

   ```bash
   tar -xzf rei_<versión>_<os>_<arch>.tar.gz
   mv rei /usr/local/bin/
   ```

4. Verifica la instalación:

   ```bash
   rei version
   ```

Si desarrollas sobre el propio repositorio y tienes Go instalado, puedes compilar
el binario localmente con `make build`; quedará en `bin/rei`.

## Actualizar el CLI

El binario `rei` puede actualizarse por sí mismo:

- `rei update` descarga la última release, verifica su `checksums.txt` (SHA-256)
  y reemplaza el ejecutable de forma atómica (con backup y restauración si algo
  falla). En Windows, donde no se puede reemplazar el `.exe` en ejecución, abre
  la página de releases para descargarla a mano.
- `rei update --check` solo informa de si hay una versión nueva, sin descargar ni
  modificar nada.

Ambos devuelven `0` cuando no hay nada que actualizar; `1` se reserva para fallos
reales (red, HTTP 5xx, checksum inválido o asset ausente).

---

# Primeros pasos en tu proyecto

1. **Instala el binario** `rei` (ver [Instalación](#instalación)).
2. **Inicializa REI** en la raíz de tu proyecto:

   ```bash
   rei init
   ```

   Despliega `AGENTS.md` y la carpeta `.rei/` (documentación, agentes, plantillas,
   adaptadores y `config.json`), crea la estructura de estado
   (`.rei/specs/`, `.rei/progress/`) e inicializa git si falta. **No sobrescribe**
   archivos existentes.

3. **(Opcional) Genera los subagentes nativos** de tu runtime:

   ```bash
   rei init opencode   # o: rei init claude
   ```

   Crea `.opencode/agents/` + el comando `/personalize`, o `.claude/agents/` +
   `/personalize` + `CLAUDE.md`. En Cursor y Codex (sin subagentes nativos) REI
   funciona por *fallback*, sin configuración.

4. **Personaliza el proyecto** (propósito, stack, arquitectura, convenciones,
   verificación):
   - Con OpenCode o Claude Code: ejecuta **`/personalize`**.
   - O pídele al agente: _"personaliza mi proyecto"_ (el Leader delega en el rol
     `initializer`, que conduce una entrevista guiada).

5. **Trabaja con el agente**: describe lo que necesitas. El Leader negocia el
   Work Item, el Spec Author planifica, tú apruebas y el Implementer/Reviewer
   ejecutan y verifican el ciclo.

Comprueba el estado con `rei init status` (personalización pendiente), `rei doctor`
(diagnóstico) y `rei items status` (Work Items).

---

# Comandos

Los más usados (todos aceptan `rei <comando> --help`):

| Comando | Qué hace |
|---------|----------|
| `rei init` | Inicializa el arnés en el proyecto (esqueleto + estado + git). |
| `rei init status` | Indica qué documentación falta por personalizar. |
| `rei init opencode` / `rei init claude` | Genera los subagentes nativos del runtime. |
| `rei init --update [--force]` | Actualiza el arnés de un proyecto ya inicializado (marker-aware). |
| `rei doctor` | Diagnóstico de solo lectura del arnés. |
| `rei new <id> <feature\|task> [title]` | Crea un Work Item en estado `pending`. |
| `rei session [status]` | Muestra la sesión activa. |
| `rei session start <id> <tipo>` | Inicia la sesión de un Work Item solo si existe, el tipo coincide y está en `pending`. |
| `rei session archive` / `rei session reset` | Archiva una sesión cerrada o restablece `current.md`. |
| `rei commit set <id> <campo>` | Registra el commit actual en `base_commit` o `last_review_commit`. |
| `rei github <enable\|disable\|status>` | Integración opcional con GitHub; `enable` acepta `--repo owner/name`, `--base <rama>` y `--strategy squash\|merge\|rebase`. |
| `rei items status` | Lista tipo, estado, issue y base de cada Work Item. |
| `rei item show <id>` | Ficha de un Work Item (metadatos, enlaces, documentos, sesión y validación). |
| `rei status set <id> <estado> [--force]` | Cambia el estado de un Work Item; `done` es terminal y salir de él requiere `--force`. |
| `rei validate [<id>]` | Comprueba la consistencia interna de un Work Item (un `links.json` inválido es un aviso recuperable). |
| `rei review-diff <id>` | Genera el paquete de revisión por diff. |
| `rei view [<id>]` | Abre el visor local de solo lectura del Work Item (documentos, validación, sesión, TOC, diff y FICHA con issue/base/épica de `links.json`; estados y FICHA se refrescan en vivo y el servidor de fondo se cierra tras mostrar «terminado»). |
| `rei test` | Ejecuta los checks declarados en `.rei/config.json`. |
| `rei update [--check]` | Actualiza el binario desde GitHub Releases. |

Lista completa y detalle: `rei help` y `rei help <comando>`.

---

# Agentes

| Agente | Responsabilidad |
|---------|-----------------|
| **Leader** | Comprender la solicitud, coordinar el workflow y delegar. |
| **Spec Author** | Transformar un Work Item en una planificación técnica. |
| **Implementer** | Implementar únicamente la planificación aprobada. |
| **Reviewer** | Validar que el trabajo cumple la planificación y las reglas del proyecto. |
| **Initializer** | Conducir la personalización guiada del proyecto (wizard). |

Cada rol vive en `.rei/agents/<rol>.md` como un **`## Contrato`** conciso (lo que
recibe el subagente) más una **`## Referencia`** opcional. Los adaptadores traducen
ese Contrato al formato nativo de cada runtime.

---

# Estructura

Este repositorio es **el código fuente de la herramienta**:

```text
.
├── AGENTS.md              # Punto de entrada para los agentes
├── README.md
├── go.mod
├── cmd/rei/               # Punto de entrada del CLI
├── internal/              # Paquetes (paths, meta, state, gitx, check, adapter, update, ...)
├── skeleton/              # Esqueleto neutral que despliega `rei init` (AGENTS.md, config y docs de proyecto)
├── embed.go               # Empaqueta el esqueleto del arnés dentro del binario
├── .goreleaser.yaml       # Release multi-SO/arch
├── .github/workflows/     # CI (formato, vet, tests) y publicación de releases con goreleaser
└── .rei/                  # Arnés (docs, agentes, plantillas, adaptadores) + su propio estado
```

Un **proyecto** que usa REI solo recibe (lo crea `rei init`):

```text
AGENTS.md
.rei/
├── agents/      # Roles (Contrato + Referencia)
├── adapters/    # Plantillas de adaptadores por runtime
├── config.json  # Checks de verificación del proyecto
├── docs/        # Documentación del arnés y del proyecto
├── templates/   # Plantillas canónicas
├── specs/       # Work Items y planificaciones
└── progress/    # Sesión actual, historial y reportes por Work Item

.opencode/agents/ + .opencode/commands/     # Si usas OpenCode
.claude/agents/ + .claude/commands/ + CLAUDE.md   # Si usas Claude Code
```

---

# Cómo funciona

## 1. El repositorio es la memoria

Los agentes no dependen del historial del chat. Toda la información relevante
(planificación, progreso, historial, estados) vive dentro del proyecto. Una
conversación puede perderse; el repositorio no.

## 2. Un agente, una responsabilidad

Cada agente posee un único objetivo y no sustituye el trabajo de otro. El Leader
transmite a cada subagente solo el `## Contrato` de su rol.

## 3. Spec Driven Development

Todo trabajo sigue el mismo flujo base:

```
Usuario → Leader → meta.json → Spec Author → Planificación
        → ⏸ Aprobación humana → Implementer → Reviewer → Finalización
```

El código nunca se implementa antes de existir una planificación aprobada.

> Este diagrama muestra la ruta principal. Las ramificaciones (rechazo en
> revisión, bloqueos) están en `.rei/docs/harness/workflow.md`.

## 4. El humano siempre mantiene el control

Ningún agente avanza automáticamente entre etapas críticas. Toda planificación
debe ser aprobada explícitamente antes de implementar.

---

# Work Items

REI Harness trabaja sobre **Work Items**, de dos tipos: **Feature** (planificación
completa: `requirements.md`, `design.md`, `tasks.md`) y **Task** (planificación
breve: `plan.md`). Los criterios, estados y transiciones están en
`.rei/docs/harness/workflow.md`.

---

# Documentación

## Del arnés (`.rei/docs/harness/`)

Define el funcionamiento de REI Harness. **No requiere personalización.**

| Documento | Propósito |
|-----------|-----------|
| `workflow.md` | Flujo completo: tipos de Work Item, estados y transiciones. |
| `specs.md` | Cómo se construyen las especificaciones (Features). |
| `task.md` | Formato del plan de una Task. |
| `progress.md` | Sistema de progreso y plantillas. |
| `meta.md` | Estructura y significado de `meta.json`. |
| `adapters.md` | Formato canónico de rol y adaptadores de runtime. |
| `integrations/github.md` | Índice de la integración opcional con GitHub Issues; el detalle vive en `integrations/github/core.md`, `backlog.md`, `templates.md` y `errors.md`. |

## Del proyecto (`.rei/docs/project/`)

Describe las reglas de **tu repositorio**. **Debes personalizarla** (con
`/personalize` o el rol `initializer`).

| Documento | Propósito |
|-----------|-----------|
| `architecture.md` | Principios de arquitectura del proyecto. |
| `conventions.md` | Convenciones de desarrollo. |
| `verification.md` | Checkpoints y reglas de validación. |

Los agentes cargan únicamente la documentación necesaria para su etapa.

---

# Sistema de progreso

El progreso vive en `.rei/progress/`: la sesión actual (`current.md`), el
historial permanente (`history.md`, append-only) y una carpeta por Work Item
(`work-items/<id>/`) con los reportes de implementación y revisión. Esto permite
continuar sesiones interrumpidas y mantener trazabilidad.

---

# Revisión por diff (git)

REI Harness funciona sin git, pero está optimizado para usarlo. Con git, el
Reviewer recibe un **paquete de revisión** (`rei review-diff`) con los cambios
reales del Work Item en lugar de releer toda la especificación. Al aprobar se
registra un `base_commit`; en un `changes_requested`, un `last_review_commit`
para que la próxima revisión vea solo los cambios pedidos. Sin git, el flujo
sigue funcionando en modo lectura.

---

# Filosofía

Este repositorio no pretende construir un asistente autónomo, sino un sistema
donde los agentes tienen responsabilidades claras, el contexto está distribuido,
la documentación es la fuente de verdad y el humano conserva siempre el control
del proceso. La IA no reemplaza el proceso de desarrollo: lo sigue.
