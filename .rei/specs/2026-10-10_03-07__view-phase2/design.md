# Design — Visor de especificaciones (Fase 2)

> Work Item: `2026-10-10_03-07__view-phase2` (feature).
>
> Este documento describe **cómo** se implementa la Feature sobre la v1
> (`internal/view/`). Las referencias `R<n>` apuntan a `requirements.md`.

---

# 1. Estrategia

Fase de lectura y pulido sobre el visor existente. No se introduce ningún
componente nuevo de infraestructura: se extienden `internal/gitx` (nuevo helper
de diff por archivo) y `internal/view` (modelo, render y assets), y se alinea el
sistema de tokens (`DESIGN.md` / `design.json` ⇄ `assets/style.css`).

Las tres piezas de comportamiento:

1. **Cambios desde la última revisión.** Se reutiliza la resolución de base que ya
   usa `rei review-diff` (`gitx.ReviewDiff`) y se expone, por documento, el diff
   unificado respecto de esa base. El visor ya tiene un selector de documento: se
   añade un modo "Cambios" por documento, más un indicador en el índice. (R1–R8)
2. **Índice de secciones.** El Markdown ya se parsea con `goldmark` y los
   encabezados ya reciben `id` (`parser.WithAutoHeadingID`). Se recorre el AST para
   extraer los encabezados (nivel, `id`, texto) y se renderiza un TOC por
   documento. (R9–R12)
3. **Pulido.** Reconexión del SSE, masthead sin redundancia, más contraste del
   documento activo y alineación de tokens. (R13–R20)

Sin dependencias Go nuevas: todo se resuelve con la biblioteca estándar y
`goldmark`. (R25)

---

# 2. Archivos afectados

## Nuevos

| Archivo | Propósito |
|---------|-----------|
| `internal/view/changes.go` | Cálculo del base de revisión y de los cambios por documento; render del diff a HTML. |
| `internal/view/changes_test.go` | Tests de degradación (sin git / sin base) y de composición del diff. |
| `internal/view/toc.go` | Extracción de encabezados del AST de `goldmark` para el TOC. |
| `internal/view/toc_test.go` | Tests de extracción de encabezados (niveles, ids, texto). |

## Modificados

| Archivo | Cambio |
|---------|--------|
| `internal/gitx/gitx.go` | `ReviewBase` (extraído de `ReviewDiff`) + `FileChange`/`FileChanges` (diff por archivo respecto de una base). |
| `internal/gitx/gitx_test.go` | Tests de `ReviewBase` y `FileChanges` (repo git temporal; `t.Skip` si no hay git). |
| `internal/view/render.go` | `buildModel` calcula base y cambios; `renderMarkdown` → `renderDoc` (HTML + secciones); helpers de etiqueta. |
| `internal/view/assets/page.html` | Masthead sin redundancia; base de revisión; indicador de cambios en el índice; TOC por documento; botón "Cambios" y bloque de diff. |
| `internal/view/assets/app.js` | Reconexión del `EventSource` (`onerror`); alternar el modo "Cambios". |
| `internal/view/assets/style.css` | Tokens de color y tipografía; contraste del activo; estilos de TOC y diff; limpieza de `.brand-sub`. |
| `internal/view/render_test.go` | Ajustar a los nuevos campos del modelo; cubrir composición con cambios. |
| `internal/view/server_test.go` | Cubrir `/content` con la información de cambios. |
| `DESIGN.md` | Añadir el token `primary-edge` y documentar TOC/cambios; ajustar los valores de los roles tipográficos a los realizados. |
| `.impeccable/design.json` | Solo si alguna divergencia de tokens lo exige (aditivo); `DESIGN.md` es la fuente de verdad. |

> No se toca `internal/view/server.go`, `lifecycle.go` ni `view.go`: no hay rutas
> nuevas (los diffs viajan dentro de `/content`) ni cambios de ciclo de vida.

---

# 3. Componentes

## 3.1 `internal/gitx` — base de revisión y diff por archivo

Se extrae de `ReviewDiff` la resolución de base y se añade el diff por archivo:

```go
// ReviewBase devuelve el commit base de revisión de un Work Item
// (last_review_commit; en su defecto base_commit) y su origen.
// ok=false si no hay ninguno; err si no se puede leer meta.json.
func ReviewBase(p *paths.Project, id string) (base, origin string, ok bool, err error)

// FileChange describe el cambio de un archivo respecto de un commit base.
type FileChange struct {
    Changed bool   // el contenido difiere del base
    New     bool   // no existía en el base (todo añadido)
    Added   int    // líneas añadidas
    Removed int    // líneas eliminadas
    Diff    string // diff unificado (--no-color); "" si no cambió
}

// FileChanges calcula el cambio de cada ruta (relativa a la raíz del proyecto)
// respecto del commit base. Reutiliza gitOut.
func FileChanges(p *paths.Project, base string, relPaths []string) (map[string]FileChange, error)
```

- `ReviewBase` replica el `switch` que hoy vive en `ReviewDiff` (`last_review_commit`
  → `base_commit`) y valida el commit con `cat-file -e <base>^{commit}`. `ReviewDiff`
  se refactoriza para usarlo (sin cambiar su salida). (R1)
- `FileChanges`, por ruta: si no está rastreada (`git ls-files --error-unmatch`) y
  existe, se marca `New` y todo el contenido como añadido; si está rastreada, se
  consulta `git diff --numstat --no-color <base> -- <ruta>` y, cuando cambió,
  `git diff --no-color <base> -- <ruta>`. (R2, R8)
- No filtra `.rei/progress/`: el visor muestra `impl.md`/`review.md`/`spec.md` y
  quiere sus cambios (esto es una decisión del visor, distinta del paquete de
  revisión del CLI). (R2)
- Si no hay git (`!IsRepo`) o no hay base, `FileChanges` no se invoca y el visor
  degrada (ver 3.2). (R6, R7)

## 3.2 `internal/view` — modelo y render

`render.go` gana los campos de cambios y el TOC:

```go
type sectionView struct{ Level int; ID, Text string }

type docView struct {
    Name     string
    Present  bool
    HTML     template.HTML
    Pos      int
    Sections []sectionView
    Changed  bool
    New      bool
    Added    int
    Removed  int
    Diff     template.HTML // diff renderizado (solo si Changed)
}

type viewModel struct {
    // … campos de la v1 …
    GitAvailable bool
    HasBase      bool
    BaseSHA      string
    BaseOrigin   string
}
```

- `renderMarkdown` pasa a `renderDoc(src []byte) (template.HTML, []sectionView, error)`:
  se parsea una sola vez (`mdRenderer.Parser().Parse(text.NewReader(src))`), se
  recogen los encabezados (`ast.Walk` sobre `*ast.Heading`, leyendo `Level`, el
  atributo `id` y el texto de los nodos `*ast.Text`/`*ast.String`) y se renderiza
  con `mdRenderer.Renderer().Render(...)` — el mismo HTML que hoy produce
  `mdRenderer.Convert`. (R9, R10)
- `buildModel`: además de lo actual, resuelve `gitx.ReviewBase`; si hay base y git,
  llama a `gitx.FileChanges` con las rutas relativas de los siete documentos y
  vuelca `Changed`/`New`/`Added`/`Removed`/`Diff` en cada `docView`; si no, deja
  `GitAvailable`/`HasBase` en `false` para que la plantilla informe. (R1–R8)
- El diff unificado se compone a HTML en `changes.go`: se escapa cada línea
  (`html.EscapeString`) y se clasifica con clases propias (`dl-add`, `dl-del`,
  `dl-hunk`, `dl-meta`, `dl-ctx`). Como el diff lo genera git, el render nunca
  interpreta el contenido de los documentos como HTML. (R2, R21)

## 3.3 `page.html`

- **Cambios** (R3–R8):
  - En el índice, cada `.doc` cambiado recibe la clase `changed` y una insignia
    "cambios"; los nuevos, además `new`.
  - En el colofón se añade un ítem `cambios desde` con el SHA corto del base y su
    origen; cuando no aplica, se muestra la nota correspondiente
    (`Sin git: …` / `sin base de revisión`).
  - Cada `.doc-body` con cambios incluye, en el running head, un botón
    `changes-toggle` (`aria-pressed`) y un bloque `.doc-diff` con el diff
    renderizado y el recuento `+A / −R`. Sin cambios, se muestra una nota
    "(sin cambios)". Con JS el diff se oculta por defecto y se alterna; sin JS se
    muestra tras el contenido. (R5)
- **TOC** (R9–R12): dentro de cada `.doc-body`, antes de `.markdown`, un
  `<nav class="toc">` con la lista anidada por nivel (`toc-l1`…). Solo se emite si
  el documento tiene 2 o más encabezados. Los enlaces son `href="#<id>"`, que ya
  encajan con el manejo de hash y los enlaces profundos de la v1.
- **Masthead** (R16): se elimina la descripción duplicada (ver D5).
- Se añade al `funcMap` un helper `short` (primeros 8 caracteres del SHA).

## 3.4 `app.js`

- **Reconexión del SSE** (R13–R15): se envuelve el `EventSource` en una función
  `connect()`. En `onerror` se cierra la instancia, se avisa de forma discreta en
  `#notice` ("Reconectando…") y se reintenta con espera creciente (de ~1 s a un
  tope de ~15 s). En `onopen` se limpia el aviso y se reinicia la espera.
- **Modo "Cambios"** (R2): delegación de `click` para `.changes-toggle`, que
  alterna la clase `show-changes` del `.doc-body` y actualiza `aria-pressed`.
- Se conservan el selector de documento, la aprobación en dos pasos, los enlaces
  profundos y la conservación del scroll. (R23)

## 3.5 `assets/style.css`

- **Contraste del activo** (R17): `.doc.active { background: var(--ground-3); }`
  (antes `--ground-2`). Ajustar el color del enlace activo si es necesario para
  mantener el contraste.
- **TOC** (R9–R12): `.toc` / `.toc-list` / `.toc-item` coherentes con el mundo
  (etiqueta `.sec-h`, enlaces mono/serif, hairline), sin sombras.
- **Diff** (R3): `.doc-diff` en mono; `dl-add`/`dl-del` con tinte de señal
  (verde/rojo) y el resto en tonos neutros; la insignia del índice usa
  `--accent-dim` (rol `primary-tint`).
- **Tokens** (R18–R20): ver 3.6.

## 3.6 Tokens (R18–R20)

`DESIGN.md` es la fuente de verdad. El bloque `:root` de `style.css` debe
declarar una variable por cada rol declarado y consumirse en las reglas.

### Colores (completar y consumir)

| Rol `DESIGN.md` | Variable CSS | Valor |
|-----------------|--------------|-------|
| `on-primary` | `--on-primary` | `#f3f0ff` |
| `primary-hover` | `--primary-hover` | `#6755d8` |
| `on-signal-warn` | `--on-warn` | `#241a06` |
| `signal-warn-hover` | `--warn-hover` | `#e9b351` |
| `scrollbar-hover` | `--scrollbar-hover` | `#4a515e` |
| `on-accent` | `--on-accent` | `#ffffff` |
| `primary-edge` (nuevo en `DESIGN.md`) | `--accent-edge` | `#6f5ae0` |

Se eliminan todos los literales de color fuera de `:root`:

- `.action button` → `color: var(--on-primary); border: 1px solid var(--accent-edge)`;
  `:hover` → `var(--primary-hover)`; `.confirming` → `var(--on-warn)`;
  `.confirming:hover` → `var(--warn-hover)`.
- `::-webkit-scrollbar-thumb:hover` → `var(--scrollbar-hover)`;
  `::selection` → `color: var(--on-accent)`.
- Los `rgba(...)` derivados (bordes de estado y de callout, selección) se expresan
  con `color-mix(in srgb, var(--<rol>) N%, transparent)`, de modo que no quede
  ningún valor de color hardcodeado.
- `--accent-dim` (rol `primary-tint`) se usa en la insignia de "cambios".

### Tipografía (completar y consumir)

Se añaden variables por rol (familia / tamaño / peso / interlineado / tracking) y
se consumen en las reglas que realizan cada rol. Para no alterar el mundo visual,
`DESIGN.md` adopta los valores ya realizados en el CSS (los de la v1, aprobados),
resolviendo así las divergencias (p. ej. `display` line-height 1.25 y
`title` 1.32rem). Mapeo rol → regla:

| Rol | Regla(s) | Familia | Tamaño | Peso | Otros |
|-----|----------|---------|--------|------|-------|
| `base` | `body` | `--ui` | 16px | 400 | lh 1.6 |
| `ui` | `.action button` | `--ui` | 15px | 700 | |
| `ui-small` | `.doc a.nm` | `--mono` | 14px | 400 | |
| `label` | `.sec-h`, `.badge`, `.callout .tag`, `.facts dt`, `.col-k` | `--ui` | 13px | 700 | tracking 0.1em |
| `mono` | `.wid`, `.session`, `.notice`, `.masthead-tag`, `.col-v` | `--mono` | 13px | 400 | |
| `display` | `.titleband h1` | `--ui` | `clamp(1.3rem,2.4vw,1.75rem)` | 650 | lh 1.25, -0.01em |
| `doc-title` | `.markdown > h1:first-child` | `--read` | `clamp(1.6rem,2.4vw,1.95rem)` | 600 | lh 1.22 |
| `title` | `.markdown h1` | `--read` | 1.32rem | 600 | |
| `headline` | `.markdown h2` | `--read` | 1.3rem | 600 | lh 1.25 |
| `subtitle` | `.markdown h3` | `--read` | 1.12rem | 600 | |
| `body` | `.markdown p, li` | `--read` | 1.0625rem | 400 | lh 1.72 |
| `caption` | `.markdown h4` | `--ui` | 0.95rem | 700 | |
| `dense` | `.markdown table` | `--ui` | 0.92rem | 400 | |
| `code` | `.markdown pre code` | `--mono` | 0.86rem | 400 | |

`design.json` declara un subconjunto de tokens cuyos valores ya coinciden con el
CSS; solo se toca (de forma aditiva) si fuese necesario. `DESIGN.md` recibe
`primary-edge` en `colors` y se documentan los componentes nuevos (TOC y cambios)
en su sección de componentes. (R20)

---

# 4. Decisiones de diseño

**D1 — El visor no introduce rutas nuevas.** Los diffs se calculan en cada
construcción del modelo y viajan dentro de `/content` (y por tanto se refrescan
con el SSE). Evita un endpoint y JS adicional, a cambio de recalcular git por
petición: aceptable en un servicio local. El servidor no cambia de contrato.
(R2, R13)

**D2 — Reutilizar la base de revisión existente.** `last_review_commit` /
`base_commit` ya son el lenguaje del proyecto (`gitx.ReviewDiff`, `rei commit
set`). El visor usa la misma resolución (`ReviewBase`) en lugar de inventar una
noción de "última vista". (R1)

**D3 — Vista de diff, no resaltado en contexto.** Se muestra el diff unificado por
documento (modo "Cambios"), no un resaltado de bloques dentro del Markdown
renderizado. El resaltado exigiría mapear líneas de origen a bloques del AST y un
renderer propio: más frágil y más código para el mismo "ver qué cambió". El diff
es determinista y reutiliza git. (R2)

**D4 — Alcance al material del Work Item.** Los cambios se calculan solo para los
documentos que el visor muestra (los siete), no para todo el repositorio. La red
flag es la spec; el diff de todo el repo ya lo cubre `rei review-diff`. (R2)

**D5 — Masthead: una sola descripción.** Se conserva la leyenda de la derecha
(`Manual de referencia · REI Harness`) y se elimina el subtítulo del brand
(`· Visor de especificaciones`), junto con su regla CSS. Así el masthead no repite
el objeto y no duplica "REI" en el lado del brand. (R16)

**D6 — TOC por documento, sin JS.** El índice de secciones se renderiza en el
servidor y usa anclas `#<id>` nativas; encaja con el manejo de hash que ya existe
y funciona sin JavaScript. Se emite solo con 2+ encabezados para no añadir ruido.
(R9–R12)

**D7 — Tokens con per-property variables.** Para no alterar el interlineado
heredado, cada rol se expresa con variables por propiedad (familia/tamaño/peso/…)
en lugar del atajo `font`, que resetea propiedades. (R18)

**D8 — `DESIGN.md` adopta los valores realizados.** Donde el sistema de diseño y el
CSS divergen en cifras concretas, gana el valor realizado en la v1 (el mundo
visual aprobado); así la alineación no cambia la apariencia. (R20)

---

# 5. Alternativas descartadas

**A1 — Resaltar bloques modificados dentro del HTML renderizado.** Descartada
(ver D3): exige mapear posiciones del AST a líneas y un renderer propio, con más
superficie de error que un diff.

**A2 — Diff de todo el repositorio en el visor.** Descartada (ver D4): mezcla
código y bookkeeping con la spec; el visor existe para leer el Work Item.

**A3 — Endpoint `/diff` y carga perezosa por JS.** Descartada: añade JS y una
ruta para ahorrar unos kilobytes en un servicio local, y complica el refresco por
SSE (habría que refrescar también el diff activo).

**A4 — `font` shorthand como token de rol.** Descartada (ver D7): resetea
`line-height`, lo que alteraría el ritmo vertical donde hoy se hereda.

**A5 — Renombrar las variables CSS a los nombres de `DESIGN.md`.** Descartada:
churn grande en todo el archivo sin beneficio; basta con añadir los roles que
faltan y consumirlos, dejando un mapeo explícito documentado.

**A6 — TOC en el raíl izquierdo (junto a la lista de documentos).** Descartada:
exigiría sincronizar el TOC con el documento activo por JS y no aporta en la
vista sin JS, donde los documentos se apilan. El TOC dentro del documento funciona
en ambos casos. (R9)
