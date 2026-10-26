# Plan — Specs concisas, issue body sin design y contexto bajo control

## Objetivo

Fijar un presupuesto explícito de planificación (Feature con `requirements.md`/`design.md`/`tasks.md`, Task con `plan.md`), hacer que `rei validate` lo vigile con un `WARN` blando y corregir dos hábitos que inflan el contexto: duplicar el `design.md` completo en la issue y leer/escanear documentos que no corresponden a la etapa.

## Archivos

Docs del arnés (se embeben vía `Harness` en `embed.go`, sin cambios de código):

- `.rei/docs/harness/specs.md`
- `.rei/docs/harness/task.md`
- `.rei/docs/harness/workflow.md`
- `.rei/docs/harness/progress.md`
- `.rei/docs/harness/integrations/github/templates.md`
- `.rei/docs/harness/integrations/github/core.md`
- `.rei/docs/harness/integrations/github/backlog.md` (§3, publicación de la spec)
- `.rei/docs/harness/integrations/github.md` (tabla de marcadores)
- `.rei/agents/leader.md`
- `.rei/agents/spec_author.md`
- `.rei/agents/implementer.md`
- `.rei/agents/reviewer.md`
- `AGENTS.md` (§9)

Código y tests:

- `internal/validate/validate.go`
- `internal/validate/validate_test.go`

No se toca `skeleton/`, `internal/cli/`, comandos, flujos, labels ni marcadores existentes.

## Cambios

### 1. Presupuesto en `specs.md`

- Añadir una sección «Presupuesto de planificación» al principio, con la tabla Feature: requisitos objetivo 10–20 y máximo 25–30; `design.md` objetivo 150–300 líneas y máximo 400; `tasks.md` objetivo 8–12 y máximo 15. Regla única: si se excede, **dividir el Work Item o justificarlo en una línea dentro del propio documento**.
- `requirements.md`: requisitos centrales, observables y verificables; prohibir relleno EARS y repetir el prompt/`description`; cada requisito con una única responsabilidad.
- `design.md`: solo decisiones con peso, alternativas descartadas y archivos afectados; sin código, boilerplate, arquitectura general ni repetición de requisitos.
- `tasks.md`: pasos ordenados, concretos y ejecutables; sin tareas de relleno ni duplicadas.
- Sección Task: criterio de tipo (documentación, sin cambio de comportamiento o cambio pequeño) y remisión a `task.md`; comportamiento o arquitectura ⇒ Feature.
- Nota: `rei validate` emite `WARN` (nunca `FAIL`) al superar el máximo en estados de planificación activos.

### 2. Presupuesto en `task.md`

- Explicitar cuándo usar Task: documentación/contratos, sin cambio de comportamiento o cambio pequeño y localizado; si hay comportamiento o arquitectura, es Feature.
- `plan.md` ≤150–200 líneas; sin `requirements.md` ni `design.md`; mantener el formato actual (Objetivo, Archivos, Cambios, Restricciones, Pasos) y la regla de marcar pasos al terminarlos.
- Retirar cualquier expectativa de diseño detallado; el plan documenta solo lo necesario para implementar.

### 3. Criterio de tipo y presupuesto en `workflow.md`

- En «Tipos de Work Item», añadir que **documentación, contratos o cambios pequeños sin cambio de comportamiento se planifican como Task**; un cambio de comportamiento o de arquitectura es Feature. Mantener la regla de desempate actual (duda ⇒ Feature) matizada por ese criterio.
- Añadir una referencia breve al presupuesto (`specs.md`/`task.md`) y a que `rei validate` avisa (`WARN`) sin bloquear.

### 4. Roles

- `leader.md`:
  - «Antes de delegar»: elegir Task por defecto en documentación/sin comportamiento; no inflar specs ni convertir una Task en Feature; el presupuesto vive en `specs.md`/`task.md`.
  - Contexto: no leer `history.md`, no escanear `.rei/specs/**`; usar `rei session`/`rei items status` y abrir solo el Work Item activo o el indicado.
  - Integración GitHub: sustituir «publica el resumen y `design.md` plegado» por «publica el resumen y las decisiones con peso; el design completo solo como comentario idempotente si `.rei` no se versiona o el usuario pide trazabilidad».
- `spec_author.md`:
  - Casos A y B: respetar el presupuesto; si se excede, dividir o justificar; no rellenar para llegar al objetivo ni repetir el prompt.
  - Reglas duras: no escribir `history.md` ni escanear specs; leer solo el Work Item activo.
- `implementer.md`/`reviewer.md`:
  - Leer solo el Work Item activo; no `history.md` (salvo petición explícita del usuario), no escanear `.rei/specs/**`.
  - Reviewer: verificar el presupuesto además de la implementación; un `WARN` de presupuesto exige justificación en el documento o división, pero no es rechazo automático.

### 5. `rei validate` (WARN blando)

En `internal/validate/validate.go`:

- Constantes: `maxFeatureRequirements = 30`, `maxDesignLines = 400`, `maxTasks = 15`, `maxPlanLines = 200` (junto a los topes de reportes, documentando objetivo y máximo).
- Helper `budgetApplicable(status)`: solo `ready`, `in_progress`, `review`, `changes_requested` (no `done`/`pending`/`blocked`). No reutilizar `isActiveStatus`, que excluye `ready`.
- Feature: contar requisitos con `regexp` sobre líneas `- **Rn**`/`Rn`, líneas de `design.md` con `reportSize`, y checkboxes `- [ ]`/`- [x]` de `tasks.md`. Si el archivo no existe o no se puede leer, omitir.
- Task: líneas de `plan.md` con `reportSize`.
- Mensajes con el presupuesto y la acción: p. ej. `requirements.md excede el presupuesto (máximo 30; objetivo 10–20): 34 requisitos. Justifica o divide.` Análogos para `design.md`, `tasks.md` y `plan.md`. Siempre `LevelWarn`, nunca `LevelFail`; no alterar el resto de comprobaciones ni el código de salida (los `WARN` ya no fallan).

### 6. Tests de `validate`

En `internal/validate/validate_test.go`, con el patrón existente (`writeMeta`, `mustWrite`, `countLevel`):

- Feature en `ready` dentro del presupuesto (30 requisitos, 400 líneas de design, 15 tareas): 0 `FAIL` y ningún `WARN` que contenga «presupuesto».
- Feature en `ready` fuera del presupuesto (31 requisitos, 401 líneas, 16 tareas): 0 `FAIL` y 3 `WARN` de presupuesto.
- Task en `ready` con `plan.md` de 201 líneas: 0 `FAIL` y `WARN` de presupuesto.
- Feature en `done` fuera del presupuesto (con `impl.md`/`review.md`): sin `WARN` de presupuesto, para fijar el filtro de estados.

### 7. Issue body sin design (`templates.md`, `core.md`, `backlog.md`)

- `templates.md §1`: el cuerpo queda en contexto/decisión, requisitos centrales, decisiones con peso, fuera de alcance y referencias (`.rei/specs/<id>/`); eliminar el `<details>` con `design.md` íntegro. Nota: si `.rei` no se versiona o el usuario pide trazabilidad completa, publicar el design como **comentario** idempotente `<!-- rei:comment:design:<id> -->`, nunca en el cuerpo.
- `templates.md §2`: el bloque de spec publicado deja de incrustar el design; resume requisitos/tareas y enlaza; mismo criterio de comentario opcional.
- `core.md`: en la creación de issue, indicar que el cuerpo sigue §1 **sin design completo**; en la publicación de spec, remitir al criterio de comentario idempotente.
- `backlog.md §3`: leer requisitos y decisiones; construir el bloque sin design; publicar/actualizar el comentario de design solo bajo la condición anterior, reutilizando el marcador (sin duplicar).
- `github.md`: añadir a la tabla de marcadores el comentario de design opcional.

### 8. Reglas de contexto

- `AGENTS.md §9`: no leer `.rei/progress/history.md` salvo petición explícita; no escanear `.rei/specs/**`; usar `rei items status`/`rei item show` y abrir solo el Work Item activo o el indicado.
- `progress.md`: misma regla en «Reglas»; `history.md` solo a petición del usuario.
- Roles (`leader`, `spec_author`, `implementer`, `reviewer`): misma regla, coherente con lo indicado en el punto 4.

## Restricciones

- Sin dependencias nuevas (solo biblioteca estándar; `regexp` ya es estándar).
- Sin cambios de flujo, comandos, labels ni marcadores existentes; el marcador de design es aditivo y opcional.
- No tocar `skeleton/`; los docs del arnés se embeben automáticamente vía `Harness`.
- Los `WARN` nunca cambian el código de salida ni provocan `FAIL`.
- Mantener `go test ./...`, `go vet ./...`, `make build`, `rei check`, `doctor`/`init status`/`validate` en verde.

## Pasos

- [x] 1. Actualizar `specs.md` con el presupuesto Feature, las reglas de concisión y el criterio Task.
- [x] 2. Actualizar `task.md` (cuándo Task, `plan.md` ≤150–200 líneas, sin requirements/design).
- [x] 3. Actualizar `workflow.md` con el criterio de tipo y la referencia al presupuesto.
- [x] 4. Actualizar `leader.md`, `spec_author.md`, `implementer.md` y `reviewer.md` (Task por defecto, presupuesto, contexto y design fuera de la issue).
- [x] 5. Implementar en `internal/validate/validate.go` constantes, filtro de estados y `WARN` de presupuesto Feature/Task.
- [x] 6. Añadir los cuatro tests de presupuesto en `internal/validate/validate_test.go`.
- [x] 7. Ajustar `templates.md`, `core.md`, `backlog.md` y `github.md` para el cuerpo sin design y el comentario idempotente.
- [x] 8. Añadir las reglas de contexto en `AGENTS.md §9` y `progress.md`.
- [x] 9. Verificar: `gofmt`, `go test ./...`, `go vet ./...`, `make build`, `rei check`, `rei validate` de un Feature y una Task dentro/fuera de presupuesto.
