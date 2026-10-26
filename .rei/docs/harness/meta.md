# meta.json

> `meta.json` describe un Work Item y permite a los agentes conocer su tipo, estado e información básica.

Cada carpeta dentro de `.rei/specs/` debe contener un único archivo `meta.json`. El nombre de la carpeta es el `id` inmutable del Work Item.

```text
.rei/specs/
└── <work-item-id>/
    ├── meta.json
    └── ...
```

---

# Estructura

```json
{
  "id": "",
  "title": "",
  "description": "",
  "type": "",
  "status": "",
  "created_at": "",
  "base_commit": "",
  "last_review_commit": ""
}
```

---

# Campos

## id

Identificador inmutable del Work Item. Deberá coincidir exactamente con el mismo nombre de su carpeta en `.rei/specs/` y `.rei/progress/work-items/`.

Formato obligatorio:

```text
YYYY-MM-DD_HH-mm__slug-en-kebab-case
```

La fecha y hora corresponden al momento de creación del Work Item. El identificador no cambia aunque el trabajo se reanude o cambie su título.

Si ya existe un identificador igual, añade un sufijo numérico al slug (`-2`, `-3`, ...) para evitar colisiones sin modificar la fecha y hora originales.

## title

Nombre corto y descriptivo del Work Item.
Debe resumir claramente el objetivo del trabajo.

---

## description

Descripción técnica del trabajo.
Debe resumir el acuerdo alcanzado con el usuario y proporcionar el contexto necesario para continuar el workflow.
No debe copiar literalmente la conversación.

---

## type

Define el tipo de Work Item.
Valores válidos:

- `feature`
- `task`

---

## status

Define la etapa actual del workflow.
Los estados válidos y sus transiciones se definen en `workflow.md`.

`done` es un estado **terminal**: las transiciones estándar no salen de él
(`done → blocked` no está permitida). La única vía es `rei status set <id>
<estado> --force`, reservada a una transición no estándar con una orden explícita
del usuario.

> `meta.json` es el **registro duradero** de cada Work Item y la referencia para
> diagnosticar el conjunto de Work Items (`rei items status`).
> El contexto vivo de la sesión activa vive en `.rei/progress/current.md` (para
> retomarla); ambos se mantienen y son complementarios.

---

## created_at

Fecha y hora exactas de creación del Work Item en formato ISO 8601, incluyendo la zona horaria. Debe corresponder al momento utilizado para construir el prefijo del campo `id`.

## base_commit

SHA del commit que marca el **inicio de la implementación**. Lo registra el Leader al aprobar el Work Item:

```bash
rei commit set <work-item-id> base_commit
```

Vacío si el proyecto no usa git. `rei review-diff` lo usa como punto base del paquete de revisión.

## last_review_commit

SHA del punto ya revisado. Lo registra el Reviewer al rechazar (`changes_requested`):

```bash
rei commit set <work-item-id> last_review_commit
```

Si está presente, `rei review-diff` calcula el diff desde ese punto, para que la próxima revisión vea solo los cambios pedidos y no revalide lo anterior. Vacío si no aplica.

> Para que acote de verdad, la implementación debe estar **commiteada** antes de solicitar la revisión. Si no lo está, el diff cubre todo lo pendiente desde `base_commit`: sigue siendo correcto, solo menos acotado.
