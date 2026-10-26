# Requisitos — P2: UX del CLI, validación offline y documentación

> Feature de P2. Cubre los puntos 1–8 y 11–13 del acuerdo; 9 y 10 están fuera de
> alcance, igual que P3, release y push. Sin dependencias nuevas y sin tocar
> `skeleton/`.

## Enlaces locales (`links.json`)

### R1 — Ficha con enlaces válidos

CUANDO se ejecuta `rei item show <id>` y `.rei/specs/<id>/links.json` existe y es
válido, el sistema DEBE mostrar el número de issue, la URL, la base (`branch` y
`sha`) y la épica (`slug`, issue y URL) registrados.

### R2 — Enlaces ausentes o inválidos en la ficha

SI `links.json` no existe, no se puede leer o no cumple la forma esperada,
ENTONCES `rei item show` DEBE señalarlo como una nota informativa, DEBE seguir
mostrando el resto de la ficha y NO DEBE alterar su código de salida por ese
motivo.

### R3 — Tipo y enlaces en el listado

CUANDO se ejecuta `rei items status`, el sistema DEBE mostrar por cada Work Item
su tipo y, si `links.json` es válido, la issue y la base (`branch@sha`) cuando
exista; un `links.json` inválido NO DEBE impedir el listado y DEBE indicarse como
nota.

### R4 — Invariante de un único `in_progress`

CUANDO exista más de un Work Item en `in_progress`, `rei items status` DEBE salir
con código 1 y nombrar los identificadores implicados; con cero o un Work Item
activo DEBE salir con código 0.

## Configuración de GitHub

### R5 — Mensaje correcto de `github status`

CUANDO `rei github status` no detecta acciones de configuración pendientes, el
sistema DEBE indicar que la configuración está completa y que no hay acciones de
configuración pendientes, sin afirmar que no existen issues, PRs u otras acciones
de GitHub.

### R6 — Flags de `github enable`

El comando `rei github enable` DEBE aceptar opcionalmente `--repo owner/name`,
`--base <rama>` y `--strategy squash|merge|rebase`, y DEBE seguir funcionando sin
ningún flag.

### R7 — Validación de flags de `github enable`

SI un flag de `rei github enable` es desconocido, le falta valor o su valor no
cumple el formato (`owner/name`, rama no vacía o estrategia permitida), ENTONCES
el sistema DEBE rechazar la orden con código 2, indicar el uso correcto y no
escribir `.rei/config.json`; `disable` y `status` DEBEN rechazar flags con código
2.

### R8 — Precedencia de valores

CUANDO `rei github enable` resuelve `repo`, `base` y `mergeStrategy`, el sistema
DEBE aplicar la precedencia flag explícito > valor previo de
`integrations.github` > detección desde git > valor por defecto. `repo` no tiene
valor por defecto y su ausencia DEBE producir un error con siguiente paso; `base`
usa `main` y avisa cuando la detección falla; `mergeStrategy` usa `squash`.

### R9 — Preservación y escritura atómica de la configuración

CUANDO `rei github enable` escribe la configuración, DEBE conservar `checks`,
`viewer`, otras integraciones y cualquier clave ajena. SI `.rei/config.json` es
inválido o no se puede escribir, ENTONCES DEBE salir con código 1 y dejar el
archivo intacto.

## Documentación

### R10 — Ayuda por comando

CUANDO se ejecuta `rei help <comando>` para `item`, `items`, `session`, `status`
o `github`, el sistema DEBE describir, según corresponda, la salida de enlaces,
la validación de `session start`, el carácter terminal de `done` y los flags de
`github enable`.

### R11 — README y guía de uso

La documentación de usuario (`README.md` y `.rei/docs/usage.md`) DEBE listar
`rei new`, `rei session`, `rei commit` y `rei github` en la tabla de comandos,
describir los flags de `enable`, los estados resultantes y que `session start`
solo inicia Work Items en `pending`.

### R12 — Documentación de la integración GitHub

La guía de GitHub (`github.md`, `core.md` y `errors.md`) DEBE documentar los
flags de `enable` con su precedencia, la validación de `links.json` como aviso
recuperable y la recuperación ante un índice inválido.

### R13 — Repositorio canónico

El sistema DEBE usar `k1wi777/REI-Harness` como repositorio canónico en la
auto-actualización y en los enlaces de instalación de `README.md` y
`.rei/docs/usage.md`, con una prueba que detecte una regresión al nombre
anterior.

## Transiciones y sesión

### R14 — `done → blocked` no estándar

SI se intenta la transición `done → blocked` sin `--force`, ENTONCES el sistema
DEBE rechazarla con causa y siguiente paso; con `--force` DEBE permitirla como
transición no estándar. `done` DEBE quedar documentado como estado terminal en
`workflow.md`, `meta.md`, `help.go` y `leader.md`.

### R15 — Task sin issue ni rama

La documentación de GitHub y el contrato del Leader DEBEN establecer que una
Task no crea issue ni rama, que no existe opt-in para Tasks, que la puerta
issue+rama aplica solo a Features y que no se usa una label `rei:task`.

### R16 — Validación de `links.json`

CUANDO `rei validate` encuentra `links.json`, DEBE validar `provider`, `repo`,
`issue`, `url`, `base.branch`/`base.sha` si existen y `epic` como objeto o
`null`. SI el archivo es inválido, ENTONCES DEBE emitir `WARN`, nunca `FAIL`, y
continuar con el resto de la validación.

### R17 — Validación de `session start`

CUANDO se ejecuta `rei session start <id> <type>`, el sistema DEBE comprobar que
el Work Item existe, que el tipo coincide con `meta.json` y que su estado es
`pending`. SI no se cumple, ENTONCES DEBE rechazar la orden con causa y siguiente
paso, sin escribir `current.md`; solo `pending` inicia una sesión nueva.

## Limpieza

### R18 — `package-lock.json` huérfano

El repositorio NO DEBE contener el `package-lock.json` huérfano ni referencias a
él en la build, la configuración o la documentación vigente.
