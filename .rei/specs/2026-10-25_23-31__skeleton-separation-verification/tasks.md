# Tasks

> Implementación ordenada de la Feature. El Implementer marca cada tarea como
> completada inmediatamente al terminarla.

- [x] **T1 — Crear `skeleton/` neutral** (R1, R2, R3, R4, R5, R7)
  - Crear `skeleton/AGENTS.md` desde `git show HEAD:AGENTS.md`.
  - Crear `skeleton/.rei/docs/project/README.md`, `architecture.md`,
    `conventions.md` y `verification.md` desde `git show HEAD:<ruta>`.
  - Escribir `skeleton/.rei/config.json` con `checks: []`, sin `integrations` y
    con `viewer.autoOpenOnReady: true`.
  - Verificar que `AGENTS.md` y los tres documentos de proyecto contienen
    `<!-- REI:PENDIENTE -->` y que no se crea `skeleton/.rei/specs` ni
    `skeleton/.rei/progress`.

- [x] **T2 — Separar las dos raíces embebidas en `embed.go`** (R1, R6, R7)
  - Declarar `Skeleton` con `//go:embed all:skeleton`.
  - Declarar `Harness` con `all:.rei/docs/harness`, `all:.rei/docs/usage.md`,
    `all:.rei/agents`, `all:.rei/adapters` y `all:.rei/templates`.
  - Dejar de embeber `AGENTS.md`, `.rei/config.json` y `.rei/docs/project/*`
    vivos; actualizar el comentario del paquete/FS.
  - Mantener el paquete raíz sin lógica.

- [x] **T3 — Mapear el esqueleto en `InstallSkeleton`** (R8, R9)
  - Añadir `internal/initwizard/embedded.go` con `walkEmbedded`, que recorre
    `skeleton/` (quitando el prefijo) y `Harness` (ruta intacta) y entrega
    `(dest, data)`.
  - Sustituir el `fs.WalkDir` de `installer.go` por `walkEmbedded`, usando `dest`
    en mensajes, aviso de `AGENTS.md`, comprobación de existencia y manifiesto.
  - Conservar la política de no sobrescritura y los contadores/resumen.

- [x] **T4 — Mapear el esqueleto en `Update`** (R13, R14, R15)
  - Sustituir el `fs.WalkDir` de `update.go` por `walkEmbedded`.
  - Mantener la regla marker-aware para documentos de personalización, el
    baseline del manifiesto y `--force`.
  - Comprobar que las claves del manifiesto siguen siendo rutas destino.

- [x] **T5 — Actualizar `embed_test.go` y añadir neutralidad** (R2, R3, R4, R5,
  R6, R7, R22)
  - Repartir los asserts entre `Skeleton` (prefijo `skeleton/`) y `Harness`.
  - Añadir asserts de config neutral y de marcadores en los documentos
    neutrales.
  - Añadir asserts de que ningún FS contiene `AGENTS.md`,
    `.rei/config.json` ni `.rei/docs/project/*` vivos, y de que no hay
    `.rei/specs` ni `.rei/progress`.
  - Verificar que `Harness` contiene harness docs, `usage.md`, agentes,
    adaptadores y plantillas.

- [x] **T6 — Ajustar los tests de `internal/adapter`** (R6, R22)
  - Cambiar las lecturas de roles/adaptadores de `reiskel.Skeleton` a
    `reiskel.Harness`.

- [x] **T7 — Corregir y ampliar los tests de `internal/initwizard`** (R8, R9,
  R10, R11, R12, R13, R14, R15, R21, R22)
  - Adaptar `embeddedFile` para resolver `skeleton/<rel>` y, si no existe,
    `<rel>` en `Harness`.
  - Hacer pasar sin debilitar asserts: `TestInitMencionaPersonalize`,
    `TestInstalacionLimpiaIntegra` y
    `TestUpdateConservaPersonalizacionYModificados`.
  - Añadir test de instalación limpia neutral: marcadores en los 4 documentos,
    config sin checks ni `integrations.github`, `Pending` con 4 documentos.
  - Ampliar el test de no sobrescritura para cubrir también los documentos de
    proyecto personalizados.

- [x] **T8 — Separar verificación rápida y lenta en el proyecto vivo**
  (R17, R18, R19)
  - En `.rei/config.json`, dejar solo V1 (`go vet ./...`); conservar
    `integrations` y `viewer` vivos.
  - En `.rei/docs/project/verification.md`, mover V2 (`go test ./...`) a la tabla
    de checkpoints lentos junto a V3 (`make build`) y V4 (`make cross`); dejar el
    ejemplo de config solo con V1 y ajustar la nota de «no ejecutar V3/V4» a
    V2/V3/V4.

- [x] **T9 — Alinear la documentación viva con `skeleton/`** (R24)
  - Añadir `skeleton/` al mapa de estructura de `AGENTS.md` (§4).
  - Actualizar `.rei/docs/project/architecture.md`: organización de la raíz,
    dos `embed.FS` y flujo de `rei init`.
  - Añadir `skeleton/` al árbol de estructura de `README.md` (raíz).
  - No alterar el resto de la personalización viva.

- [x] **T10 — Ejecutar la verificación completa** (R18, R20, R21, R22, R23)
  - `gofmt` sobre los archivos tocados y `go vet ./...`.
  - `go test ./...` completo (incluye los tests hoy fallidos y los nuevos).
  - `make build` y `make cross`.
  - `rei check` en el repo debiendo ejecutar solo V1.
  - `rei doctor` y `rei init status` sin fallos ni pendientes.
  - Smoke test en un `t.TempDir()`: `rei init`, comprobar placeholders/config
    neutral/`/personalize`, y `rei doctor` sin fallos.
  - Confirmar que `go.mod`/`go.sum` no incorporan dependencias nuevas.

- [x] **T11 — Comprobar que el proyecto vivo se conserva** (R16, R20)
  - Registrar `git diff --name-only` de `AGENTS.md`, `.rei/config.json` y
    `.rei/docs/project/`.
  - Ejecutar `rei init --update` sobre el repositorio y confirmar que esos
    archivos personalizados no cambian (solo se permiten los cambios
    intencionados de T8/T9).
  - Confirmar idempotencia con una segunda ejecución.
