# Plan — Archive: incluir el veredicto de review en history

## Objetivo

Impedir que `rei session archive` cierre un Work Item que aún no está en estado
terminal, de modo que la entrada de `history.md` refleje siempre el veredicto
final (y no el estado previo del Implementer), y reordenar el contrato del
Reviewer para que el cierre ocurra en el orden correcto.

Caso real que motiva el cambio: `2026-10-25_18-00__github-issues-v1` se archivó
en `review` antes de `rei status set done`, y `history.md` quedó con «Esperando
revisión» para un Work Item ya `done`.

## Decisiones

- **El guard vive en `ArchiveSession` (`internal/template`), no en el CLI.** Así
  protege a cualquier llamador (CLI, tests, futuros comandos) y no solo a
  `rei session archive`. `cmdSession` ya convierte el error en salida 1 con
  `ERROR: ...`, por lo que el mensaje debe incluir causa y siguiente paso.
- **No se añade `--force`.** Un escape hatch para archivar estados no terminales
  reabriría exactamente el bug. Para descartar una sesión sin archivarla ya
  existe `rei session reset`; para transiciones no estándar ya existe
  `rei status set <id> <status> --force`, tras el cual el archivado sí procede.
- **`blocked` cuenta como terminal**, además de `done` (lo exige la descripción
  del Work Item).
- **La comprobación es determinista y previa a escribir `history.md`:** estado
  de `meta.json` ∈ {`done`, `blocked`} **y** `Estado` de `current.md` idéntico a
  `meta.json.status`. Cualquier otra combinación no archiva y no toca
  `history.md` ni `current.md`.
- **El veredicto en `history.md` se consigue por orden de operaciones, no por
  código nuevo:** el Reviewer vuelca el veredicto en `current.md` y pone el
  Work Item en `done` antes de archivar; el archivado copia el contenido tal
  cual (no se altera el formato de la entrada).

## Archivos

| Archivo | Acción |
|---------|--------|
| `internal/template/render.go` | Editar: guard en `ArchiveSession`. |
| `internal/template/render_test.go` | Editar: actualizar el test existente y añadir tests del guard. |
| `internal/cli/help.go` | Editar: detalle del subcomando `session archive`. |
| `.rei/agents/reviewer.md` | Editar: reordenar el **Cierre** y las reglas duras; ajustar el racional de `## Referencia`. |
| `.rei/docs/harness/progress.md` | Editar: documentar la precondición y el orden de cierre. |
| `.opencode/agents/reviewer.md` | Regenerar con `rei init opencode` (no editar a mano). |
| `.claude/agents/reviewer.md` | Regenerar con `rei init claude` (no editar a mano). |

## Cambios

### 1. Guard determinista en `ArchiveSession`

En `internal/template/render.go`, tras leer `current.md` y resolver un
`workItem` activo (y **antes** de abrir `history.md`):

1. Cargar `meta.json` con `meta.Load(p.MetaFile(workItem))`. Si falla (p. ej. no
   existe), devolver error con causa y siguiente paso (`rei session reset` para
   descartar la sesión, o corregir el Work Item).
2. Exigir `m.Status` ∈ {`meta.StatusDone`, `meta.StatusBlocked`}. Si no, error
   que indique el estado actual y el comando correctivo
   (`rei status set <id> done` **después** de registrar el veredicto en
   `current.md`).
3. Leer `Estado` de `current.md` con el helper `field` ya existente y exigir que
   sea idéntico a `m.Status`. Si difiere, error que muestre ambos valores y cómo
   resolverlo (actualizar `current.md` o ejecutar `rei status set <id> <estado>`,
   que alinea `current.md`).
4. Definir un error centinela exportado (p. ej. `ErrArchiveNotAllowed`) y
   envolver con `%w` los tres casos, para que los tests usen `errors.Is`.
5. Mantener la firma `ArchiveSession(p *paths.Project) (string, error)` y el
   comportamiento actual cuando no hay sesión activa (`"", nil`).
6. No escribir nada (ni `history.md` ni `current.md`) cuando el guard rechaza.

Importar `internal/meta` en `internal/template` no crea ciclos (`meta` solo
importa la stdlib).

### 2. Tests del guard

En `internal/template/render_test.go`:

- Helper para crear el esqueleto temporal (plantilla `current.md` con
  `Work Item`/`Estado`, `history.md` vacío) y un `meta.json` con el `status`
  indicado bajo `.rei/specs/<id>/`.
- **Actualizar `TestStartAndArchive`**: hoy archiva una sesión `pending` sin
  `meta.json`, algo que el guard ahora rechaza. Convertirlo en el camino feliz:
  crear `meta.json` con `done`, alinear `Estado` de `current.md` y comprobar que
  archiva, que `history.md` contiene la entrada con `Estado: done` y que
  `current.md` queda restablecido.
- Añadir tests (tabla) para:
  - `done` archiva; `blocked` archiva;
  - `review`, `in_progress`, `ready`, `pending` y `changes_requested` no
    archivan y devuelven `ErrArchiveNotAllowed`;
  - `current.md` incoherente (`meta=done`, `Estado=review`, y el caso inverso) no
    archiva e informa ambos valores;
  - sesión activa sin `meta.json` no archiva;
  - sin sesión activa → `("", nil)` (regresión).
- En cada caso de rechazo, comprobar que `history.md` no recibió la entrada y
  que `current.md` sigue intacto.

### 3. Contrato del Reviewer

En `.rei/agents/reviewer.md`, sección **Cierre** (aprobación):

1. Escribir `review.md` (sin cambios).
2. Actualizar `current.md`: **Estado** `done`, **Bitácora** con el veredicto y
   las verificaciones realizadas, **Próximo paso** indicando el cierre.
3. `rei status set <work-item-id> done`.
4. `rei validate <work-item-id>`.
5. `rei session archive`.
6. DETENTE.

Ajustar las reglas duras:

- Reemplazar «SIEMPRE cierra la sesión con `rei session archive` al aprobar» por
  «`rei session archive` **solo** tras `rei status set <id> done`».
- Añadir: «NUNCA archives antes de que `meta.json` esté en `done` (o `blocked`) y
  `current.md` coincida».

Actualizar el racional de `## Referencia`: `rei session archive` exige estado
terminal y coherencia con `current.md`, y copia el resumen tal cual, por lo que
el veredicto debe estar escrito **antes** de archivar.

No tocar el flujo de rechazo (`changes_requested`) ni el de bloqueo.

### 4. Documentación de progreso y ayuda

- `.rei/docs/harness/progress.md`:
  - en `current.md` (línea del cierre): indicar que `rei session archive` solo
    archiva con `meta.json` en `done`/`blocked` y `Estado` coherente, y que
    rechaza con código 1 en caso contrario;
  - en `history.md`: la entrada refleja el estado final registrado en
    `current.md` (veredicto del Reviewer), porque el archivado ocurre después de
    `rei status set done`;
  - en la tabla de **Responsabilidades**: reflejar el orden
    (veredicto en `current.md` → `status done` → `validate` → `archive`).
- `internal/cli/help.go`: detalle del subcomando `archive` →
  «archiva la sesión si el Work Item está en `done`/`blocked` y `current.md` es
  coherente; si no, falla».

### 5. Regenerar agentes nativos

El cambio del Reviewer afecta a `## Contrato`, que es lo que consumen los
adaptadores. Regenerar con `rei init opencode` y `rei init claude` (idempotentes;
solo reescriben archivos con la marca `GENERATED`) y verificar con
`rei init opencode --check` / `rei init claude --check`.

## Restricciones

- Cambio mínimo: sin dependencias nuevas, sin nuevos comandos ni flags.
- No alterar el formato de `history.md` ni reescribir entradas pasadas.
- No cambiar `rei status set`, ni auto-archivar, ni el `WARN` de `rei validate`.
- No romper el archivado cuando la sesión está correctamente cerrada
  (`done`/`blocked` + `current.md` coherente).
- No editar a mano los agentes nativos ni las plantillas de `.rei/templates/`.
- No ejecutar `rei session archive` sobre la sesión viva del repositorio durante
  la verificación: probar el guard solo en directorios temporales (tests).

## Pasos

- [x] 1. Implementar el guard en `ArchiveSession` (`internal/template/render.go`)
  con el error centinela y mensajes de causa + siguiente paso.
- [x] 2. Actualizar `TestStartAndArchive` y añadir los tests del guard en
  `internal/template/render_test.go`.
- [x] 3. Ejecutar `go test ./internal/template/` y comprobar que pasan.
- [x] 4. Reordenar la sección **Cierre** y las reglas duras de
  `.rei/agents/reviewer.md`; ajustar el racional de `## Referencia`.
- [x] 5. Actualizar `.rei/docs/harness/progress.md` y el detalle de `archive` en
  `internal/cli/help.go`.
- [x] 6. Regenerar los agentes nativos con `rei init opencode` y
  `rei init claude`.
- [x] 7. Verificar: `rei init opencode --check`, `rei init claude --check`,
  `go vet ./...`, `go test ./...`, `gofmt -l .` (sin salida) y
  `rei check --quiet`.
- [x] 8. Documentar el trabajo en
  `.rei/progress/work-items/2026-10-25_18-29__archive-review-verdict/impl.md`,
  con la evidencia de cada comando ejecutado.
