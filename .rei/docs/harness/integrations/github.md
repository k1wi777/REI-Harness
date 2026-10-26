# Integración opcional con GitHub Issues (v2)

> **Índice de enrutado del playbook.** GitHub es una **proyección** del trabajo, no
> su fuente de verdad. `.rei` manda sobre la ejecución del Work Item; la issue manda
> sobre el problema, el alcance y la discusión. Toda la integración es **opcional**:
> si está desactivada o no configurada, el flujo es idéntico al que no usa GitHub.
>
> La capa determinista es el comando `rei github`. Las acciones reales sobre GitHub
> las ejecuta el **Leader** con `gh` y `git`, siempre bajo las reglas de este
> playbook: nunca commit, push, PR ni merge sin una orden explícita del usuario. La
> única excepción es la prueba de campo de un Work Item, que puede ejecutar el
> Implementer con la autorización explícita del usuario registrada por el Leader
> ([`github/errors.md §3`](github/errors.md)).
>
> **El CLI nunca llama a GitHub.** No crea, lee ni modifica issues, labels, assignees,
> comentarios, PR ni épicas. No hay subcomandos, flags ni dependencias Go nuevas por
> esta integración.
>
> **Este archivo es solo el índice** (alcance, configuración, vocabulario, puerta y
> enrutado). Los comandos, procedimientos y plantillas viven en un único archivo de
> dominio: carga **solo** el que indique la tabla de §5.

# 1. Alcance

Incluido en v1: configuración y diagnóstico, vínculo issue ↔ Work Item con marcador
y `links.json`, puerta única de las features (issue + rama), comentarios automáticos,
PR y merge bajo orden, adopción, labels y errores con causa. Detalle:
[`github/core.md`](github/core.md).

Incluido en v2: épicas, backlog de hijas, adopción y promoción con publicación de la
spec, base por Work Item en la puerta, asignación, progreso y cierre de la épica, y
errores/reintentos ampliados. Detalle: [`github/backlog.md`](github/backlog.md).

Fuera de alcance: Projects, milestones, sub-issues nativas, PRs desde
fork→upstream, sincronización bidireccional, Work Items `pending` locales como
sustituto del backlog, autoasignación por carga, time tracking, proveedores distintos
de GitHub, commits/push o merge automáticos, `--admin`/auto-merge y flags de
acciones GitHub (issues, PR, merge) en `rei`.

**Alcance por tipo de Work Item.** La puerta «issue + rama» es **solo para Work
Items de tipo `feature`**. Una `task` **no** genera issue ni rama: se trabaja en
la rama base sin tocar GitHub y **no existe opt-in** para Tasks. No se pregunta la
puerta a una `task` «por si acaso» ni se usa una label `rei:task`.

**Regla dura:** si `integrations.github` no existe o `enabled` es `false`, el Leader
**no menciona GitHub**, no ejecuta `gh`, no lee ni crea issues y trabaja sobre la rama
base como siempre.

# 2. Configuración y diagnóstico

La configuración vive en `.rei/config.json`, bajo `integrations.github`:

```json
{
  "integrations": {
    "github": {
      "enabled": true,
      "repo": "owner/name",
      "base": "main",
      "mergeStrategy": "squash"
    }
  }
}
```

- **Clave ausente** ⇒ integración **desactivada**. El Leader no menciona GitHub ni
  ejecuta `gh`.
- `enabled: false` ⇒ integración **desactivada**; se conservan `repo`, `base`,
  `mergeStrategy` y `links.json`.
- `base` es solo el **fallback** de la base por Work Item: se usa únicamente si el
  Work Item no tiene base registrada (`links.json.base` ni la línea `Base:` de la
  issue) y no se puede preguntar al usuario. La base efectiva es **por Work Item** y
  se decide en la puerta. Si `base` está vacío ⇒ `main`. `mergeStrategy` vacío ⇒
  `squash`.

La base por Work Item **no** se escribe en `.rei/config.json` ni la modifica `rei
github enable`; vive en `links.json` y en el cuerpo de la issue. `rei github status`
sigue siendo offline y **no** conoce la base efectiva de cada Work Item.

Comandos:

```bash
rei github enable [--repo owner/name] [--base <rama>] [--strategy squash|merge|rebase]
rei github disable   # enabled: false; conserva el resto
rei github status    # solo lectura: estado, valores, remoto, gh y siguiente paso
```

`enable` resuelve cada valor con la precedencia **flag explícito > valor previo de
`integrations.github` > detección desde git > valor por defecto**:

- `--repo owner/name`: si falta, se usa el valor previo; si tampoco existe, se
  detecta con `git remote get-url origin` (SSH o HTTPS de `github.com`). Sin
  flag, valor previo ni remoto resoluble es un error operativo (1) y no escribe.
- `--base <rama>`: si falta, se usa la base previa; si tampoco existe, se detecta
  con `git symbolic-ref --quiet --short refs/remotes/origin/HEAD` y, si falla, se
  usa `main` avisando. Ese valor es solo el **fallback** de la base por Work Item
  y no decide la base de cada Work Item.
- `--strategy squash|merge|rebase`: si falta, se usa el valor previo; por defecto
  `squash`.

No se detecta `origin` cuando ya hay un `repo` previo: un `enable` sin flags no
sobrescribe valores ya configurados. Un flag desconocido, repetido, sin valor o
inválido es uso incorrecto (2) y no escribe `.rei/config.json`; los fallos
operativos (config inválida, remoto no resoluble o escritura) salen con 1. Los
avisos de `gh` ausente o sin autenticar **no bloquean** (sale 0).

`status` es **offline y de solo lectura**: no llama a `gh` ni a la API de GitHub.
Además de estado, valores, remoto y disponibilidad de `gh`, **avisa si el remoto de
`origin` no coincide con `integrations.github.repo`** (rename, clon viejo o fork) y
propone el siguiente paso: resolver el nombre canónico con `gh repo view` y ponerlo
en la config, o actualizar `origin`. Con esa discrepancia **no** concluye que la
configuración esté completa; cuando no hay nada que configurar indica
«Configuración completa; sin acciones de configuración pendientes.», sin afirmar
nada sobre issues o PRs. `disable` y `status` rechazan flags (2).

# 3. Vocabulario

## 3.1 Marcadores

| Elemento | Valor |
|----------|-------|
| Issue de Work Item | `<!-- rei:work-item:<id> -->` |
| Épica | `<!-- rei:epic:<epic-slug> -->` |
| Hija de backlog | `<!-- rei:backlog:<epic-slug>/<item-slug> -->` |
| Comentario de implementación | `<!-- rei:comment:impl:<id> -->` |
| Comentario de revisión | `<!-- rei:comment:review:<id> -->` |
| Comentario de design (opcional) | `<!-- rei:comment:design:<id> -->` |
| Delimitador de la task list de la épica | `<!-- rei:children -->` … `<!-- /rei:children -->` |
| Delimitador de la spec publicada | `<!-- rei:spec:<id> -->` … `<!-- /rei:spec:<id> -->` |

Reglas:

- Los marcadores de identidad (`work-item`, `epic`, `backlog`) van en la **primera
  línea** del cuerpo.
- Los delimitadores `children` y `spec` son **auxiliares**: permiten reemplazar un
  bloque sin destruir el resto del cuerpo. No identifican issues.
- **El marcador identifica; la label es solo una pista humana.** Ningún paso depende
  de la label. La ausencia de marcador **no** se suple con labels: si hay duda, se
  pregunta.
- La búsqueda de idempotencia usa el **marcador exacto**, nunca la label ni el título.
- Un marcador de identidad debe aparecer **una sola vez** en el cuerpo. Si aparece
  más de una vez **o hay más de una issue con el mismo marcador**, se detiene la
  acción y se reporta la anomalía. Nunca se crea ni se sobrescribe nada en ese estado.
- Cambiar el título de una issue **no** cambia su slug ni su marcador.

## 3.2 Labels

| Label | Color | Descripción |
|-------|-------|-------------|
| `rei` | `1D76DB` | Gestionado por REI |
| `rei:epic` | `6F42C1` | Épica: planificación de módulo (no Work Item) |
| `rei:backlog` | `FBCA04` | Hija de backlog pendiente de adopción |
| `rei:feature` | `0E8A16` | Work Item de tipo feature |
| `rei:blocked` | `B60205` | Bloqueado |

Se aseguran con `gh label create --force` antes de usarlas (idempotente: crea o
actualiza color y descripción). Composición:

- **épica**: `rei` + `rei:epic`;
- **hija de backlog**: `rei` + `rei:backlog`;
- **hija adoptada**: conserva `rei`, recibe `rei:feature` y pierde `rei:backlog`;
- **Work Item normal**: `rei` + `rei:feature` (v1);
- **bloqueado**: añade `rei:blocked` sin comentario (v1).

## 3.3 `links.json`

`.rei/specs/<id>/links.json` es un **índice local recuperable**, nunca la fuente de
verdad. Esquema, resolución de la base y recuperación:
[`github/core.md §2`](github/core.md).

## 3.4 Nombre canónico del repositorio

`<repo>` significa siempre el **nombre canónico vigente** `owner/name`, resuelto **al
inicio de cada acción GitHub** (un rename a mitad de sesión invalidaría un valor
cacheado). Si la resolución falla, se informa de causa y siguiente paso y se detiene
la acción; nunca se continúa con el nombre viejo. Comando y detalle:
[`github/core.md §3`](github/core.md).

## 3.5 Rename y fork

GitHub redirige el nombre viejo (301), pero el search API no sigue ese redirect: la
idempotencia **nunca** se apoya en la búsqueda por texto. En un fork se opera siempre
sobre el repositorio configurado (o `origin`), nunca se cambia solo a `upstream`.
Detalle: [`github/core.md §4`](github/core.md).

# 4. Puerta solo Features

**Disparador:** `meta.json` está en `ready`, el usuario aprueba,
`integrations.github.enabled` es `true` **y** `type` es `feature`.

La puerta «issue + rama» **no** se ejecuta para Work Items de tipo `task`: se trabajan
en la rama base sin crear issue ni rama, no se pregunta nada y **no existe opt-in**
para Tasks. No preguntes la puerta a una `task` «por si acaso» ni uses una label
`rei:task`. Procedimiento, captura de base, clasificación y tipos:
[`github/core.md §5`](github/core.md).

# 5. Qué leer según la acción

| Acción | Archivo |
|--------|---------|
| Orientación, configuración, vocabulario, `rei github` (enable/disable/status) | [`github.md`](github.md) (este archivo) |
| Puerta issue+rama y base, `links.json`, nombre canónico, adopción v1, comentarios impl/review, PR/merge | [`github/core.md`](github/core.md) |
| Épicas, backlog/hijas, entrevista/borrador/aprobación, promoción, publicación de spec, asignación, cierre de épica | [`github/backlog.md`](github/backlog.md) |
| Escribir cuerpo de issue, comentario, épica, hija o bloque de spec | [`github/templates.md`](github/templates.md) (además del archivo del flujo) |
| Error, ambigüedad, reintento, recordatorios en `current.md`, prueba de campo | [`github/errors.md`](github/errors.md) |

Carga **solo** el archivo que corresponda a la acción en curso (como máximo `templates.md`
junto al archivo del flujo cuando vayas a escribir). **Nunca** cargues el playbook
completo ni todos los archivos a la vez.

# 6. Enlaces

- [`github/core.md`](github/core.md) — núcleo v1; [`github/backlog.md`](github/backlog.md) — épicas y backlog v2.
- [`github/templates.md`](github/templates.md) — plantillas; [`github/errors.md`](github/errors.md) — errores y recordatorios.
- [`../README.md`](../README.md) — índice de la documentación del arnés.
