# Design — Integración opcional con GitHub Issues (v1)

> Work Item: `2026-10-25_18-00__github-issues-v1` (feature).
>
> Este documento describe **cómo** se implementa la Feature. Las referencias
> `R<n>` apuntan a `requirements.md`.

---

# 1. Estrategia

La Feature tiene dos capas con responsabilidades separadas:

1. **Capa determinista (Go): `rei github enable|disable|status`.**
   Solo lee y escribe la configuración `integrations.github` de
   `.rei/config.json`. Puede invocar `git remote get-url origin`,
   `git symbolic-ref` y `gh auth status` para diagnosticar, pero **nunca** crea
   issues, ramas, comentarios, PR ni merges. Es el único código Go de la
   Feature (R14).

2. **Capa de agente (playbook + contratos).**
   El Leader ejecuta las acciones reales con `gh` y `git switch -c` siguiendo
   `.rei/docs/harness/integrations/github.md`. El `initializer` solo pregunta y
   configura. Esta separación evita acoplar el arnés a GitHub, mantiene las
   acciones bajo control humano y permite que `rei` siga funcionando sin `gh`.

**Fuentes de verdad:**

- El **marcador** `<!-- rei:work-item:<id> -->` en el cuerpo de la issue es la
  fuente de verdad del vínculo.
- `.rei/specs/<id>/links.json` es solo un índice local recuperable (R27, R28).
- `.rei` manda sobre la ejecución; la issue manda sobre el problema y la
  discusión.
- Los fallos de GitHub nunca cambian el estado de REI (R46).

**Vocabulario fijado en v1** para no tener que retrofitear épicas/backlog:

| Elemento | Valor |
|----------|-------|
| Marcador de issue | `<!-- rei:work-item:<id> -->` |
| Comentario de implementación | `<!-- rei:comment:impl:<id> -->` |
| Comentario de revisión | `<!-- rei:comment:review:<id> -->` |
| Labels v1 | `rei`, `rei:feature`, `rei:blocked` |
| Labels reservadas v2 | `rei:epic`, `rei:backlog` (no se crean en v1) |

El marcador es invisible y la label es solo una pista humana: la búsqueda y la
idempotencia se hacen **siempre por marcador** (R55).

---

# 2. Archivos afectados

## Nuevos

| Archivo | Propósito |
|---------|-----------|
| `internal/github/github.go` | Parseo de remotos, detección de repo/base, `Runner`, `Enable`, `Disable`, `Status`. |
| `internal/github/github_test.go` | Tests con `Runner` falso (sin red ni `gh` real). |
| `.rei/docs/harness/integrations/github.md` | Playbook completo de la integración. |

## Modificados

| Archivo | Cambio |
|---------|--------|
| `internal/config/config.go` | Modelo `Integrations`/`GitHub`, helpers de valores efectivos y persistencia (`SetGitHub`, `SetGitHubEnabled`). |
| `internal/config/config_test.go` | Presencia/ausencia, preservación de claves, idempotencia, JSON inválido. |
| `internal/cli/cli.go` | Dispatch de `github` y `cmdGitHub`. |
| `internal/cli/help.go` | Entrada `github` en la tabla de comandos. |
| `internal/cli/help_test.go` | Ayuda y uso incorrecto de `github`. |
| `internal/cli/cli_test.go` | Casos de uso incorrecto de `cmdGitHub` (código `2`). |
| `internal/check/check.go` | Añadir `.rei/docs/harness/integrations/github.md` a `RequiredFiles`. |
| `internal/check/check_test.go` | Ajustar si comprueba la lista (usa `RequiredFiles`, no la duplica). |
| `embed_test.go` | Verificar que el nuevo playbook está embebido. |
| `.rei/docs/harness/README.md` | Fila del nuevo playbook en el índice. |
| `.rei/agents/leader.md` | Sección de integración GitHub con la puerta, comentarios, PR/merge, adopción, errores y recordatorios. |
| `.rei/agents/initializer.md` | Pregunta de configuración junto al visor. |
| `.opencode/agents/*.md` | Regenerados por `rei init opencode`. |
| `.claude/agents/*.md` | Regenerados por `rei init claude`. |
| `.rei/install-manifest.json` | Regenerado por los comandos anteriores. |

No se modifica el comportamiento de `check.Run`, `doctor`, `view`, las
plantillas ni el workflow de estados. `.rei/specs/**` y `.rei/progress/**` no se
tocan salvo el propio Work Item.

---

# 3. Configuración

## 3.1 Modelo

```go
// GitHub es la configuración de la integración con GitHub Issues.
type GitHub struct {
    Enabled       bool   `json:"enabled"`
    Repo          string `json:"repo,omitempty"`
    Base          string `json:"base,omitempty"`
    MergeStrategy string `json:"mergeStrategy,omitempty"`
}

// Integrations agrupa integraciones externas opcionales.
type Integrations struct {
    GitHub *GitHub `json:"github,omitempty"`
}

type Config struct {
    Checks       []Check      `json:"checks"`
    Viewer       Viewer       `json:"viewer"`
    Integrations Integrations `json:"integrations"`
}
```

- `GitHub == nil` ⇒ clave ausente ⇒ integración desactivada (R4).
- Valores efectivos:

```go
func (g *GitHub) EffectiveBase() string          // "main" si Base == ""
func (g *GitHub) EffectiveMergeStrategy() string // "squash" si MergeStrategy == ""
```

`config.Load` no cambia: un `integrations` ausente deja el puntero en `nil` sin
error.

## 3.2 Persistencia

Se añade a `internal/config`:

```go
// SetGitHub escribe integrations.github preservando el resto del archivo.
func SetGitHub(path string, gh GitHub) error

// SetGitHubEnabled cambia solo enabled; devuelve ErrGitHubNotConfigured si la
// clave no existe.
func SetGitHubEnabled(path string, enabled bool) error
```

Implementación:

- Leer el archivo como `map[string]any` (o crear uno vacío si no existe).
- Si el JSON es inválido, devolver error **sin escribir** (R12).
- Conservar `checks`, `viewer`, otras integraciones y claves desconocidas.
- Escribir `integrations.github` con las cuatro claves (`enabled`, `repo`,
  `base`, `mergeStrategy`) y serializar indentado con `SetEscapeHTML(false)`.
- Escritura atómica con el patrón ya usado por `SetAutoOpenOnReady`
  (`tmp` + `rename`); si falla, el original queda intacto (R13).

`SetGitHubEnabled` lee primero la configuración tipada; si `GitHub == nil`
devuelve `ErrGitHubNotConfigured` y no crea la clave. `disable` lo interpreta
como «ya estaba desactivada» (R10).

Ejemplo resultante:

```json
{
  "checks": [],
  "viewer": { "autoOpenOnReady": true },
  "integrations": {
    "github": {
      "enabled": true,
      "repo": "k1wi777/my-harness-SDD",
      "base": "main",
      "mergeStrategy": "squash"
    }
  }
}
```

---

# 4. Detección de repositorio y rama base

## 4.1 `internal/github`

```go
type Runner interface {
    LookPath(file string) (string, error)
    Run(dir, name string, args ...string) (stdout, stderr string, err error)
}

type ExecRunner struct{} // implementación con os/exec

type Remote struct {
    Host  string
    Owner string
    Name  string
}

func ParseRemote(raw string) (Remote, error)
func DetectOrigin(dir string, r Runner) (Remote, error)
func DetectBase(dir string, r Runner) (base string, fallback bool)
```

`Runner` permite tests deterministas sin `git`, `gh` ni red. `ExecRunner` usa
`exec.LookPath` y `exec.Command` con `Dir` y captura separada de stdout/stderr.

## 4.2 Parseo de remotos (R6)

Formatos aceptados:

- SCP-like: `git@github.com:owner/name.git`
- URL: `https://github.com/owner/name.git`, `http://github.com/owner/name`,
  `ssh://git@github.com/owner/name.git`

Algoritmo:

1. Si contiene `://`, parsear con `net/url`; host = `u.Hostname()`, ruta =
   `u.Path`.
2. Si no, aplicar `^[A-Za-z0-9._-]+@([^:]+):(.+)$`; host = grupo 1, ruta =
   grupo 2.
3. Limpiar la ruta: quitar `/` inicial y final y el sufijo `.git`.
4. Exigir exactamente dos segmentos `owner/name`; ambos deben casar con
   `^[A-Za-z0-9._-]+$` y no ser `.` ni `..`.
5. Exigir `strings.EqualFold(host, "github.com")`. Cualquier otro host se
   rechaza con causa «v1 solo soporta github.com».

Los errores incluyen el remoto original **sin credenciales** (si hubiera
`user:pass@`, se descarta la userinfo).

## 4.3 Detección

- `DetectOrigin`: ejecuta `git remote get-url origin` en `p.Root`. Si falla,
  devuelve un error con causa y siguiente paso (revisar `git remote -v`).
- `DetectBase`: ejecuta
  `git symbolic-ref --quiet --short refs/remotes/origin/HEAD`; si devuelve
  `origin/<rama>`, usa `<rama>`. Si falla o está vacío, devuelve `main` con
  `fallback = true` para que `enable` lo informe (R8).

No se consulta `gh repo view` ni la API: `enable`/`status` deben funcionar sin
`gh` y sin red.

---

# 5. CLI `rei github`

## 5.1 Dispatch

`internal/cli/cli.go`:

```go
case "github":
    return cmdGitHub(rest)
```

`cmdGitHub` valida los argumentos **antes** de resolver el proyecto (para que un
uso incorrecto sea `2`):

- `len(args) != 1` o subcomando desconocido ⇒ uso por `stderr` y `2`.
- `enable`, `disable`, `status` ⇒ `project()` y delegación.
- Error de proyecto ⇒ mensaje de `projectMessage` y `1`.

```go
func cmdGitHub(args []string) int {
    const usage = "uso: rei github <enable|disable|status>"
    if len(args) != 1 {
        fmt.Fprintln(os.Stderr, usage)
        return 2
    }
    switch args[0] {
    case "enable", "disable", "status":
    default:
        fmt.Fprintln(os.Stderr, usage)
        return 2
    }
    p, code := project()
    if p == nil {
        return code
    }
    switch args[0] {
    case "enable":
        return github.Enable(p, github.ExecRunner{}, os.Stdout)
    case "disable":
        return github.Disable(p, os.Stdout)
    default:
        return github.Status(p, github.ExecRunner{}, os.Stdout)
    }
}
```

## 5.2 `enable` (R5–R9, R13)

Flujo:

1. `config.Load(.rei/config.json)`. Si falla ⇒ `[FAIL]` con causa y siguiente
   paso, `1`, sin escribir.
2. `DetectOrigin(p.Root, runner)`. Si falla ⇒ `[FAIL]` con causa y siguiente
   paso, `1`, sin escribir.
3. Calcular valores:
   - `repo` = detectado (`owner/name`).
   - `base` = valor previo no vacío; si no, `DetectBase` (con aviso si es
     `main` por fallback).
   - `mergeStrategy` = valor previo no vacío; si no, `squash`.
4. `SetGitHub(path, gh)`; si falla ⇒ `[FAIL]`, `1`.
5. Sondear `gh`:
   - `LookPath("gh")` falla ⇒ `[WARN] gh no está instalado...`.
   - `Run("gh", "auth", "status")` falla ⇒ `[WARN] gh no está autenticado...`.
   - Los avisos no cambian el código de salida (R9).
6. Imprimir resumen (repo, base, mergeStrategy) y siguiente paso; `0`.

Salida orientativa:

```text
== rei github ==
[OK]    Integración GitHub activada.
        repo: k1wi777/my-harness-SDD
        base: main
        mergeStrategy: squash
[WARN]  `gh` no está autenticado. Siguiente paso: `gh auth login`.
[OK]    Configuración escrita en .rei/config.json.
```

## 5.3 `disable` (R10)

1. `config.Load`; si falla ⇒ `[FAIL]`, `1`.
2. Si `GitHub == nil` o `Enabled == false` ⇒ informar «ya estaba desactivada»,
   `0`, sin escribir.
3. `SetGitHubEnabled(path, false)`; conserva repo/base/mergeStrategy y el resto
   de claves. Si falla ⇒ `[FAIL]`, `1`.
4. Informar de que los enlaces locales se conservan; `0`.

## 5.4 `status` (R11)

Solo lectura. Informa:

- Estado: `no configurada` (`GitHub == nil`), `desactivada` (`enabled: false`) o
  `activa`.
- Valores efectivos de `repo`, `base`, `mergeStrategy`.
- Remoto detectado y comparación con `repo` (`coincide`, `no coincide`, `no
  detectado`).
- `gh`: disponible/no instalado; auth: autenticado/sin autenticar/no
  comprobable.
- Siguiente paso recomendado según el estado.

Código `0` en todos los estados legibles, incluida «no configurada» o
«desactivada» (R11). Solo `1` si el proyecto o el JSON no son legibles (R12) o
`2` por uso.

## 5.5 Ayuda

Nueva entrada en `internal/cli/help.go`:

- `usage: "github <enable|disable|status>"`.
- `short`: «Configura la integración opcional con GitHub Issues».
- `detail`: qué hace cada subcomando, que `enable` detecta el repo desde
  `origin`, que preserva `checks`/`viewer`, que los avisos de `gh` no bloquean,
  códigos `0/1/2` y referencia a `.rei/docs/harness/integrations/github.md`.

---

# 6. Playbook del agente

`.rei/docs/harness/integrations/github.md` es la referencia operativa. Contiene
las secciones y comandos siguientes.

## 6.1 Configuración y diagnóstico

- `rei github enable` / `disable` / `status`.
- Lectura de `integrations.github`.
- Regla: si la clave falta o `enabled` es `false`, el Leader **no menciona
  GitHub** (R20).

## 6.2 Puerta única (R21–R26)

Disparador: planificación en `ready`, el usuario aprueba y la integración está
activa. Antes de `rei status set <id> in_progress`, el Leader:

1. Lee `.rei/config.json`.
2. Comprueba si ya hay `links.json`; si existe, no pregunta.
3. Comprueba en `current.md` si la decisión ya se registró; si existe, no
   vuelve a preguntar.
4. Formula **una sola** pregunta:
   «¿Creo la issue y la rama `feat/<n>-<slug>`?» (con `<n>` aún desconocido; se
   explica que será el número de la issue).

### Camino «sí»

Orden exacto:

```bash
# 1. ¿Existe ya? (idempotencia por marcador)
gh issue list --repo <repo> --state all --limit 10 \
  --search 'in:body "rei:work-item:<id>"' \
  --json number,url,state

# 2. Labels idempotentes
gh label create rei --color 1D76DB --description "Gestionado por REI" --force
gh label create 'rei:feature' --color 0E8A16 \
  --description "Work Item de tipo feature" --force

# 3. Crear la issue (solo si no existe)
gh issue create --repo <repo> \
  --title "<título del Work Item>" \
  --body-file /tmp/rei-issue-<id>.md \
  --label rei --label rei:feature

# 4. Verificar número/URL y estado
gh issue view <n> --repo <repo> --json number,url,state,body
```

5. Escribir `.rei/specs/<id>/links.json` **antes** de crear la rama, para no
   perder el vínculo si la rama falla.
6. Crear la rama:

```bash
git rev-parse --abbrev-ref HEAD     # debe ser <base>
git switch -c feat/<n>-<slug>
```

Si la rama actual no es `<base>`, el Leader se detiene y pregunta; no crea la
rama desde un punto arbitrario. Si la rama ya existe, no la fuerza (R25).
7. Reportar issue, rama y `links.json`.

`links.json`:

```json
{
  "provider": "github",
  "repo": "k1wi777/my-harness-SDD",
  "issue": 123,
  "url": "https://github.com/k1wi777/my-harness-SDD/issues/123",
  "epic": null
}
```

### Plantilla del cuerpo de la issue (R23)

```md
<!-- rei:work-item:<id> -->

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

El marcador va en la **primera línea**. El título es `meta.json.title`.

### Camino «no»

- No se crea issue ni rama.
- Se trabaja sobre la rama base, como hoy.
- Se registra en `current.md` (Bitácora): «GitHub: issue/rama declinadas; se
  trabaja en `<base>`».
- No se vuelve a preguntar (R24).

## 6.3 Comentarios (R29–R33)

Disparadores: el Leader acaba de recibir al Implementer y al Reviewer.

Búsqueda idempotente del comentario (REST, para obtener el `id` numérico):

```bash
gh api repos/<repo>/issues/<n>/comments --paginate \
  --jq '.[] | select(.body | contains("<!-- rei:comment:impl:<id> -->")) | .id'
```

- Si devuelve un id: actualizar en sitio con
  `gh api --method PATCH repos/<repo>/issues/comments/<comment-id> -F body=@/tmp/rei-comment-impl-<id>.md`.
- Si no: crear con `gh issue comment <n> --repo <repo> --body-file <archivo>`.
- Verificar releyendo el comentario o comprobando que el marcador aparece una
  sola vez.

Plantilla de implementación:

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

Plantilla de revisión:

```md
<!-- rei:comment:review:<id> -->

## Revisión: <done|changes_requested|blocked>

- Work Item: `<id>`
- Verificaciones: ...
- Observaciones: ...
- Detalle: `.rei/progress/work-items/<id>/review.md`
```

Reglas:

- Nunca comentarios de cambio de estado (R32).
- La label `rei:blocked` se aplica/retira con `gh issue edit <n> --add-label` /
  `--remove-label`, sin comentario (R34); si la label no existe, se crea con
  `gh label create rei:blocked --force`.
- Si la revisión es `changes_requested` y luego se repite, se **actualiza** el
  mismo comentario de review.
- Un fallo de comentario se reporta; no cambia el estado de REI (R46).

## 6.4 PR y merge (R35–R39)

Solo tras orden explícita y tras la revisión:

```bash
# Si hay cambios del Work Item sin commitear, la orden de PR autoriza
# commitearlos (nunca `git add -A` ciego) y publicar SOLO esta rama.
git add <archivos del Work Item>
git commit -m "<título>"
git push -u origin feat/<n>-<slug>

gh pr create --repo <repo> --base <base> --head feat/<n>-<slug> \
  --title "<título>" --body-file /tmp/rei-pr-<id>.md

gh pr view <pr> --repo <repo> --json number,url,state,mergeable
```

Cuerpo del PR: resumen, cambios, verificación y `Closes #<n>`.

Merge solo con orden explícita:

```bash
gh pr merge <pr> --repo <repo> --squash --delete-branch \
  --subject "<título> (#<n>)"
```

- Nunca `--admin`, nunca `--auto` (R36).
- Si branch protection, checks o permisos lo impiden: informar, dejar el PR
  abierto y parar (R37).
- Después:

```bash
gh issue view <n> --repo <repo> --json state,url
gh issue close <n> --repo <repo> --reason completed   # solo si sigue OPEN
gh issue view <n> --repo <repo> --json state          # verificar CLOSED
```

- Si ya estaba cerrada por `Closes #N`, informarlo.
- Nunca commit/push/merge sin orden (R39).

## 6.5 Adopción (R40–R43)

Flujo:

1. Comprobar que la integración está activa; si no, no tocar GitHub y ofrecer
   `rei github enable` (R43).
2. `gh issue view <n> --repo <repo> --json number,title,body,state,labels,url`.
3. Si está cerrada, preguntar antes de continuar.
4. Buscar `<!-- rei:work-item:<id> -->` en el cuerpo:
   - Marcador de un Work Item existente ⇒ reutilizarlo.
   - Marcador de otro Work Item ⇒ detenerse y reportar conflicto (R42).
   - Sin marcador ⇒ preguntar lo que falte (objetivo, alcance, criterios, fuera
     de alcance, restricciones), crear el Work Item con `rei new`, rellenar
     `title`/`description` con el acuerdo, añadir el marcador al cuerpo sin
     destruir el contenido (`gh issue view --json body` → archivo temporal →
     `gh issue edit --body-file`) y asegurar labels (R41).
5. Escribir `links.json`.
6. Continuar con el workflow normal (`spec_author`).

## 6.6 Errores y reintentos (R44–R49)

Tabla obligatoria del playbook:

| Situación | Detección | Mensaje (causa) | Siguiente paso |
|-----------|-----------|-----------------|----------------|
| Sin `.rei` | `rei github` / Leader | «REI no está inicializado aquí» | `rei init` |
| Integración no configurada | `GitHub == nil` | «GitHub no está configurado» | `rei github enable` |
| Integración desactivada | `enabled: false` | «GitHub está desactivado» | `rei github enable` si se desea |
| Config incompleta | `repo` vacío o inválido | «Falta `integrations.github.repo`» | `rei github enable` o editar la config |
| `gh` ausente | `LookPath` falla | «GitHub CLI no está instalado» | Instalar `gh` |
| `gh` sin auth | `gh auth status` falla | «GitHub CLI no está autenticado» | `gh auth login` |
| Repo no detectado / ambiguo | `git remote get-url origin` falla o host no válido | «No se pudo detectar el repositorio» | Revisar `git remote -v` / configurar `repo` |
| Issue inexistente | `gh issue view` 404 | «La issue #N no existe en `<repo>`» | Verificar número/repo |
| Rama existente | `git switch -c` falla | «La rama ya existe» | Reutilizarla o elegir otra; nunca forzar |
| PR inexistente | `gh pr view` 404 | «No hay PR para esa rama» | Crear el PR solo si hay orden |
| Permisos insuficientes | 403/404 de API | «Sin permisos sobre `<repo>`» | Revisar acceso y `gh auth status` |
| Branch protection | merge rechazado | «El merge está bloqueado por protección de rama/checks» | Completar checks/reviews; nunca `--admin` |
| Red | error de conexión | «No se pudo contactar con GitHub» | Reintentar cuando haya red |
| Rate limit | mensaje de límite | «Límite de GitHub alcanzado» | Esperar al reset indicado; no reintentar en bucle |

Reglas transversales:

- Todo mensaje lleva **causa** y **siguiente paso** (R44).
- Verificar antes de afirmar éxito (R47).
- Un fallo de GitHub no cambia `meta.json`, el estado de `current.md` ni la
  etapa; se informa de lo hecho y lo pendiente y se reintenta sin duplicar
  (R46).
- Ante rate limit, no reintentar en bucle; mostrar el reset si `gh` lo da (R48).

## 6.7 Recordatorios en `current.md` (R54)

Mientras haya sesión activa, el Leader actualiza «Próximo paso» con líneas como:

- `GitHub: pendiente crear rama feat/<n>-<slug> (issue #<n> ya creada).`
- `GitHub: pendiente comentario de implementación (reintentar).`
- `GitHub: pendiente comentario de revisión (reintentar).`
- `GitHub: pendiente PR de la rama feat/<n>-<slug>.`
- `GitHub: pendiente merge del PR #<pr>.`

Al completarse, se retira la línea. Con la integración desactivada no se escribe
ninguna línea. Tras `done`, el Reviewer archiva la sesión; las acciones
pendientes (comentario de review, PR o merge) se reportan en la respuesta del
Leader, sin recrear `current.md` (R54).

---

# 7. Contratos de los roles

## 7.1 Leader (`.rei/agents/leader.md`)

Insertar una sección «Integración con GitHub (opcional)» después del bloque de
Casos y antes de «Delegación», y un paso en Caso C:

- **Caso C:** entre la aprobación y `rei status set in_progress`, aplicar la
  puerta única si `integrations.github.enabled` es `true`.
- **Caso C (tras Implementer):** publicar el comentario de implementación y
  actualizar «Próximo paso».
- **Caso E (tras Reviewer):** publicar el comentario de revisión.
- **Adopción:** al arrancar, si el usuario pide «trabajemos en la #N», seguir el
  flujo de adopción antes de `rei new`.
- **PR/merge:** solo bajo orden explícita; nunca `--admin`; verificar la issue
  después del merge.
- **Errores:** causa + siguiente paso; no bloquean REI; reintento idempotente.
- **Recordatorios:** `current.md` mientras haya sesión.
- **Referencia:** remitir a `.rei/docs/harness/integrations/github.md` para los
  comandos exactos; el Contrato conserva las reglas no negociables y no copia
  toda la tabla de errores.

No se añaden herramientas nuevas: el Leader ya dispone de `shell`, `write` y
`edit`.

## 7.2 Initializer (`.rei/agents/initializer.md`)

Añadir, tras «Configuración del visor», una subsección «Configuración de GitHub
Issues»:

- Preguntar en el mismo paso de configuración si se activa la integración.
- Si sí: ejecutar `rei github enable`, mostrar repo/base y trasladar avisos.
- Si no: no escribir `integrations.github` (ausencia = desactivado).
- Si ya existe: mostrar la configuración y pedir confirmación antes de cambiarla.
- Nunca editar el JSON a mano; usar el CLI.
- Un fallo de `gh` se informa y no bloquea la personalización.

El alcance de escritura del `initializer` no cambia: el CLI es quien escribe
`.rei/config.json`.

---

# 8. Tests

## 8.1 `internal/config`

- `Load` sin `integrations` ⇒ `GitHub == nil`; con `enabled:false` ⇒ no nil y
  desactivada.
- `SetGitHub` crea la clave y preserva `checks`, `viewer`, claves desconocidas y
  otras integraciones.
- `SetGitHubEnabled(false)` conserva `repo`, `base` y `mergeStrategy`.
- JSON inválido ⇒ error y archivo intacto.
- Idempotencia: aplicar dos veces el mismo valor no cambia el resto.

## 8.2 `internal/github`

- `ParseRemote` table-driven: SSH, HTTPS con/sin `.git`, barra final, host
  distinto, ruta de tres segmentos, cadena vacía, URL con credenciales.
- `DetectOrigin` con `Runner` falso: éxito, sin `origin`, salida con espacios.
- `DetectBase`: `origin/main` → `main`; error → `main` + fallback.
- `Enable`: escritura correcta; repo/base conservados; aviso si `gh` falta; aviso
  si auth falla y salida `0`; fallo de detección ⇒ `1` sin escribir; JSON
  inválido ⇒ `1`.
- `Disable`: conserva enlaces; ya desactivada ⇒ `0`; config inválida ⇒ `1`.
- `Status`: no configurada/desactivada/activa ⇒ `0`; remoto coincidente y no
  coincidente; config inválida ⇒ `1`; no escribe archivos.

`Runner` falso:

```go
type fakeRunner struct {
    paths map[string]string
    runs  map[string]runResult // clave: "git remote get-url origin", etc.
}
```

## 8.3 `internal/cli`

- `cmdGitHub` sin subcomando, con subcomando desconocido o argumentos extra ⇒
  `2` (sin resolver proyecto).
- La ayuda general incluye `github enable|disable|status`; la ayuda del comando
  menciona los tres subcomandos, los códigos `0/1/2` y el playbook.
- Tests de dispatch que no ejecuten red: los casos válidos se cubren en el
  paquete `github` con `Runner` falso.

## 8.4 Embed y harness

- `embed_test.go` comprueba que existe
  `.rei/docs/harness/integrations/github.md`.
- `rei check --quiet` pasa con el nuevo archivo en `RequiredFiles`.
- `rei init opencode --check` y `rei init claude --check` pasan tras regenerar.

---

# 9. Alternativas descartadas

| Alternativa | Motivo del descarte |
|-------------|---------------------|
| Crear issues/PR desde Go (SDK o `gh` embebido) | Añade dependencia/acoplamiento, impide el control humano y contradice «solo `rei github` es Go». |
| Usar `links.json` como fuente de verdad | Se pierde al copiar el repo o al no versionar `.rei`; el marcador es recuperable. |
| Preguntar por GitHub en cada sesión | La puerta es única por Work Item; preguntar de nuevo es ruido. |
| Publicar comentarios de cambio de estado | El histórico de la issue se llenaría de ruido; solo importan implementación y revisión. |
| Crear PR/merge automáticamente | El usuario debe ordenarlo explícitamente; nunca `--admin`. |
| Flag `--repo` en `enable` | No está en el alcance v1; si la detección falla, se edita `integrations.github.repo` y se reintenta. |
| Crear rama desde el HEAD actual sin comprobar la base | Puede partir de una rama equivocada; se exige estar en `<base>`. |
| Actualizar el cuerpo de la issue en cada cambio de spec | Pisaría ediciones humanas; se conserva la issue y se actualiza `links.json`. |
| Implementar épicas/backlog/asignación ahora | Fuera de alcance v1; el vocabulario queda reservado. |
| Reabrir la sesión después de `done` para PR/merge | Rompe el modelo de `current.md`; se reporta en la respuesta del Leader. |

---

# 10. Verificación

- `gofmt -l internal` sin salida.
- `make vet`.
- `make test`.
- `make build`.
- `rei check --quiet` ⇒ `0` (incluye el nuevo archivo requerido).
- `./bin/rei init opencode --check` y `./bin/rei init claude --check` ⇒ `0`.
- Prueba manual en un proyecto temporal:
  - sin `.rei` ⇒ `rei github status` sale `1` con causa y siguiente paso;
  - con `.rei` y sin config ⇒ `status` sale `0` e informa «no configurada»;
  - con `origin` SSH/HTTPS ⇒ `enable` escribe repo/base/mergeStrategy,
    preserva `checks`/`viewer` y avisa si `gh` falta;
  - `disable` conserva los valores; `status` lo refleja;
  - JSON inválido ⇒ `1` sin sobrescribir;
  - argumentos inválidos ⇒ `2`.
- Revisión del playbook contra la tabla de errores y las plantillas de issue,
  comentarios y PR.
