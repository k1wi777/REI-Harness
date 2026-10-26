# Verification

> **Documento del proyecto** — define cómo se verifica que un Work Item de REI Harness ha sido implementado correctamente.
>
> El objetivo no es indicar **qué** se implementó, sino **cómo demostrar que funciona**.

## Estrategia de verificación

La verificación es **local y reproducible**: no hay workflow de CI que ejecute los
tests (`.github/workflows/release.yml` solo publica releases en tags `v*`).

Se apoya en:

- **`go vet`** para errores estáticos evidentes.
- **Tests automatizados** con `go test ./...` sobre todos los paquetes
  `internal/**` y `cmd/rei`.
- **Build** del binario real (`make build`), que además valida `go:embed` del
  esqueleto.
- **Build multiplataforma** (`make cross`) antes de un release o cuando el cambio
  toque código específico de plataforma (`process_unix.go`, `process_windows.go`).
- **Validación manual del CLI** cuando el cambio afecta a la interacción: ejecutar
  el comando afectado sobre un proyecto de prueba (`t.TempDir()` o un repo
  desechable) y comprobar salida y código de salida.

## Checkpoints de verificación

Cada checkpoint posee un identificador estable (`V1`, `V2`, ...). El Implementer y
el Reviewer deben referenciar estos IDs en sus reportes.

### Checkpoints rápidos (se declaran en `.rei/config.json`)

Verificaciones de bajo costo que pueden ejecutarse muchas veces al día. `rei check`
se ejecuta al iniciar sesión, al implementar y al revisar.

| ID | Comando | Descripción |
|----|---------|-------------|
| `V1` | `go vet ./...` | Análisis estático; detecta errores evidentes de compilación y uso incorrecto de la API. |

### Checkpoints lentos (NO van en `.rei/config.json`)

Se ejecutan manualmente antes de cerrar el Work Item (o en el release), se
documentan aquí y se referencian en la evidencia.

| ID | Comando | Descripción |
|----|---------|-------------|
| `V2` | `go test ./...` | Suite completa de tests unitarios (~13 s); compila todos los paquetes y valida el comportamiento. Se ejecuta antes de cerrar o revisar el Work Item, no en cada `rei check`. |
| `V3` | `make build` | Compila y enlaza `bin/rei` con `-ldflags`; valida el `go:embed` del esqueleto y el punto de entrada. |
| `V4` | `make cross` | Build para linux/darwin/windows y amd64/arm64; obligatorio si el cambio toca código por plataforma o el proceso de release. |

> Los checkpoints rápidos se declaran en `.rei/config.json`:
>
> ```json
> {
>   "checks": [
>     { "id": "V1", "description": "Análisis estático", "command": ["go", "vet", "./..."] }
>   ]
> }
> ```
>
> `rei check` ejecuta cada `command` (sin shell intermedio) y refleja su resultado
> en el código de salida. Si añades o modificas un checkpoint, actualiza también
> `.rei/config.json`.

> **Recomendación de personalización:** durante la iteración puedes ejecutar solo el
> paquete afectado (`go test ./internal/view/...`) para ir rápido, pero la evidencia
> final debe incluir `V2` completo: `go test ./...`.

## Evidencia

Como mínimo, por cada checkpoint ejecutado:

- ID del checkpoint (`V1`, `V2`, `V3`, `V4`);
- resultado (pasa / falla);
- comando ejecutado;
- observaciones relevantes (logs, capturas, validaciones manuales) cuando el
  checkpoint no sea automatizable.

La evidencia debe documentarse en:

```
.rei/progress/work-items/<work-item-id>/impl.md
```

Si un checkpoint no aplica, se indica explícitamente el motivo (por ejemplo, `V4`
no aplica porque el cambio es solo de documentación), no se omite en silencio.

## Bloqueos

Un Work Item pasa a `blocked` cuando alguna verificación falla y no puede
resolverse dentro del alcance de la sesión.

El protocolo exacto para declarar un bloqueo (qué documentar, dónde, y cómo
detenerse) está definido en el archivo de rol correspondiente (`.rei/agents/implementer.md`, `.rei/agents/reviewer.md`).

Nunca continúes la implementación ni solicites revisión ignorando una verificación
fallida.

## Qué NO hacer

- No asumir que un cambio funciona sin verificarlo contra los checkpoints definidos.
- No omitir checkpoints definidos para el proyecto.
- No solicitar revisión sin evidencia suficiente por cada checkpoint.
- No describir una verificación de forma libre si ya existe un ID (`V1`, `V2`, ...) para ella.
- No declarar `V2` como pasado ejecutando solo un subconjunto de paquetes.
- No ejecutar `V2`/`V3`/`V4` desde `rei check`: son lentos y no deben ir en la configuración.
