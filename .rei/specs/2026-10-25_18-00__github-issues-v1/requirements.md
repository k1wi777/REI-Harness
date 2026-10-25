# Requirements — Integración opcional con GitHub Issues (v1)

> Work Item: `2026-10-25_18-00__github-issues-v1` (feature).
>
> Los requisitos describen **qué** debe hacer el sistema, no **cómo**. Se
> redactan en EARS. El sujeto «el sistema» designa al CLI `rei`, al rol
> `initializer` o al rol `Leader`, según el requisito.
>
> Principio rector: GitHub es una **proyección** del Work Item, no su fuente de
> verdad. `.rei` manda sobre la ejecución; la issue manda sobre el problema y la
> discusión. Toda la integración es opcional y su ausencia debe dejar el flujo
> actual intacto.

---

# 1. Comando `rei github`

**R1.** El sistema DEBE ofrecer el comando `rei github` con los subcomandos
`enable`, `disable` y `status`.

**R2.** SI `rei github` se ejecuta sin subcomando, con un subcomando desconocido
o con argumentos sobrantes, ENTONCES el sistema DEBE mostrar el uso por
`stderr` y devolver código de salida `2`.

**R3.** SI no existe un proyecto REI (`.rei/`) en el directorio actual ni en un
ancestro, ENTONCES el sistema DEBE informar de la causa y del siguiente paso
(ejecutar `rei init`) y devolver código de salida `1`.

**R4.** El sistema DEBE almacenar la configuración de la integración en
`.rei/config.json`, bajo `integrations.github`, con las claves `enabled`
(booleano), `repo` (`owner/name`), `base` (rama base) y `mergeStrategy`
(estrategia de merge). La ausencia de la clave `integrations.github` DEBE
significar «integración desactivada».

**R5.** CUANDO se ejecuta `rei github enable`, el sistema DEBE detectar el
repositorio desde el remoto `origin` del proyecto y escribir `enabled: true`,
`repo`, `base` y `mergeStrategy`. La ejecución repetida DEBE ser idempotente.

**R6.** El sistema DEBE aceptar remotos SSH con forma
`git@github.com:owner/name(.git)` y remotos HTTPS con forma
`https://github.com/owner/name(.git)`, admitiendo una barra final. DEBE
descartar el sufijo `.git`, validar `owner` y `name`, y rechazar hosts distintos
de `github.com`, rutas con más de dos segmentos o valores vacíos.

**R7.** SI no se puede detectar el repositorio o el remoto no es válido,
ENTONCES `rei github enable` DEBE informar de la causa y del siguiente paso
(revisar `git remote -v` o editar `integrations.github.repo`), NO DEBE escribir
la configuración y DEBE devolver código de salida `1`.

**R8.** CUANDO se ejecuta `rei github enable`, el sistema DEBE determinar la
rama base a partir de la rama por defecto de `origin`; si no puede hacerlo,
DEBE usar `main` e informarlo. La estrategia de merge por defecto DEBE ser
`squash`. Si ya existían `base` o `mergeStrategy`, DEBE conservarlos.

**R9.** CUANDO se ejecuta `rei github enable` y `gh` no está instalado o no está
autenticado, el sistema DEBE emitir un aviso no bloqueante con la causa y el
siguiente paso (`gh auth login` o instalación de GitHub CLI), DEBE escribir la
configuración y DEBE devolver código de salida `0`.

**R10.** CUANDO se ejecuta `rei github disable`, el sistema DEBE escribir
`enabled: false`, conservar `repo`, `base`, `mergeStrategy`, el resto de claves
de `config.json` y los `links.json` existentes, y NO DEBE contactar con GitHub.
Si ya estaba desactivada o no configurada, DEBE informarlo y devolver `0`.

**R11.** CUANDO se ejecuta `rei github status`, el sistema DEBE ser de solo
lectura y DEBE informar de: si la integración está no configurada, desactivada o
activa; los valores efectivos de `repo`, `base` y `mergeStrategy`; si `gh` está
disponible y autenticado; y si el `repo` configurado coincide con el remoto
`origin` detectado. DEBE devolver `0` también cuando la integración esté
desactivada o no configurada.

**R12.** SI `.rei/config.json` no es JSON válido o no se puede leer, ENTONCES
`rei github` DEBE informar de la causa y del siguiente paso, NO DEBE sobrescribir
el archivo y DEBE devolver código de salida `1`.

**R13.** `rei github enable` y `rei github disable` DEBEN preservar `checks`,
`viewer` y cualquier otra clave existente de `.rei/config.json`. SI la escritura
falla, ENTONCES el archivo DEBE quedar intacto.

**R14.** El comando `rei github` DEBE implementarse sin dependencias Go nuevas y
NO DEBE crear issues, comentarios, ramas, PR ni merges: esas acciones
corresponden al agente.

---

# 2. Entrevista del `initializer`

**R15.** CUANDO el rol `initializer` llega al paso de configuración (el mismo en
el que pregunta por el visor), DEBE preguntar al usuario si desea activar la
integración con GitHub Issues.

**R16.** SI el usuario responde que sí, ENTONCES el rol `initializer` DEBE
configurar la integración mediante `rei github enable`, mostrar el repo y la
rama base detectados y trasladar los avisos de `gh` con su siguiente paso.

**R17.** SI el usuario responde que no, ENTONCES el rol `initializer` NO DEBE
escribir `integrations.github`; la ausencia de la clave DEBE quedar como
integración desactivada.

**R18.** SI `integrations.github` ya existe, ENTONCES el rol `initializer` DEBE
mostrar la configuración actual y pedir confirmación antes de cambiarla; NUNCA
DEBE modificarla en silencio.

**R19.** SI `gh` no está instalado o no está autenticado, ENTONCES el rol
`initializer` DEBE informarlo con causa y siguiente paso, sin bloquear la
personalización ni el cierre del paso.

---

# 3. Puerta única al aprobar la planificación

**R20.** MIENTRAS la integración esté desactivada o no configurada, el Leader NO
DEBE mencionar GitHub, NO DEBE ejecutar `gh` y NO DEBE crear issues ni ramas; el
flujo DEBE ser idéntico al actual y trabajar sobre la rama base.

**R21.** CUANDO la integración esté activa, el usuario apruebe la planificación
y el Work Item no tenga issue vinculada, el Leader DEBE formular una única
pregunta: «¿Creo la issue y la rama `feat/<n>-<slug>`?», donde `<n>` será el
número de la issue y `<slug>` el slug del Work Item. NO DEBE volver a
preguntarlo en la misma sesión.

**R22.** SI el usuario responde que sí, ENTONCES el Leader DEBE: buscar primero
el marcador `<!-- rei:work-item:<id> -->` para no duplicar la issue; crearla si
no existe; crear la rama `feat/<n>-<slug>` desde la rama base; y escribir
`.rei/specs/<id>/links.json`.

**R23.** El cuerpo de la issue DEBE ser autosuficiente e incluir: el marcador
`<!-- rei:work-item:<id> -->`, el contexto y la decisión acordada, un resumen de
los requisitos, las decisiones de diseño relevantes, el fuera de alcance, la
referencia a `.rei/specs/<id>/` y el contenido completo de `design.md` plegado en
un bloque `<details>`. La issue DEBE llevar las labels `rei` y `rei:feature`.

**R24.** SI el usuario responde que no, ENTONCES el Leader NO DEBE crear issue ni
rama, DEBE continuar sobre la rama base como hoy y DEBE registrar la decisión en
`.rei/progress/current.md` para no volver a preguntar.

**R25.** SI la issue ya existe por el marcador, ENTONCES el Leader DEBE
reutilizarla, informar de su número y URL y NO DEBE crear otra. SI la rama ya
existe, NO DEBE forzarla ni recrearla; DEBE informar de la causa y del siguiente
paso.

**R26.** SI la creación de la issue tiene éxito pero falla la rama o
`links.json`, ENTONCES el Leader DEBE informar de qué quedó hecho y qué está
pendiente, DEBE conservar la issue y DEBE permitir reintentar sin duplicarla.

**R27.** `links.json` DEBE contener `provider: "github"`, `repo`, `issue`, `url`
y `epic` (vacío o `null` en v1). Es un índice local, NO es fuente de verdad.

**R28.** SI `links.json` falta, está corrupto o se desincroniza, ENTONCES el
sistema NO DEBE bloquear el Work Item; el Leader DEBE recuperar la issue
buscando el marcador y reescribir `links.json`.

---

# 4. Comentarios automáticos

**R29.** CUANDO el Implementer termina la implementación, el Leader DEBE publicar
sin preguntar un comentario en la issue vinculada, identificado por
`<!-- rei:comment:impl:<id> -->`, con un resumen de lo implementado, los archivos
principales, el resultado de la verificación y la referencia a
`.rei/progress/work-items/<id>/impl.md`.

**R30.** CUANDO el Reviewer termina la revisión, el Leader DEBE publicar sin
preguntar un comentario identificado por `<!-- rei:comment:review:<id> -->`, con
el resultado (`done`, `changes_requested` o `blocked`), las verificaciones y
observaciones relevantes, y la referencia al `review.md`.

**R31.** CUANDO exista ya un comentario con el marcador correspondiente, el
sistema DEBE actualizarlo en lugar de crear otro. NUNCA DEBE duplicar
comentarios.

**R32.** El Leader NUNCA DEBE publicar comentarios por cambios de estado del
Work Item; los únicos comentarios automáticos son los de implementación y
revisión.

**R33.** SI falla la publicación o actualización de un comentario, ENTONCES el
Leader DEBE informar de la causa y del siguiente paso, NO DEBE alterar el estado
del Work Item y DEBE poder reintentar sin duplicar el comentario.

**R34.** CUANDO un Work Item con issue vinculada pase a `blocked`, el Leader DEBE
aplicar la label `rei:blocked`; cuando deje de estar bloqueado, DEBE retirarla.
Estos cambios de label NO DEBEN generar comentarios.

---

# 5. Pull request y merge

**R35.** El Leader SOLO DEBE crear un pull request tras la revisión y bajo una
orden explícita del usuario. El PR DEBE usar la rama base configurada, la rama
del Work Item, un cuerpo con resumen, verificación e `Closes #<n>`, y DEBE
informar de su número y URL. La orden de PR DEBE habilitar únicamente los commits
y el push de la rama del Work Item; cualquier otro commit o push requiere una
orden propia.

**R36.** El Leader SOLO DEBE fusionar bajo una orden explícita del usuario,
mediante squash, con asunto `<título> (#<n>)` y borrado de la rama. NUNCA DEBE
usar `--admin` ni auto-merge.

**R37.** SI la branch protection, los checks requeridos o los permisos impiden
el merge, ENTONCES el Leader DEBE informar de la causa y del siguiente paso, DEBE
dejar el PR abierto y NO DEBE reintentar en bucle ni forzar el merge.

**R38.** Después de un merge, el Leader DEBE verificar el estado de la issue y,
si sigue abierta, cerrarla como completada, informando del resultado. SI la
verificación o el cierre falla, DEBE informar de qué quedó pendiente.

**R39.** El Leader NUNCA DEBE ejecutar commit, push, merge ni borrado de rama sin
una orden explícita del usuario.

---

# 6. Adopción de issues existentes

**R40.** CUANDO el usuario pida trabajar en una issue existente (`#N`) y la
integración esté activa, el Leader DEBE leer la issue, comprobar si está abierta,
buscar el marcador de Work Item y, si no existe, preguntar lo que falte para
definir el alcance antes de crear el Work Item. SI la issue está cerrada,
ENTONCES DEBE preguntar antes de continuar.

**R41.** CUANDO se adopte una issue sin marcador, el Leader DEBE crear el Work
Item, registrar el acuerdo en `description`, añadir el marcador sin destruir el
cuerpo existente, asegurar las labels y escribir `links.json`.

**R42.** SI la issue ya tiene un marcador de otro Work Item, ENTONCES el Leader
DEBE detenerse, informar del conflicto y NO DEBE crear un Work Item duplicado ni
sobrescribir el marcador.

**R43.** SI la integración está desactivada, el Leader NO DEBE leer ni modificar
issues; DEBE ofrecer activar la integración como siguiente paso.

---

# 7. Manejo de errores de GitHub

**R44.** Todo fallo de `gh`, de `git` o de configuración DEBE comunicarse con
causa y siguiente paso. NUNCA DEBE silenciarse y NUNCA DEBE afirmarse un éxito
sin verificarlo.

**R45.** El sistema DEBE cubrir, como mínimo, estos fallos: proyecto sin `.rei`,
integración desactivada o configuración incompleta, `gh` ausente, `gh` sin
autenticar, repo no detectado o remoto ambiguo, issue/rama/PR inexistente,
permisos insuficientes, branch protection que bloquea el merge, y red o rate
limit.

**R46.** Los fallos de GitHub NO DEBEN alterar el estado de REI (`meta.json`,
estado de `current.md` ni la etapa del workflow). El Leader DEBE informar de qué
quedó hecho y qué está pendiente, y el reintento NO DEBE duplicar issues ni
comentarios.

**R47.** Tras cada acción GitHub, el Leader DEBE verificar el resultado (número y
URL de issue, identificador de comentario, PR, estado del merge y estado de la
issue) antes de informar de éxito.

**R48.** SI se alcanza un rate limit o falla la red, ENTONCES el Leader DEBE
informar de la causa, del momento de reintento si está disponible y DEBE evitar
reintentos en bucle.

**R49.** `rei github` DEBE usar los códigos de salida `0` (correcto, incluso con
avisos), `1` (error operativo o de configuración) y `2` (uso incorrecto).

---

# 8. Documentación, contratos y recordatorios

**R50.** El sistema DEBE documentar el playbook completo en
`.rei/docs/harness/integrations/github.md`: configuración y comandos, vocabulario
de marcadores y labels, puerta de aprobación, plantilla del cuerpo de la issue,
comentarios, PR y merge, adopción, tabla de errores y reintentos, y recordatorios
en `current.md`.

**R51.** El playbook DEBE declararse en el índice `.rei/docs/harness/README.md` y
en la lista de archivos requeridos del harness, de modo que `rei check` detecte
su ausencia.

**R52.** El contrato del Leader DEBE incorporar la integración: consulta de la
configuración, puerta única, issue/rama/`links.json`, comentarios idempotentes,
PR/merge bajo orden, adopción, errores no bloqueantes y recordatorios. El
contrato del `initializer` DEBE incorporar la pregunta de configuración.

**R53.** Los agentes nativos de los runtimes soportados DEBEN regenerarse a partir
de los contratos actualizados y verificar que coinciden.

**R54.** MIENTRAS exista una sesión activa, el Leader DEBE listar en «Próximo
paso» de `.rei/progress/current.md` toda acción GitHub pendiente (issue, rama,
comentario, PR o merge) y DEBE retirarla cuando se complete. Cuando el Work Item
pase a `done` y la sesión se archive, DEBE reportar en su respuesta las acciones
que sigan pendientes, sin inventar un estado de sesión.

**R55.** El vocabulario DEBE quedar fijado en v1: marcador de issue
`<!-- rei:work-item:<id> -->`; marcadores de comentario
`<!-- rei:comment:impl:<id> -->` y `<!-- rei:comment:review:<id> -->`; labels
`rei`, `rei:feature` y `rei:blocked`. `rei:epic` y `rei:backlog` quedan
reservadas para v2 y NO DEBEN crearse en v1. El marcador DEBE ser invisible y la
label DEBE ser solo una pista humana: NINGÚN paso del protocolo DEBE depender de
la label.

**R56.** La Feature DEBE incluir tests del comando Go (`rei github` y detección
de remoto) que no requieran red ni un `gh` real, y DEBE mantener `gofmt`,
`make vet`, `make test`, `make build`, `rei check --quiet` y las verificaciones
de adaptadores en verde. NO DEBE introducir dependencias Go nuevas.
