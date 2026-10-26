# Tareas — P2: UX del CLI, validación offline y documentación

- [x] T1 — Crear `internal/links` (modelo, `Read`, `Problems` y tests) con las reglas de forma acordadas, incluyendo ausencia, JSON roto y `epic` objeto/`null` (R1, R2, R3, R16)
- [x] T2 — Mostrar la sección `Enlaces:` en `rei item show` y cubrir válido, ausente e inválido sin alterar el código de salida (R1, R2)
- [x] T3 — Añadir `Type` a `state.ItemStatus` y mostrar tipo, issue y base en `rei items status`, conservando el marcador y el fallo con más de un `in_progress` (R3, R4)
- [x] T4 — Cambiar el cierre de `rei github status` por «Configuración completa; sin acciones de configuración pendientes.» y actualizar su test (R5)
- [x] T5 — Añadir opciones a `github.Enable`, parsear y validar `--repo`/`--base`/`--strategy` en `cmdGitHub`, aplicar la precedencia, preservar la config y cubrir errores sin escritura (R6, R7, R8, R9)
- [x] T6 — Integrar `links.json` en `validate.WorkItem` como `WARN` (nunca `FAIL`) y probar JSON roto, forma inválida y archivo ausente (R16)
- [x] T7 — Retirar `done → blocked` de las transiciones estándar en `meta.TransitionAllowed`, verificar el rechazo sin `--force` y el permiso con `--force`, y ajustar tests (R14)
- [x] T8 — Validar existencia, tipo y estado `pending` en `StartSession` antes de escribir; añadir tests de rechazo sin escritura parcial y actualizar el fixture (R17)
- [x] T9 — Actualizar `repoPath`/`releasesPage` a `k1wi777/REI-Harness` y añadir la prueba de regresión canónica (R13)
- [x] T10 — Actualizar `help.go`, `README.md` y `.rei/docs/usage.md` (comandos, flags, estados y enlaces canónicos) y sus tests (R10, R11, R13)
- [x] T11 — Actualizar `github.md`, `core.md`, `errors.md`, `workflow.md`, `meta.md`, `progress.md` y `leader.md`; regenerar los agentes nativos del Leader y verificar con `--check` (R12, R14, R15, R17)
- [x] T12 — Eliminar `package-lock.json`, comprobar que no hay referencias vigentes y ejecutar la verificación final (`go test`, `go vet`, `make build`, `rei check`) (R18)
