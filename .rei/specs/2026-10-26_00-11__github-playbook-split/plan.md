# Plan — GitHub: dividir el playbook y cargar solo la sección necesaria

## Objetivo

Convertir `.rei/docs/harness/integrations/github.md` (1.170 líneas / 48,6 KB) en
un **índice corto de orientación y enrutado** (≤ ~200 líneas, objetivo 150–200)
y mover todo el detalle operativo a cuatro archivos por dominio bajo
`.rei/docs/harness/integrations/github/`:

- `core.md` — flujo v1 (puerta issue+rama, base, `links.json`, nombre canónico,
  adopción genérica, comentarios automáticos, PR/merge).
- `backlog.md` — flujo v2 (épicas, hijas, entrevista/borrador/aprobación,
  idempotencia, creación parcial, promoción, publicación de spec, asignación,
  progreso y cierre de épica).
- `templates.md` — plantillas de issue, comentario, épica, hija y bloque de spec.
- `errors.md` — tabla de errores, reintentos, aislamiento, recordatorios en
  `current.md` y prueba de campo.

El resultado debe ser **solo reorganización y reglas de carga**: no cambian
comandos, marcadores, labels, tipos de base, mensajes de error ni el flujo. El
Leader deja de cargar el playbook completo y carga únicamente el archivo que
indique la tabla de enrutado.

## Archivos

| Archivo | Acción |
|---------|--------|
| `.rei/docs/harness/integrations/github.md` | Reescribir como índice (alcance, config, vocabulario, puerta solo Features, tabla de carga y enlaces). Sin comandos `gh`/`git` detallados. |
| `.rei/docs/harness/integrations/github/core.md` | Nuevo: detalle v1 migrado del playbook. |
| `.rei/docs/harness/integrations/github/backlog.md` | Nuevo: detalle v2 migrado del playbook. |
| `.rei/docs/harness/integrations/github/templates.md` | Nuevo: plantillas extraídas del playbook. |
| `.rei/docs/harness/integrations/github/errors.md` | Nuevo: errores, reintentos, aislamiento, recordatorios y prueba de campo. |
| `.rei/agents/leader.md` | Contrato: tabla acción → archivo, prohibición de cargar el playbook entero y referencias internas actualizadas. |
| `.rei/agents/initializer.md` | Auditar; solo se edita si referencia el playbook (hoy no lo hace). |
| `internal/check/check.go` | Añadir los cuatro archivos nuevos a `RequiredFiles`. |
| `embed_test.go` | Añadir los cuatro archivos a `TestHarnessContieneMaterialDelArnes`. |
| `.rei/docs/harness/README.md` | Actualizar la fila de `integrations/github.md` y añadir las filas de los cuatro archivos. |
| `AGENTS.md` | §7: añadir `task.md` (falta) y la entrada de integración GitHub con el índice. |
| `internal/cli/help.go` | Aclarar que `github.md` es el índice y que el detalle vive en `integrations/github/`; conservar la cadena `integrations/github.md` que exige `help_test.go`. |
| `README.md` | Añadir la integración GitHub a la tabla de documentación del arnés con el índice y los archivos por dominio. |
| `.rei/docs/usage.md` | §1.1: mención breve del índice y de la carga por acción. |
| `.opencode/agents/leader.md`, `.claude/agents/leader.md` | Regenerar desde `.rei/agents/leader.md` (no editar a mano). |

## Cambios

### 1. `github.md` como índice (≤ 200 líneas)

Contenido y orden propuesto:

1. **Cabecera y reglas globales** (actual 1–20): GitHub como proyección, `.rei`
   manda, integración opcional, el CLI no llama a GitHub. Añadir una frase que
   declare este archivo **índice de enrutado** y remita al archivo del dominio.
2. **Alcance** (actual 22–63): resumen de v1 y v2 en una línea cada uno, lista
   completa de **fuera de alcance**, alcance por tipo de Work Item (la puerta es
   solo para `feature`; `task` sin issue/rama salvo opt-in) y regla dura de
   integración desactivada. El detalle de v1 se migra a `core.md` y el de v2 a
   `backlog.md`; el índice no los duplica.
3. **Configuración y diagnóstico** (actual 66–118): claves de
   `integrations.github`, semántica de `enabled`/`base`/`mergeStrategy`, comandos
   `rei github enable|disable|status` y aviso de remoto. Es configuración de CLI,
   no acciones GitHub; se queda completo.
4. **Vocabulario**:
   - tabla de **marcadores** y reglas de identidad (actual 123–151);
   - tabla de **labels** y composición (actual 152–171);
   - **`links.json`**: solo una línea de qué es y enlace a `core.md`; el esquema
     y la resolución de base se migran;
   - **nombre canónico**: regla breve (<repo> = `owner/name` canónico resuelto al
     inicio de cada acción; si falla, se detiene) y enlace a `core.md`; sin
     comando `gh`;
   - **rename/fork**: una línea y enlace a `core.md`.
5. **Puerta solo Features**: disparador y exclusiones de `task` (resumen de
   actual 253–262), con enlace a `core.md` para el procedimiento.
6. **Tabla «qué leer según la acción»** (obligatoria):

   | Acción | Archivo |
   |--------|---------|
   | Orientación, configuración, vocabulario, activar/desactivar/status | `github.md` (este archivo) |
   | Puerta issue+rama y base, `links.json`, nombre canónico, adopción v1, comentarios impl/review, PR/merge | `core.md` |
   | Épicas, backlog/hijas, entrevista/borrador/aprobación, promoción, publicación de spec, asignación, cierre de épica | `backlog.md` |
   | Escribir cuerpo de issue, comentario, épica, hija o bloque de spec | `templates.md` (además del archivo del flujo) |
   | Error, ambigüedad, reintento, recordatorios en `current.md`, prueba de campo | `errors.md` |

7. **Enlaces** a `core.md`, `backlog.md`, `templates.md`, `errors.md` y
   `.rei/docs/harness/README.md`.

No debe contener bloques de `gh api`, `gh issue`, `gh pr`, `git switch` ni
plantillas completas: esos viven en un único archivo de dominio.

### 2. `core.md` (v1)

Migrar literalmente, renumerando como secciones propias y actualizando las
referencias cruzadas:

- alcance v1 detallado (actual 24–34);
- `links.json` completo (actual 172–216);
- nombre canónico completo con comando (actual 217–236) y rename/fork (actual
  237–250);
- puerta única completa: disparador, captura de base, tabla de clasificación y
  tipo de base, pregunta única (actual 253–306), camino «sí» (307–371) y camino
  «no» (417–424);
- adopción: lectura y conflictos (428–455) y adopción genérica v1 (551–587);
- comentarios automáticos: búsqueda idempotente, actualización/creación,
  verificación y labels `rei:blocked` (908–936 y 966–981); las plantillas van a
  `templates.md`;
- PR y merge completos (985–1066), incluida la limitación de fork y los tipos de
  base.

### 3. `backlog.md` (v2)

- alcance v2 detallado (36–46);
- promoción de hija de backlog (456–513) y publicación de la spec (514–550),
  dejando el bloque Markdown en `templates.md`;
- épicas: modelo (590–603), creación de backlog (648–724) y reglas del bloque
  `children` (644–646);
- backlog de hijas: modelo (730–741 y 774–776) y creación/ampliación (778–795);
- asignación (798–850);
- progreso/task list (853–867) y cierre de épica (869–904);
- plantillas de épica e hija: solo la referencia a `templates.md`, sin duplicar
  el Markdown.

### 4. `templates.md`

Contiene únicamente las plantillas, copiadas sin cambios:

- cuerpo de issue de Work Item (373–415);
- bloque de spec publicado (523–540);
- épica (604–647);
- hija de backlog (742–772);
- comentario de implementación (937–951) y de revisión (952–965).

Cada plantilla conserva su marcador, encabezados, líneas `Base:` y bloques
`<details>`. Los archivos de flujo enlazan aquí en vez de repetirlas.

### 5. `errors.md`

- tabla completa de errores y reintentos (1068–1107) sin quitar filas ni
  cambiar causa/siguiente paso;
- reglas transversales: causa+siguiente paso, aislamiento, reintentos
  idempotentes, rate limit/red y lote parcial (1109–1121);
- recordatorios v1 y v2 en `current.md` (1124–1150);
- prueba de campo (1152–1166) y cierre de sesión tras `done` (1168–1170).

### 6. Mapa de migración (verificación de no pérdida)

| Sección actual (aprox.) | Destino |
|---|---|
| Cabecera 1–20 | `github.md` |
| §1 Alcance 22–63 | `github.md` (resumen/fuera de alcance), `core.md` (v1), `backlog.md` (v2) |
| §2 Configuración 66–118 | `github.md` |
| §3.1 Marcadores 123–151 | `github.md` |
| §3.2 Labels 152–171 | `github.md` |
| §3.3 `links.json` 172–216 | `core.md` |
| §3.4 Nombre canónico 217–236 | `github.md` (regla) + `core.md` (comando/detalle) |
| §3.5 Rename y fork 237–250 | `core.md` |
| §4 Puerta 253–306 | `core.md` |
| §4.1 Camino «sí» 307–371 | `core.md` |
| §4.2 Plantilla de issue 373–415 | `templates.md` |
| §4.3 Camino «no» 417–424 | `core.md` |
| §5.1 Lectura y conflictos 428–455 | `core.md` |
| §5.2 Promoción 456–513 | `backlog.md` |
| §5.3 Publicación de spec 514–550 | `backlog.md` (procedimiento) + `templates.md` (bloque) |
| §5.4 Adopción genérica 551–587 | `core.md` |
| §6.1 Modelo de épica 590–603 | `backlog.md` |
| §6.2 Plantilla de épica 604–647 | `templates.md` (plantilla); `backlog.md` (regla del bloque `children`) |
| §6.3 Creación de backlog 648–724 | `backlog.md` |
| §7.1 Modelo y plantilla de hija 728–777 | `backlog.md` (modelo) + `templates.md` (plantilla) |
| §7.2 Creación y ampliación 778–795 | `backlog.md` |
| §8 Asignación 798–850 | `backlog.md` |
| §9.1 Task list 853–867 | `backlog.md` |
| §9.2 Cierre 869–904 | `backlog.md` |
| §10 Comentarios 908–981 | `core.md` (procedimiento y labels) + `templates.md` (plantillas) |
| §11 PR y merge 985–1066 | `core.md` |
| §12 Errores 1068–1121 | `errors.md` |
| §13 Recordatorios 1124–1150 | `errors.md` |
| §13.1 Prueba de campo 1152–1166 | `errors.md` |
| Cierre tras `done` 1168–1170 | `errors.md` |

Reglas de migración:

- Copiar el texto íntegro; solo se renumeran encabezados y se reescriben las
  referencias `§N` para apuntar al archivo correcto (p. ej. `§12` →
  `errors.md §1`). No se reformulan reglas ni mensajes.
- Cada comando `gh`/`git` y cada plantilla aparece **una sola vez**; el resto de
  archivos enlaza.
- Los marcadores, labels, colores, tipos de base (`default`, `integración`,
  `apilada`) y filas de la tabla de errores se conservan literalmente.
- Si alguna sección no encaja en el mapa, se detiene la implementación y se
  reporta; no se inventa un destino.

### 7. Contrato del Leader

En `.rei/agents/leader.md`:

- Sustituir «Los comandos exactos están en
  `.rei/docs/harness/integrations/github.md`» por: los comandos y
  procedimientos están en los archivos de dominio; `github.md` es solo el
  índice de orientación/configuración.
- Añadir una tabla **acción → archivo exacto** equivalente a la del índice:
  configuración/orientación → `github.md`; puerta, base, `links.json`, nombre
  canónico, adopción v1, comentarios, PR/merge → `core.md`; épicas, backlog,
  promoción, publicación, asignación, cierre → `backlog.md`; escritura de
  cuerpos/bloques → `templates.md`; error/ambigüedad/reintento/recordatorios/
  prueba de campo → `errors.md`.
- Añadir la prohibición explícita: **no cargar el playbook completo ni todos los
  archivos a la vez**; cargar solo el archivo que indique la tabla para la acción
  en curso (como máximo `templates.md` junto al archivo del flujo cuando se vaya
  a escribir).
- Actualizar las referencias de «Arranque» (línea ~199) y «Prueba de campo»
  (línea ~259) para que apunten al archivo de dominio, no al playbook genérico.
- Mantener intactas las reglas de negocio del Contrato (puerta solo Features,
  silencio no es aprobación, sin autoasignación, etc.).

En `.rei/agents/initializer.md`: verificar que no cita el playbook. Hoy solo
menciona `integrations.github` y `rei github enable|disable`; si sigue así, no se
edita.

### 8. Registro, embebido y documentación

- `internal/check/check.go`: añadir a `RequiredFiles`, junto a
  `.rei/docs/harness/integrations/github.md`:
  `.../github/core.md`, `.../github/backlog.md`, `.../github/templates.md`,
  `.../github/errors.md`.
- `embed_test.go`: añadir esas cuatro rutas a
  `TestHarnessContieneMaterialDelArnes`. El patrón `all:.rei/docs/harness` ya las
  embebe; no se toca `embed.go`.
- `.rei/docs/harness/README.md`: cambiar la fila de `integrations/github.md` por
  «índice de la integración» y añadir `integrations/github/core.md`,
  `backlog.md`, `templates.md` y `errors.md` con su propósito.
- `AGENTS.md` §7: añadir `| Planificación de Tasks | .rei/docs/harness/task.md |`
  (hoy falta) y una fila de integración:
  `| Integración GitHub (opcional) | .rei/docs/harness/integrations/github.md (índice; carga el archivo del dominio que indique) |`.
- `internal/cli/help.go`: mantener la subcadena `integrations/github.md` (la
  exige `help_test.go`) y aclarar que es el índice; el detalle está en
  `integrations/github/`.
- `README.md`: añadir la fila de integración GitHub a la tabla de documentación
  del arnés, indicando índice + archivos por dominio.
- `.rei/docs/usage.md` §1.1: una frase que explique que el playbook GitHub es un
  índice y que los agentes cargan solo el archivo de la acción.

### 9. Adaptadores nativos

El cambio en el `## Contrato` de `leader.md` invalida los agentes generados. No
editarlos a mano: ejecutar `rei init opencode` y `rei init claude` para
regenerarlos, y comprobar después con `rei init opencode --check` y
`rei init claude --check`.

## Restricciones

- Sin dependencias nuevas ni cambios de comportamiento funcional: solo
  organización de documentación y reglas de carga.
- No tocar `skeleton/`: los docs de harness vivos se embeben vía `Harness`.
- No tocar `.rei/specs/**` ni `.rei/progress/**` históricos.
- No ejecutar ninguna acción GitHub (`gh issue`, `gh pr`, etc.).
- Mantener `github.md` como ruta estable para enlaces, `RequiredFiles` y
  `help_test.go`.
- Cada comando/plantilla en un solo archivo; el índice enlaza, no duplica.
- No cambiar marcadores, labels, colores, tipos de base, errores ni flujo.

## Pasos

- [x] 1. Crear `.rei/docs/harness/integrations/github/` y migrar el contenido a
  `core.md`, `backlog.md`, `templates.md` y `errors.md` según el mapa de la
  sección 6, conservando texto, comandos, marcadores, labels, tipos de base y
  errores; renumerar encabezados y referencias cruzadas.
- [x] 2. Reescribir `github.md` como índice (≤ 200 líneas) con alcance,
  configuración, vocabulario, puerta solo Features, tabla de carga y enlaces; sin
  comandos `gh`/`git` detallados ni plantillas.
- [x] 3. Verificar el mapeo: comparar encabezados del original con los nuevos
  archivos, comprobar que no falta ninguna sección/fila de error y que ningún
  comando o plantilla está duplicado; confirmar `wc -l github.md` ≤ 200.
- [x] 4. Actualizar `.rei/agents/leader.md` con la tabla acción → archivo, la
  prohibición de cargar todo y las referencias de Arranque/Prueba de campo.
  Auditar `initializer.md` y editarlo solo si cita el playbook.
- [x] 5. Añadir los cuatro archivos a `internal/check.RequiredFiles` y a
  `embed_test.go`; actualizar `.rei/docs/harness/README.md`, `AGENTS.md` §7,
  `internal/cli/help.go`, `README.md` y `.rei/docs/usage.md` §1.1.
- [x] 6. Regenerar los adaptadores con `rei init opencode` y `rei init claude`;
  no editar `.opencode/` ni `.claude/` a mano.
- [x] 7. Verificación final:
  - `go test ./...` (V2), `go vet ./...` (V1), `make build` (V3);
  - `rei check` (incluye V1), `rei doctor`, `rei init status`, `rei validate`;
  - `rei init opencode --check` y `rei init claude --check` (0);
  - confirmar con `go test ./ -run TestHarness` que los cuatro archivos se
    embeben y con `fs.Stat`/`RequiredFiles` que existen;
  - `git grep` de `integrations/github.md` para comprobar que todas las
    referencias apuntan al índice o al archivo correcto;
  - V4 (`make cross`) no aplica: cambio de documentación sin código por
    plataforma.
