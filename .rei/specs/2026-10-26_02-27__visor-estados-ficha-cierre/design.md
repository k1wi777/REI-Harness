# Diseño — Visor: estados reales, FICHA con enlaces y cierre en done

## 1. Punto de partida y causa raíz

El visor ya renderiza el estado con `statusLabel` y la FICHA vive dentro de
`#content`, el mismo fragmento que `app.js` sustituye tras un `reload`. Sin
embargo:

- `Server.snapshot` solo vigila el `mtime` de los siete documentos del Work Item.
  Los cambios de `meta.json` (estado) y de `links.json` (enlaces) no alteran
  ninguna instantánea, por lo que el SSE nunca emite `reload` al cambiar el
  estado o los enlaces. El problema no es el refresco, sino su disparador.
- `watchStatus` y `handleEvents` cierran el servidor en cuanto observan `done`,
  sin emitir un evento final ni esperar a que el cliente renderice. La última
  imagen útil queda en `review`.
- La FICHA solo compone tipo, estado y fecha; `internal/links` ya existe y
  resuelve lectura y validación de `links.json`.

El diseño ataca esas tres causas sin tocar el modelo de estados, el flujo de
aprobación ni el carácter read-only del visor.

## 2. Decisiones de diseño

### D1 — Una única fuente de etiquetas

`statusLabel` sigue siendo la única traducción de estados. Se corrigen dos
valores (`ready` → «en aprobación», `done` → «terminado») y el badge y la fila
de la FICHA siguen invocando la misma función de plantilla. Las clases CSS
continúan usando el estado crudo (`status-ready`, `status-done`), de modo que no
hay cambios de estilo. Se añade una prueba de tabla con los siete estados y un
valor desconocido.

*Alternativa descartada:* duplicar etiquetas en la plantilla o en JavaScript.
Rompería la coherencia badge/FICHA y obligaría a mantener dos mapas.

### D2 — Vigilar metadatos y enlaces, no solo documentos

Se amplía la instantánea del SSE para incluir, además de los documentos, las
rutas `meta.json` y `links.json` del Work Item. La comparación y el evento
`reload` no cambian; cambia el conjunto vigilado. Como `/content` ya renderiza
el fragmento completo (cabecera con badge, índice, lector y FICHA), el refresco
existente actualiza badge y FICHA sin una ruta nueva.

Se mantiene el sondeo periódico (1 s) y no se introduce `fsnotify`: la detección
por `mtime` ya es suficiente, no añade dependencias y las escrituras del CLI usan
reemplazo atómico, que actualiza el `mtime`. El coste de `Stat` sobre dos rutas
más cada segundo es despreciable.

*Alternativa descartada:* emitir un evento dedicado solo con la etiqueta de
estado. Actualizaría el badge, pero no la FICHA ni el resto del fragmento, y
abriría un segundo camino de render que puede desincronizarse.

### D3 — FICHA de enlaces a partir de `internal/links`

`buildModel` lee `.rei/specs/<id>/links.json` con `links.Read` y lo expone en el
modelo de la vista como una ficha independiente:

- archivo inexistente → nota «sin links.json»;
- error de lectura o `links.Problems` no vacío → nota «links.json inválido» con
  el resumen de problemas y sin usar ningún campo;
- archivo válido → issue (número y URL), base (`branch@sha`) y épica
  (slug, número, URL). `Base` o `Epic` nulos se muestran como «sin registrar» y
  «sin épica», respectivamente.

La vista **no** duplica las reglas de validación: solo consume `Read` y
`Problems`. Los campos se guardan como datos planos y la plantilla los escapa;
las URL se emiten como enlaces únicamente cuando la ficha es válida, de modo que
el escapado contextual de `html/template` y la validación http/https de
`internal/links` actúan en serie. No se usa `template.URL` ni ningún bypass de
escapado.

La «base» de la FICHA (branch@sha de `links.json`) es distinta del «cambios
desde» del colofón (commit base de revisión de git); ambas coexisten con
etiquetas diferenciadas para no confundir al lector.

*Alternativa descartada:* calcular enlaces desde `meta.json` o desde GitHub. El
visor es local y read-only; `links.json` ya es la fuente acordada y no requiere
red.

### D4 — Protocolo de cierre en `done`

Se centraliza el ciclo terminal en el `Server` con estado idempotente:

- `done` se marca **una sola vez** (`sync.Once`) al detectarlo `watchStatus`.
- `watchStatus` es el **único** responsable de cerrar por `done`. Se elimina el
  cierre desde `handleEvents`, que hoy compite con el vigilante y puede cortar
  antes de emitir nada.
- Al marcarse `done`, cada conexión SSE emite un evento final con nombre `final`
  y un `data` JSON con el estado crudo, su etiqueta y el HTML completo del
  fragmento `#content` ya renderizado con el estado `done`. El HTML viaja en el
  evento para que el cliente no dependa de un `fetch` que podría perder la
  carrera contra el cierre.
- El cliente aplica ese HTML, restaura la interfaz y responde con una
  confirmación `POST /ack` (autorizada por cookie/token, sin escrituras). Al
  recibirla, el servidor cierra.
- Si no hay cliente, el cliente no confirma o el render falla, el servidor
  cierra igualmente tras una gracia por defecto de 2 s contada desde la marca de
  `done`. La gracia es el respaldo; la confirmación es el camino determinista.
- Para no cortar a otras pestañas que aún están emitiendo su evento final, el
  cierre efectivo nunca ocurre antes de un intervalo mínimo de vaciado (~200 ms)
  desde la marca, aunque la confirmación llegue antes.
- `--stop`, el cierre por inactividad y Ctrl+C siguen llamando a
  `requestClose`, que es idempotente y no depende del protocolo de `done`.

En primer plano (`closeOnDone` falso) no se inicia `watchStatus` ni se emite
`final`: el refresco normal por `reload` muestra «terminado» y el servidor sigue
hasta Ctrl+C.

*Alternativas descartadas:*

- **Solo `reload` y cerrar**: el `fetch("/content")` puede fallar si el servidor
  ya cerró; es exactamente la carrera que deja la página en `review`.
- **Solo una espera fija**: no garantiza que un cliente lento haya renderizado y
  penaliza siempre el cierre.
- **Mantener el servidor indefinidamente tras `done`**: rompe el ciclo de vida
  del servidor de fondo y obliga a `--stop` manual.
- **Recargar la página completa**: perdería scroll, documento activo, diffs y
  enlace profundo, además de la conexión SSE.

### D5 — Conservación del estado de lectura en el cliente

`refresh` ya guarda el scroll del lector, índice y FICHA y reaplica la selección
de documento. Se amplía para:

- capturar qué documentos tienen el diff abierto y reaplicarlo tras el reemplazo;
- reaplicar el hash vigente (documento o sección) después del refresco;
- compartir la misma rutina de restauración entre `reload` y `final`.

El evento `final` se registra con `addEventListener` (los eventos con nombre no
llegan a `onmessage`). Tras aplicarlo se marca la sesión como finalizada, se
cierra el `EventSource` y se inhiben reconexión y aviso «Reconectando…»: el
servidor va a desaparecer por diseño. Si el HTML final no se pudiera aplicar, se
hace una recarga completa como último recurso, todavía dentro de la gracia.

### D6 — Confirmación sin mutación

`POST /ack` solo señala que la página aplicó el estado final: valida token o
cookie, responde 204 y no lee ni escribe documentos, `meta.json` ni
`current.md`. No es una ruta de aprobación ni cambia el estado; la transición
sigue siendo exclusiva de `rei status set`. Se añade junto a las rutas existentes
sin alterar el resto del mux.

## 3. Archivos afectados

| Archivo | Cambio |
|---------|--------|
| `internal/view/render.go` | Etiquetas corregidas; modelo de FICHA de enlaces; carga de `links.json` en `buildModel`; campos de la ficha en `viewModel`. |
| `internal/view/server.go` | Instantánea con `meta.json`/`links.json`; estado terminal idempotente; emisión de `final`; cierre centralizado con confirmación y gracia; ruta `POST /ack`. |
| `internal/view/assets/page.html` | Filas de issue/URL, base y épica; nota de enlaces ausentes o inválidos; enlaces solo en fichas válidas. |
| `internal/view/assets/app.js` | Restauración compartida de scroll/selección/diffs/hash; manejo del evento `final`; confirmación; supresión de reconexión tras el cierre. |
| `internal/view/assets/style.css` | Ajuste menor para la nota de la FICHA (reutilizando las clases existentes si es posible). |
| `internal/view/render_test.go` | Pruebas de etiquetas y de la FICHA (válida, ausente, inválida, campos opcionales). |
| `internal/view/server_test.go` | Pruebas de refresco por estado/enlaces, secuencia `final` → confirmación → cierre, gracia sin cliente, cierre no anticipado y ausencia de mutaciones. |
| `internal/cli/help.go` | Ayuda de `view`: actualización en vivo y cierre tras mostrar «terminado». |
| `README.md`, `.rei/docs/usage.md` | Descripción del visor alineada con estados, FICHA y cierre. |

`internal/links` se reutiliza sin cambios. No se tocan `skeleton/`, `internal/state`,
`internal/meta` ni el flujo de aprobación.

## 4. Contrato del evento final

El evento se describe aquí para que servidor, cliente y pruebas compartan la
misma forma; no se prescribe el código.

| Campo | Contenido |
|-------|-----------|
| nombre del evento | `final` |
| `status` | estado crudo (`done`) |
| `label` | etiqueta mostrada («terminado») |
| `html` | fragmento `#content` completo, ya renderizado con el estado final |

El `data` se serializa como JSON de una sola línea para no romper el framing de
SSE. El cliente solo sustituye `#content`; cabecera, scripts y conexión quedan
intactos. El evento se emite una vez por conexión y solo cuando el servidor de
fondo va a cerrarse por `done`.

## 5. Secuencia de cierre

1. `watchStatus` observa `meta.json` con estado `done`.
2. Marca el cierre terminal (idempotente) y notifica a las conexiones SSE.
3. Cada conexión construye el modelo en ese instante (ya `done`), renderiza el
   fragmento y emite `final`.
4. El cliente aplica el HTML, restaura la lectura y envía `POST /ack`.
5. El servidor cierra al recibir la confirmación, respetando el intervalo mínimo
   de vaciado.
6. Si no llega confirmación, cierra al vencer la gracia de 2 s.
7. `requestClose` elimina el descriptor del servidor de fondo como hoy.

`--stop`, inactividad y Ctrl+C pueden disparar el paso 7 en cualquier momento,
sin esperar a `done`.

## 6. Estrategia de pruebas

- `statusLabel`: tabla de los siete estados y un valor desconocido.
- FICHA: `links.json` válido con base y épica completas; sin base/épica; archivo
  ausente; JSON mal formado; campos inválidos; verificación de que las URL
  válidas aparecen como enlaces y las inválidas no.
- Refresco: cambiar `meta.json` y `links.json` con una conexión SSE abierta y
  exigir `reload`; comprobar que `/content` refleja el nuevo estado y la nueva
  ficha.
- Cierre: servidor de fondo con cliente SSE; llevar a `done`; verificar que
  `final` llega con «terminado» y HTML `status-done`, que el servidor no se ha
  cerrado antes, que tras `/ack` se cierra y que `Serve` termina sin error.
- Respaldo: sin confirmación, el cierre ocurre tras la gracia y nunca antes del
  evento final; en `in_progress`/`review` no se cierra.
- Read-only: `/approve` sigue devolviendo 404 y no cambia `meta.json`; `/ack` no
  modifica documentos ni estado.
- Regresión: autenticación, `/static`, TOC, diffs, validación, sesión y enlaces
  profundos.

Los tests usan intervalos reducidos (`pollInterval`, gracia) para no alargar la
suite; los valores por defecto quedan para producción.

## 7. Riesgos y bordes

- **Sin pestañas abiertas**: el servidor tarda la gracia en cerrar (2 s) en vez
  de hacerlo de inmediato. Es el precio de garantizar el render y no afecta a
  `--stop` ni al idle.
- **Varias pestañas**: todas reciben `final`; la primera confirmación cierra,
  protegida por el intervalo mínimo de vaciado. Una pestaña muy lenta puede
  quedar con la última imagen si no llega a aplicar el evento, pero el estado
  final sigue disponible en disco y al reabrir el visor se muestra `done`.
- **Resolución del `mtime`**: se mantiene la comparación por `mtime`; las
  escrituras del CLI son reemplazos atómicos, por lo que el cambio se observa.
- **Coste del render final**: se renderiza una sola vez por conexión y por
  evento terminal; los documentos son locales y pequeños.
- **Fallback de recarga completa**: solo si el HTML final no se puede aplicar;
  es preferible a quedarse en `review`.

## 8. Fuera de alcance

Modo directo sin Work Item, manifest, release, P3, cambios en `internal/links`,
nuevas dependencias y cualquier ruta que apruebe o mute el Work Item.
