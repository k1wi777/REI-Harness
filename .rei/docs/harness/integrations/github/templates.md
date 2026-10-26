# GitHub — plantillas (issue, spec, épica, hija y comentarios)

> **Archivo de dominio del playbook de GitHub.** Cárgalo solo cuando la tabla
> «Qué leer según la acción» de [`../github.md`](../github.md) indique que vas a
> **escribir** un cuerpo, comentario o bloque; se usa además del archivo del
> flujo correspondiente.
>
> Referencias: [índice](../github.md) · [núcleo v1](core.md) ·
> [backlog](backlog.md) · [errores](errors.md)

---

# 1. Cuerpo de la issue de Work Item

El marcador va en la **primera línea**. El título es `meta.json.title`. El
cuerpo debe ser autosuficiente: alguien sin acceso al repo debe entender el
problema, la decisión y el alcance. La línea `Base:` (`<rama>` es la base elegida
en la puerta y `<sha>` su sha corto) solo se escribe en issues **nuevas**; nunca
se añade a issues ya creadas.

```md
<!-- rei:work-item:<id> -->

Base: <rama> @ <sha>

## Contexto y decisión

<description de meta.json: problema, decisión y acuerdo>

## Requisitos (resumen)

- **R1** — ...
- **R2** — ...

## Decisiones de diseño relevantes

- ...

## Fuera de alcance

- ...

## Referencias

- Work Item: `.rei/specs/<id>/`
- Requisitos: `.rei/specs/<id>/requirements.md`
- Tareas: `.rei/specs/<id>/tasks.md`
```

El cuerpo **no** incluye el `design.md` completo: basta el contexto, los
requisitos centrales, las decisiones con peso, el fuera de alcance y los enlaces.
Si `.rei` **no se versiona** o el usuario pide trazabilidad completa, publica el
design como **comentario** idempotente con el marcador
`<!-- rei:comment:design:<id> -->` ([§7](#7-comentario-de-design-opcional)),
nunca en el cuerpo.

---

# 2. Bloque de spec publicado

```md
<!-- rei:spec:<id> -->

## Plan (REI)

- Work Item: `.rei/specs/<id>/`
- Requisitos: <resumen R1..Rn>
- Tareas: <resumen>
- Referencias: `.rei/specs/<id>/requirements.md`, `design.md`, `tasks.md`

<!-- /rei:spec:<id> -->
```

El bloque resume requisitos y tareas y enlaza al Work Item; no incrusta el
`design.md`. Mismo criterio que §1: el design completo solo como comentario
idempotente `<!-- rei:comment:design:<id> -->` si `.rei` no se versiona o el
usuario pide trazabilidad.

---

# 3. Épica

El marcador va en la **primera línea**; el bloque `children` puede crearse vacío
y rellenarse después de crear las hijas.

```md
<!-- rei:epic:modulo-x -->

## Contexto del módulo

<problema y situación actual>

## Objetivo

<qué debe conseguirse>

## No-objetivos

- ...

## Hijas

<!-- rei:children -->
- [ ] #12 <título de la hija>
- [ ] #13 <título de la hija>
<!-- /rei:children -->

## Orden y dependencias

1. #12 → #13: <motivo>

## Criterio de éxito

<condición verificable>

<details>
<summary>Plan completo (opcional)</summary>

<contenido opcional>

</details>
```

---

# 4. Hija de backlog

El marcador va en la **primera línea**.

```md
<!-- rei:backlog:modulo-x/hija-a -->

## Contexto

<por qué existe esta hija>

## Alcance

- ...

## Fuera de alcance

- ...

## Criterios de aceptación (borrador)

- [ ] ...

## Épica

#100 — <título de la épica>

## Assignee

<@login | sin asignar>

## Notas y dependencias

- ...
```

---

# 5. Comentario de implementación

```md
<!-- rei:comment:impl:<id> -->

## Implementación terminada

- Work Item: `<id>`
- Estado: `review`
- Resumen: ...
- Archivos principales: ...
- Verificación: ...
- Detalle: `.rei/progress/work-items/<id>/impl.md`
```

---

# 6. Comentario de revisión

```md
<!-- rei:comment:review:<id> -->

## Revisión: <done|changes_requested|blocked>

- Work Item: `<id>`
- Verificaciones: ...
- Observaciones: ...
- Detalle: `.rei/progress/work-items/<id>/review.md`
```

---

# 7. Comentario de design (opcional)

Solo si `.rei` **no se versiona** o el usuario pide trazabilidad completa. Nunca
sustituye al cuerpo de la issue ni se duplica.

```md
<!-- rei:comment:design:<id> -->

## Design (REI)

- Work Item: `<id>`
- Referencias: `.rei/specs/<id>/design.md`

<contenido íntegro de design.md>
```
