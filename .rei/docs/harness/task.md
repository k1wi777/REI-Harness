# Task — plan.md

> Las Tasks utilizan una planificación simplificada. El Spec Author genera
> únicamente `plan.md`; no hay `requirements.md`, `design.md` ni `tasks.md`.
>
> Este documento es la referencia del Spec Author cuando el Work Item es de
> `type == task`. Para Features, consulta `specs.md`.

---

# Cuándo usar una Task

Una Task es adecuada cuando el trabajo es documentación o contratos, o un cambio
pequeño y localizado cuyo alcance cabe en un plan breve, **siempre que no cambie
el comportamiento ni la arquitectura**. Si hay un cambio de comportamiento o de
arquitectura, es una **Feature** (ver `workflow.md`). En caso de duda, elige
Feature.

No conviertas una Task en una Feature encubierta: si el plan necesita requisitos o
un diseño detallado, el tipo correcto es Feature.

---

# Presupuesto

`plan.md` debe mantenerse en **150–200 líneas** (máximo 200). Si se excede,
divide el Work Item o justifícalo en una línea dentro del propio plan. No incluyas
`requirements.md` ni `design.md`, ni relleno para llegar a una extensión concreta.

`rei validate` emite un `WARN` (nunca un `FAIL`) si `plan.md` supera las 200
líneas en estados de planificación activos.

---

# plan.md

Resume el trabajo que debe realizar el Implementer sin generar documentación innecesaria.

Debe contener únicamente la información necesaria para implementar correctamente el cambio.

Formato recomendado:

```text
Objetivo

...

Archivos

...

Cambios

...

Restricciones

...

Pasos

- [ ] 1. ...
- [ ] 2. ...
- [ ] 3. ...
```

El Implementer marca cada paso como completado (`[x]`) en `plan.md` **inmediatamente** al terminarlo, antes de continuar con el siguiente.
