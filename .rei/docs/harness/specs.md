# Spec Driven Development

> Este repositorio utiliza Spec Driven Development (SDD) para planificar el trabajo antes de escribir código.

El objetivo del proceso no es producir documentación extensa, sino proporcionar al Implementer el contexto necesario para realizar un cambio correctamente, minimizando ambigüedades y evitando decisiones improvisadas durante la implementación.

---

# Presupuesto de planificación

La planificación debe ser concisa pero completa: debe dar al Implementer lo
necesario para ejecutar, no documentación de relleno.

| Documento | Objetivo | Máximo | Qué evitar |
|-----------|----------|--------|------------|
| `requirements.md` | 10–20 requisitos | 30 | Relleno EARS, repetir el prompt/`description`, requisitos solapados |
| `design.md` | 150–300 líneas | 400 | Código, boilerplate, arquitectura general, repetir requisitos |
| `tasks.md` | 8–12 tareas | 15 | Tareas de relleno, duplicadas o no ejecutables |

**Regla única:** si un documento supera el máximo, **divide el Work Item o
justifícalo en una línea dentro del propio documento**. No rellenes para llegar al
objetivo ni recortes requisitos necesarios para bajar del máximo.

`rei validate` emite un `WARN` (nunca un `FAIL`) al superar el máximo en los
estados de planificación activos (`ready`, `in_progress`, `review`,
`changes_requested`). Un `WARN` exige justificación o división, pero no es un
rechazo automático.

---

# Objetivo

Todo Work Item debe ser planificado antes de implementarse.

La planificación siempre es responsabilidad del **Spec Author** y debe ser aprobada explícitamente por el usuario antes de comenzar la implementación.

El tipo de planificación depende del tipo de Work Item.

---

# Estructura

Cada Work Item dispone de una carpeta propia dentro de `.rei/specs/`. El nombre de la carpeta es su `id` inmutable.

```
.rei/specs/
└── <work-item-id>/
    ├── meta.json
    ├── requirements.md   (solo Features)
    ├── design.md         (solo Features)
    ├── tasks.md          (solo Features)
    └── plan.md           (solo Tasks)
```

---

# Tipos de planificación

## Feature

Las Features requieren una planificación completa.

El Spec Author debe generar:

- `requirements.md`
- `design.md`
- `tasks.md`

Cada documento cumple un propósito distinto.

---

### requirements.md

Describe **qué debe hacer** el sistema.
No contiene decisiones técnicas.
Los requisitos deben redactarse utilizando **EARS (Easy Approach to Requirements Syntax)**.

Reglas:

- [ ] Cada requisito tiene un identificador estable (`R1`, `R2`, ...)
- [ ] Cada requisito representa una única responsabilidad
- [ ] Cada requisito es verificable
- [ ] Ningún requisito describe decisiones de implementación
- [ ] Los requisitos son centrales, observables y testables
- [ ] Sin relleno EARS ni repetición del prompt o de `meta.json.description`

---

### EARS

Los requisitos pueden escribirse utilizando cualquiera de los siguientes patrones.

| Tipo | Plantilla |
|------|-----------|
| Ubicuo | El sistema DEBE... |
| Evento | CUANDO... el sistema DEBE... |
| Estado | MIENTRAS... el sistema DEBE... |
| Opcional | DONDE... el sistema DEBE... |
| Error | SI... ENTONCES el sistema DEBE... |

---

### design.md

Describe **cómo se implementará** la Feature.

Debe documentar únicamente las decisiones necesarias para implementar el cambio:
las decisiones con peso, las alternativas descartadas y los archivos afectados.

Puede incluir:

- archivos afectados;
- componentes implicados;
- decisiones técnicas;
- restricciones;
- alternativas descartadas.

No debe convertirse en documentación de arquitectura general. Evita el código, el
boilerplate y la repetición de los requisitos: eso ya vive en `requirements.md`.

---

### tasks.md

Divide la implementación en pasos concretos y ordenados.

Cada tarea representa una unidad de trabajo que el Implementer ejecutará de forma secuencial.

Las tareas deben ser ejecutables y no duplicarse: nada de tareas de relleno.

Formato recomendado (cada tarea indica los requisitos que implementa):

```md
- [ ] T1 — ... (R1, R3)
- [ ] T2 — ... (R2)
- [ ] T3 — ...
```

El Implementer marcará cada tarea como completada (`[x]`) en `tasks.md` **inmediatamente** al terminarla, antes de continuar con la siguiente.

---

## Task

Las Tasks usan una planificación simplificada: el Spec Author genera únicamente
`plan.md`. El formato y el presupuesto se definen en
`.rei/docs/harness/task.md`.

**Criterio de tipo:** una Task cubre documentación o contratos, o un cambio
pequeño y localizado, en ambos casos **sin cambio de comportamiento ni de
arquitectura**. Cualquier cambio de comportamiento o de arquitectura es una
**Feature**, aunque el diff sea pequeño.

---

# Estados

Todos los Work Items utilizan los mismos estados.

Los estados válidos y sus transiciones se definen en `workflow.md`.
---

# Aprobación humana

La planificación nunca se implementa automáticamente.

Cuando el Spec Author finaliza su trabajo, el Leader debe detener el workflow y esperar una aprobación explícita del usuario.

Solo después de esa aprobación el Work Item puede pasar a `in_progress`.

---

# Reglas generales

- Todo Work Item comienza en estado `pending`.
- Todo Work Item debe pasar por el Spec Author.
- Ningún Work Item puede implementarse sin planificación aprobada.
- El Implementer nunca modifica la planificación, salvo marcar tareas o pasos como completados en `tasks.md` o `plan.md`.
- El Reviewer verifica la implementación contra la planificación aprobada.
