# Workflow

> Este documento define los flujos de trabajo disponibles dentro del repositorio.
> El Leader debe consultarlo para decidir qué proceso aplicar antes de delegar cualquier trabajo.

---

# Objetivo

No todos los cambios requieren el mismo nivel de planificación.

Antes de crear un nuevo Work Item, el Leader debe decidir si la solicitud corresponde a una **Feature**, a una **Task** o a un **cambio directo** autorizado, aplicando siempre el flujo más simple que permita mantener la calidad del proyecto.

---

# Modo directo (fuera del workflow)

El modo directo permite al Leader aplicar un cambio trivial solicitado
explícitamente, sin planificación ni estado. **No es un tipo de Work Item ni un
estado del ciclo**: es una vía fuera del workflow, sobre la rama y el árbol de
trabajo actuales.

**Disparador.** Solo se activa cuando el usuario pide un «cambio directo» o «sin
Work Item». El Leader nunca lo inicia por iniciativa propia ni a partir de una
petición que no lo solicite de forma explícita.

**Elegibilidad.** El cambio debe limitarse a documentación, erratas, formato,
comentarios o ajustes triviales sin cambio de comportamiento.

**Negativa.** Si el cambio afecta comportamiento, seguridad, datos, dependencias,
API pública, arquitectura, más de un archivo o cualquier cosa que necesite
especificación o revisión, el Leader **debe negarse** a aplicarlo en modo directo
y ofrecer continuar como **Task** o **Feature** mediante el workflow. La
insistencia del usuario no es autorización: no existe override de la negativa.

**Autorización.** Incluso si el cambio es elegible, el Leader pide confirmación
explícita antes de tocar nada e informa de que no habrá planificación, rama ni
registro. Si duda de la elegibilidad, avisa y exige esa confirmación. El silencio
o una respuesta ambigua no son autorización; el defecto es **no** aplicar el modo
directo.

**Sin registro.** El modo directo no crea Work Item, spec, `meta.json`,
`current.md`, `history.md`, `direct.md`, sesión, issue ni rama, y no modifica el
estado ni los documentos de ningún Work Item existente. La única traza es el
diff/commit que el usuario decida y la conversación. Si hay un Work Item en
`in_progress`, el Leader avisa de que el cambio compartirá árbol y diff antes de
aplicarlo.

**Verificación.** Siempre `rei check`; además `go test` o `make build` según el
cambio. El Leader reporta los comandos ejecutados y su resultado. El modo directo
no lanza al Reviewer salvo petición explícita del usuario.

**Entrega.** Sin orden explícita del usuario no hay commit, push, PR ni merge.

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
| `done` | Trabajo completado y aprobado (estado terminal). |
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
- Si en cualquier punto el trabajo no puede continuar: `* → blocked`, salvo desde `done`, que es terminal: `done → blocked` solo se permite con `--force` como transición no estándar.
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
- El modo directo autorizado es una vía ajena a los Work Items y no exime del
  resto de reglas: no crea ni altera Work Items y se rige por la sección «Modo
  directo (fuera del workflow)».
- `done` es terminal: no se sale de él por una transición estándar. `rei status
  set <id> blocked --force` es la vía explícita para una transición no estándar y
  solo se usa con una orden del usuario.
