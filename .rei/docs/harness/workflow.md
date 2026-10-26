# Workflow

> Este documento define los flujos de trabajo disponibles dentro del repositorio.
> El Leader debe consultarlo para decidir qué proceso aplicar antes de delegar cualquier trabajo.

---

# Objetivo

No todos los cambios requieren el mismo nivel de planificación.

Antes de crear un nuevo Work Item, el Leader debe decidir si la solicitud corresponde a una **Feature** o a una **Task**, aplicando siempre el flujo más simple que permita mantener la calidad del proyecto.

---

# Tipos de Work Item (criterios de decisión)

## Feature

Una Feature representa un cambio funcional que requiere una fase completa de planificación antes de implementarse.

Normalmente implica uno o varios de los siguientes casos:

- incorporación de nuevas funcionalidades;
- cambios de arquitectura;
- modificaciones del comportamiento existente, por menores que sean;
- pequeños ajustes de interfaz que cambian lo que el usuario ve o recibe
  (mensajes, validaciones, textos visibles);
- refactorizaciones relevantes;
- trabajo cuyo alcance necesita definirse antes de implementarse.

Las Features utilizan el flujo completo de Spec Driven Development y generan `requirements.md`, `design.md` y `tasks.md`.

**Selecciona Feature cuando:**

- el trabajo todavía necesita ser especificado;
- existen varias decisiones de diseño por tomar;
- el cambio afecta distintas partes del proyecto;
- el trabajo requiere documentar requisitos;
- hay un cambio de comportamiento o de arquitectura, aunque el diff sea pequeño.

La documentación, los contratos y los cambios pequeños sin cambio de
comportamiento **no** son Features: se planifican como Task.

---

## Task

Una Task representa documentación, contratos o un cambio pequeño y localizado,
siempre que no altere el comportamiento ni la arquitectura y su alcance pueda
definirse completamente mediante una planificación breve.

Algunos ejemplos son:

- documentación y contratos;
- cambios que no alteran el comportamiento (refactor interno, formato,
  comentarios);
- corrección de erratas o ajustes de texto que no cambian lo que el usuario ve o
  recibe;
- modificaciones puntuales sobre código existente sin cambio de comportamiento.

Las Tasks utilizan un flujo simplificado y generan únicamente `plan.md`.

**Selecciona Task cuando:**

- el objetivo está claramente definido;
- el cambio es pequeño y localizado;
- no hay cambio de comportamiento ni de arquitectura;
- no es necesario generar requisitos ni diseño detallado.

---

**Regla de desempate:** si existe cualquier duda sobre el tipo, selecciona **Feature**.

Ambos tipos tienen un presupuesto de concisión: Feature en
[`specs.md`](specs.md) (requisitos, design y tareas) y Task en
[`task.md`](task.md) (`plan.md`). Si se excede el máximo, divide el Work Item o
justifícalo en el propio documento. `rei validate` avisa (`WARN`) sin bloquear.

---

# Aprobación humana

Ningún Work Item puede avanzar desde `ready` hasta `in_progress` sin una aprobación explícita del usuario.

El Leader debe detener siempre el workflow en este punto y esperar la decisión humana.

---

# Estados

Todos los Work Items utilizan los mismos estados.

| Estado | Significado |
|---------|-------------|
| `pending` | Trabajo creado, pendiente de planificación. |
| `ready` | Planificación terminada y esperando aprobación humana. |
| `in_progress` | Implementación en curso. |
| `review` | Implementación finalizada, pendiente de revisión. |
| `done` | Trabajo completado y aprobado. |
| `blocked` | El trabajo no puede continuar. |
| `changes_requested` | La revisión rechazó la implementación y requiere modificaciones. |

> `meta.json.status` es el registro duradero de la etapa de cada Work Item (para
> diagnóstico global); `.rei/progress/current.md` refleja la sesión activa para
> retomarla. Son complementarios.

---

# Flujo general

El flujo es el mismo para Features y Tasks — únicamente cambia la planificación que genera el Spec Author (ver "Tipos de Work Item").

pending → spec_author → ready → ⏸ aprobación humana → in_progress → implementer → review → reviewer → done

Durante todo el flujo, `.rei/progress/current.md` refleja el estado vivo del Work Item: el **Spec Author** lo inicializa en `pending` y lo deja en `ready`; el **Implementer** lo actualiza en `in_progress` y `review`. Los reportes de cada agente se almacenan en `.rei/progress/work-items/<work-item-id>/`, usando el mismo `id` de `.rei/specs/<work-item-id>/`.

**Ramificaciones:**

- Si el Reviewer rechaza el trabajo: `review → changes_requested`. El Leader relanza al Implementer y el Work Item vuelve a `in_progress`.
- Si en cualquier punto el trabajo no puede continuar: `* → blocked`.
---

# Reglas generales

- Solo puede existir un Work Item en estado `in_progress`. El invariante es
  duro: lo aplican `rei status set <id> in_progress` (también con `--force`),
  `rei check`, `rei doctor` e `rei items status`. Si se rechaza, cierra
  (`review`/`done`), bloquea o revisa el Work Item activo antes de continuar.
- Todo Work Item comienza en `pending`.
- Todo Work Item debe pasar por `spec_author`.
- Ningún Work Item puede omitir la aprobación humana.
- Ningún Work Item puede finalizar sin pasar por `reviewer`.
