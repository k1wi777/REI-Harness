# Requirements — Modo directo: cambios sin Work Item autorizados

> Requisitos en EARS. Cada uno tiene un identificador estable y describe un
> comportamiento observable del arnés (documentación y contrato del Leader).

## Definición

**R1** — La documentación del arnés DEBE presentar el modo directo como un modo
fuera del workflow de Work Items, no como un tercer tipo de Work Item ni como un
estado del ciclo.

**R2** — CUANDO el usuario solicite explícitamente un «cambio directo» o «sin
Work Item», el Leader DEBE evaluar el alcance y la elegibilidad antes de modificar
ningún archivo.

**R3** — El Leader NO DEBE iniciar el modo directo por iniciativa propia ni a
partir de una petición que no lo solicite de forma explícita.

## Elegibilidad y autorización

**R4** — DONDE el cambio se limite a documentación, erratas, formato, comentarios
o ajustes triviales sin cambio de comportamiento, el Leader DEBE considerarlo
elegible para el modo directo.

**R5** — CUANDO el cambio sea elegible, el Leader DEBE solicitar confirmación
explícita antes de aplicarlo e informar de que no habrá planificación, rama ni
registro.

**R6** — SI el Leader no puede determinar con certeza la elegibilidad, ENTONCES
DEBE avisar al usuario y requerir confirmación explícita antes de continuar.

## Negativa

**R7** — SI el cambio afecta comportamiento, seguridad, datos, dependencias, API
pública, arquitectura, más de un archivo o requiere especificación o revisión,
ENTONCES el Leader DEBE negarse a aplicarlo en modo directo.

**R8** — CUANDO el Leader se niegue, DEBE ofrecer continuar el cambio como Task o
Feature mediante el workflow.

**R9** — El Leader NO DEBE aceptar la insistencia del usuario como autorización
para saltarse una negativa.

## Sin registro

**R10** — El modo directo NO DEBE crear Work Item, spec, `meta.json`,
`current.md`, `history.md`, `direct.md`, sesión, issue ni rama; DEBE operar sobre
la rama y el árbol de trabajo actuales.

**R11** — El modo directo NO DEBE modificar el estado, la sesión, el historial ni
los documentos de ningún Work Item existente.

## Verificación y entrega

**R12** — CUANDO se aplique un cambio directo, el Leader DEBE ejecutar `rei check`
y, según el cambio, `go test` o `make build`, y reportar los comandos ejecutados
y su resultado.

**R13** — El modo directo NO DEBE lanzar al Reviewer salvo petición explícita del
usuario.

**R14** — El Leader NO DEBE crear commit, push, PR ni merge de un cambio directo
sin una orden explícita del usuario.

## Límites de la excepción

**R15** — La excepción a «NUNCA implementes código» DEBE limitarse al modo directo
autorizado del Leader; los demás roles y el resto de reglas duras DEBEN permanecer
vigentes.

## Documentación

**R16** — `AGENTS.md` y `skeleton/AGENTS.md` DEBEN documentar el modo directo con
reglas equivalentes, sin referencias específicas del repositorio en la plantilla.

**R17** — `.rei/docs/harness/workflow.md`, `.rei/agents/leader.md`,
`.rei/docs/usage.md` y `README.md` DEBEN documentar disparador, elegibilidad,
negativa, ausencia de registro, verificación y ausencia de commit.

**R18** — CUANDO se modifique el contrato del Leader, los agentes nativos de
OpenCode y Claude Code DEBEN regenerarse y sus `--check` DEBEN devolver `0`.

**R19** — El cambio NO DEBE modificar código Go ni el comportamiento del CLI; y
`go test ./...`, `go vet ./...`, `make build`, `rei check`, `rei doctor`,
`rei init status`, `rei validate` y los `--check` de adaptadores DEBEN quedar en
verde.
