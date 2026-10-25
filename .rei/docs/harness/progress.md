# Progress

> Este documento define cómo se documenta el progreso del trabajo dentro del proyecto.
>
> La carpeta `.rei/progress/` constituye el registro vivo del estado de desarrollo y permite que cualquier agente o persona pueda comprender qué ocurrió durante una sesión, retomarla si fue interrumpida y consultar el historial del proyecto.
>
> Ningún agente debe inventar nuevos formatos. Todos los archivos de `.rei/progress/` deben respetar las plantillas canónicas de `.rei/templates/`.
>
> Las plantillas se generan y restablecen con el CLI (`rei`) — **nunca las escribas a mano**.

---

# Estructura

```text
.rei/progress/
│
├── current.md              # Sesión activa (único Work Item en curso)
├── history.md              # Bitácora histórica (append-only)
│
└── work-items/             # Detalle de cada Work Item
    └── <work-item-id>/     # Usa el mismo id que .rei/specs/<work-item-id>/
        ├── impl.md         # Reporte del Implementer
        ├── review.md       # Reporte del Reviewer
        └── spec.md         # Bloqueo del Spec Author (solo si aplica)
```

Los archivos de sesión (`current.md`, `history.md`) viven en la raíz de `.rei/progress/`.

Cada Work Item tiene su propia carpeta `.rei/progress/work-items/<work-item-id>/`, análoga a `.rei/specs/<work-item-id>/`, donde se concentran todos los documentos generados durante su ciclo de vida.

---

# current.md

Representa el estado **actual** de la sesión.

> La plantilla canónica vive en `.rei/templates/current.md` (fuente única de verdad).
> No la copies a mano: usa el CLI (`rei`).

> Este archivo es el **contexto vivo de la sesión activa**: sirve para retomar un
> Work Item exactamente donde quedó si la sesión se interrumpe. El registro
> duradero y el diagnóstico global del conjunto de Work Items viven en
> `.rei/specs/<work-item-id>/meta.json` (consulta `rei items status`).

El **Spec Author** lo inicializa ejecutando
`rei session start <work-item-id> <type>` al comenzar la
planificación (`pending`) y lo deja en `ready` al terminar, esperando aprobación humana.

El **Implementer** lo actualiza al iniciar la implementación (`in_progress`) y lo mantiene durante toda la ejecución hasta pasar a `review`.

Debe mantenerse actualizado durante todo el ciclo de vida del Work Item — desde `pending` hasta `review`.

No debe rellenarse únicamente al finalizar el trabajo.

Su propósito es permitir que una sesión pueda retomarse en cualquier momento, incluso durante la planificación o la espera de aprobación.

Al cerrar el Work Item, `rei session archive` mueve su resumen a `history.md` y restablece este archivo a su estado inicial. Solo archiva si `meta.json` está en un estado terminal (`done` o `blocked`) y el `Estado` de este archivo coincide con él; en caso contrario rechaza el archivado con código 1 y explica la causa.

---

# history.md

Es la bitácora histórica del proyecto.

> La plantilla canónica vive en `.rei/templates/history.md` (fuente única de verdad).
> No la copies a mano: `rei session archive` la usa.

Su contenido es **append-only**.

Nunca deben modificarse entradas anteriores.

Al finalizar correctamente un Work Item, `rei session archive` añade el resumen de `current.md` al final de este archivo. La entrada refleja el estado final registrado en `current.md` (el veredicto del Reviewer), porque el archivado ocurre después de `rei status set <id> done`.

Cada entrada debe resumir:

- fecha;
- Work Item;
- agente;
- trabajo realizado;
- archivos modificados;
- resultado de la verificación;
- estado final.

---

# .rei/progress/work-items/<work-item-id>/impl.md

Documento generado por el **Implementer** en `.rei/progress/work-items/<work-item-id>/impl.md`.

Describe el trabajo realizado durante la implementación.

Debe ser **conciso**: tope blando de **~80 líneas / ~600 palabras**. Si se
supera, `rei validate` emite un `WARN` (nunca un `FAIL`). Su estructura puede
adaptarse según la naturaleza del Work Item, pero siempre debe incluir como
mínimo:

- resumen de la implementación;
- archivos modificados;
- cambios realizados;
- proceso de verificación;
- observaciones relevantes.

Este archivo constituye la evidencia principal para la revisión.

---

# .rei/progress/work-items/<work-item-id>/review.md

Documento generado por el **Reviewer** en `.rei/progress/work-items/<work-item-id>/review.md`.

Resume el resultado de la revisión realizada sobre el Work Item.

Debe ser **conciso**: tope blando de **~80 líneas / ~600 palabras**; superarlo
produce un `WARN` en `rei validate` (nunca un `FAIL`). Debe incluir como mínimo:

- estado final (`done`, `changes_requested` o `blocked`);
- verificaciones realizadas;
- observaciones;
- acciones requeridas si la revisión fue rechazada.

---

# .rei/progress/work-items/<work-item-id>/spec.md

Documento generado únicamente cuando el **Spec Author** no puede completar correctamente la planificación.

Debe explicar claramente:

- motivo del bloqueo;
- información faltante;
- decisiones que requieren intervención humana.

Si no existen bloqueos, este archivo no debe crearse.

---

# Responsabilidades

| Archivo | Responsable | Cómo |
|----------|-------------|------|
| `current.md` (inicio: `pending`) | Spec Author | `rei session start <id> <type>` |
| `current.md` (implementación: `in_progress` → `review`) | Implementer | edición directa |
| `history.md` + reset de `current.md` | Reviewer | veredicto en `current.md` → `rei status set <id> done` → `rei validate` → `rei session archive` |
| `work-items/<work-item-id>/impl.md` | Implementer | edición directa |
| `work-items/<work-item-id>/review.md` | Reviewer | edición directa |
| `work-items/<work-item-id>/spec.md` | Spec Author | edición directa (solo en bloqueo) |

---

# Reglas

- Mantén `current.md` actualizado durante todo el ciclo de vida del Work Item.
- Crea `.rei/progress/work-items/<work-item-id>/` al iniciar el trabajo sobre un Work Item.
- Nunca sobrescribas el historial.
- No elimines documentación existente.
- Utiliza siempre las plantillas canónicas de `.rei/templates/` (vía el CLI `rei`).
- Todo Work Item debe dejar evidencia suficiente para poder comprender qué ocurrió sin depender del historial del chat.
