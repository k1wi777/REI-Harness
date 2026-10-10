# Requirements — Visor de especificaciones en el navegador

> Work Item: `2026-10-09_23-45__spec-browser-viewer` (feature).
>
> Los requisitos describen **qué** debe hacer el sistema, no **cómo**. Se redactan
> en EARS. El sujeto "el sistema" designa al CLI `rei` o a los Contratos de rol
> del arnés, según el requisito.

---

# 1. Comando y modos

**R1.** El sistema DEBE ofrecer el comando `rei view [<id>]`.

**R2.** CUANDO se ejecuta `rei view` sin `<id>` y existe una sesión activa, el sistema DEBE mostrar el Work Item de la sesión activa.

**R3.** SI se ejecuta `rei view` sin `<id>` y no existe una sesión activa, ENTONCES el sistema DEBE informar de la ausencia de sesión y devolver un código de salida distinto de `0`.

**R4.** El sistema DEBE admitir, además del modo de primer plano (por defecto), los modos `--open`, `--stop`, `--enable` y `--disable`, que DEBEN ser mutuamente excluyentes.

**R5.** SI `rei view` recibe un modo desconocido o una combinación de modos incompatible, ENTONCES el sistema DEBE mostrar el uso por `stderr` y devolver código de salida `2`.

**R6.** CUANDO se ejecuta `rei view` en primer plano, el sistema DEBE servir el visor y mantener el proceso activo hasta que el usuario lo interrumpa (Ctrl+C).

**R7.** CUANDO se ejecuta `rei view <id> --open`, el sistema DEBE arrancar el servidor en segundo plano, abrir el navegador y devolver el control sin bloquear el proceso ni la conversación.

**R8.** CUANDO se ejecuta `rei view --stop`, el sistema DEBE detener el servidor en segundo plano del proyecto.

**R9.** CUANDO se ejecuta `rei view --enable`, el sistema DEBE activar la apertura automática del visor; CUANDO se ejecuta `rei view --disable`, el sistema DEBE desactivarla.

---

# 2. Contenido mostrado

**R10.** El visor DEBE mostrar los metadatos del Work Item (`id`, `title`, `type`, `status`, `created_at`, `base_commit`, `last_review_commit`).

**R11.** El visor DEBE mostrar los documentos de planificación existentes del Work Item: `requirements.md`, `design.md` y `tasks.md` para un `feature`, o `plan.md` para un `task`.

**R12.** El visor DEBE mostrar los documentos de progreso existentes del Work Item: `impl.md`, `review.md` y `spec.md`.

**R13.** El visor DEBE mostrar el resultado de la validación del Work Item equivalente a `rei validate <id>`.

**R14.** El visor DEBE mostrar el estado de la sesión cuando el Work Item mostrado sea la sesión activa.

**R15.** El visor DEBE renderizar el contenido de los documentos Markdown a HTML, sin mostrarlos en crudo.

**R16.** SI un documento esperado no existe, ENTONCES el visor DEBE indicarlo explícitamente sin fallar la carga de la página.

---

# 3. Render y seguridad

**R17.** El render de Markdown DEBE realizarse con la biblioteca `goldmark` en Go y NO DEBE interpretar HTML en crudo contenido en los documentos (modo seguro por defecto).

**R18.** El visor DEBE usar únicamente JavaScript propio mínimo; NO DEBE cargar JavaScript de terceros.

---

# 4. Recarga en vivo

**R19.** MIENTRAS el visor está abierto, CUANDO cambia el contenido de cualquiera de los documentos mostrados del Work Item, el navegador DEBE recargar automáticamente mediante Server-Sent Events (SSE).

**R20.** La detección de cambios DEBE limitarse a los documentos del Work Item mostrado.

---

# 5. Botón Aprobar

**R21.** CUANDO el Work Item mostrado está en estado `ready`, el visor DEBE ofrecer un botón "Aprobar".

**R22.** CUANDO el usuario pulsa "Aprobar", el sistema DEBE aplicar la transición validada `ready -> in_progress` sobre `meta.json`, reutilizando la misma validación de transición que `rei status set`.

**R23.** CUANDO se aplica una aprobación y el Work Item es la sesión activa, el sistema DEBE alinear el campo **Estado** de `.rei/progress/current.md` con el nuevo estado.

**R24.** SI la transición `ready -> in_progress` no está permitida para el Work Item, ENTONCES el sistema NO DEBE modificar `meta.json` y DEBE informar del motivo.

**R25.** El botón "Aprobar" NO DEBE introducir un modelo de aprobación ni un estado paralelos: el registro durable es `meta.json` y la sesión es `.rei/progress/current.md`.

**R26.** CUANDO se aprueba desde el botón, el visor DEBE indicar al usuario que debe reanudar la conversación en el chat con una palabra breve (limitación del modelo por turnos).

---

# 6. Servidor en segundo plano

**R27.** El servidor DEBE escuchar únicamente en `127.0.0.1` y en un puerto efímero asignado por el sistema operativo.

**R28.** El sistema DEBE generar un token secreto por instancia de servidor, incluirlo en la URL y exigirlo para acceder al visor.

**R29.** CUANDO se ejecuta `rei view <id> --open` y ya existe un servidor en segundo plano vivo para el proyecto, el sistema DEBE reutilizarlo en lugar de arrancar otro.

**R30.** El servidor en segundo plano DEBE cerrarse automáticamente cuando el Work Item mostrado deja de estar en estado `ready`.

**R31.** El servidor en segundo plano DEBE cerrarse automáticamente tras un periodo de inactividad sin peticiones.

**R32.** CUANDO se ejecuta `rei view --stop`, el sistema DEBE terminar el servidor y eliminar su registro temporal.

---

# 7. Configuración

**R33.** `.rei/config.json` DEBE admitir la clave `viewer.autoOpenOnReady` (booleano) y su valor por defecto DEBE ser `false`.

**R34.** SI la clave `viewer` o `autoOpenOnReady` no existe en `.rei/config.json`, ENTONCES el sistema DEBE asumir `false` sin fallar.

**R35.** CUANDO se ejecuta `rei view --enable` o `--disable`, el sistema DEBE actualizar únicamente la clave `viewer.autoOpenOnReady`, preservando el resto del archivo.

---

# 8. Disparo desde el arnés

**R36.** El `## Contrato` del Leader DEBE indicar que, en la puerta de aprobación (Caso A), si `viewer.autoOpenOnReady` está activo, ejecute `rei view <work-item-id> --open` sin bloquear la conversación.

**R37.** MIENTRAS `viewer.autoOpenOnReady` está desactivado, el flujo de aprobación DEBE ser idéntico al actual, sin visor.

**R38.** El disparo del visor DEBE realizarse mediante comando CLI desde el `## Contrato` del Leader; NO DEBE depender de hooks nativos de un runtime como mecanismo central.

**R39.** El `## Contrato` del Initializer DEBE preguntar al usuario si desea activar la apertura automática del visor y DEBE registrar el valor en `viewer.autoOpenOnReady` de `.rei/config.json`.

---

# 9. Dependencias, ayuda e integridad

**R40.** El sistema DEBE incorporar `goldmark` como única dependencia Go nueva y DEBE evitar dependencias transitivas.

**R41.** Los nuevos componentes DEBEN ser verificables con `go test`, incluido el servidor HTTP (mediante `httptest`).

**R42.** `rei help` DEBE listar el comando `view` y `rei help view` DEBE documentar sus modos y códigos de salida.

**R43.** El visor NO DEBE crear estado durable paralelo al Work Item; cualquier registro operativo del servidor DEBE ser temporal y eliminarse al cerrar.

**R44.** La aprobación por chat DEBE seguir funcionando exactamente igual que antes de esta Feature.
