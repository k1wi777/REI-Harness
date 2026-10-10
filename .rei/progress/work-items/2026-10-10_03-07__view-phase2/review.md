# Review — Visor de especificaciones (Fase 2)

- **Work Item:** `2026-10-10_03-07__view-phase2` (feature)
- **Resultado:** ✅ **Aprobado** — cumple la planificación (R1–R26).
- **Base revisada:** `a27c9cab` (`base_commit`).
- **Paquete:** `rei review-diff 2026-10-10_03-07__view-phase2` (13 archivos, +892/−109).

## Verificaciones

| ID | Comando | Resultado |
|----|---------|-----------|
| V1 | `gofmt -l internal` | PASA (sin salida) |
| V2 | `make vet` | PASA |
| V3 | `make test` | PASA (todas las suites `ok`, incl. `view`, `gitx`) |
| V4 | `make build` | PASA |
| V5 | `rei check --quiet` | PASA (`exit 0`) |
| V6 | `node --check app.js` | PASA |
| V7 | literales de color fuera de `:root` | PASA (grep sin resultados) |

> `verification.md` sigue sin personalizar (placeholder): los checkpoints de facto
> son los de `tasks.md` T12 / `AGENTS.md` §3, todos en verde. `config.json`
> declara `checks: []`, por lo que `rei check` no añade checks propios.

## Trazabilidad

- **T1–T4 / R1–R8:** `gitx.ReviewBase` (last_review_commit → base_commit, validado
  con `cat-file -e`) y `FileChanges`; `changes.go` compone el diff escapado línea
  a línea con clases `dl-*`; degrada sin git (`GitAvailable=false` → "sin git") y
  sin base (`HasBase=false` → "sin base de revisión"); el índice marca
  cambiados/nuevos; colofón `cambios desde <sha> · <origen>`; documento sin
  cambios muestra "sin cambios". Confirmado en `changes_test.go`
  (`TestBuildModelSinGit`, `TestBuildModelConCambios`) y `gitx_test.go`
  (`TestReviewBase`, `TestFileChanges`).
- **T5 / R9–R12:** `toc.go` extrae encabezados del AST (nivel/id/texto);
  `page.html` emite `<nav class="toc">` solo con 2+ encabezados, jerarquía por
  `toc-lN`, enlaces `#<id>` nativos (funciona sin JS). `toc_test.go` en verde.
- **T6 / R13–R15:** SSE envuelto en `connectEvents()`; `onerror` cierra y
  reintenta con espera creciente (1s→15s) avisando "Reconectando…"; `onopen`
  limpia el aviso.
- **T7 / R16:** masthead con una única descripción; `.brand-sub` eliminado
  (plantilla y CSS). **T8 / R17:** `.doc.active` a `--ground-3` (surface-raised).
- **T9–T10 / R18–R20:** roles de color y tipografía declarados en `:root` y
  consumidos; `rgba(...)` → `color-mix(...)`; botón sin literales
  (`--on-primary`/`--accent-edge`/`--primary-hover`/`--on-warn`/`--warn-hover`);
  `DESIGN.md` añade `primary-edge` y alinea los valores realizados; `design.json`
  aditivo (ya coincidía).
- **R21–R25:** sin terceros (assets propios embebidos); sin hooks de runtime;
  sin dependencias Go nuevas (`go.mod`/`go.sum` intactos); UI en español.
  **R26:** fuera de alcance, no implementado.

## No regresión (R23)

Selector de documento, aprobación en dos pasos, enlaces profundos (documento y
sección) y conservación del scroll presentes en `app.js`; el manejo de `click`
de `.changes-toggle` no interfiere con esas rutas.

## Alineación de tokens sin cambio de apariencia

Revisado el `git diff` de `style.css`: cada literal sustituido por su variable
conserva el valor; las nuevas propiedades tipográficas (`font-weight`/`line-height`
añadidos) reproducen lo heredado (400 / 1.6). El único cambio visual es el
intencionado de R17.

## Observaciones (no bloqueantes)

- Formato y niveles de sombra/estilo coherentes con el mundo de `DESIGN.md`.
- Archivos sin rastrear ajenos al deliverable en el árbol (`package-lock.json`
  vacío, `.impeccable/**`). No forman parte del cambio comprometido; se
  recomienda limpieza/ignorado en una tarea de higiene aparte.
- Un documento rastreado añadido *y commiteado* tras la base se marca "cambios"
  (no "nuevo"), pero su diff lo presenta como material añadido íntegro, conforme
  a R8.

## Acciones requeridas

Ninguna.
