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

<details>
<summary>design.md completo</summary>

<contenido íntegro de design.md>

</details>
```

---

# 2. Bloque de spec publicado

```md
<!-- rei:spec:<id> -->

## Plan (REI)

- Work Item: `.rei/specs/<id>/`
- Requisitos: <resumen R1..Rn>
- Tareas: <resumen>
- Referencias: `.rei/specs/<id>/requirements.md`, `design.md`, `tasks.md`

<details>
<summary>design.md completo</summary>

<contenido íntegro de design.md>

</details>
<!-- /rei:spec:<id> -->
```

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
