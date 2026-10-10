# Implementation — Visor de especificaciones (Fase 2)

> Work Item: `2026-10-10_03-07__view-phase2` (feature).
> Implementación completa siguiendo `tasks.md` (T1–T13, R1–R26).

## Resumen

Segunda fase del visor `rei view` sobre la v1. Se añade el modo "Cambios" por
documento (diff respecto de `last_review_commit`; en su defecto `base_commit`),
el índice de secciones (TOC) por documento, la reconexión del SSE, el masthead
sin redundancia, más contraste del documento activo y la alineación de tokens
(`DESIGN.md`/`design.json` ⇄ `style.css`). Sin dependencias Go nuevas: solo la
biblioteca estándar y `goldmark`. El HTML de los documentos nunca se interpreta:
el diff se escapa línea a línea.

## Archivos

**Nuevos**

- `internal/view/changes.go` + `changes_test.go`: base de revisión, cambios por
  documento y render del diff a HTML escapado.
- `internal/view/toc.go` + `toc_test.go`: extracción de encabezados del AST.

**Modificados**

- `internal/gitx/gitx.go` + `gitx_test.go`: `ReviewBase` (extraído de
  `ReviewDiff`, que ahora lo reutiliza) y `FileChange`/`FileChanges`.
- `internal/view/render.go`: `renderDoc` (parseo único → HTML + secciones),
  nuevos campos del modelo y helper `short`.
- `internal/view/assets/page.html`: masthead sin redundancia; base de revisión e
  insignias en el índice; TOC y bloque de diff por documento; colofón con
  `cambios desde`.
- `internal/view/assets/app.js`: reconexión del `EventSource` con espera
  creciente (1s→15s) y alternancia del modo "Cambios".
- `internal/view/assets/style.css`: tokens de color y tipografía, contraste del
  activo (`--ground-3`), estilos de TOC y diff; se elimina `.brand-sub`.
- `internal/view/render_test.go` / `server_test.go`: adaptados al nuevo contrato.
- `DESIGN.md`: token `primary-edge`, valores tipográficos realizados y
  componentes TOC/cambios documentados.

## Cambios por fase

- **F1 (T1–T4):** `gitx.ReviewBase` (last_review_commit → base_commit, validado
  con `cat-file`) y `FileChanges`; el diff viaja en `/content` (sin rutas nuevas)
  y degrada sin git o sin base con aviso en el colofón.
- **F2 (T5):** TOC servido por documento (`ast.Walk` sobre `*ast.Heading`), solo
  con 2+ encabezados.
- **F3 (T6):** SSE reconectable con `onerror`/`onopen` ("Reconectando…").
- **F4 (T7–T8):** masthead con una única descripción; activo a `--ground-3`.
- **F5 (T9–T10):** roles de color y tipografía como variables en `:root`, cero
  literales de color fuera de él (los `rgba` pasan a `color-mix`), y `DESIGN.md`
  alineado a los valores realizados.
- **F6 (T11–T12):** tests de gitx y del visor; validación final.

## Verificación

- `gofmt -l internal` → sin salida. **PASA**
- `make vet` → OK. **PASA**
- `make test` → todas las suites `ok` (incl. `internal/view`, `internal/gitx`).
  **PASA**
- `make build` → OK. **PASA**
- `rei check --quiet` → exit 0. **PASA**
- Comprobación de literales de color fuera de `:root` en `style.css`: sin
  resultados. **PASA**
- Manual (render real del propio Work Item vía test temporal): base `a27c9cab`
  (`base_commit`); requirements/design/tasks marcados como nuevos con su diff;
  TOC anidado por niveles; insignias "nuevo"; botón `Cambios +A / −R`; colofón
  `cambios desde a27c9cab · base_commit`; masthead con una sola descripción;
  diff escapado (`&gt;`, sin HTML interpretado). **PASA**
- `node --check internal/view/assets/app.js` → OK. **PASA**

## Observaciones

- `gofmt`/`go` no estaban en el `PATH`; se usó el toolchain
  `~/.local/go-toolchain/go` (fuera del repo).
- `DESIGN.md` `headline.lineHeight` se fijó al valor realizado (1.6, heredado);
  no se alteró la apariencia en ningún token (D8: gana el valor realizado).
- `design.json` no requirió cambios: sus valores canónicos ya coincidían y
  `primary-edge` (#6f5ae0) ya aparecía en su `tonalRamp`.
