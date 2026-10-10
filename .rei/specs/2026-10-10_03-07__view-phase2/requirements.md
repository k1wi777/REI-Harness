# Requirements — Visor de especificaciones (Fase 2)

> Work Item: `2026-10-10_03-07__view-phase2` (feature).
>
> Los requisitos describen **qué** debe hacer el sistema, no **cómo**. Se redactan
> en EARS. El sujeto "el visor" designa la vista servida por `rei view`.
>
> Esta fase mejora la lectura y pule lo que quedó fuera de la v1
> (`2026-10-09_23-45__spec-browser-viewer`).

---

# 1. Ver qué cambió desde la última revisión

**R1.** El visor DEBE determinar el commit base de revisión de un Work Item:
`last_review_commit` cuando esté definido; en su defecto, `base_commit`.

**R2.** CUANDO el commit base de revisión existe, el visor DEBE permitir ver, para
cada documento del Work Item que haya cambiado respecto de ese commit, el material
añadido y eliminado (diff).

**R3.** El visor DEBE indicar, en el índice de documentos, cuáles cambiaron desde
el commit base de revisión.

**R4.** El visor DEBE mostrar de forma visible cuál es el commit base de revisión
utilizado, con su identificador y su origen (`last_review_commit` o `base_commit`).

**R5.** CUANDO un documento mostrado no ha cambiado desde el commit base de
revisión, el visor DEBE indicarlo explícitamente.

**R6.** SI no hay commit base de revisión registrado, ENTONCES el visor DEBE
indicarlo y DEBE seguir mostrando los documentos con normalidad.

**R7.** SI el proyecto no es un repositorio git o git no está disponible, ENTONCES
el visor DEBE informar de que no puede calcular qué cambió y DEBE cargar la página
con normalidad.

**R8.** SI un documento existe en disco pero no en el commit base de revisión
(documento nuevo), ENTONCES el visor DEBE presentarlo como material añadido por
completo.

---

# 2. Índice de secciones (TOC)

**R9.** El visor DEBE mostrar, dentro de cada documento renderizado, un índice de
sus secciones construido a partir de sus encabezados.

**R10.** El índice de secciones DEBE respetar la jerarquía de los encabezados,
anidando los niveles inferiores bajo los superiores.

**R11.** CUANDO el usuario activa una entrada del índice de secciones, el visor
DEBE desplazarse hasta esa sección del documento.

**R12.** SI un documento tiene menos de dos encabezados, ENTONCES el visor NO DEBE
mostrar un índice de secciones para ese documento.

---

# 3. Reconexión de la recarga en vivo

**R13.** SI la conexión de recarga en vivo se interrumpe o falla, ENTONCES el
visor DEBE intentar reconectarla automáticamente, sin intervención del usuario.

**R14.** MIENTRAS la conexión de recarga en vivo esté interrumpida, el visor DEBE
indicarlo de forma discreta y DEBE retirar el aviso al restablecerse.

**R15.** CUANDO la conexión de recarga en vivo se restablece, el visor DEBE
reanudar la recarga automática de los documentos.

---

# 4. Masthead

**R16.** El masthead DEBE mostrar una única descripción del visor; NO DEBE
presentar dos descripciones del mismo objeto.

---

# 5. Contraste del estado activo del índice

**R17.** El fondo del documento activo en el índice DEBE usar la capa tonal
`surface-raised` (en lugar de `surface`), de modo que el estado activo se
distinga de forma perceptible con el fondo del índice.

---

# 6. Alineación de tokens

**R18.** El CSS del visor DEBE declarar, como variables de token, los roles de
color y de tipografía declarados en `DESIGN.md`, y DEBE consumirlos en las reglas
que los realizan.

**R19.** El CSS NO DEBE contener valores literales de color fuera del bloque de
definición de tokens; en particular, el botón de aprobación NO DEBE usar colores
hardcodeados.

**R20.** Los valores de los tokens compartidos DEBEN coincidir entre `DESIGN.md`,
`design.json` (para los tokens que declara) y el CSS; toda divergencia DEBE
resolverse de forma que el sistema de diseño y el CSS queden alineados.

---

# 7. No regresión y restricciones

**R21.** El visor NO DEBE cargar JavaScript ni CSS de terceros; DEBE seguir
usando únicamente assets propios embebidos.

**R22.** El visor DEBE seguir siendo runtime-agnóstico: su funcionamiento NO DEBE
depender de hooks nativos de un runtime concreto.

**R23.** El visor DEBE conservar el comportamiento actual: el selector de
documento, la aprobación en dos pasos, los enlaces profundos (a documento y a
sección) y la conservación de la posición de lectura en la recarga.

**R24.** La interfaz del visor DEBE estar en español.

**R25.** La Feature DEBE resolverse con la biblioteca estándar y las dependencias
ya presentes (`goldmark`); SI introduce una dependencia Go nueva, ENTONCES DEBE
justificarlo y evitar dependencias transitivas.

**R26.** El feedback y la selección de cambios sobre las specs (que mapea al Caso C
del Spec Author) NO forman parte de esta Feature.
