# Review — Visor de especificaciones en el navegador

> Work Item: `2026-10-09_23-45__spec-browser-viewer` (feature).
> Paquete de revisión: `rei review-diff` (base `5e86b46`), 34 archivos, +2706/-44.
> Resultado: **done**.

## Resultado

Re-revisión tras `changes_requested`. Las dos acciones requeridas quedaron
resueltas:

1. **`tasks.md` al día:** T9, T10, T11 y T12 están marcadas `[x]`. Las 17 tareas
   aparecen completadas y T17 (trazabilidad) es coherente con el contenido.
2. **Cosmético resuelto:** en `internal/cli/cli.go` el comentario doc de
   `cmdUpdate` (línea 319) volvió a su función, separado del de `cmdView`
   (línea 261); cada bloque doc queda adosado a su función.

Con ello desaparece la desviación frente a la planificación que motivó el
rechazo. La implementación cubre R1–R44, todas las verificaciones ejecutables
pasan y no quedan incumplimientos.

## Verificaciones

`verification.md` y `.rei/config.json` siguen siendo plantillas sin personalizar
(checkpoints `V*` de ejemplo, `checks: []`); se re-ejecutó la verificación
declarada en T16:

- `gofmt -l internal` → sin salida. **PASA**
- `make vet` → OK. **PASA**
- `make test` → todas las suites `ok` (incl. `internal/view`, `browser`,
  `config`, `state`, `cli`). **PASA**
- `make build` → OK. **PASA**
- `rei check --quiet` → exit 0. **PASA**
- `rei validate <id>` → `Resultado: OK`. **PASA**
- `./bin/rei help view` → documenta uso, modos y códigos de salida (R42). **PASA**

## Trazabilidad / requisitos

Sin cambios respecto a la revisión previa. El diff (nuevos `internal/browser`,
`internal/view/*`; modificados `config`/`state`/`cli`/`help`/`update`, contratos
de Leader/Initializer, `.rei/config.json`; eliminado `internal/update/browser.go`)
responde al diseño y cubre R1–R44 (comandos/modos y códigos, render seguro
goldmark, servidor+token+SSE, ciclo de vida del servidor de fondo, config
`viewer.autoOpenOnReady`, transición compartida, contratos del arnés). T1–T17
marcadas y verificadas.

## Observaciones

- **No bloqueante (WARN de tope blando):** `rei validate` avisa de que `impl.md`
  tiene 90 líneas (tope ~80) aunque 599 palabras (< 600). Es un aviso de
  verbosidad documental, no un fallo funcional ni una desviación del plan; la
  causa es la sección "Rework" añadida. No impide la aprobación. Recomendación
  futura: condensar `impl.md` al archivarse.
- **Artefacto de entorno:** `/usr/local/bin/rei` está desactualizado (no
  reconoce `view`); el binario recién construido `./bin/rei` sí lo expone. Ajeno
  al Work Item.
- `package-lock.json` sin rastrear en la raíz: preexistente (mtime 2026-09-10),
  ajeno al Work Item.

## Acciones requeridas

Ninguna. El Work Item queda aprobado.
