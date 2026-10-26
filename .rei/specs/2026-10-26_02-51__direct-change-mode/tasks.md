# Tasks — Modo directo: cambios sin Work Item autorizados

> Orden de ejecución. Cada tarea indica los requisitos que implementa.
> No hay código Go: todas las tareas son edición documental y verificación.

- [x] T1 — Añadir en `.rei/docs/harness/workflow.md` la sección «Modo directo
  (fuera del workflow)» antes de «Tipos de Work Item»: definición (no es tipo ni
  estado), disparador, elegibilidad, negativa con oferta de Task/Feature sin
  override, confirmación explícita, ausencia de registro y rama actual,
  verificación, sin Reviewer salvo petición y sin commit/push. Ajustar el párrafo
  de «Objetivo» y añadir el bullet correspondiente en «Reglas generales» sin
  alterar el flujo Feature/Task. (R1, R2, R4, R5, R7, R8, R9, R10, R11, R12, R13,
  R14, R15)

- [x] T2 — En `.rei/agents/leader.md`, actualizar el `description` del frontmatter
  y la Identidad para acotar «NUNCA implementa» al modo directo autorizado; añadir
  el paso de triaje en «Arranque» y el subprotocolo «Modo directo autorizado (sin
  Work Item)» antes de «Personalización del proyecto»: evaluación previa,
  elegibilidad, duda con aviso y confirmación, negativa con oferta Task/Feature,
  no override, comprobación de `in_progress`, sin registro ni rama, verificación
  reportada, sin Reviewer salvo petición y sin commit/push/PR/merge sin orden.
  (R2, R3, R4, R5, R6, R7, R8, R9, R10, R11, R12, R13, R14)

- [x] T3 — En `.rei/agents/leader.md`, actualizar «Reglas duras»: excepción
  acotada a «NUNCA implementes código», prohibición de aplicar el modo directo sin
  disparador o tras una negativa, y prohibición de crear registro, rama o issue.
  Revisar «Documentos de referencia» para que apunten a `workflow.md`. (R3, R9,
  R10, R15)

- [x] T4 — En `AGENTS.md`, añadir la subsección «Modo directo autorizado» en §5
  (disparador, elegibilidad, negativa sin override, confirmación, sin registro,
  verificación, sin Reviewer salvo petición, sin commit), la frase de excepción en
  §6 y el bullet en §9 que aclara que no crea Work Item ni sesión. Mantener el
  resto de secciones intacto. (R1, R16)

- [x] T5 — Replicar en `skeleton/AGENTS.md` la misma subsección de §5 y el bullet
  de §9 con redacción neutral: sin comandos del repositorio fuente ni rutas
  específicas; la verificación se expresa como los checks del proyecto (`rei check`
  y los declarados en `.rei/config.json`/`verification.md`). Conservar los
  marcadores `<!-- REI:PENDIENTE -->`. (R16)

- [x] T6 — En `.rei/docs/usage.md`, añadir dentro de §3 la subsección «Atajo:
  cambios directos sin Work Item» (cómo pedirlo, qué es elegible, negativa y
  oferta de Task/Feature, confirmación, ausencia de registro, verificación, sin
  Reviewer salvo petición, sin commit) y una entrada en la FAQ. No renumerar las
  secciones existentes. (R17)

- [x] T7 — En `README.md`, acotar en «Cómo funciona §3» la frase de que el código
  nunca se implementa sin planificación aprobada y añadir en «Work Items» un
  párrafo que presente el modo directo como vía fuera de los dos tipos, con enlace
  a `.rei/docs/harness/workflow.md`. (R17)

- [x] T8 — Revisar la coherencia cruzada de los seis documentos: misma formulación
  de elegibilidad y negativa, ninguna afirmación absoluta que contradiga la
  excepción, ninguna mención a registro/sesión/rama en modo directo y ninguna
  alteración del flujo Feature/Task. Corregir las contradicciones detectadas. (R1,
  R2, R15, R17)

- [x] T9 — Regenerar los agentes nativos del Leader con `rei init opencode` y
  `rei init claude`; verificar con `rei init opencode --check` y
  `rei init claude --check` que devuelven `0`. No editar a mano archivos con la
  marca `GENERATED`. (R18)

- [x] T10 — Ejecutar la verificación del arnés: `go test ./...`, `go vet ./...`,
  `make build`, `rei check`, `rei doctor`, `rei init status` y
  `rei validate 2026-10-26_02-51__direct-change-mode`. Corregir cualquier fallo
  documental o de formato (sin tocar código Go) y registrar la evidencia en
  `.rei/progress/work-items/<id>/impl.md`. (R19)

- [x] T11 — Confirmar que no se ha modificado código Go, el CLI, `.rei/adapters/`
  ni `skeleton/` fuera de `skeleton/AGENTS.md`; reportar en `impl.md` el diff
  acotado y los resultados de verificación. (R19)
