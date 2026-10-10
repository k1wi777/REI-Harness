# Tasks — Visor de especificaciones (Fase 2)

> Work Item: `2026-10-10_03-07__view-phase2` (feature).
>
> Orden de ejecución. Cada tarea indica los requisitos que implementa.
> El Implementer marca `[x]` **inmediatamente** al terminar cada tarea.

---

## Fase 1 — Ver qué cambió desde la última revisión

- [x] **T1 — `internal/gitx`: base de revisión y diff por archivo.** Extraer `ReviewBase(p, id) (base, origin string, ok bool, err error)` del `switch` de `ReviewDiff` y refactorizar `ReviewDiff` para usarlo (sin cambiar su salida). Añadir `FileChange{Changed, New bool; Added, Removed int; Diff string}` y `FileChanges(p, base, relPaths)` reutilizando `gitOut`: sin rastrear → `New` y todo añadido; rastreado → `git diff --numstat --no-color` y, si cambió, `git diff --no-color`. No filtrar `.rei/progress/`. (R1, R2, R8)

- [x] **T2 — `internal/view`: modelo y render del diff.** En `render.go`, convertir `renderMarkdown` en `renderDoc` (parseo único del AST → HTML + encabezados; ver T4) y ampliar `docView` (`Changed`, `New`, `Added`, `Removed`, `Diff`) y `viewModel` (`GitAvailable`, `HasBase`, `BaseSHA`, `BaseOrigin`). En `changes.go`, resolver `gitx.ReviewBase` y `gitx.FileChanges` para los siete documentos (rutas relativas a la raíz) y renderizar el diff unificado a HTML escapado con clases `dl-add`/`dl-del`/`dl-hunk`/`dl-meta`/`dl-ctx`. Degradar sin git (`GitAvailable=false`) o sin base (`HasBase=false`). (R1–R8)

- [x] **T3 — `page.html` y `app.js`: modo "Cambios".** En el índice, marcar los documentos cambiados/nuevos; en el colofón, el ítem `cambios desde <sha> (<origen>)` con las notas `Sin git` / `sin base`; en cada `.doc-body` con cambios, el botón `changes-toggle` (`aria-pressed`) y el bloque `.doc-diff` con el recuento `+A / −R` (o "(sin cambios)"). En `app.js`, delegación de `click` para alternar `show-changes` y actualizar `aria-pressed`; sin JS el diff se muestra tras el contenido. Añadir el helper `short` al `funcMap`. (R2, R3, R4, R5, R6, R7)

- [x] **T4 — `style.css`: estilos de cambios.** Estilo del indicador del índice (insignia con `--accent-dim`), del bloque de diff y de sus líneas (añadido/eliminado/hunk), coherentes con el mundo (mono, sin sombras, color de señal solo para estado). (R3)

## Fase 2 — Índice de secciones (TOC)

- [x] **T5 — `internal/view/toc.go` + plantilla.** Extraer los encabezados del AST (`ast.Walk` sobre `*ast.Heading`: `Level`, atributo `id`, texto de `*ast.Text`/`*ast.String`) y poblarlos en `docView.Sections`. Renderizar en `page.html` un `<nav class="toc">` por documento con la lista anidada por nivel (`toc-l1`…), con enlaces `#<id>`, solo si hay 2+ encabezados. Estilos del TOC en `style.css`. (R9, R10, R11, R12)

## Fase 3 — Reconexión de la recarga en vivo

- [x] **T6 — `app.js`: reconexión del SSE.** Envolver el `EventSource` en `connect()`: en `onerror`, cerrar, avisar en `#notice` ("Reconectando…") y reintentar con espera creciente (de ~1 s a un tope de ~15 s); en `onopen`, limpiar el aviso y reiniciar la espera. Conservar la recarga al recibir `reload`. (R13, R14, R15)

## Fase 4 — Pulido visual

- [x] **T7 — Masthead sin redundancia.** Eliminar el subtítulo duplicado del brand en `page.html` y su regla en `style.css` (`.brand-sub`), conservando una única descripción (la leyenda derecha). (R16)

- [x] **T8 — Contraste del documento activo.** En `style.css`, `.doc.active { background: var(--ground-3); }` (antes `--ground-2`) y ajustar el color del enlace activo si hiciera falta para mantener el contraste. (R17)

## Fase 5 — Alineación de tokens

- [x] **T9 — Tokens de color y literales.** Añadir a `:root` los roles de color que faltan (`--on-primary`, `--primary-hover`, `--on-warn`, `--warn-hover`, `--scrollbar-hover`, `--on-accent`, `--accent-edge`) y eliminar todo literal de color fuera de `:root`: botón (`#f3f0ff`, `#6f5ae0`, `#6755d8`, `#241a06`, `#e9b351`), scrollbar (`#4a515e`), selección (`#fff`) y los `rgba(...)` derivados vía `color-mix(...)`. Usar `--accent-dim` en la insignia de cambios. Añadir `primary-edge` a `colors` en `DESIGN.md`. (R18, R19, R20)

- [x] **T10 — Tokens de tipografía y alineación de `DESIGN.md`.** Declarar una variable por rol tipográfico y consumirla en la regla correspondiente (tabla de 3.6 en `design.md`), sin alterar el resultado visual. Ajustar en `DESIGN.md` los valores de rol divergentes a los realizados (p. ej. `display` line-height 1.25, `title` 1.32rem, `ui` peso 700, `base` line-height 1.6) y documentar los componentes TOC y cambios. Verificar `design.json` (solo aditivo si hace falta). (R18, R20)

## Fase 6 — Verificación

- [x] **T11 — Tests.** `internal/gitx/gitx_test.go`: `ReviewBase` (last_review_commit, respaldo base_commit, sin base) y `FileChanges` (cambiado, sin cambios, nuevo) con un repo git temporal (`t.Skip` si no hay git). `internal/view`: `renderDoc` (HTML + secciones), composición con cambios y degradación sin git/sin base, `/content` con la información de cambios; ajustar `render_test.go`/`server_test.go` a los nuevos campos. (R1–R15)

- [x] **T12 — Validación final** y corrección de lo que falle:
  `gofmt -l internal` (sin salida), `make vet`, `make test`, `make build`,
  `rei check --quiet` (`exit 0`).
  Verificación manual (con un Work Item cuyo `last_review_commit`/`base_commit` esté fijado):
  - el índice marca los documentos cambiados y el colofón muestra el base usado;
  - el modo "Cambios" muestra el diff por documento; un documento sin cambios lo dice;
  - un Work Item sin base o sin git degrada con aviso, sin romper la página;
  - el TOC de cada documento enlaza a sus secciones y respeta la jerarquía;
  - al morir y revivir el servidor, la página avisa "Reconectando…" y se recupera;
  - el masthead muestra una sola descripción; el activo del índice contrasta más;
  - el selector, la aprobación en dos pasos, los enlaces profundos y el scroll siguen funcionando;
  - no hay literales de color en `style.css` fuera de `:root`. (R1–R26)

- [x] **T13 — Marcar como completadas (`[x]`)** todas las tareas de este archivo a medida que se terminan, dejando `tasks.md` al día. (trazabilidad)
