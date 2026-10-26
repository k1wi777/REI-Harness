# Design

> Feature: **Separar skeleton del proyecto vivo y verificación rápida/lenta**

## Contexto y problema

El repositorio es a la vez la plantilla de REI Harness y un proyecto vivo ya
personalizado. `embed.go` embebe los archivos vivos (`AGENTS.md`,
`.rei/config.json`, `.rei/docs/**`), por lo que `rei init` despliega en cualquier
proyecto nuevo la personalización de este repositorio: documentación de REI,
`integrations.github.repo = k1wi777/REI-Harness` y los checks `go vet` +
`go test`.

Eso rompe la instalación limpia (los checks de REI se ejecutan en un proyecto sin
su módulo) y hace fallar los tests de `initwizard`, que esperan placeholders y
configuración neutral. La causa no es el initializer: falta separar el material
desplegable del material personalizado.

## Decisión: dos raíces embebidas y mapeo en `initwizard`

Se separa el contenido en dos orígenes:

| Origen | Contenido | Destino en `rei init` |
|--------|-----------|------------------------|
| `skeleton/` | `AGENTS.md`, `.rei/config.json`, `.rei/docs/project/{README,architecture,conventions,verification}.md` neutrales | Raíz del proyecto, sin el prefijo `skeleton/` |
| Archivos vivos | `.rei/docs/harness/**`, `.rei/docs/usage.md`, `.rei/agents/**`, `.rei/adapters/**`, `.rei/templates/**` | Se mantienen en su ruta actual |

`embed.go` declara dos `embed.FS` y sigue sin contener lógica (restricción
arquitectónica del paquete raíz):

```go
//go:embed all:skeleton
var Skeleton embed.FS

//go:embed all:.rei/docs/harness all:.rei/docs/usage.md all:.rei/agents all:.rei/adapters all:.rei/templates
var Harness embed.FS
```

Notas:

- `all:` es necesario porque `.rei` empieza por punto.
- `.rei/docs/usage.md` se incluye explícitamente para no perder la guía humana
  que antes entraba con `all:.rei/docs`.
- Los docs de proyecto vivos no se embeben: `skeleton/.rei/docs/project/*` es la
  única copia desplegable.
- `skeleton/` no puede contener archivos de apoyo (por ejemplo un `README.md`
  propio) porque el mapeo los desplegaría en la raíz del proyecto. Solo contiene
  archivos desplegables.

## Contenido neutral del esqueleto

Los tres documentos de proyecto y `AGENTS.md` se restauran desde la versión
neutral de git: `git show HEAD:AGENTS.md` y
`git show HEAD:.rei/docs/project/<archivo>.md`. Esas versiones ya contienen los
marcadores `<!-- REI:PENDIENTE -->` (2 en `AGENTS.md`; 1 en cada documento de
proyecto).

`.rei/docs/project/README.md` se copia también desde `HEAD` (es guía genérica,
sin marcador).

`skeleton/.rei/config.json` del esqueleto se escribe neutral:

```json
{
  "checks": [],
  "viewer": {
    "autoOpenOnReady": true
  }
}
```

No se incluye la clave `integrations`: su ausencia significa «GitHub
desactivado» en `internal/config`. `autoOpenOnReady: true` conserva el default
aprobado en `2026-10-13_15-51__viewer-lifecycle`: el visor se abre
automáticamente en la puerta de aprobación de los proyectos nuevos.

## Mapeo en `initwizard`

Se añade un helper interno (nuevo `internal/initwizard/embedded.go`) que recorre
ambos `embed.FS` y entrega a `InstallSkeleton` y `Update` pares
`(rutaDestino, contenido)`:

```go
func walkEmbedded(fn func(dest string, data []byte) error) error
```

- Recorre `Skeleton` desde `skeleton/` y calcula
  `dest = strings.TrimPrefix(path, "skeleton/")`.
- Recorre `Harness` desde `.` y usa la ruta tal cual como `dest`.
- Solo entrega archivos (los directorios se crean con `MkdirAll` al escribir).
- No hay colisiones por construcción: el esqueleto aporta la raíz y
  `.rei/docs/project/*`; el harness aporta `.rei/docs/harness/**`,
  `.rei/docs/usage.md`, `.rei/agents/**`, `.rei/adapters/**` y
  `.rei/templates/**`.

`InstallSkeleton` (`installer.go`) sustituye su `fs.WalkDir(reiskel.Skeleton, …)`
por `walkEmbedded`, conservando:

- la política de no sobrescribir por existencia;
- el aviso `[WARN]` para `AGENTS.md` (ahora comparando `dest == "AGENTS.md"`);
- los contadores `[OK]`/`[SKIP]` y el resumen;
- el manifiesto con claves de destino.

`Update` (`update.go`) hace el mismo cambio. `isPersonalizationDoc(dest)` sigue
funcionando porque las rutas destino son las canónicas (`AGENTS.md`,
`.rei/docs/project/…`). La regla marker-aware y el baseline del manifiesto no
cambian. `Pending`, `Init` y `Status` no requieren cambios de lógica: sus rutas
canónicas ya son las de destino y el esqueleto ahora contiene los marcadores.

`internal/adapter` no cambia: reutiliza `InstallSkeleton` y recibe el mismo
mapeo.

## Compatibilidad con instalaciones existentes

El manifiesto (`.rei/install-manifest.json`) ya usa rutas destino, así que no hay
migración de formato. En el repositorio vivo, `rei init --update` verá los docs
personalizados sin marcador y los conservará; verá `.rei/config.json` distinto
del esqueleto y también lo conservará. Los archivos del harness que sigan
intactos se refrescarán como hasta ahora.

## Tests

### `embed_test.go`

- Los tests de estructura se reparten entre `Skeleton` (rutas con prefijo
  `skeleton/`) y `Harness` (rutas `.rei/...`).
- Nuevos asserts:
  - `Skeleton` contiene `skeleton/AGENTS.md`, `skeleton/.rei/config.json` y los
    cuatro documentos de `.rei/docs/project/`.
  - `Harness` contiene `.rei/docs/harness/integrations/github.md` y
    `.rei/docs/usage.md`, además de agentes, adaptadores y plantillas.
  - Ningún `embed.FS` contiene `AGENTS.md`, `.rei/config.json` ni
    `.rei/docs/project/*` sin el prefijo del esqueleto.
  - El `AGENTS.md` y los tres documentos de proyecto del esqueleto contienen
    `<!-- REI:PENDIENTE -->`.
  - El config del esqueleto, leído con `internal/config`, tiene `checks` vacío,
    `Integrations.GitHub == nil` y `Viewer.AutoOpenOnReady == true`.
  - Ningún `embed.FS` contiene `.rei/specs` ni `.rei/progress`.

### `internal/initwizard`

- `embeddedFile(t, rel)` resuelve primero `skeleton/<rel>` y, si no existe, `<rel>`
  en `Harness`. Así los tests existentes de docs de proyecto y de harness siguen
  siendo válidos.
- Se mantienen y deben pasar los tests hoy fallidos:
  `TestInitMencionaPersonalize`, `TestInstalacionLimpiaIntegra` y
  `TestUpdateConservaPersonalizacionYModificados` (sus 4 asserts).
- Nuevo test de instalación limpia neutral: `Init` en `t.TempDir()` deja los
  cuatro documentos con marcador, config sin checks ni integración y
  `Pending` con los 4 documentos.
- Se amplía el test de no sobrescritura para incluir también los documentos de
  proyecto personalizados, no solo `AGENTS.md` y `.rei/config.json`.

### `internal/adapter`

Los tests que leen roles y adaptadores del esqueleto pasan a leer de
`reiskel.Harness`.

## Documentación viva

Además de separar el material, se corrige la documentación del proyecto vivo:

- `.rei/config.json`: se elimina V2 de `checks`; se conservan `integrations` y
  `viewer` vivos.
- `.rei/docs/project/verification.md`: V2 pasa a la tabla de checkpoints lentos
  (junto a V3 y V4); el ejemplo de config queda solo con V1 y se ajusta la nota
  de «no ejecutar V3/V4» a V2/V3/V4.
- `.rei/docs/project/architecture.md`: la organización refleja `skeleton/` y los
  dos `embed.FS`; se mantiene la regla «el paquete raíz no contiene lógica».
- `AGENTS.md` (§4): se añade `skeleton/` al mapa del repositorio.
- `README.md` (raíz): se añade `skeleton/` al árbol de estructura del código
  fuente.

No se tocan los docs personalizados más allá de estas correcciones puntuales.

## Archivos afectados

| Archivo | Cambio |
|---------|--------|
| `skeleton/AGENTS.md` | Nuevo (neutral desde `HEAD`) |
| `skeleton/.rei/config.json` | Nuevo (neutral) |
| `skeleton/.rei/docs/project/{README,architecture,conventions,verification}.md` | Nuevos (neutrales desde `HEAD`) |
| `embed.go` | Dos `embed.FS`; patrones selectivos |
| `internal/initwizard/embedded.go` | Nuevo helper de mapeo |
| `internal/initwizard/installer.go` | Usa `walkEmbedded` |
| `internal/initwizard/update.go` | Usa `walkEmbedded` |
| `embed_test.go`, `internal/initwizard/*_test.go`, `internal/adapter/{adapter,claude}_test.go` | Ajustes y tests nuevos |
| `.rei/config.json` | Solo V1 |
| `.rei/docs/project/verification.md` | V2 como checkpoint lento |
| `.rei/docs/project/architecture.md`, `AGENTS.md`, `README.md` | Reflejan `skeleton/` |

## Alternativas descartadas

- **Mantener una sola raíz embebida y filtrar en tiempo de ejecución**: el
  binario seguiría conteniendo la personalización viva; no resuelve el problema
  de fondo y hace imposible verificar la neutralidad del esqueleto.
- **Generar el esqueleto desde `HEAD` en build time**: añade complejidad de build
  y no funciona al empaquetar una release desde un checkout modificado.
- **Mover el esqueleto a `.rei/skeleton/`**: `rei init` lo desplegaría dentro del
  proyecto destino como parte del arnés, contaminando la instalación.
- **Duplicar `embed.go` con patrones negativos**: `go:embed` no admite
  exclusiones; la separación por directorios es la forma soportada.

## Riesgos y mitigaciones

- **Deriva entre esqueleto y proyecto vivo**: son copias con propósitos
  distintos. Los tests de neutralidad y de actualización detectan olvidos; los
  cambios de estructura del harness deben reflejarse en ambas copias cuando
  apliquen.
- **Pérdida de `.rei/docs/usage.md`**: se incluye explícitamente en `Harness` y
  se verifica en `embed_test`.
- **Regresión en instalaciones existentes**: el manifiesto no cambia de formato y
  `Update` conserva la política marker-aware/baseline; se cubre con los tests de
  update y una comprobación manual sobre el repositorio.

## Fuera de alcance

- Tasks P1 (visor read-only, un solo `in_progress`, playbook dividido, specs
  concisas/issue body).
- CI y release.
- Cambios en el contrato de roles o en el comportamiento de `rei check` más allá
  de la configuración del proyecto.
