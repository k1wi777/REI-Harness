# Plan — GitHub: idempotencia por API y aviso de remoto

## Objetivo

Eliminar la causa raíz del duplicado de issues y del diagnóstico engañoso de
`rei github status` cuando el remoto de GitHub no es el nombre canónico del
repositorio (repo renombrado, clon con remoto viejo o fork).

1. **Idempotencia autoritativa por API.** El playbook deja de buscar el marcador
   con `gh issue list --search` (índice de search, no sigue renames, devuelve
   `[]` con exit 0 y provoca duplicados) y pasa a listar las issues del
   repositorio vía REST (`gh api .../issues?state=all&per_page=100 --paginate`)
   filtrando por el marcador. El listado REST **sí** sigue el redirect del
   rename.
2. **Nombre canónico explícito.** Al inicio de cada acción GitHub se resuelve
   `nameWithOwner` con `gh repo view <repo> --json nameWithOwner --jq
   .nameWithOwner` y ese valor se usa en todos los comandos (labels, issue,
   comentarios, PR). Si la resolución falla, se informa de la causa y del
   siguiente paso; nunca se cae en silencio al nombre viejo.
3. **Aviso en `rei github status`.** Si el remoto detectado no coincide con
   `integrations.github.repo`, `status` lo advierte y no concluye «Sin acciones
   pendientes». Sin red en el CLI: la resolución canónica es responsabilidad del
   agente.

Evidencia que motiva el cambio: con remoto `k1wi777/my-harness-SDD` (renombrado
a `REI-Harness`), `gh issue list --search 'in:body "rei:work-item:..."'`
devolvía `[]` con exit 0, mientras `gh api
"repos/<repo>/issues?state=all&per_page=100" --paginate` encontraba la issue #1
con ambos nombres.

## Decisiones

- **La fuente de verdad de la idempotencia es el listado REST filtrado por
  marcador**, no el search API. Motivos a documentar en el playbook: el índice
  de búsqueda se actualiza con retraso, no resuelve renames y falla en silencio
  (exit 0 + lista vacía) → camino directo al duplicado. El listado REST sigue el
  redirect 301 del rename y `--paginate` evita perder issues antiguas.
- **`<repo>` pasa a significar «nombre canónico vigente»** en todo el playbook.
  Se define una sola vez (nueva subsección) y se recuerda en cada sección, en
  lugar de reescribir a mano cada comando: los comandos ya usan el placeholder
  `<repo>` y así el cambio es mínimo y no introduce erratas en rutas REST.
- **La resolución canónica ocurre al inicio de cada acción GitHub, no una vez por
  sesión.** Cuesta una llamada barata y evita reutilizar un nombre cacheado tras
  un rename a mitad de sesión.
- **Fork vs. rename:** `gh repo view` devuelve el repositorio al que apunta el
  nombre consultado. Para un rename devuelve el nombre nuevo del mismo
  repositorio; para un fork devuelve el fork (no el upstream). Por tanto la
  integración opera siempre sobre el repositorio configurado/enlazado; si se
  quiere operar sobre el upstream hay que ponerlo explícitamente en
  `integrations.github.repo`. Esto se documenta como comportamiento esperado, no
  como error.
- **`status` sigue siendo offline y de solo lectura.** El aviso compara el
  remoto ya detectado con `git remote get-url origin` y el valor de la config;
  no se añade ninguna llamada a `gh api` ni `gh repo view`. El aviso reutiliza la
  línea `remoto origin: ... (coincide|no coincide ...)` ya existente y añade una
  advertencia + siguiente paso.
- **Prioridad del «Siguiente paso»:** se conserva el orden actual (`no
  configurada` → `desactivada` → `gh` no instalado → `gh` sin auth → remoto no
  detectado) y el caso nuevo de discrepancia se inserta **antes** del `default`.
  Así nunca se imprime «Sin acciones pendientes» con un remoto no canónico,
  pero un problema bloqueante de `gh` sigue teniendo prioridad.
- **Sin cambios de contrato de rol.** `.rei/agents/leader.md` delega los comandos
  exactos en el playbook y solo exige «busca el marcador antes de crear», sin
  nombrar el search API. No se toca ningún `## Contrato`, así que **no** se
  regeneran agentes nativos; solo se verifica que siguen sincronizados.

## Archivos

| Archivo | Acción |
|---------|--------|
| `.rei/docs/harness/integrations/github.md` | Editar: nombre canónico, idempotencia por API, fork/rename, fila(s) de la tabla de errores y matices de `status`. |
| `internal/github/github.go` | Editar: aviso y siguiente paso por remoto no canónico en `Status`. |
| `internal/github/github_test.go` | Editar: tests del aviso de `status`. |
| `.rei/specs/2026-10-25_19-01__github-repo-canonico-idempotencia/plan.md` | Editar: marcar pasos al completarlos. |

No se modifican `.rei/agents/*.md`, `.opencode/agents/*`, `.claude/agents/*` ni
`.rei/specs/2026-10-25_18-00__github-issues-v1/design.md` (documento histórico).

## Cambios

### 1. Playbook — nombre canónico (`.rei/docs/harness/integrations/github.md`)

En la sección **3. Vocabulario fijo**, añadir una subsección
**«Nombre canónico del repositorio»** con este contenido efectivo:

- Regla: en este documento, `<repo>` significa siempre el **nombre canónico
  vigente** `owner/name`, resuelto al inicio de cada acción GitHub:

  ```bash
  canonical=$(gh repo view <repo> --json nameWithOwner --jq .nameWithOwner)
  ```

- Ese valor se usa en **todos** los comandos: `--repo <canonical>` en
  `gh label create`, `gh issue view|create|edit|close|comment` y
  `gh pr create|view|merge`, y en las rutas REST
  `repos/<canonical>/issues/...`.
- Si la resolución falla (sin red, sin auth, sin permisos, repo inexistente), se
  informa de **causa y siguiente paso** y se detiene la acción. **Nunca** se
  continúa con el nombre viejo ni con el valor de la config sin resolver.
- `<repo>` de partida es `integrations.github.repo`; si está vacío, el `owner/name`
  detectado en `origin`.

Y un párrafo **«Rename y fork»** que explique: GitHub redirige el nombre viejo al
nuevo (por eso el listado REST funciona con ambos), pero el search API no lo
hace; en un fork `gh repo view` resuelve al fork, de modo que la integración
opera sobre el repositorio configurado y no sobre el upstream salvo que se
configure explícitamente.

### 2. Playbook — idempotencia por API (sección 4.1)

Sustituir el bloque actual del paso 1 (`gh issue list --repo <repo> --state all
--limit 10 --search ...`) por:

```bash
# 0. Nombre canónico (obligatorio al inicio de la acción)
canonical=$(gh repo view <repo> --json nameWithOwner --jq .nameWithOwner)

# 1. ¿Existe ya la issue? (fuente de verdad: el marcador, vía REST)
gh api "repos/$canonical/issues?state=all&per_page=100" --paginate \
  --jq '.[] | select(.pull_request == null)
        | select((.body // "") | contains("<!-- rei:work-item:<id> -->"))
        | [.number, .state, .html_url] | @tsv'
```

- Añadir una nota breve: el endpoint de issues incluye PRs, por eso se descartan
  los que traen `pull_request`; el marcador se busca en `.body` y el resultado es
  autoritativo aunque el remoto sea un nombre viejo.
- Mantener el resto del flujo (reutilizar si hay resultado; crear solo si no lo
  hay; verificar antes de afirmar éxito), pero con `<canonical>` en todos los
  comandos de `--repo`.
- En las secciones **5 (adopción)**, **6 (comentarios)** y **7 (PR/merge)**,
  añadir un recordatorio de una línea al inicio de cada una: «`<repo>` es el
  nombre canónico resuelto con `gh repo view`, no el remoto literal». No es
  necesario reescribir cada comando: el placeholder ya está en todos.

### 3. Playbook — `rei github status` y tabla de errores

- En **2. Configuración y diagnóstico**, precisar la descripción de
  `rei github status`: además de estado/valores/remoto/`gh`, **avisa si el remoto
  no coincide con `integrations.github.repo`** y propone siguiente paso; sigue
  sin consultar la red.
- En **8. Errores y reintentos**, añadir a la tabla:

  | Situación | Detección | Mensaje (causa) | Siguiente paso |
  |---|---|---|---|
  | Remoto ≠ nombre canónico | `git remote get-url origin` frente a `integrations.github.repo` | «El remoto no coincide con `integrations.github.repo`» | Resolver el canónico con `gh repo view <repo> --json nameWithOwner --jq .nameWithOwner` y ponerlo en `integrations.github.repo`, o actualizar el remoto a ese repositorio |
  | Fallo al resolver el canónico | `gh repo view` falla (red/auth/permisos/404) | «No se pudo resolver el nombre canónico de `<repo>`» | Revisar `gh auth status` y la red; no continuar con el nombre viejo |

### 4. `rei github status` (`internal/github/github.go`)

En `Status`, dentro del bloque del remoto y tras la línea
`remoto origin: ... (coincide|no coincide con integrations.github.repo)`:

- Calcular una vez un booleano `mismatch` =
  `gh != nil && gh.Repo != "" && remoteErr == nil && !strings.EqualFold(gh.Repo, remote.String())`.
- Si `mismatch`, imprimir una advertencia con la misma estética que el resto
  (`[WARN]  `), p. ej.:

  ```text
  [WARN]  El remoto no es el nombre canónico configurado; la idempotencia debe
          resolverse con `gh repo view <repo> --json nameWithOwner --jq .nameWithOwner`.
  ```

- En el `switch` final de «Siguiente paso», insertar `case mismatch:` **antes**
  del `default`, con el texto:

  ```text
  Siguiente paso: pon el nombre canónico (`gh repo view <repo> --json nameWithOwner --jq .nameWithOwner`)
  en integrations.github.repo, o actualiza `origin` a ese repositorio.
  ```

- Sin llamadas nuevas al `Runner` (nada de `gh repo view`/`gh api`), sin escribir
  archivos y sin cambiar el código de salida (sigue siendo 0 en estados legibles).

### 5. Tests (`internal/github/github_test.go`)

Ampliar `TestStatusRemoto` (mantener el estilo `fakeRunner` + `gitOK()` +
`ghOK()`):

- **coincide:** la salida contiene `coincide` y `Sin acciones pendientes.`, y
  **no** contiene el aviso nuevo.
- **no coincide:** la salida contiene el aviso nuevo (una subcadena estable, p.
  ej. `nombre canónico`), contiene el siguiente paso con `gh repo view`, y **no**
  contiene `Sin acciones pendientes.`.
- **no coincide con integración desactivada** (`enabled:false` + `repo`
  distinto): sigue apareciendo el aviso y el siguiente paso es
  `rei github enable` (documenta la prioridad del `switch`).
- **`repo` vacío:** no se compara ni se avisa (`gh.Repo == ""`), se mantiene el
  comportamiento actual.
- En los casos nuevos: código de salida 0, `config.json` intacto y **ningún
  comando de red simulado** (si la implementación invocase `gh api`/`gh repo
  view`, el `fakeRunner` devolvería «comando no simulado» y la prueba debe seguir
  pasando porque no se usa).

## Restricciones

- **Fuera de alcance:** flags `--repo`, `--base`, `--strategy`; cliente HTTP de
  GitHub en Go; auto-canonicalización silenciosa de `.rei/config.json`; borrar o
  tocar la issue #1; otros proveedores; cambios en `gh issue list --search` fuera
  del playbook.
- No introducir dependencias nuevas ni llamadas de red en `rei github status`.
- No cambiar la firma de `Status` ni el código de salida; no tocar `Enable`,
  `Disable`, `DetectOrigin`, `ParseRemote` ni la config.
- No editar a mano agentes nativos; si (contra lo previsto) hubiera que cambiar
  un `## Contrato`, regenerar con `rei init opencode` / `rei init claude`, nunca
  a mano.
- No reescribir documentos históricos de Work Items anteriores y no tocar la
  issue #1.
- Mantener el estilo del proyecto: mensajes con **causa** y **siguiente paso**,
  sin afirmar éxito sin verificar.

## Pasos

- [x] 1. Reescribir en `.rei/docs/harness/integrations/github.md` el paso de
  idempotencia de la sección 4.1 con la resolución canónica
  (`gh repo view ... --json nameWithOwner --jq .nameWithOwner`) y el listado REST
  paginado filtrado por marcador; eliminar el `gh issue list --search`.
- [x] 2. Añadir la subsección «Nombre canónico del repositorio» y el párrafo
  «Rename y fork» (incluido el motivo por el que no se usa el search API) en la
  sección 3, y los recordatorios de `<repo>` = canónico en las secciones 5, 6 y 7.
- [x] 3. Actualizar la descripción de `rei github status` (sección 2) y añadir
  las dos filas nuevas a la tabla de errores (sección 8).
- [x] 4. Implementar en `Status` (`internal/github/github.go`) el cálculo de
  `mismatch`, la advertencia y el `case` de siguiente paso antes del `default`;
  sin llamadas al `Runner` adicionales.
- [x] 5. Ampliar `TestStatusRemoto` en `internal/github/github_test.go` con los
  casos descritos en el apartado 5 de este plan.
- [x] 6. Verificar: `gofmt -l .` (sin salida), `go vet ./...`,
  `go test ./internal/github/`, `go test ./...` y `rei check --quiet`.
- [x] 7. Verificar que no hay contratos de rol afectados
  (`grep -rn "issue list\|--search" .rei/agents/` sin resultados) y que los
  agentes nativos siguen sincronizados: `rei init opencode --check` y
  `rei init claude --check`. Regenerar **solo** si alguno falla por un contrato
  modificado.
- [x] 8. Documentar la implementación en
  `.rei/progress/work-items/2026-10-25_19-01__github-repo-canonico-idempotencia/impl.md`
  con la evidencia de cada comando ejecutado.
