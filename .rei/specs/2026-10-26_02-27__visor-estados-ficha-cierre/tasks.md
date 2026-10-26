# Tareas — Visor: estados reales, FICHA con enlaces y cierre en done

- [x] T1 — Corregir `statusLabel`: `ready` → «en aprobación» y `done` →
  «terminado», manteniendo las demás etiquetas y el valor desconocido sin
  traducir; añadir la prueba de tabla de estados. (R1, R2)
- [x] T2 — Ampliar `Server.snapshot` para vigilar también `meta.json` y
  `links.json` junto a los documentos, sin cambiar la comparación ni el evento
  `reload`. (R3, R4)
- [x] T3 — Cargar `links.json` en `buildModel` con `internal/links` (`Read` y
  `Problems`), exponer issue/URL, base y épica en la FICHA, y renderizar la nota
  cuando falte, no se pueda leer o sea inválido. (R6, R7, R8)
- [x] T4 — Ajustar `page.html` para las filas de enlaces de la FICHA, con enlaces
  solo en fichas válidas y nota explícita en caso contrario; conservar el resto
  de filas y el colofón. (R6, R7, R8)
- [x] T5 — Compartir en `app.js` la restauración de scroll, documento activo,
  diffs abiertos, TOC y hash entre el refresco normal y el evento final. (R5)
- [x] T6 — Implementar en `server.go` el estado terminal idempotente, la emisión
  del evento `final` con estado y HTML renderizado, y el cierre centralizado en
  `watchStatus` con confirmación y gracia; eliminar el cierre desde
  `handleEvents`. (R9, R10, R12, R13)
- [x] T7 — Añadir `POST /ack` autorizado y sin escrituras, y en `app.js` el
  manejo de `final`: aplicar HTML, confirmar, cerrar el `EventSource` e inhibir
  reconexión y aviso de reconexión. (R9, R10, R11, R14)
- [x] T8 — Añadir pruebas de FICHA: enlaces válidos con base y épica, campos
  opcionales ausentes, archivo ausente, JSON mal formado y campos inválidos;
  comprobar que solo se enlazan URL http/https válidas. (R6, R7, R8)
- [x] T9 — Añadir pruebas de servidor/SSE: `reload` al cambiar estado y enlaces,
  secuencia `final` → confirmación → cierre, cierre por gracia sin confirmación,
  no cierre en `in_progress`/`review`, comportamiento en primer plano, `/approve`
  inexistente y `/ack` sin mutaciones. (R3, R4, R9, R10, R11, R12, R13, R14)
- [x] T10 — Actualizar `README.md`, `.rei/docs/usage.md` y la ayuda de `rei view`
  para describir las etiquetas, la FICHA de enlaces y el cierre tras «terminado».
  (R15)
- [x] T11 — Ejecutar `gofmt`, `go vet ./...`, `go test ./...`, `make build`,
  `rei check`, `rei validate` y las comprobaciones de adaptadores/doctor/init
  status; verificar que TOC, diff, sesión, assets y autenticación no han
  regresionado. (R16)
