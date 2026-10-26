# Requisitos — Visor: estados reales, FICHA con enlaces y cierre en done

## Estados y etiquetas

**R1 — Etiquetas de estado.** El sistema DEBE mostrar cada estado del Work Item
con la etiqueta: `pending` → «pendiente», `ready` → «en aprobación»,
`in_progress` → «en curso», `review` → «en revisión», `changes_requested` →
«cambios solicitados», `done` → «terminado» y `blocked` → «bloqueado». Un valor
desconocido DEBE mostrarse sin traducir.

**R2 — Coherencia badge/FICHA.** El sistema DEBE usar la misma etiqueta de estado
en el badge de la cabecera y en la fila «estado» de la FICHA para un mismo
Work Item.

## Actualización en vivo

**R3 — Estado en vivo.** CUANDO cambia el estado de un Work Item mientras el
visor está abierto, el sistema DEBE actualizar el badge y la FICHA sin recargar
la página manualmente ni dejar el estado anterior visible.

**R4 — Enlaces en vivo.** CUANDO se crea, modifica o elimina `links.json`
mientras el visor está abierto, el sistema DEBE actualizar la FICHA de enlaces.

**R5 — Continuidad de lectura.** CUANDO el visor refresca su contenido, el
sistema DEBE conservar el documento activo, las posiciones de desplazamiento del
lector, del índice y de la FICHA, el estado abierto/cerrado de los diffs, el TOC,
la conexión SSE y el enlace profundo (documento o sección) vigente.

## Ficha de enlaces

**R6 — Enlaces válidos.** DONDE `links.json` es válido, la FICHA DEBE mostrar el
issue (número y URL), la base (`branch@sha`) y la épica (slug, número y URL).

**R7 — Enlaces ausentes o inválidos.** SI `links.json` no existe, no se puede
leer o no supera la validación de enlaces, ENTONCES la FICHA DEBE mostrar una
nota explícita que indique que el índice no está disponible o es inválido, sin
enlaces rotos y sin impedir el resto de la página.

**R8 — Enlaces seguros.** El sistema DEBE renderizar como enlace únicamente las
URL de `links.json` que sean http/https absolutas y válidas, y DEBE escapar
cualquier otro contenido.

## Cierre en `done`

**R9 — Estado final.** CUANDO el Work Item llega a `done` con el servidor de
fondo activo, el sistema DEBE emitir a las páginas conectadas el estado final
«terminado» junto con el contenido actualizado.

**R10 — Cierre posterior al render.** El sistema DEBE mantener el servidor y el
contenido accesibles hasta que la página haya podido renderizar el estado final,
y solo entonces cerrar el servidor. El cierre NO DEBE adelantarse al evento
final.

**R11 — Página en done.** CUANDO el servidor se cierra tras `done`, la página
DEBE quedar mostrando «terminado» y el contenido final, sin volver a «en
revisión» ni entrar en un bucle de reconexión.

**R12 — Ciclo de vida conservado.** CUANDO el usuario ejecuta `--stop` o expira
el tiempo de inactividad, el sistema DEBE cerrar el servidor aunque el Work Item
no haya llegado a `done`, y ambos mecanismos DEBEN seguir funcionando tras
`done`.

**R13 — Modo primer plano.** MIENTRAS el visor se ejecuta en primer plano, el
sistema DEBE reflejar el estado `done` y seguir sirviendo hasta Ctrl+C, sin
cerrarse automáticamente.

## Solo lectura y compatibilidad

**R14 — Sin aprobación.** El visor DEBE seguir siendo de solo lectura: NO DEBE
exponer botón ni ruta de aprobación, y ninguna ruta de actualización o
confirmación DEBE modificar documentos, `meta.json` ni el estado del Work Item.

**R15 — Documentación.** El sistema DEBE actualizar `README`, `.rei/docs/usage.md`
y la ayuda de `rei view` cuando mencionen los estados, la FICHA o el cierre del
visor, para que describan el comportamiento resultante.

**R16 — Sin regresiones.** El sistema DEBE conservar autenticación por token y
cookie, render seguro de Markdown, TOC, diff, validación, sesión, enlaces
profundos, assets embebidos y los modos `--open/--stop/--enable/--disable`.
