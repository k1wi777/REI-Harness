---
name: REI Harness — Visor de especificaciones
description: Manual de referencia oscuro para revisar la planificación de un Work Item en el navegador.
colors:
  primary: "#9a86f5"
  primary-deep: "#5a49c9"
  primary-tint: "#332a56"
  canvas: "#14161a"
  surface: "#1a1d22"
  surface-raised: "#23272e"
  border: "#2b3039"
  border-strong: "#3a404b"
  text: "#eae7e0"
  text-soft: "#d6d2c9"
  text-muted: "#9aa0a8"
  signal-warn: "#e0a33a"
  signal-ok: "#7fb98a"
  signal-fail: "#e5646b"
  signal-review: "#6bb0d6"
  signal-warn-hover: "#e9b351"
  on-signal-warn: "#241a06"
  primary-hover: "#6755d8"
  on-primary: "#f3f0ff"
  scrollbar-hover: "#4a515e"
  on-accent: "#ffffff"
typography:
  display:
    fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif'
    fontSize: "clamp(1.3rem, 2.4vw, 1.75rem)"
    fontWeight: 650
    lineHeight: 1.2
    letterSpacing: "-0.01em"
  headline:
    fontFamily: '"Iowan Old Style", "Palatino Linotype", Palatino, "Book Antiqua", "Source Serif 4", Georgia, serif'
    fontSize: "1.6rem"
    fontWeight: 600
    lineHeight: 1.25
  body:
    fontFamily: '"Iowan Old Style", "Palatino Linotype", Palatino, "Book Antiqua", "Source Serif 4", Georgia, serif'
    fontSize: "1.0625rem"
    fontWeight: 400
    lineHeight: 1.72
  label:
    fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif'
    fontSize: "13px"
    fontWeight: 700
    letterSpacing: "0.16em"
  mono:
    fontFamily: 'ui-monospace, "SFMono-Regular", Menlo, Consolas, "JetBrains Mono", monospace'
    fontSize: "13px"
    fontWeight: 400
  title:
    fontFamily: '"Iowan Old Style", "Palatino Linotype", Palatino, "Book Antiqua", "Source Serif 4", Georgia, serif'
    fontSize: "1.3rem"
    fontWeight: 600
  subtitle:
    fontFamily: '"Iowan Old Style", "Palatino Linotype", Palatino, "Book Antiqua", "Source Serif 4", Georgia, serif'
    fontSize: "1.12rem"
    fontWeight: 600
  caption:
    fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif'
    fontSize: "0.95rem"
    fontWeight: 700
  dense:
    fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif'
    fontSize: "0.92rem"
    fontWeight: 400
  code:
    fontFamily: 'ui-monospace, "SFMono-Regular", Menlo, Consolas, "JetBrains Mono", monospace'
    fontSize: "0.86rem"
    fontWeight: 400
  ui:
    fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif'
    fontSize: "15px"
    fontWeight: 400
  ui-small:
    fontFamily: 'ui-monospace, "SFMono-Regular", Menlo, Consolas, "JetBrains Mono", monospace'
    fontSize: "14px"
    fontWeight: 400
  base:
    fontFamily: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif'
    fontSize: "16px"
    fontWeight: 400
rounded:
  xs: "3px"
  sm: "4px"
  hairline: "1px"
  micro: "2px"
  pill: "999px"
spacing:
  xs: "8px"
  sm: "14px"
  md: "24px"
  lg: "36px"
components:
  button-primary:
    backgroundColor: "{colors.primary-deep}"
    textColor: "#f3f0ff"
    rounded: "{rounded.sm}"
    padding: "13px 0"
  button-primary-hover:
    backgroundColor: "{colors.primary}"
  badge:
    backgroundColor: "{colors.surface-raised}"
    textColor: "{colors.text-soft}"
    rounded: "{colors.xs}"
    padding: "3px 9px"
---

# Design System: REI Harness — Visor de especificaciones

## Overview

**Creative North Star: "El manual de servicio"**

El visor de specs se comporta como un manual de referencia técnica: una
cabecera, una portada con el identificador del trabajo, un índice, un apartado
de validación y un colofón. Es un documento pensado para revisarse, no un
panel de métricas ni una pila de tarjetas. La voz visual es técnica y sobria,
pero hospitalaria: fondo grafito cálido, tipografía de lectura serif y
etiquetas rotundas, para que una persona lea con comodidad una spec larga antes
de aprobarla.

El mundo posee el **frame** —cabecera, banda de título, índice, validación,
colofón— y lo lleva con carácter; la **columna de lectura** se mantiene calmada:
serif, medida real y contraste, sin nada actuando detrás del texto. La
profundidad se resuelve por **capas tonales y hairlines**, nunca por sombras ni
cristal. El acento es un único **violeta**; los colores de señal (ámbar, verde,
rojo, azul) solo comunican estado.

Decisiones confirmadas con el usuario: fondo **oscuro** (no blanco), acento
**violeta** (no naranja) y tipografía pequeña **generosa**. Restricción de
proyecto: **sin activos de terceros** (solo stacks de sistema y assets propios
embebidos).

**Key Characteristics:**
- Frame de manual: masthead, banda de título, índice, validación, colofón.
- Grafito cálido de fondo; violeta como único acento de marca.
- Columna de lectura serif a ~70ch; cuerpo a 17px.
- Etiquetas generosas: ninguna por debajo de 13px.
- Sin sombras ni desenfoques: profundidad por capas tonales y hairlines.
- Sin fuentes de terceros: stacks de sistema.

## Colors

Paleta monocroma oscura —grafito cálido— con un solo acento violeta y un
conjunto reducido de colores de señal para el estado.

### Primary
- **Violeta de instrumento** (#9a86f5): foco, enlaces, marca y realces; el
  violeta legible sobre el grafito.
- **Violeta de acción** (#5a49c9): relleno del botón primario (Aprobar).
- **Violeta sombra** (#332a56): reservado a tintes, no a superficies grandes.

### Neutral
- **Grafito cálido** (#14161a): lienzo general.
- **Grafito elevado** (#1a1d22): masthead, bloques de código y callouts.
- **Superficie** (#23272e): badges, cabeceras de tabla y código en línea.
- **Hairline** (#2b3039) y **hairline fuerte** (#3a404b): separadores y bordes.
- **Tinta** (#eae7e0): texto principal.
- **Tinta suave** (#d6d2c9): cuerpo de lectura.
- **Gris apagado** (#9aa0a8): etiquetas y texto secundario.

### Tertiary — Señal
- **Ámbar de precaución** (#e0a33a): hallazgos `WARN` y estado `ready`.
- **Verde de conformidad** (#7fb98a): hallazgo `OK` y estado `done`.
- **Rojo de fallo** (#e5646b): hallazgos `FAIL` y estados `blocked`/`changes_requested`.
- **Azul de revisión** (#6bb0d6): estado `review`.

### Named Rules
**La Regla del Único Acento.** El violeta es el único color de marca. Los
colores de señal solo significan estado; nunca se usan para decorar.

## Typography

**Display Font:** system-ui (con reserva por SO: `-apple-system`, `"Segoe UI"`,
`Roboto`, `Helvetica`, `Arial`)
**Body Font:** serif de lectura (`"Iowan Old Style"`, `Palatino`, `"Book Antiqua"`,
`"Source Serif 4"`, `Georgia`)
**Label/Mono Font:** monoespaciada del sistema (`ui-monospace`, `SFMono-Regular`,
`Menlo`, `Consolas`)

**Character:** la serif de lectura da el tono de "documento que se lee"; la sans
de sistema firma las etiquetas en mayúsculas y la mono fija los datos. La
elección de stacks de sistema es una decisión del proyecto (sin activos de
terceros), no una ausencia de criterio; se compensa con jerarquía por peso y
tamaño.

### Hierarchy
- **Display** (650, `clamp(1.3rem, 2.4vw, 1.75rem)`, 1.2, tracking -0.01em): el
  título del Work Item en la banda de título.
- **Headline** (600, 1.6rem / 1.3rem / 1.12rem, 1.25, serif): los encabezados
  de los documentos renderizados (`h1`/`h2`/`h3`).
- **Body** (400, 1.0625rem / 17px, 1.72, serif, máx ~70ch): la columna de
  lectura.
- **Label** (700, 13px, tracking 0.16em, MAYÚSCULAS, sans): encabezados de
  sección (índice, validación, ficha) y etiquetas de campo.
- **Mono** (400, 13px): identificadores, commits, nombres de archivo y código.

### Named Rules
**La Regla de la Etiqueta Grande.** Ninguna etiqueta baja de 13px; el suelo
pisa sobre terreno blando, no en letra enana.

## Layout

Rejilla de tres columnas `15rem | minmax(0, 1fr) | 19rem`, con ancho máximo de
`92rem` y centrada: **índice** a la izquierda, **columna de lectura** en el
centro, **ficha y validación** a la derecha. El masthead es *sticky* (64px) y la
banda de título lleva el nombre del Work Item. Un colofón cierra la página con
`id`, `base_commit` y `last_review_commit`.

Por debajo de `1080px` la rejilla colapsa a una columna: el índice pasa a una
lista horizontal que envuelve, la ficha baja bajo el lector y el masthead
oculta su leyenda. La columna de lectura se limita a `70ch` en todo breakpoint.

El ritmo de espacio es amplio: `2rem` de separación entre bloques y más aire
**encima** de un encabezado que debajo. Los documentos se separan con una
hairline superior y `2.5rem` de relleno.

## Elevation & Depth

El sistema es **plano**: no usa sombras. La profundidad se expresa con capas
tonales —`canvas` → `surface` → `surface-raised`— y con hairlines de 1px que
separan columnas y filas. Cualquier sombra, desenfoque o cristal decorativo
rompe el mundo del manual impreso.

### Named Rules
**La Regla del Papel sin Sombra.** La profundidad es tonal y por filetes. El
relieve se gana con capas, nunca con sombras.

## Shapes

Radios pequeños y esquinas casi rectas: `3px` en badges, `4px` en callouts,
código y botón. Los bordes son de 1px en todo el sistema. El único gesto
redondeado es la barra de desplazamiento (`999px`). La forma dominante es la
regla y la banda, la de un instrumento impreso, no la de una tarjeta.

## Components

### Buttons
- **Shape:** esquina apenas suavizada (`4px`), sin sombra.
- **Primary:** fondo `{colors.primary-deep}`, texto `#f3f0ff`, padding `13px 0`,
  ancho completo, MAYÚSCULAS con tracking `0.05em`.
- **Hover / Focus:** el fondo sube a `{colors.primary}`; el foco usa un contorno
  violeta de 2px con `offset: 2px`.
- **Disabled:** opacidad `0.6`.

### Badges (estado)
- **Style:** fondo `{colors.surface-raised}`, texto `{colors.text-soft}`, borde
  1px, radio `3px`, MAYÚSCULAS 13px.
- **State:** color por estado — `ready` ámbar, `in_progress` violeta, `review`
  azul, `done` verde, `blocked`/`changes_requested` rojo.

### Callouts (validación)
- **Style:** fondo `{colors.surface}`, borde 1px teñido por el nivel, radio
  `4px`; etiqueta `WARN`/`FAIL`/`OK` + mensaje.
- **Regla:** sin filetes laterales gruesos; el color vive en el borde y la
  etiqueta, no en una pestaña.

### Index (navegación)
- **Style:** lista de documentos en mono (14px); los ausentes en gris con
  `(no existe)`; los presentes enlazan a su sección. En móvil, lista horizontal
  que envuelve.

### Colophon
- **Style:** fila de pares etiqueta/valor (`id`, `base_commit`,
  `last_review_commit`) separada por una hairline superior.

## Do's and Don'ts

### Do:
- **Do** mantener el acento violeta en ≤10% de la pantalla; su rareza es el punto.
- **Do** dejar que el frame (masthead, índice, validación, colofón) cargue el
  carácter, y mantener la columna de lectura calmada.
- **Do** usar capas tonales y hairlines de 1px para la profundidad.
- **Do** mantener toda etiqueta en ≥13px.
- **Do** usar los colores de señal solo para estado.

### Don't:
- **Don't** introducir sombras, desenfoques ni cristal decorativo.
- **Don't** usar el violeta (ni ningún color) como decoración: solo marca o señal.
- **Don't** añadir filetes laterales gruesos a callouts o listas.
- **Don't** poner un "kicker"/eyebrow encima de un encabezado.
- **Don't** cargar fuentes, CSS o JS de terceros: solo stacks de sistema y assets
  propios embebidos.
- **Don't** poner el texto de lectura en monoespaciada: la mono es para código y datos.
