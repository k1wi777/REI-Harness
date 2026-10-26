# Convenciones

> **Documento del proyecto** — define las convenciones de desarrollo de REI Harness.
>
> El Implementer debe seguirlas durante la implementación y el Reviewer las utilizará para validar la calidad del código.

---

# Objetivo

Mantener un estilo de desarrollo consistente en todo el proyecto.

Las convenciones describen **cómo escribir el código**, no cómo está organizada la arquitectura del sistema.

---

# Organización del código

- Un paquete por directorio bajo `internal/<área>`, con nombre corto en minúsculas
  y sin guiones (`initwizard`, `workitem`, `view`).
- Los tests viven junto al código, en el mismo paquete, como `<archivo>_test.go`.
- El código específico de plataforma usa sufijos `_unix.go` / `_windows.go` en
  lugar de condicionales en tiempo de ejecución.
- `cmd/rei/main.go` solo llama a `cli.Run` y devuelve el código de salida; no
  contiene lógica de dominio.
- Los assets del visor (HTML, CSS, JS) van en `internal/view/assets/` y se embeben
  desde `internal/view/render.go`.
- Las plantillas canónicas de documentos viven en `.rei/templates/`; no se
  duplican literales de plantilla en Go.
- Las utilidades compartidas se extraen a un paquete `internal/` con API pequeña;
  no se crean paquetes `utils` genéricos.

---

# Nombres

- Paquetes en minúsculas, una sola palabra cuando sea posible.
- Identificadores exportados en `PascalCase`; no exportados en `camelCase`.
- Errores centinela con prefijo `Err` (`ErrNotFound`, `ErrNoGit`,
  `ErrTransitionNotAllowed`).
- Constantes y variables de configuración en `camelCase`; no se usa
  `UPPER_SNAKE_CASE` salvo que la constante represente un valor externo fijo.
- Nombres de test `Test<Comportamiento>` descriptivos, en español o inglés según el
  área, p. ej. `TestFindFromNotFound`, `TestHelpViewDocumentaModos`.
- Archivos en minúsculas con guion bajo solo para plataforma (`process_unix.go`);
  el resto, una palabra o CamelCase de Go (`initwizard.go`, `github_test.go`).
- Nombres de variables y funciones descriptivos; evitar abreviaturas opacas salvo
  las idiomáticas (`cfg`, `err`, `ctx`, `id`).

---

# Estilo de implementación

- `gofmt` es obligatorio: el código debe formatearse con `make fmt` antes de
  solicitar revisión.
- Funciones pequeñas y con una única responsabilidad; si una función de `cli`
  crece, se extrae un `cmdX` o un paquete de dominio.
- Preferir composición y funciones puras; el estado se pasa explícitamente
  (structs de opciones como `view.Options`).
- La salida de usuario va por `io.Writer` inyectable cuando se testea; no se
  escribe directamente a `os.Stdout` en lógica reutilizable.
- Los comandos nuevos se añaden en `internal/cli/cli.go` y su ayuda en
  `internal/cli/help.go`, con test que verifique que la ayuda los documenta.
- Evitar duplicación: una transición, validación o formato se implementa una vez y
  se reutiliza desde CLI y visor.
- No introducir complejidad especulativa ni abstracciones para un único uso.

## Tests

- Tests con `testing` de la biblioteca estándar; sin frameworks externos.
- Usar `t.TempDir()` para cualquier operación sobre ficheros.
- Casos simples como tests independientes; tabla (`t.Run` sobre slice de casos)
  cuando se repite el mismo comportamiento con varias entradas.
- Mensajes de fallo con el formato `campo = %q, want %q` o `esperaba ...`.
- Un test debe fallar por una razón clara y no depender de red, hora ni estado
  global.

## Commits

- Formato Conventional Commits: `<tipo>(<scope>): <resumen>`.
- Resumen en inglés, imperativo y breve; el cuerpo (si hace falta) puede ir en
  español.
- Tipos usados: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`.
- Un commit por cambio coherente; no mezclar refactor y feature.

---

# Manejo de errores

- Devolver `error` explícito; no usar `panic` para fallos esperables.
- Definir errores centinela en el paquete que los origina y comprobarlos con
  `errors.Is` en las capas superiores (`cli` traduce a mensajes y códigos de
  salida).
- Envolver errores con contexto usando `%w`:
  `fmt.Errorf("no se puede escribir en %s: %w", dir, err)`.
- Mensajes en español, en minúscula y sin punto final; incluir el dato relevante
  (ruta, nombre, estado) para que el usuario pueda actuar.
- Validar entradas en el límite del comando (`cli`) y volver a validar en el
  dominio cuando el paquete se pueda usar por otra vía.
- No ignorar errores: si se decide ignorar uno, debe ser explícito y justificado.

---

# Dependencias

- Biblioteca estándar primero. La única dependencia externa actual es
  `github.com/yuin/goldmark` (render de Markdown).
- No añadir una dependencia para algo que se resuelve en pocas líneas de código
  propio.
- Toda dependencia nueva debe justificarse en el Work Item (qué aporta, por qué no
  basta la stdlib, impacto en el binario y en la licencia).
- Mantener `go.mod` y `go.sum` actualizados; no introducir dependencias
  transitivas pesadas ni generadores en tiempo de build.
- Prohibido introducir JS/CSS/fuentes de terceros en `internal/view/assets`.

---

# Comentarios

- Todo paquete exportado lleva comentario de paquete (`// Package x ...`).
- Documentar la API exportada y las decisiones no obvias; explicar **por qué**, no
  **qué**.
- Comentarios en español, coherentes con el resto de la documentación.
- Las decisiones de build (por ejemplo, la inyección de versión con `-ldflags`) se
  documentan junto a la variable afectada.
- No dejar comentarios obsoletos: si el código cambia, el comentario cambia o se
  elimina.

---

# Qué NO hacer

- Duplicar lógica entre CLI, visor y paquetes de dominio.
- Dejar código muerto, comentado o `TODO` sin Work Item asociado.
- Utilizar nombres ambiguos (`data`, `manager`, `helper`) sin contexto.
- Ignorar errores o convertirlos en mensajes genéricos sin causa.
- Añadir dependencias sin justificación.
- Escribir directamente en `os.Stdout` desde lógica que deba testearse.
- Acceder a la red fuera de `internal/update` e `internal/github`.
- Saltarse `make fmt` o `make vet` antes de pedir revisión.
