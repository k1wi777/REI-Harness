# Design — Modo directo: cambios sin Work Item autorizados

## Contexto

El arnés afirma hoy, en varios puntos, que ningún trabajo se implementa sin Work
Item y que el Leader nunca escribe código. El modo directo introduce una
excepción acotada: el Leader puede aplicar un cambio trivial solicitado
explícitamente por el usuario, sin planificación ni estado, sobre la rama actual.
La implementación es **exclusivamente documental y de contrato**: no hay código
Go, ni comandos nuevos, ni cambios en el CLI.

El riesgo principal no es técnico sino de **coherencia**: si la excepción se
documenta en un solo archivo, el resto del arnés seguirá afirmando lo contrario y
el Leader quedará sin criterio operativo.

## Estrategia

1. **Una norma canónica.** El procedimiento completo vive una sola vez en
   `.rei/docs/harness/workflow.md`. El resto de documentos lo resumen o lo
   aplican, sin redefinir criterios.
2. **Operacionalizar en el Leader.** `.rei/agents/leader.md` recibe el triaje
   (disparador, elegibilidad, autorización, negativa) y los límites de ejecución
   (sin registro, verificación, sin commit). Además se ajustan su `description`
   de frontmatter, su Identidad y sus Reglas duras, que hoy prohíben de forma
   absoluta implementar.
3. **Reflejar en los puntos de entrada.** `AGENTS.md` (vivo) y
   `skeleton/AGENTS.md` (plantilla) exponen la excepción en §5 y su relación con
   §9. La plantilla se redacta en términos neutrales, sin comandos del repo.
4. **Explicar al usuario.** `.rei/docs/usage.md` y `README.md` indican cómo pedir
   un cambio directo, qué es elegible, qué ocurre si se niega y qué trazabilidad
   queda.
5. **Eliminar contradicciones.** Cada afirmación previa del tipo «nunca se
   implementa sin planificación» se revisa y se acota con la excepción.
6. **Regenerar y verificar.** Se regeneran los agentes nativos de OpenCode y
   Claude Code desde el contrato editado y se comprueba con `--check`.

Criterio de redacción: **una sola formulación** de elegibilidad y negativa,
replicada literalmente en los documentos que la necesiten. Se evita copiar el
procedimiento completo en seis sitios: `workflow.md` norma, `leader.md`
ejecución, `AGENTS.md` resumen y excepción, `usage.md`/`README.md` uso, skeleton
norma neutral.

## Archivos afectados

| Archivo | Cambio | Motivo |
|---------|--------|--------|
| `.rei/docs/harness/workflow.md` | Nueva sección «Modo directo (fuera del workflow)» antes de «Tipos de Work Item»; bullet en «Reglas generales»; ajuste del párrafo de «Objetivo». | Fuente canónica de la política. |
| `.rei/agents/leader.md` | `description` de frontmatter; Identidad; nuevo subprotocolo «Modo directo autorizado»; Reglas duras. | El Leader necesita criterio ejecutable, no solo doctrina. |
| `AGENTS.md` | Subsección en §5; frase en §6; bullet en §9. | Punto de entrada universal: evita contradicción inmediata. |
| `skeleton/AGENTS.md` | Misma subsección y bullet, en versión neutral. | La plantilla que despliega `rei init` debe incluir la norma. |
| `.rei/docs/usage.md` | Subsección «Atajo: cambios directos sin Work Item» dentro de §3; entrada en FAQ. | El usuario debe saber cómo pedirlo y qué esperar. |
| `README.md` | Nota en «Cómo funciona §3» y párrafo en «Work Items». | Presentación pública del arnés. |
| `.opencode/agents/leader.md`, `.claude/agents/leader.md` | Regenerados con `rei init opencode` / `rei init claude`. | Derivan del contrato; quedarían obsoletos. |

Sin cambios en `cmd/`, `internal/`, `embed.go`, `.rei/adapters/` ni en el resto de
`skeleton/`.

## Decisiones de diseño

### D1 — La política vive en el contrato, no en el CLI

El modo directo es una **decisión de triaje del Leader**, no una capacidad del
binario. Añadir un subcomando o un flag (`rei direct`) obligaría a tocar Go,
estado y adaptadores, y daría una falsa sensación de control automático: el CLI no
puede evaluar si un cambio es trivial. Se descarta.

Consecuencia: la garantía es **contractual y verificable por revisión de
documentos**, no mecánica. Se compensa con una lista de elegibilidad estrecha, la
negativa obligatoria y una verificación posterior explícita.

### D2 — `workflow.md` como fuente canónica y ubicación de la sección

La sección se inserta **antes** de «Tipos de Work Item», porque el triaje del
modo directo ocurre antes de decidir Feature o Task. Debe quedar claro que:

- no es un tipo ni un estado;
- no pasa por `spec_author`, aprobación de planificación ni `reviewer`;
- no altera el invariante de un único Work Item `in_progress`, porque no crea ni
  cambia Work Items.

Se añade un bullet en «Reglas generales» que remite a la sección, para que las
reglas duras del workflow no parezcan incompletas.

### D3 — Ubicación en el contrato del Leader

Se añade `#### Modo directo autorizado (sin Work Item)` como **primer caso del
protocolo**, antes de «Personalización del proyecto» y «Antes de delegar». El
motivo es que el disparador llega con la petición: si el Leader entra primero en
la clasificación Feature/Task, ya está creando un Work Item cuando debía aplicar
el atajo.

En «Arranque» se añade un paso previo: si la petición invoca un cambio directo, se
consulta ese caso antes de la comprobación de `in_progress`. El resto del
arranque no cambia.

### D4 — Elegibilidad: lista cerrada más cláusula abierta

La lista de casos elegibles es cerrada y conservadora: documentación, erratas,
formato, comentarios y ajustes triviales sin comportamiento. La lista de negativa
incluye los dominios del acuerdo (comportamiento, seguridad, datos, dependencias,
API pública, arquitectura, multiarchivo) **más** la cláusula abierta «cualquier
cosa que necesite especificación o revisión».

Esto resuelve el caso límite sin enumerar infinitas situaciones: ante duda, la
regla es avisar y exigir confirmación (R6); ante un criterio de negativa, negarse.
La asimetría es deliberada: **el defecto es no aplicar el modo directo**.

### D5 — Confirmación explícita incluso en cambios elegibles

Aunque el cambio sea claramente elegible, el Leader no lo aplica hasta que el
usuario confirma. La confirmación debe informar de las tres ausencias que sorprenden
al usuario: sin planificación, sin rama y sin registro. El silencio o una respuesta
ambigua no son autorización.

### D6 — Sin registro: prohibición de llamadas de estado

El modo directo no ejecuta `rei new`, `rei session start`, `rei status set`,
`rei commit set`, `rei github` ni escribe `current.md`, `meta.json`, `history.md`
ni un `direct.md`. La única traza es el diff/commit que el usuario decida y la
conversación.

Esta es una decisión del usuario y acepta una tensión con el principio «la
documentación es la fuente de verdad». Se acota por tres vías: alcance trivial,
autorización explícita y verificación reportada. No se crea un registro parcial
porque un artefacto sin gate de aprobación daría apariencia de workflow sin serlo.

### D7 — Verificación proporcional y reportada

`rei check` es obligatorio siempre. `go test` y `make build` se añaden según el
cambio: tests si se toca documentación que el código embebe o comportamiento
observable; build si el cambio afecta al binario o al esqueleto embebido. El
Leader reporta **comando y resultado**, no una afirmación genérica de que «todo
pasa».

### D8 — Sin Reviewer automático y sin entrega no autorizada

El Reviewer presupone un Work Item con planificación y criterios; en modo directo
no existe. Solo se activa si el usuario lo pide explícitamente, y en ese caso la
revisión se limita al diff, sin crear artefactos. Commit, push, PR y merge
requieren orden explícita, igual que en el flujo normal.

### D9 — Paridad vivo/plantilla con redacción neutral

`AGENTS.md` y `skeleton/AGENTS.md` deben contener la misma norma, pero la plantilla
no puede citar `go test`, `make build` ni rutas del repositorio fuente. En la
plantilla la verificación se expresa como «los checks del proyecto (`rei check` y
los declarados en `.rei/config.json`/`verification.md`)». Se conserva el marcador
`<!-- REI:PENDIENTE -->` de las secciones personalizables para no romper
`TestSkeletonNeutralTieneMarcadores`.

### D10 — Regenerar ambos adaptadores

En este repositorio existen agentes nativos de OpenCode y Claude Code. Ambos
derivan del `## Contrato` y del `description` del frontmatter, por lo que editar
solo `leader.md` deja deriva. Se regeneran los dos y se verifican con
`rei init opencode --check` y `rei init claude --check`. Los archivos generados
no se editan a mano.

### D11 — No tocar el flujo Feature/Task

No cambian estados, transiciones, presupuestos, plantillas, CLI ni el
comportamiento de `spec_author`, `implementer` o `reviewer`. El modo directo es
una excepción de triaje, no una modificación del ciclo.

## Afirmaciones existentes que hay que ajustar

| Documento | Afirmación actual | Ajuste |
|-----------|-------------------|--------|
| `.rei/agents/leader.md` | `description`: «NUNCA implementa directamente». | Excepción acotada al modo directo autorizado. |
| `.rei/agents/leader.md` | Identidad: «NUNCA implementes código directamente». | Misma excepción, remitiendo al subprotocolo. |
| `.rei/agents/leader.md` | Reglas duras: «NUNCA implementes código». | Excepción + bullets de negativa, sin override y sin registro. |
| `AGENTS.md` / `skeleton/AGENTS.md` | §5: «No omitas ninguna etapa del workflow». | Añadir que el modo directo es la excepción y dónde se define. |
| `AGENTS.md` / `skeleton/AGENTS.md` | §9: «Trabaja sobre un único Work Item por sesión». | Aclarar que el modo directo no abre Work Item ni sesión. |
| `.rei/docs/harness/workflow.md` | «No todos los cambios requieren el mismo nivel de planificación» y solo dos tipos. | Introducir el modo directo como vía ajena a los tipos. |
| `.rei/docs/harness/workflow.md` | «Ningún Work Item puede finalizar sin pasar por `reviewer`». | Sigue siendo cierto: el modo directo no es un Work Item. Añadir remisión. |
| `README.md` | «El código nunca se implementa antes de existir una planificación aprobada». | Acotar con «salvo en modo directo autorizado para cambios triviales». |
| `.rei/docs/usage.md` | «Cada Work Item pasa por las mismas etapas». | Añadir el atajo y sus límites sin reescribir el ciclo. |

## Alternativas descartadas

- **Tercer tipo de Work Item (`direct`).** Requeriría `rei new`, estados, sesión,
  plantilla de planificación y validaciones; contradice el acuerdo y el propio
  concepto de Work Item.
- **Artefacto ligero (`direct.md`) o entrada en `history.md`.** El usuario decidió
  no dejar registro; además crearía un estado sin gate de aprobación.
- **Documentar solo en `leader.md`.** El resto del arnés seguiría afirmando lo
  contrario y el usuario no sabría cómo pedirlo.
- **Comando/flag `rei direct`.** Código Go, dependencia de runtime y falsa
  automatización; fuera de alcance.
- **Permitir override del usuario tras una negativa.** Elimina la salvaguarda que
  justifica la existencia del modo; se ofrece Task/Feature en su lugar.

## Riesgos y mitigaciones

- **Ambigüedad de elegibilidad.** Lista cerrada, cláusula abierta y defecto
  negativo; ante duda, confirmación explícita.
- **Presión para colar cambios no triviales.** Negativa obligatoria, sin override y
  oferta de Task/Feature; el Leader no negocia los criterios.
- **Contaminación del diff de un Work Item `in_progress`.** El modo directo no
  toca su estado; antes de aplicar el cambio, el Leader comprueba
  `rei items status` y, si hay un Work Item activo, advierte de que el cambio
  compartirá árbol y diff y pide confirmación.
- **Deriva de agentes nativos.** Regeneración de ambos runtimes y `--check` como
  puerta final.
- **Deriva documental entre seis archivos.** Una formulación canónica, réplica
  literal de criterios y una tarea específica de coherencia cruzada.
- **Expectativa de trazabilidad.** `usage.md` y `README.md` explican que la única
  traza es el diff/commit y el chat, para que no se perciba como un fallo.

## Verificación del propio Work Item

- `rei validate 2026-10-26_02-51__direct-change-mode` sin `FAIL`.
- `go test ./...` y `go vet ./...` en verde.
- `make build` compila y valida el `go:embed` del esqueleto.
- `rei check`, `rei doctor`, `rei init status` sin errores.
- `rei init opencode --check` y `rei init claude --check` devuelven `0`.
- Revisión manual de coherencia: ninguna afirmación absoluta contradice la
  excepción; `skeleton/AGENTS.md` no cita comandos ni rutas del repo fuente.

## Fuera de alcance

Registro del cambio directo, Reviewer automático, cambios de comportamiento,
seguridad, datos, dependencias, API pública, arquitectura o multiarchivo, y
cualquier modificación del flujo Feature/Task, del CLI o del resto del skeleton.
