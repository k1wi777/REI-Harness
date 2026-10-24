# Visor de especificaciones (`rei view`)

- **Superficie:** la página que `rei view` sirve para un Work Item.
- **Modo:** Read (documentación) con una acción de aprobación.
- **Audiencia y trabajo:** devs que revisan la planificación de un agente antes de aprobarla.
- **Acción/tarea:** entender la spec y, si procede, aprobar (ready → in_progress).
- **Contenido/prueba real:** documentos del Work Item (requirements/design/tasks o plan, impl, review), metadatos, validación y sesión.
- **Restricciones:** sin terceros en JS/CSS; runtime-agnóstico; UI en español; **en oscuro**; **tipografía pequeña generosa** (etiquetas ≥13-14px, cuerpo 16-17px).

## Dirección elegida

**Manual de referencia técnica**, sobre campo oscuro cálido y acento **violeta oscuro** (no naranja).
Momento memorable: los **callouts de validación** (WARN/OK) y la **tabla de historial de revisión**;
el frame se lee como un manual de servicio.

## Decisiones abiertas

- Ninguna: esta ronda cierra la dirección visual.

## Direction contract

THESIS: Un manual de servicio para revisar specs: el frame hace de cabecera, índice y aparato; la columna de lectura es un instrumento, no un póster. Refusa el shell de "tarjetas en un dashboard".

OWN-WORLD: Grafito cálido oscuro de fondo, tinta hueso, atenuado gris; **acento violeta oscuro** (realce más claro para interacción), ámbar para WARN y verde apagado para OK; callouts, tablas de revisión, hairlines y numerales de sección. Reconocible con todo el contenido borrado.

STORY: el lector entiende de un vistazo qué Work Item es, en qué estado está, qué documentos tiene y qué dice la validación; y puede aprobar sin salir de la página.

FIRST VIEWPORT: masthead con marca y metadatos del Work Item; índice de documentos; columna de lectura con el título y el primer encabezado; panel de validación y estado; botón Aprobar. El título y el estado mandan la jerarquía.

FORM: manual de referencia técnica (posición 1 de la lista ordenada); seed key 4756bb8e.

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance.
