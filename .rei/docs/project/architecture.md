# Arquitectura

> **Documento del proyecto** — define las decisiones arquitectónicas de REI Harness.
>
> El Implementer debe respetarlas durante la implementación y el Reviewer las utilizará como criterio de validación.

---

# Objetivo

Describe cómo está organizado el proyecto y las reglas arquitectónicas que deben mantenerse durante su evolución.

No documenta requisitos funcionales ni decisiones específicas de una Feature.

---

# Principios

- **CLI delgada, lógica en paquetes internos.** `cmd/rei` y `internal/cli` solo
  despachan comandos, validan argumentos y formatean salida; la lógica de dominio
  vive en paquetes `internal/<área>` con tests propios.
- **Núcleo determinista.** Inicialización, estado, validación, checks y
  diagnóstico son deterministas: no usan IA ni dependen de un modelo. Los agentes
  son consumidores del arnés, no parte del binario.
- **El disco es la única fuente de verdad.** Planificación, progreso, metadatos y
  configuración son ficheros del repositorio. No hay base de datos ni estado en
  memoria que sobreviva al proceso.
- **Contrato de rol canónico, runtime aparte.** `.rei/agents/<rol>.md` define el
  comportamiento; `internal/adapter` solo traduce ese contrato al formato de cada
  runtime. Añadir un runtime no cambia el contrato.
- **El visor observa y reutiliza.** `rei view` es de solo lectura sobre los
  documentos; la única mutación (Aprobar) pasa por la misma transición de estado
  validada que usa el CLI.
- **Simplicidad y dependencias mínimas.** Biblioteca estándar primero; toda
  dependencia nueva debe justificarse frente a código propio.
- **Bajo acoplamiento y alta cohesión.** Cada paquete expone una API pequeña y no
  importa paquetes de presentación (`cli`, `view`, `show`).

---

# Organización del proyecto

```text
cmd/rei/                 Punto de entrada: llama a cli.Run y devuelve el exit code.
skeleton/                Esqueleto neutral de `rei init`: AGENTS.md, .rei/config.json y .rei/docs/project/**.
embed.go                 (raíz) Declara los dos embed.FS: Skeleton (skeleton/) y Harness (.rei/...).
internal/
    cli/                 Despacho de comandos, ayuda, códigos de salida.
    initwizard/          `rei init`: despliegue del esqueleto, detección de pendientes, update.
    paths/               Rutas canónicas del arnés dentro del proyecto.
    config/              Lectura/escritura de `.rei/config.json`.
    meta/                Lectura/escritura de `meta.json` de cada Work Item.
    state/               Transiciones de estado validadas (pending → ready → …).
    workitem/            Modelo y localización de Work Items.
    validate/            Validación de documentos y estructura del Work Item.
    check/               Ejecución de los checkpoints rápidos declarados en config.
    doctor/              Diagnóstico del arnés (estructura, roles, personalización).
    show/                Ficha de Work Item para la CLI (`rei item show`).
    template/            Render de las plantillas canónicas de `.rei/templates/`.
    adapter/             Contrato de rol → formato de cada runtime (opencode, claude).
    view/                Visor HTTP local: render, TOC, diff, SSE y ciclo de vida.
    view/assets/         HTML, CSS y JS propios embebidos en el binario.
    browser/             Apertura de URLs en el navegador del sistema.
    gitx/                Wrapper mínimo sobre `git` (rama, diff, commit).
    github/              Integración opcional con GitHub Issues (`gh`).
    update/              Auto-actualización desde GitHub Releases (semver, checksums).
```

La raíz del módulo solo contiene `embed.go`: `go:embed` no admite patrones con
`..`, por eso los `embed.FS` deben declararse allí. Hay dos raíces embebidas y
sin lógica:

- `Skeleton` (`skeleton/`): esqueleto neutral con el que `rei init` crea
  `AGENTS.md`, `.rei/config.json` y `.rei/docs/project/**` en el proyecto
  destino. Sus rutas llevan el prefijo `skeleton/`, que `initwizard` elimina.
- `Harness` (`.rei/docs/harness/**`, `.rei/docs/usage.md`, `.rei/agents/**`,
  `.rei/adapters/**`, `.rei/templates/**`): material del arnés que se instala
  con su ruta canónica.

Los archivos vivos del proyecto (`AGENTS.md`, `.rei/config.json` y
`.rei/docs/project/*` personalizados) no se embeben: el esqueleto es su única
copia desplegable.

---

# Flujo general

```text
Usuario / agente
    ↓
cmd/rei → internal/cli (despacho, flags, formato)
    ↓
paquetes de dominio (initwizard, meta, state, validate, check, doctor,
                     workitem, show, template, adapter, github)
    ↓
sistema de ficheros del proyecto destino (.rei/**, AGENTS.md, .rei/config.json)
```

Flujos particulares:

- **`rei init`**: `embed.FS` (`Skeleton` mapeado a la raíz + `Harness` en su ruta
  canónica) → `initwizard` → escritura sin sobrescribir existentes en el proyecto
  destino → detección de marcadores pendientes → plan de pasos.
- **Operación diaria**: `cli` → `paths`/`config`/`meta` → `state` (transición
  validada) → disco. `check` y `doctor` solo leen y ejecutan comandos.
- **`rei view`**: `view` carga el Work Item → `render` (goldmark) → `net/http` con
  assets embebidos y SSE; el botón Aprobar invoca la transición de `state` y
  notifica a los clientes conectados.
- **`rei update`**: `update` consulta GitHub Releases, verifica `checksums.txt` y
  reemplaza el ejecutable de forma atómica.

---

# Restricciones arquitectónicas

- **Sin red salvo `rei update` y `rei github`.** El resto del CLI debe funcionar
  sin conexión; ninguna ruta de inicialización, validación o estado puede depender
  de servicios externos.
- **Ninguna operación escribe fuera del proyecto destino.** Todo se limita a rutas
  derivadas de la raíz detectada por `paths`.
- **Sin base de datos.** Prohibido introducir almacenamiento opaco: el estado debe
  seguir siendo legible y versionable en Markdown/JSON.
- **Dependencias mínimas.** La única dependencia externa es `goldmark`. Cualquier
  otra requiere justificación y aprobación explícita.
- **Visor sin terceros.** Nada de JavaScript, CSS ni fuentes de CDN; los assets son
  propios y van embebidos (`internal/view/assets`).
- **El visor no muta documentos.** Solo puede ejecutar la transición de estado ya
  validada por `state`; nunca edita `requirements.md`, `design.md` ni `tasks.md`.
- **El paquete raíz no contiene lógica.** `embed.go` solo declara `embed.FS`
  (`Skeleton` y `Harness`).
- **No acoplar el contrato de rol a un runtime.** `internal/adapter` no puede
  convertirse en dependencia del núcleo.
- **Sin dependencias circulares** entre paquetes `internal/`.
- **La versión se inyecta en build time** (`-X .../internal/cli.version`), nunca se
  hardcodea en el código.

---

# Decisiones importantes

- **Lenguaje y distribución:** Go 1.27, binario único sin `CGO`, multiplataforma
  vía GoReleaser. El esqueleto se embebe con `go:embed` para que `rei init` no
  necesite clonar el repositorio.
- **Patrón de capas:** presentación (`cli`, `view`, `show`) → dominio → ficheros.
  Los paquetes de dominio no conocen la CLI.
- **Estado:** ficheros Markdown como contenido y `meta.json` como estado
  estructurado; las transiciones se centralizan en `state` para que CLI y visor
  compartan validación.
- **Visor:** servidor HTTP local (`net/http`) con `html/template` y render de
  Markdown con `goldmark`; recarga en vivo por SSE; modo primer plano y modo
  servidor de fondo con registro en `.rei/.view-server.json`.
- **Integración con GitHub:** opcional y desactivable; la ausencia de
  `integrations.github` en `.rei/config.json` significa «desactivada», y se
  gestiona solo mediante `rei github enable|disable`.
- **Errores y salida:** las funciones de dominio devuelven `error`; `cli.Run`
  traduce a códigos de salida (`0` éxito, `1` fallo real).

---

# Qué NO documentar

Este archivo NO debe contener:

- requisitos funcionales;
- decisiones temporales;
- planificación de Features;
- detalles de implementación concretos;
- cambios específicos de un Work Item.
