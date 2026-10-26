// Package rei expone los sistemas de archivos de REI Harness embebidos en el
// binario: el esqueleto neutral que despliega `rei init` y el material del
// harness que se comparte con el proyecto destino.
//
// El paquete raíz no tiene dependencias ni lógica: solo declara los `embed.FS`
// que consume `internal/initwizard`. Debe vivir en la raíz del módulo porque
// `go:embed` no admite patrones con `..`.
package rei

import "embed"

// Skeleton contiene el esqueleto desplegable y neutral de `rei init`:
// AGENTS.md, .rei/config.json y .rei/docs/project/**. Sus rutas llevan el
// prefijo `skeleton/`; `internal/initwizard` lo elimina al instalarlo en la
// raíz del proyecto destino.
//
// Es independiente de los archivos vivos del repositorio: no hereda su
// personalización (documentación de REI, integración GitHub ni checks).
//
//go:embed all:skeleton
var Skeleton embed.FS

// Harness contiene el material del arnés que se instala tal cual, con sus rutas
// canónicas: .rei/docs/harness/**, .rei/docs/usage.md, .rei/agents/**,
// .rei/adapters/** y .rei/templates/**.
//
// El patrón selectivo con prefijo `all:` incluye los subárboles completos bajo
// `.rei/` (incluidos nombres que empiezan por `.`) y excluye deliberadamente
// .rei/specs/ y .rei/progress/ (estado del proyecto destino, no del arnés), así
// como los documentos vivos del proyecto.
//
//go:embed all:.rei/docs/harness all:.rei/docs/usage.md all:.rei/agents all:.rei/adapters all:.rei/templates
var Harness embed.FS
