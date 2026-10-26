# Verification

> **Documento del proyecto** — define cómo se verifica que un Work Item de REI Harness ha sido implementado correctamente.
>
> El objetivo no es indicar **qué** se implementó, sino **cómo demostrar que funciona**.

## Estrategia de verificación

La verificación combina tres vías complementarias y es reproducible en local:

- **CI** (`.github/workflows/ci.yml`): en cada push a `main`/`dev`, cada PR contra
  esas ramas y bajo demanda (`workflow_dispatch`), un único job `test` comprueba
  formato (`gofmt -l .`), `go vet ./...` y `go test -count=1 ./...`. El CI
  **valida**: no compila ni publica binarios.
- **CD** (`.github/workflows/release.yml`): al publicar un tag `v*`, GoReleaser
  construye y publica los binarios de release. El CD **publica**.
- **Manual**: `make build` (`V3`) y `make cross` (`V4`) se ejecutan a mano antes de
  cerrar un Work Item cuando el cambio lo justifica. El CI no los ejecuta.

Se apoya en:

- **Formato** con `gofmt -l .`: puerta del job `test` (falla si lista archivos) y
  comprobación obligatoria antes de revisar (`make fmt` para corregir).
- **`go vet`** para errores estáticos evidentes; también se ejecuta en CI.
- **Tests automatizados** con `go test ./...` sobre todos los paquetes
  `internal/**` y `cmd/rei`; CI los ejecuta con `-count=1`, sin caché.
- **Build** del binario real (`make build`), manual: además de enlazar el binario
  valida `go:embed` del esqueleto. El CI no compila.
- **Build multiplataforma** (`make cross`), manual o cubierto por el release:
  obligatorio si el cambio toca código específico de plataforma
  (`process_unix.go`, `process_windows.go`) o el proceso de release. El CI no lo
  ejecuta.
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

`V1` sigue siendo el único checkpoint declarado en `rei check`. El job `test` de
CI ejecuta además `go vet` como refuerzo, pero `.rei/config.json` no cambia.

### Checkpoints lentos (NO van en `.rei/config.json`)

Se ejecutan **manualmente o en el CD**, nunca desde `rei check`; el job de CI solo
ejecuta `V2`. Se documentan aquí y se referencian en la evidencia.

| ID | Comando | Descripción |
|----|---------|-------------|
| `V2` | `go test ./...` | Suite completa de tests unitarios (~13 s); compila todos los paquetes y valida el comportamiento. La ejecuta el job `test` de CI (`go test -count=1 ./...`) y también puede ejecutarse a mano antes de revisar el Work Item. |
| `V3` | `make build` | Compila y enlaza `bin/rei` con `-ldflags`; valida el `go:embed` del esqueleto y el punto de entrada. **Manual**: el CI no compila binarios. |
| `V4` | `make cross` | Build multiplataforma **manual** (o cubierto por el release). Produce cinco binarios: linux/darwin × amd64/arm64 y windows × amd64; GoReleaser cubre las seis combinaciones en release, incluida windows/arm64. Obligatorio si el cambio toca código por plataforma o el proceso de release. |

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

Para `V2`, el run de CI es evidencia válida: se indica la URL del run, el job
`test` y el commit SHA. El run debe corresponder al commit que se está revisando.
Para `V3`/`V4` se documenta la ejecución manual (comando, resultado y salida
relevante) o el run de release del CD cuando aplique.

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
- No citar un run de CI de otro commit como evidencia del Work Item revisado.
- No declarar verificado un job de CI que ha fallado.
- No atribuir a CI la ejecución de `V3`/`V4`: el CI no compila binarios.
