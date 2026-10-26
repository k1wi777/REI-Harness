# Requirements — GitHub Issues v2: épicas, backlog y asignación

> Work Item: `2026-10-25_19-33__github-epicas-backlog-v2` (feature).
>
> Los requisitos describen **qué** debe hacer el sistema, no **cómo**. Se
> redactan en EARS. El sujeto «el sistema» designa al conjunto formado por el
> CLI `rei`, el rol `Leader` y el playbook de la integración, según el requisito.
>
> Principio rector: GitHub es la **proyección** del trabajo, no la fuente de
> verdad de `.rei`. La issue manda sobre el problema, el alcance y la discusión;
> `.rei` manda sobre la ejecución del Work Item. La integración es opcional y su
> ausencia debe dejar el flujo actual intacto.
>
> v2 construye sobre la v1 ya implementada (issue por feature, puerta única,
> adopción genérica, comentarios, PR/merge bajo orden, nombre canónico e
> idempotencia por marcador).

---

# 1. Alcance y principios

**R1.** La integración con GitHub Issues DEBE ser opcional y estar desactivada
por defecto. MIENTRAS `integrations.github` no exista o `enabled` sea `false`,
el sistema NO DEBE mencionar GitHub, NO DEBE ejecutar `gh`, NO DEBE leer ni
modificar issues y el flujo DEBE ser idéntico al que no usa GitHub.

**R2.** v2 DEBE construir sobre la v1 sin alterar su comportamiento: issue por
feature, puerta única al aprobar, adopción genérica, comentarios de
implementación y revisión, PR y merge bajo orden, nombre canónico del
repositorio e idempotencia por marcador.

**R3.** El CLI NO DEBE crear, leer ni modificar issues, labels, assignees,
comentarios, PR ni épicas, y NO DEBE añadir subcomandos, flags ni dependencias
Go nuevas. Las acciones sobre GitHub DEBEN ejecutarse con `gh` y `git` desde el
playbook del agente, no desde el CLI.

**R4.** Un fallo de GitHub NO DEBE alterar `meta.json`, el estado de
`.rei/progress/current.md` ni la etapa del workflow.

**R5.** GitHub DEBE ser la fuente de verdad de las issues, sus estados, labels,
assignees y épicas. REI NO DEBE modelar personas ni DEBE crear Work Items
`pending` locales como sustituto del backlog.

---

# 2. Modelo de épica y de hija

**R6.** Una épica DEBE ser una issue de GitHub, NO un Work Item. DEBE llevar el
marcador `<!-- rei:epic:<epic-slug> -->` en la primera línea del cuerpo y las
labels `rei` y `rei:epic`. NO DEBE tener `requirements.md`, `design.md`,
`tasks.md` ni requisitos EARS.

**R7.** El cuerpo de la épica DEBE incluir: contexto del módulo, objetivo,
no-objetivos, task list de hijas, orden y dependencias, y criterio de éxito. El
plan completo DEBE ser opcional y, si existe, DEBE plegarse en un bloque
`<details>`.

**R8.** El slug de la épica DEBE ser kebab-case, estable y único en el
repositorio. Cambiar el título de la issue NO DEBE cambiar el slug ni el
marcador.

**R9.** Una hija DEBE ser una issue completa de GitHub, NO un Work Item hasta
que se adopte. DEBE llevar el marcador
`<!-- rei:backlog:<epic-slug>/<item-slug> -->` en la primera línea del cuerpo y
las labels `rei` y `rei:backlog`.

**R10.** El cuerpo de la hija DEBE incluir: contexto, alcance, fuera de alcance,
criterios de aceptación en borrador, referencia a la épica, assignee y
notas/dependencias. NO DEBE incluir requisitos EARS, `design.md`, `tasks.md` ni
plan de archivos.

**R11.** El slug de la hija DEBE ser kebab-case, estable y único dentro de la
épica. La identidad de la hija DEBE ser su marcador, no su título.

**R12.** Las labels de la integración DEBEN ser exactamente `rei`, `rei:epic`,
`rei:backlog`, `rei:feature` y `rei:blocked`. El marcador DEBE ser invisible y
la label DEBE ser solo una pista humana: ningún paso DEBE depender de la label
para identificar una issue. La ausencia de marcador NO DEBE suplirse con
labels; si hay duda, el sistema DEBE preguntar.

**R13.** `links.json` DEBE contener `provider`, `repo`, `issue`, `url` y `epic`.
Cuando la issue pertenezca a una épica, `epic` DEBE ser un objeto con `slug`,
`issue` y `url`; cuando no, DEBE ser `null`. `links.json` DEBE ser un índice
local recuperable, nunca la fuente de verdad.

---

# 3. Creación del backlog

**R14.** CUANDO el usuario pida crear o planificar el backlog de un módulo y la
integración esté activa, el sistema DEBE iniciar una entrevista y NO DEBE crear
ninguna issue hasta completarla. La entrevista DEBE cubrir alcance,
no-objetivos, features hijas, orden/dependencias y responsables.

**R15.** Durante la entrevista, el sistema DEBE preguntar solo la información
que falte y NO DEBE inventar alcance, hijas, responsables ni criterios.

**R16.** Antes de crear nada, el sistema DEBE mostrar el borrador completo de la
épica y de cada hija, incluyendo título, labels, marcador, cuerpo y orden, y
DEBE pedir una aprobación explícita.

**R17.** SI el usuario no aprueba o pide cambios, ENTONCES el sistema NO DEBE
crear issues, DEBE actualizar el borrador y DEBE repetir la aprobación. NUNCA
DEBE interpretar el silencio o una respuesta ambigua como aprobación.

**R18.** CUANDO el usuario apruebe, el sistema DEBE crear las issues de forma
idempotente por marcador sobre el nombre canónico del repositorio, incluyendo
issues cerradas y descartando pull requests. SI encuentra más de una issue con
el mismo marcador, DEBE detenerse y reportar la anomalía sin crear ni
sobrescribir nada.

**R19.** SI no se puede resolver el nombre canónico, listar las issues o buscar
el marcador, ENTONCES el sistema DEBE detenerse con causa y siguiente paso y NO
DEBE crear nada usando un nombre sin resolver.

**R20.** Antes de crear issues, el sistema DEBE asegurar de forma idempotente las
labels `rei`, `rei:epic` y `rei:backlog`; al adoptar, DEBE asegurar además
`rei:feature`.

**R21.** La creación DEBE seguir el orden épica → hijas → task list final. Cada
issue creada o reutilizada DEBE verificarse (número, URL, estado, cuerpo y
labels) antes de continuar. La task list de la épica DEBE escribirse con los
números reales de las hijas.

**R22.** SI falla la creación de una hija después de crear la épica o algunas
hijas, ENTONCES el sistema DEBE informar de lo creado, lo reutilizado y lo
pendiente, DEBE conservar lo creado y DEBE permitir reintentar sin duplicar.
En el reintento DEBE reutilizar por marcador todo lo existente y crear solo lo
que falte.

**R23.** CUANDO el usuario pida añadir una hija a una épica existente, el
sistema DEBE aplicar el mismo flujo de borrador, aprobación e idempotencia y
DEBE actualizar la task list. SI la épica está cerrada, ENTONCES NO DEBE añadir
la hija y DEBE informar de la causa y del siguiente paso.

**R24.** SI las Issues están deshabilitadas en el repositorio, ENTONCES el
sistema DEBE informar de la causa y del siguiente paso (activarlas o configurar
otro repositorio) y NO DEBE afirmar éxito.

**R25.** MIENTRAS la integración esté desactivada, el sistema NO DEBE leer ni
crear issues y DEBE ofrecer `rei github enable` como siguiente paso. NO DEBE
crear Work Items locales como sustituto del backlog.

---

# 4. Adopción y promoción de una hija

**R26.** CUANDO el usuario pida «trabajemos en la #N» y la integración esté
activa, el sistema DEBE leer la issue, sus comentarios, labels, assignees y
marcadores antes de crear un Work Item.

**R27.** SI la issue es una épica, ENTONCES el sistema NO DEBE adoptarla y DEBE
explicar que las épicas no son Work Items y ofrecer sus hijas.

**R28.** SI la issue o la épica a la que pertenece están cerradas, ENTONCES el
sistema DEBE preguntar antes de continuar.

**R29.** SI la issue ya tiene un marcador `rei:work-item:<id>`, ENTONCES el
sistema NO DEBE crear otro Work Item ni sobrescribir el marcador. SI el
marcador corresponde a un Work Item existente en este clon, DEBE reutilizarlo
e informar; SI corresponde a otro clon u otro Work Item, DEBE detenerse e
informar del conflicto.

**R30.** SI la issue tiene un marcador de backlog, ENTONCES la adopción DEBE
promoverla. SI no tiene marcador, la adopción genérica de la v1 DEBE seguir
funcionando sin cambios.

**R31.** La adopción DEBE preguntar solo la información que falte para definir
el Work Item (objetivo, alcance, fuera de alcance, criterios de aceptación y
restricciones), usando el cuerpo y los comentarios como contexto. NO DEBE
inventar información.

**R32.** CUANDO se adopte una hija, el sistema DEBE crear el Work Item con
`rei new` como `feature` y rellenar `title` y `description` con el acuerdo
alcanzado, la referencia a la épica y la URL de la issue.

**R33.** CUANDO se adopte una hija, el sistema DEBE sustituir el marcador de
backlog por `rei:work-item:<id>` en el cuerpo, preservando el resto del
contenido, y DEBE verificar que el marcador de backlog ya no existe.

**R34.** CUANDO se adopte una hija, el sistema DEBE añadir la label
`rei:feature`, retirar `rei:backlog`, conservar `rei` y verificar las labels
resultantes.

**R35.** CUANDO se adopte una hija de backlog, el sistema DEBE resolver la
épica por su marcador antes de crear el Work Item. SI la épica no existe o su
marcador es ambiguo, ENTONCES NO DEBE iniciar la adopción y DEBE informar del
conflicto. CUANDO la adopción se complete, `links.json` DEBE contener
`provider`, `repo`, `issue`, `url` y el objeto `epic` resuelto desde el
marcador de la épica.

**R36.** SI `links.json` falta, está corrupto o se desincroniza, ENTONCES el
sistema NO DEBE bloquear el Work Item y DEBE recuperar la issue por su marcador
y reescribir el archivo.

**R37.** SI la hija está sin asignar al adoptarla, ENTONCES el sistema DEBE
ofrecer asignarla a quien la trabaja y NO DEBE asignarla sin confirmación.

**R38.** SI la hija está asignada a otra persona, ENTONCES el sistema DEBE
preguntar antes de adoptarla o de reasignarla.

**R39.** CUANDO `spec_author` termine la planificación de un Work Item adoptado
de backlog, el sistema DEBE publicar en el cuerpo de la issue un resumen de los
requisitos y el `design.md` completo plegado en `<details>`, preservando el
contenido previo de la issue.

**R40.** La publicación de la spec DEBE ser idempotente: si ya existe la sección
del Work Item, DEBE reemplazarla en sitio y NO DEBE duplicarla.

**R41.** SI falla la publicación de la spec, ENTONCES el sistema DEBE informar
de la causa y del siguiente paso, NO DEBE alterar el estado del Work Item y DEBE
permitir reintentar sin duplicar la sección.

---

# 5. Asignación de personas

**R42.** El sistema SOLO DEBE asignar o desasignar bajo una orden explícita del
usuario o la confirmación de una oferta. NUNCA DEBE autoasignar por carga,
por adopción ni por creación de backlog.

**R43.** CUANDO se ordene asignar, el sistema DEBE resolver `@me` al login
autenticado y DEBE validar que cada login existe antes de modificar la issue.
SI un login no existe o no es asignable, ENTONCES NO DEBE afirmar éxito y DEBE
informar de la causa y del siguiente paso.

**R44.** La asignación DEBE realizarse con `gh issue edit --add-assignee` y la
desasignación con `gh issue edit --remove-assignee`. El lote DEBE ser opcional y
solo DEBE usarse cuando el mismo conjunto de personas se aplique a varias
issues.

**R45.** SI una asignación o desasignación falla total o parcialmente, ENTONCES
el sistema DEBE informar de las issues afectadas, de las que quedaron sin
cambiar y del siguiente paso, sin ocultar el fallo.

**R46.** Después de asignar o desasignar, el sistema DEBE verificar los
assignees reales de cada issue y solo entonces informar de éxito. SI la issue ya
estaba en el estado pedido, DEBE informarlo como «sin cambios».

---

# 6. Progreso y cierre de la épica

**R47.** La task list de la épica DEBE usar el formato `- [ ] #N <título>` y DEBE
ser la lista de hijas de la épica. Las casillas son informativas y NO DEBEN ser
la fuente de verdad del estado de una hija.

**R48.** El sistema NO DEBE sincronizar en vivo el estado de las hijas con la
épica. Solo DEBE actualizar la task list al crear o reanudar un backlog
parcial, al añadir una hija, o bajo una orden explícita de reconciliación o
cierre.

**R49.** CUANDO el usuario ordene cerrar una épica, el sistema DEBE comprobar
que la issue tiene marcador de épica, leer su task list, consultar el estado
real de cada hija y cerrar solo si todas están cerradas.

**R50.** SI alguna hija está abierta, no existe, es un pull request, o la task
list está vacía o no se puede interpretar, ENTONCES el sistema NO DEBE cerrar
la épica y DEBE informar de las hijas pendientes o del problema.

**R51.** Como parte del cierre, el sistema DEBE reconciliar la task list con el
estado real de las hijas y, si no puede escribirla, NO DEBE cerrar la épica.

**R52.** La épica NUNCA DEBE cerrarse automáticamente al cerrarse la última
hija. El cierre SOLO DEBE ocurrir bajo orden explícita y DEBE verificar el
estado `CLOSED` después. El sistema NUNCA DEBE cerrar hijas automáticamente;
su cierre sigue su propio flujo.

---

# 7. Errores, aislamiento y reintentos

**R53.** Todo fallo de `gh`, `git` o configuración DEBE comunicarse con causa y
siguiente paso. NUNCA DEBE silenciarse ni afirmarse un éxito sin verificarlo.

**R54.** El sistema DEBE cubrir como mínimo: integración no configurada,
desactivada o incompleta; `gh` ausente o sin autenticar; nombre canónico no
resuelto; Issues deshabilitadas; épica cerrada; épica o hija inexistente; hija
ya adoptada; marcador duplicado; assignee inexistente o no asignable;
permisos insuficientes; creación parcial; task list vacía o ilegible; hijas
abiertas al cerrar; fallo al publicar la spec; red y rate limit.

**R55.** SI se alcanza un rate limit o falla la red, ENTONCES el sistema DEBE
informar de la causa y del momento de reintento si está disponible, y NO DEBE
reintentar en bucle.

**R56.** Un fallo de GitHub NO DEBE cambiar `meta.json`, el estado de
`current.md` ni la etapa del workflow. El sistema DEBE informar de lo hecho y lo
pendiente y DEBE permitir reintentar sin duplicar Work Items, issues,
marcadores, labels ni secciones de spec.

---

# 8. Documentación, contratos y verificación

**R57.** El playbook `.rei/docs/harness/integrations/github.md` DEBE documentar
v2 completo: modelo de épica e hija, vocabulario, `links.json`, creación con
entrevista y aprobación, idempotencia canónica, creación parcial, adopción y
promoción, publicación de la spec, asignación, conflictos, progreso y cierre,
tabla de errores, recordatorios y plantillas exactas.

**R58.** El contrato del Leader DEBE incorporar los disparadores de v2 (crear
backlog, adoptar una hija, asignar o desasignar, cerrar una épica), las reglas
no negociables de conflicto y ausencia de automatismos, y la autorización de la
prueba de campo cuando la planificación la incluya, remitiendo al playbook para
los comandos exactos y sin añadir herramientas nuevas.

**R59.** El playbook DEBE seguir declarado en el índice
`.rei/docs/harness/README.md` y en `check.RequiredFiles`; la fila del índice DEBE
reflejar v2.

**R60.** Los agentes nativos de los runtimes soportados DEBEN regenerarse desde
los contratos actualizados y las verificaciones `--check` DEBEN pasar.

**R61.** La Feature NO DEBE introducir dependencias Go nuevas ni cambios de
comportamiento del CLI más allá del contenido embebido actualizado. `gofmt`,
`make vet`, `make test`, `make build`, `rei check --quiet` y las verificaciones
de adaptadores DEBEN quedar en verde.

**R62.** La Feature DEBE incluir una prueba de campo en este repositorio con una
épica de prueba y dos hijas: creación idempotente, marcadores, labels, task
list, asignación y desasignación, reintento sin duplicar y cierre bajo orden.
También DEBE probar la adopción de una hija en un worktree temporal y la
publicación idempotente de su spec.

**R63.** La prueba de campo DEBE dejar evidencia en
`.rei/progress/work-items/2026-10-25_19-33__github-epicas-backlog-v2/impl.md`
con números, URLs, comandos y resultados. Las issues de prueba DEBEN quedar
cerradas al final o documentarse como evidencia permanente con su estado.

**R64.** La prueba de campo DEBE ejecutarse solo con autorización explícita del
usuario registrada por el Leader; NUNCA DEBE usarse para justificar escrituras
no aprobadas. Las acciones normales de backlog y adopción siguen siendo
responsabilidad del Leader.
