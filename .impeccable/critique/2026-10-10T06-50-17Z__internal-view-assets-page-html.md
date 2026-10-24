---
target: internal/view/assets/page.html
total_score: 24
max_score: 40
na_heuristics: 
p0_count: 0
p1_count: 2
target_identity: "file:/home/jose/Documentos/template_spec/REI-HARNESS-SDD-TEMPLATE/internal/view/assets/page.html"
target_fingerprint: "sha256:316cc6461e60e61c4c38484a31d1bf62119ce582fd37ea600ae2f081d338904e"
target_path: /home/jose/Documentos/template_spec/REI-HARNESS-SDD-TEMPLATE/internal/view/assets/page.html
timestamp: 2026-10-10T06-50-17Z
slug: internal-view-assets-page-html
---
# Critique — `rei view` (visor de especificaciones)

Method: dual-agent (A: ses_edb6e183bffe7siVOOu9AiuZVO · B: ses_edb6e1838ffeOANB7sVbEic80Z)

## Design Health Score

| # | Heuristic | Score | Key Issue |
|---|-----------|:-----:|-----------|
| 1 | Visibility of System Status | 2 | El acuse de "Aprobar" vive en `#notice` (masthead), lejos del botón; la recarga SSE es invisible. |
| 2 | Match System / Real World | 3 | UI en español, estados en crudo en inglés (`done`, `in_progress`, `WARN/FAIL/OK`). |
| 3 | User Control and Freedom | 2 | Aprobar es irreversible sin confirmación ni deshacer; sin "volver arriba"/TOC. |
| 4 | Consistency and Standards | 3 | Se llama "Índice" pero reemplaza el documento; dos hamburguesas móviles para el mismo panel. |
| 5 | Error Prevention | 2 | Aprobar a un clic; `refresh()` traga errores (`.catch` vacío); sin aviso de conexión. |
| 6 | Recognition Rather Than Recall | 3 | Doc activo resaltado y persistido (bien), pero la acción aparece/desaparece según estado sin explicar por qué. |
| 7 | Flexibility and Efficiency | 2 | Sin atajos, búsqueda, TOC ni doc anterior/siguiente; no se pueden enlazar secciones. |
| 8 | Aesthetic and Minimalist Design | 4 | Contención ejemplar; el frame cede ante la lectura. |
| 9 | Error Recovery | 2 | Mensajes crudos; los callouts de validación no enlazan al punto del documento. |
| 10 | Help and Documentation | 1 | Casi nula; `base_commit`/`last_review_commit` sin explicar. |
| **Total** | | **24/40** | |

## Design Specificity Verdict

**Casi: con un cambio de copy, sirve para cualquier producto.** El sistema —grafito cálido, un acento violeta, columna serif, badges, callouts, colofón— es un *look de "dev-tool docs"* ejecutado con gran disciplina pero **intercambiable**: no hay folio, running head, signatura ni sello de revisión. La North Star "El manual de servicio" está **declarada, no materializada**; el "índice" ni siquiera es un índice (es un selector). La coherencia CSS ↔ `DESIGN.md` es alta y las tres Named Rules se cumplen de verdad.

**Deterministic scan (B):** 1 hallazgo, `design-system-color` ("Color outside DESIGN.md"), severidad **advisory**, en `page.html:0` (`rgb(0,0,0)` en `body`). **Falso positivo confirmado**: el detector no resuelve el `<link href="/static/style.css">` de la plantilla Go (lo busca en `assets/static/`), así que evalúa contra el color por defecto del navegador; en vivo el CSS carga (HTTP 200, idéntico al repo) y el tema oscuro aplica. No hay hallazgos en `style.css`.

**Overlays:** no disponibles (Chrome headless por CLI sin `evaluate`/CDP) → fallback signal. Evidencia de navegador: escritorio 1440×900 y móvil 390×844 renderizan bien, contraste ≈ 6.9:1, sin overflow horizontal.

## Overall Impression

Un visor de documentación sobrio y bien construido, con la mejor disciplina de sistema que se puede pedir. El problema no es el acabado: es que **el momento que define el producto —aprobar— está resuelto como un botón genérico cuyo resultado aparece a un palmo de distancia**, y que la joya (recarga en vivo) **tira al usuario al principio del documento** cada vez que un agente edita. La forma aún no dice "REI".

## What's Working

1. **La columna de lectura.** Serif a ~70ch, 17px/1.72, sobre grafito cálido; el frame cede toda la atención al texto. Es exactamente "baja fricción de revisión".
2. **La disciplina del sistema.** Un acento (≤10%), señal solo para estado, profundidad puramente tonal, suelo de 13px, sin sombras. El CSS honra el manual sin deriva.
3. **La gramática visual frame/contenido.** Sans + tracking (etiquetas) / mono (datos) / serif (cuerpo): jerarquía legible de un vistazo.

## Priority Issues

### [P1] La aprobación no se siente como una puerta y su feedback está desconectado
- **Qué:** `#approve` es un botón genérico de un clic, al fondo de la columna derecha; el resultado aparece en `#notice` (masthead), a ~1000px; sin confirmación; al recargar, el botón se elimina (se pierde el foco).
- **Por qué importa:** el producto existe para una puerta "obligatoria, explícita y sin ambigüedad". El momento de máxima tensión es el peor resuelto.
- **Fix:** confirmación inline con `from→to` explícito; feedback junto al botón; tras éxito, un check persistente "Aprobado · in_progress" en la misma posición; gestionar el foco antes de la recarga.
- **Suggested command:** `/impeccable polish` o `/impeccable clarify`

### [P1] La recarga en vivo destruye la posición de lectura
- **Qué:** `refresh()` hace `#content.innerHTML = html`; `applySelection()` no restaura `scrollTop`. Cualquier edición del agente reinicia las columnas al principio.
- **Por qué importa:** el feature estrella penaliza al lector concentrado; contradice "revisar de verdad".
- **Fix:** capturar/restaurar `scrollTop` de `.reader` (y columnas) alrededor del swap; o diffear y actualizar solo el `article` afectado.
- **Suggested command:** `/impeccable optimize` o `/impeccable harden`

### [P2] El "índice" no es índice y los enlaces profundos no existen
- **Qué:** el bloque se comporta como selector, no como TOC; `app.js` escribe `#doc-…` con `replaceState` pero **nunca lee `location.hash`**; `goldmark` no usa `WithAutoHeadingID`, así que ningún heading tiene `id`.
- **Por qué importa:** rompe la metáfora del manual; impide compartir/abrir en pestaña un documento o sección y cualquier TOC futuro.
- **Fix:** habilitar `WithAutoHeadingID`, leer el hash al cargar (`hashchange` + inicial), y construir un TOC de secciones en el lector.
- **Suggested command:** `/impeccable shape` o `/impeccable polish`

### [P2] Lenguaje mezclado (UI en español, estados en inglés)
- **Qué:** `{{.Meta.Status}}` se pinta crudo (`done`) en badge y ficha; tags `WARN/FAIL/OK`.
- **Por qué importa:** `PRODUCT.md` fija "la interfaz habla en español"; grieta en el contrato de marca.
- **Fix:** mapa ES (`done→completado`, `ready→listo`, `in_progress→en curso`…), conservando el token crudo en mono como dato secundario.
- **Suggested command:** `/impeccable clarify`

### [P3] Errores y validación sin guía de recuperación
- **Qué:** "Error de red al aprobar."; los callouts `WARN/FAIL` no enlazan al punto que los origina.
- **Fix:** mensajes accionables y un enlace "ver en `impl.md`" que seleccione el documento (y salte a la sección cuando haya ids).
- **Suggested command:** `/impeccable harden`

## Persona Red Flags

- **Revisor senior (aprueba specs en serie):** no puede abrir cada documento en pestaña nueva (el hash se ignora → siempre el primer doc); no ve **qué cambió** desde `last_review_commit` (sin diff); y **pierde el scroll** (P2) cada vez que un agente edita.
- **Dev en móvil:** el `#notice` con el resultado de aprobar cae en una fila aparte del masthead (fácil de no ver); **dos hamburguesas idénticas** controlan el mismo panel (ambigüedad de control).
- **Lector de pantalla / baja visión:** outline con múltiples `h1` y saltos (h1→h4) porque el Markdown no se normaliza; sin `id` para saltar; al aprobar, el foco se pierde en `body` (nodo destruido por la recarga).

## Minor Observations

- El masthead dice "Visor de especificaciones" y el tag derecho "Manual de referencia · REI Harness": dos descripciones del mismo objeto.
- **Corrección de evidencia:** el assessment A señaló "espacio muerto en 2560 por max-width 92rem"; eso quedó **resuelto** en el build actual (commit `99fb6f4`), que llena el ancho (A revisó capturas previas al arreglo).
- Deriva de tokens: `DESIGN.md`/`design.json` declaran roles (`headline`, `title`, `caption`, `on-primary`…) que el CSS no usa como variables; y el botón tiene `#6f5ae0`/`#f3f0ff`/`#6755d8` **hardcodeados**.
- `.markdown h2` lleva filete inferior y `h1`/`h3` no; en specs donde las secciones son `h1`, no hay filete de sección.
- `EventSource` sin `onerror`: si el visor se reinicia, la página no se reconecta.
- El `.hint` (13px mono atenuado) es la única explicación de Aprobar; muy tenue para una acción irreversible.

## Questions to Consider

1. Si se llama "Índice", ¿por qué clickear reemplaza el documento en vez de llevarte a él? ¿Qué pasa con 60 secciones?
2. `PRODUCT.md` dice que aprobar es "explícito, nunca implícito". ¿Un clic sin confirmación, con el acuse en la otra punta, es explícito? ¿Cómo se haría sentir como **firmar**?
3. La North Star es "El manual de servicio". Nombra un elemento que tenga un manual impreso y un docs-site genérico no. Si no puedes, ¿la metáfora está construida o solo declarada?
4. Al revisor le importa **qué cambió**. El visor muestra `base_commit`/`last_review_commit`… ¿y el diff?
5. Si un agente edita mientras lees, la recarga te tira al principio. ¿El feature sirve al usuario o al demo?
