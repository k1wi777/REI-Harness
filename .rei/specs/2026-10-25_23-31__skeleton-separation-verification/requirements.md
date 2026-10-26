# Requirements

> Feature: **Separar skeleton del proyecto vivo y verificación rápida/lenta**
>
> Requisitos en EARS. Cada `R-id` es estable, implementable y verificable.

## Esqueleto desplegable

**R1** — El sistema DEBE mantener un esqueleto desplegable cuyos documentos de proyecto sean neutrales e independientes de los archivos vivos del repositorio.

**R2** — El esqueleto desplegable DEBE contener `AGENTS.md`, `.rei/config.json` y los documentos `.rei/docs/project/README.md`, `.rei/docs/project/architecture.md`, `.rei/docs/project/conventions.md` y `.rei/docs/project/verification.md`.

**R3** — CUANDO se inspecciona el `AGENTS.md` del esqueleto, el sistema DEBE contener al menos un marcador `<!-- REI:PENDIENTE -->`; la misma condición DEBE cumplirse en `architecture.md`, `conventions.md` y `verification.md`.

**R4** — El `.rei/config.json` del esqueleto DEBE declarar una lista `checks` vacía, NO DEBE contener `integrations.github` y DEBE declarar `viewer.autoOpenOnReady` en `true`, conservando el default aprobado por `2026-10-13_15-51__viewer-lifecycle`.

**R5** — El esqueleto NO DEBE contener `.rei/specs/**` ni `.rei/progress/**`.

**R6** — El esqueleto DEBE incluir la documentación vigente del harness (`.rei/docs/harness/**` y `.rei/docs/usage.md`), los roles (`.rei/agents/**`), los adaptadores (`.rei/adapters/**`) y las plantillas (`.rei/templates/**`).

**R7** — El esqueleto desplegable NO DEBE contener el contenido personalizado de `AGENTS.md`, `.rei/config.json` ni `.rei/docs/project/*` del repositorio vivo.

## Instalación y actualización

**R8** — CUANDO se ejecuta `rei init` en un proyecto sin arnés, el sistema DEBE desplegar el esqueleto en la raíz del proyecto, incluyendo `AGENTS.md`, `.rei/config.json` y los documentos de proyecto.

**R9** — CUANDO `rei init` encuentra un archivo ya existente en el proyecto destino, el sistema DEBE conservarlo sin sobrescribirlo y DEBE informarlo en la salida.

**R10** — CUANDO se ejecuta `rei init status` sobre un proyecto recién inicializado, el sistema DEBE listar los cuatro documentos de personalización como pendientes y DEBE devolver código de salida 1.

**R11** — CUANDO no queda ningún marcador de personalización, `rei init status` DEBE devolver código de salida 0 e indicar que la documentación está personalizada.

**R12** — CUANDO `rei init` detecta personalización pendiente, el sistema DEBE indicar el comando `/personalize` y el rol `initializer`.

**R13** — CUANDO se ejecuta `rei init --update` sobre un proyecto instalado, el sistema DEBE actualizar los archivos del harness que no hayan sido modificados y los documentos de personalización que aún conserven el marcador.

**R14** — CUANDO `rei init --update` encuentra un documento de personalización sin marcador o un archivo del harness modificado, el sistema DEBE conservarlo salvo que se use `--force`.

**R15** — `rei init --update` DEBE ser idempotente: una segunda ejecución consecutiva no DEBE modificar el árbol de archivos.

**R16** — CUANDO se ejecuta `rei init` o `rei init --update` sobre el repositorio vivo, el sistema DEBE conservar el contenido personalizado de `AGENTS.md`, `.rei/config.json` y `.rei/docs/project/*`.

## Verificación

**R17** — El `.rei/config.json` del repositorio vivo DEBE declarar únicamente el checkpoint rápido V1 (`go vet ./...`).

**R18** — CUANDO se ejecuta `rei check` en el repositorio vivo, el sistema DEBE ejecutar únicamente V1.

**R19** — La documentación de verificación del repositorio DEBE clasificar V1 como checkpoint rápido y V2 (`go test ./...`), V3 (`make build`) y V4 (`make cross`) como checkpoints lentos que se ejecutan antes de cerrar o revisar el Work Item.

**R20** — CUANDO se ejecuta `rei doctor` o `rei init status` en el repositorio vivo, el sistema NO DEBE reportar fallos de integridad ni personalización pendiente.

**R21** — Un proyecto limpio inicializado con el esqueleto DEBE superar `rei check` y `rei doctor` sin errores derivados de checks o integraciones del repositorio vivo.

## Calidad

**R22** — `go test ./...` DEBE pasar, incluidos los tests de instalación limpia, actualización marker-aware, neutralidad del esqueleto y exclusión de los archivos vivos.

**R23** — El cambio NO DEBE introducir dependencias externas nuevas.

**R24** — La documentación viva que describe la estructura del repositorio DEBE reflejar la existencia y el propósito de `skeleton/`.
