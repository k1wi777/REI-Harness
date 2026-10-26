# Uso de REI Harness

> Esta guía está dirigida a **personas**, no a agentes.
> Explica cómo operar REI Harness en el día a día: cómo arrancarlo, qué esperar en cada etapa y cómo participar en los puntos donde se requiere tu aprobación.
>
> Si eres un agente de IA trabajando en este repositorio, tu punto de entrada es `AGENTS.md`, no este documento.

---

# Tu responsabilidad en este proceso

REI Harness no reemplaza tu criterio — lo agiliza. Los agentes redactan planificación, implementan y verifican, pero la responsabilidad del proyecto sigue siendo tuya en dos momentos concretos:

- **Al negociar el Work Item.** Tú decides qué se construye, por qué y con qué alcance. El Leader no actúa hasta llegar a un acuerdo explícito contigo — su trabajo es comprender tu solicitud y formalizarla, no interpretarla por su cuenta.
- **Al aprobar la planificación.** Antes de que se escriba una sola línea de código, tú confirmas que la lógica, las decisiones de diseño y los requisitos son correctos. No se te pide revisar código — esa es justamente la parte que REI Harness te ahorra —, pero sí se te pide revisar las decisiones que determinan qué se va a construir.

Ninguna de las dos es un trámite. Son los dos puntos donde el proceso se detiene y te devuelve el control antes de avanzar.

---

# 1. Primera vez en el repositorio

1. Obtén el binario `rei` para tu sistema operativo descargándolo de la
   [página de releases](https://github.com/k1wi777/REI-Harness/releases)
   (elige el archive `rei_<versión>_<os>_<arch>`) y colócalo en tu `PATH`.
   El binario **no necesita Go ni ninguna dependencia** para ejecutarse. Si
   desarrollas sobre el propio repositorio y tienes Go, también puedes
   compilarlo con `make build`.
2. En la raíz de tu proyecto, inicializa el arnés:
   ```bash
   rei init
   ```
   Despliega `AGENTS.md` y la carpeta `.rei/` (documentación, agentes, plantillas,
   adaptadores y `config.json`), prepara la estructura de estado
   (`.rei/specs/`, `.rei/progress/`) e inicializa git si falta. **No sobrescribe**
   archivos existentes.
3. (Opcional) Genera los subagentes nativos de tu runtime:
   ```bash
   rei init opencode   # o: rei init claude
   ```
   En Cursor y Codex (sin subagentes nativos) REI funciona por *fallback*, sin configuración.
4. **Personaliza la documentación de tu proyecto** (ver Sección 1.1): ejecuta
   `/personalize` (OpenCode o Claude Code) o pídele al agente «personaliza mi
   proyecto» (el rol `initializer` conduce la entrevista).

Comprueba el estado en cualquier momento con `rei init status` (documentación
pendiente), `rei doctor` (diagnóstico) y `rei items status` (Work Items).

### Permisos de ejecución

El CLI `rei` es un binario compilado; no necesita permisos de ejecución especiales más allá de estar disponible en el `PATH`.

### Actualizar el CLI

Una vez instalado, `rei` puede actualizarse por sí mismo:

- `rei update` descarga la última release publicada, verifica la integridad del archive (SHA-256) y reemplaza el ejecutable de forma atómica. En Windows, como no se puede reemplazar el `.exe` en ejecución, abre la página de releases para que la descargues a mano.
- `rei update --check` solo comprueba si hay una versión más reciente; no descarga ni modifica nada.

Ambos devuelven `0` cuando no hay nada que actualizar (ya estás al día o todavía no hay releases publicadas). Reservan el código `1` para fallos reales, como problemas de red o un checksum inválido.

---

## 1.1 Qué personalizar y qué no

La documentación en `docs/` está dividida en dos carpetas. Solo una de ellas debes adaptarla a tu proyecto.

### Personaliza (documentación del proyecto)

| Archivo | Qué debes definir |
|---------|-------------------|
| `.rei/docs/project/architecture.md` | Decisiones arquitectónicas de tu repositorio. |
| `.rei/docs/project/conventions.md` | Estilo de código y convenciones de desarrollo. |
| `.rei/docs/project/verification.md` | Checkpoints (`V1`, `V2`, ...) y cómo demostrar que el trabajo funciona. |
| `AGENTS.md` — Sección 2 | Propósito del producto, usuarios, principios y límites. |
| `AGENTS.md` — Sección 3 | Stack técnico y comandos principales del proyecto. |
| `.rei/config.json` | Checks de verificación (`checks`) alineados con `verification.md`. |

### No personalices (documentación de REI Harness)

| Carpeta / archivo | Motivo |
|-------------------|--------|
| `.rei/docs/harness/` | Define el funcionamiento del arnés — es igual en todos los proyectos. |
| `.rei/agents/` | Roles y protocolos de los subagentes. |
| `.rei/docs/usage.md` | Esta guía. |

Si modificas un archivo de `.rei/docs/harness/` o `.rei/agents/`, estarías alterando REI Harness en sí, no la configuración de tu proyecto.

La integración con GitHub es opcional: `.rei/docs/harness/integrations/github.md`
es el índice y los agentes cargan solo el archivo de la acción en curso
(`integrations/github/core.md`, `backlog.md`, `templates.md` o `errors.md`).

---

# 2. Negociar el Work Item con el Leader

Descríbele al agente, en lenguaje natural, lo que necesitas. El agente actúa como **Leader**, y su primer trabajo no es ejecutar — es entender.

Si tu solicitud tiene cualquier ambigüedad de alcance, el Leader **debe** detenerse y preguntarte antes de continuar — no está autorizado a asumir nada por su cuenta. Esto normalmente significa una conversación de ida y vuelta: el Leader propone una interpretación, tú la ajustas, hasta llegar a un acuerdo claro sobre qué se va a construir.

Esta negociación es tu responsabilidad, no un paso automático. El `meta.json` que el Leader crea al final **representa el acuerdo alcanzado contigo** — es tu criterio quedando registrado, no una decisión que el agente tomó solo. Vale la pena tomarte este paso en serio: cuanto más claro quede el acuerdo aquí, menos ambigüedad tendrá que resolver el Spec Author después.

Solo una vez que existe ese acuerdo explícito, el Leader crea el Work Item (`.rei/specs/<work-item-id>/meta.json`) con estado `pending`, genera su `id` con fecha y hora de creación y decide si corresponde a una Feature o una Task (ver `.rei/docs/harness/workflow.md`).

---

# 3. El ciclo de trabajo

Cada Work Item pasa por las mismas etapas, sin importar si es una Feature o una Task:

```
pending → spec_author → ready → ⏸ tu aprobación → in_progress → review → reviewer → done
```

En la práctica, esto es lo que vas a ver:

1. **Planificación.** El Leader delega en el Spec Author, que redacta la planificación (`requirements.md` + `design.md` + `tasks.md` para una Feature, o `plan.md` para una Task) dentro de `.rei/specs/<work-item-id>/`.
2. **Tu aprobación.** El Leader se detiene y te pide revisar la planificación. Esto **nunca se salta** — es el punto donde tienes control total antes de que se escriba una sola línea de código.
3. **Implementación.** Solo después de tu aprobación explícita, el Implementer ejecuta la planificación tal como fue aprobada.
4. **Revisión.** El Reviewer valida el trabajo contra la planificación y los checkpoints de `.rei/docs/project/verification.md`.
5. **Finalización.** Si todo pasa, el Work Item queda `done` y su resumen se archiva en `.rei/progress/history.md`.

---

# 4. Cómo aprobar o pedir cambios

Cuando el Leader te presente una planificación lista (`ready`), te toca ejercer la parte de este proceso que nadie más puede hacer por ti: revisar que la lógica, las decisiones de diseño y los requisitos sean los correctos **antes** de que exista código. No se trata de leer por leer — es la comprobación real de que lo que se va a construir es lo que realmente acordaste en la negociación del Work Item (Sección 2).

- Si estás de acuerdo, responde algo como **"aprobado"**. El Leader avanzará el Work Item a `in_progress`.
- Si algo no refleja lo acordado, o encuentras una decisión de diseño con la que no estás de acuerdo, dilo en lenguaje natural. El Leader relanzará al Spec Author y volverá a presentarte la planificación actualizada.

La aprobación se da **en el chat o en la terminal** (el Leader ejecuta
`rei status set <id> in_progress`). Si abres el visor (`rei view`), ten en cuenta
que es de **solo lectura**: muestra documentos, validación, FICHA y sesión, con el
estado real del Work Item (badge y FICHA se refrescan en vivo y la FICHA incluye
issue, base y épica de `links.json`), pero no tiene botón ni ruta de aprobación.
El visor nunca sustituye tu respuesta. En segundo plano se cierra poco después de
mostrar «terminado» (`done`); `--stop`, el tiempo de inactividad y Ctrl+C siguen
funcionando.

REI Harness nunca avanza sin esta aprobación explícita — no por burocracia, sino porque la responsabilidad de qué se construye es tuya, no del agente.

Lo mismo aplica después de la revisión: si el Reviewer rechaza el trabajo (`changes_requested`), el Leader te lo informará y relanzará al Implementer con las correcciones necesarias — no necesitas hacer nada salvo revisar el resultado otra vez al final.

---

# 5. Continuar una sesión interrumpida

Si cierras la conversación con un Work Item a medias, no se pierde nada — vive en `.rei/specs/<work-item-id>/` y `.rei/progress/work-items/<work-item-id>/`, no en el historial del chat.

La próxima vez que ejecutes `rei check` y hables con el agente, el Leader consultará el estado mediante el CLI (`rei session`, `rei items status`) y te preguntará si quieres continuar, reiniciar o cancelar ese Work Item. Tú decides; el Leader nunca lo asume por su cuenta.

---

# 6. Preguntas frecuentes

**¿Puedo tener varios Work Items en paralelo?**
No al mismo tiempo en estado `in_progress` — REI Harness solo permite uno activo a la vez (ver `.rei/docs/harness/workflow.md`). Sí puedes tener varios `pending` esperando su turno. El límite es duro: `rei status set <id> in_progress` lo rechaza (también con `--force`), y `rei check`, `rei doctor` y `rei items status` fallan si detectan más de uno. Cierra (`review`/`done`), bloquea o revisa el Work Item activo antes de iniciar otro.

**¿Dónde declaro los checks de mi proyecto?**
En `.rei/config.json` (ver `.rei/docs/project/verification.md`). `rei check` los ejecuta y refleja el resultado en su código de salida.

**¿Qué hago si `rei check` falla por un check del proyecto?**
Significa que algún comando del proyecto (tests, build, lint) falló. Revisa el detalle en la salida y corrige antes de solicitar revisión de un Work Item.

**¿Dónde veo el historial de todo lo que se ha hecho?**
En `.rei/progress/history.md` — es un registro permanente, nunca se sobrescribe.

**¿Necesito leer toda la documentación de `.rei/docs/` para usar REI Harness?**
No. Esta guía es suficiente para el uso diario. Los agentes cargan solo lo necesario según la etapa — como humano, consulta `.rei/docs/project/` si quieres revisar las reglas de tu proyecto, o `.rei/docs/harness/` si quieres entender el detalle del arnés.

**¿REI Harness necesita git?**
No, pero funciona mejor con él. Si el repositorio usa git, el Reviewer revisa solo los cambios reales del Work Item (`rei review-diff`) en lugar de releer toda la especificación, y en `changes_requested` puede acotar la revisión a lo que cambió. Si no hay git, el flujo sigue funcionando en modo lectura. `rei check` inicializa git automáticamente si no existe.

---

# 7. Comandos

Los comandos de uso diario (todos aceptan `rei <comando> --help`):

| Comando | Qué hace |
|---------|----------|
| `rei new <id> <feature\|task> [title]` | Crea un Work Item en estado `pending`. |
| `rei session` / `rei session status` | Muestra la sesión activa. |
| `rei session start <id> <tipo>` | Inicia la sesión de un Work Item **solo si existe, el tipo coincide y su estado es `pending`**; en cualquier otro caso rechaza con causa y siguiente paso sin tocar `current.md`. |
| `rei session archive` / `rei session reset` | Archiva la sesión cuando el Work Item está cerrado o restablece `current.md`. |
| `rei commit set <id> <campo>` | Registra el commit actual en `base_commit` o `last_review_commit`. |
| `rei github enable` | Activa la integración con GitHub. Acepta `--repo owner/name`, `--base <rama>` y `--strategy squash\|merge\|rebase`; sin flags detecta `origin` y `origin/HEAD`. |
| `rei github disable` / `rei github status` | Desactiva la integración conservando los valores, o muestra el estado sin escribir. |
| `rei item show <id>` | Ficha del Work Item con la sección `Enlaces:` (issue, URL, base y épica) si existe `links.json`. |
| `rei items status` | Lista cada Work Item con su tipo y, si `links.json` es válido, su issue y base. |
| `rei status set <id> <estado> [--force]` | Cambia el estado. `done` es terminal: salir de `done` (por ejemplo `done → blocked`) solo se permite con `--force` como transición no estándar. |
| `rei validate [<id>]` | Valida el Work Item; un `links.json` inválido se informa como `WARN` recuperable, nunca como `FAIL`. |

`rei github enable` resuelve cada valor con la precedencia **flag explícito >
valor previo de `integrations.github` > detección desde git > valor por defecto**.
`base` usa `main` y `mergeStrategy` usa `squash` cuando no hay nada mejor; `repo`
no tiene valor por defecto y, sin flag ni valor previo, exige un remoto `origin`
resoluble. Un flag desconocido, repetido o inválido es uso incorrecto (código 2)
y no escribe la configuración.

---

# 8. Subagentes nativos (OpenCode y Claude Code)

REI Harness define sus roles en `.rei/agents/`: el `leader` coordina el workflow y delega en el Spec Author, el Implementer, el Reviewer y el Initializer. Si tu runtime dispone de un adaptador nativo, puedes instalarlos como **subagentes nativos** para invocarlos directamente, sin que el Leader tenga que reenviarles el contrato cada vez.

Para generarlos en **OpenCode**, ejecuta:

```bash
rei init opencode
```

Esto crea un archivo por rol en `.opencode/agents/<rol>.md` —por ejemplo `.opencode/agents/implementer.md`— a partir del `## Contrato` canónico de `.rei/agents/<rol>.md`, además del comando `/personalize`.

Para **Claude Code** el comando es análogo:

```bash
rei init claude
```

Crea `.claude/agents/<rol>.md`, el comando `/personalize` y un `CLAUDE.md` que apunta a la documentación del arnés.

En ambos casos, los archivos generados llevan la marca `GENERATED` y la instalación es idempotente: si ya están al día, volver a ejecutarla no los modifica. Estos comandos **no sobrescriben** archivos que no hayan generado ellos; si encuentran uno, te avisan y lo conservan tal cual.

Una vez generados, tienes dos formas de trabajar con ellos:

- **Invocar el subagente directamente** con `@spec_author`, `@implementer`, `@reviewer` o `@initializer`, según la etapa del ciclo (Sección 3).
- **Cambiar al agente `leader`**, que es el agente principal desde el que coordinas el workflow y que delega en el resto.

Si editas el `## Contrato` de un rol en `.rei/agents/`, los archivos generados quedan desactualizados. Para comprobarlo sin escribir nada:

```bash
rei init opencode --check   # o: rei init claude --check
```

El comando devuelve `0` si todos los archivos coinciden con la generación canónica, y `1` si falta alguno o difiere. Así detectas la **deriva** entre `.rei/agents/` y los agentes nativos antes de que afecte al trabajo.

Si tu entorno no dispone de un adaptador nativo, el resultado es equivalente: el Leader transmite al subagente únicamente el `## Contrato` del rol —nunca el archivo completo—, de modo que obtienes el mismo contrato operativo sin configuración adicional.
