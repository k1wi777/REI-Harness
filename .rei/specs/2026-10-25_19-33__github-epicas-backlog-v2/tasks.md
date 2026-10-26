# Tasks — GitHub Issues v2: épicas, backlog y asignación

> Work Item: `2026-10-25_19-33__github-epicas-backlog-v2` (feature).
>
> Orden de ejecución. Cada tarea indica los requisitos que implementa.
> El Implementer marca `[x]` **inmediatamente** al terminar cada tarea.
>
> No hay código Go nuevo: la implementación es el playbook y el contrato del
> Leader. La verificación funcional es la prueba de campo (T12).

---

## Fase 1 — Playbook v2

- [x] **T1 — Actualizar cabecera, alcance y vocabulario del playbook.**
  En `.rei/docs/harness/integrations/github.md`: título v2; alcance v1+v2;
  principio «GitHub es una proyección»; marcadores `rei:epic`,
  `rei:backlog`, delimitadores `rei:children` y `rei:spec:<id>`; tabla de
  labels con colores y descripciones; `links.json` con el objeto `epic` y
  `null` sin épica; nombre canónico; regla «el marcador identifica, la label es
  una pista»; prohibición de duplicados por marcador.
  Mantener las secciones v1 y su contenido.
  (R1–R13, R57)

- [x] **T2 — Documentar el modelo y la creación de épicas.**
  Secciones: plantilla de épica, slug estable, labels, cuerpo con contexto,
  objetivo, no-objetivos, task list delimitada, orden/dependencias y criterio
  de éxito; `<details>` opcional. Flujo de creación: entrevista, borrador
  completo, aprobación explícita, idempotencia REST canónica (issues cerradas,
  sin PRs, paginada), detección de marcadores duplicados, creación de labels,
  orden épica → hijas → task list, verificación por issue, creación parcial y
  reintento sin duplicar, épica cerrada, Issues deshabilitadas e integración
  desactivada.
  (R6–R8, R18–R25, R57)

- [x] **T3 — Documentar el modelo y la creación del backlog de hijas.**
  Secciones: plantilla de hija, slug único dentro de la épica, cuerpo con
  contexto, alcance, fuera de alcance, criterios en borrador, épica, assignee y
  notas/dependencias; sin EARS/design/tasks. Entrevista (alcance, no-objetivos,
  features, orden, responsables), borrador completo, aprobación, creación
  idempotente, creación parcial, reintento, añadir hija a una épica abierta y
  rechazo si la épica está cerrada.
  (R9–R11, R14–R25, R57)

- [x] **T4 — Documentar adopción, promoción y publicación de la spec.**
  Secciones: lectura de issue, comentarios, labels, assignees y marcadores;
  rechazo de épicas; issue cerrada; conflicto por marcador de Work Item;
  promoción de hija de backlog; preguntas solo por lo que falte; `rei new`
  feature y `description` con la épica; sustitución del marcador preservando el
  cuerpo; labels `rei`/`rei:feature` sin `rei:backlog`; `links.json` con objeto
  `epic` y recuperación; oferta de asignación si está sin asignar; pregunta si
  está asignada a otra persona; publicación de la spec tras `spec_author`
  (resumen + design plegado) y actualización idempotente por delimitadores.
  Mantener la adopción genérica v1 cuando no hay marcador.
  (R26–R41, R57)

- [x] **T5 — Documentar asignación y desasignación.**
  Comandos `gh issue edit --add-assignee`/`--remove-assignee`, resolución de
  `@me`, validación de login, lote opcional con el mismo conjunto de personas,
  verificación de assignees, informe de «sin cambios», errores de login
  inexistente/no asignable/permisos y fallo parcial. Dejar claro que GitHub es
  la fuente de verdad y que no hay autoasignación ni estado local de personas.
  (R42–R46, R57)

- [x] **T6 — Documentar progreso y cierre de la épica.**
  Secciones: formato `- [ ] #N <título>`; la task list como índice; casillas
  informativas; sin sincronización en vivo; actualización solo al crear,
  reanudar, añadir hija o bajo orden explícita. Cierre: solo épica, solo bajo
  orden, lectura de la task list, consulta real del estado de cada hija,
  rechazo con hijas abiertas/inexistentes/PR/task list vacía, reconciliación a
  `[x]`, `gh issue close`, verificación `CLOSED`, prohibición de autocierre y
  de cerrar hijas desde este protocolo.
  (R47–R52, R57)

- [x] **T7 — Ampliar errores, aislamiento y recordatorios.**
  Añadir a la tabla de errores: épica cerrada, hija ya adoptada, épica/hija
  inexistente, marcador duplicado, assignee inexistente/no asignable/permisos,
  task list vacía o ilegible, hijas abiertas al cerrar, creación parcial y
  fallo al publicar la spec. Reforzar: causa + siguiente paso, verificación
  antes de afirmar éxito, aislamiento de `meta.json`/`current.md`, reintento
  idempotente, rate limit/red sin bucles. Añadir recordatorios v2 en
  «Próximo paso» (backlog, hijas pendientes, asignación, spec, cierre).
  (R53–R57)

## Fase 2 — Contratos e índice

- [x] **T8 — Actualizar el contrato del Leader.**
  En `.rei/agents/leader.md`, dentro de «Integración con GitHub (opcional)»:
  disparadores de backlog, adopción de hija, asignación/desasignación y cierre
  de épica; entrevista + borrador + aprobación; idempotencia canónica y
  creación parcial; conflicto por épica o Work Item ya adoptado; oferta de
  asignación; publicación de la spec tras `spec_author`; cierre solo bajo orden
  con verificación de hijas; prohibición de autocierre/autoasignación. Añadir la
  regla de prueba de campo: el Leader registra la autorización explícita del
  usuario y permite que el Implementer ejecute la prueba cuando la planificación
  la incluya; las operaciones normales siguen siendo del Leader.
  Remitir al playbook para los comandos exactos y no añadir herramientas.
  Mantener intactas las reglas v1.
  (R58)

- [x] **T9 — Actualizar el índice del harness.**
  Cambiar la fila de `.rei/docs/harness/integrations/github.md` en
  `.rei/docs/harness/README.md` a v2. Verificar que el playbook sigue en
  `check.RequiredFiles` y en `embed_test.go`; no hace falta añadir archivos.
  (R57, R59)

- [x] **T10 — Regenerar agentes nativos y manifiesto.**
  Ejecutar `make build`, después `./bin/rei init opencode` y
  `./bin/rei init claude`. Verificar con `./bin/rei init opencode --check` y
  `./bin/rei init claude --check`. Revisar que solo cambian los artefactos
  esperados (`.opencode/agents/*`, `.claude/agents/*`,
  `.rei/install-manifest.json`).
  (R60)

## Fase 3 — Verificación

- [x] **T11 — Verificación local.**
  Ejecutar y dejar en verde:
  - `gofmt -l internal cmd` (sin salida);
  - `make vet`;
  - `make test`;
  - `make build`;
  - `rei check --quiet`;
  - `./bin/rei init opencode --check`;
  - `./bin/rei init claude --check`;
  - `rei validate 2026-10-25_19-33__github-epicas-backlog-v2`.
  No añadir tests Go: no hay cambio de comportamiento del CLI.
  (R61)

- [x] **T12 — Prueba de campo.**
  Con la autorización explícita del usuario registrada por el Leader en
  `current.md`, ejecutar el protocolo de `design.md` §10.2: crear la épica de
  prueba `[prueba v2]` y dos hijas; verificar marcadores, labels, task list y
  URLs; repetir la creación y comprobar que no hay duplicados; asignar y
  desasignar (`@me`), probar un login inexistente y el lote; intentar cerrar la
  épica con una hija abierta; adoptar una hija en un worktree temporal y
  publicar/actualizar una spec de prueba; cerrar las hijas y la épica bajo
  orden; verificar `CLOSED` y el rechazo de añadir una hija a una épica
  cerrada. Registrar números, URLs, comandos y resultados en
  `.rei/progress/work-items/2026-10-25_19-33__github-epicas-backlog-v2/impl.md`.
  Las issues de prueba quedan cerradas o documentadas con su estado.
  (R62–R64)
