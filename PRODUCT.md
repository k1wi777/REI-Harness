# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Desarrolladores que trabajan con agentes de IA bajo Spec Driven Development
(SDD). En la práctica, la persona que revisa y aprueba la planificación de un
Work Item (specs) antes de que se implemente, muchas veces alternando varios
runtimes (OpenCode, Claude Code, Cursor, Codex).

## Product Purpose

REI Harness impone un workflow SDD con roles explícitos y una **puerta de
aprobación humana obligatoria**. Existe para que un agente no implemente a
ciegas: primero produce una planificación revisable, el humano la aprueba, y
solo entonces se implementa. El éxito es que revisar y aprobar esa planificación
sea de baja fricción y sin ambigüedad.

## Positioning

Un arnés SDD que separa el **contrato de rol** (canónico, en
`.rei/agents/<rol>.md`) del **runtime**, y expone toda la coordinación por un
**CLI propio** (`rei`) que funciona igual en cualquier harness. La combinación
—contratos de rol runtime-agnósticos + CLI + puerta de aprobación humana
obligatoria registrada en disco— no puede copiarla un framework de agentes
acoplado a un runtime sin renunciar a ese acoplamiento.

## Operating Context

- El trabajo vive en Work Items bajo `.rei/specs/<id>/` (planificación) y
  `.rei/progress/` (estado y reportes).
- Roles: `leader`, `spec_author`, `implementer`, `reviewer`, `initializer`.
- Documentos en Markdown: `requirements.md`, `design.md`, `tasks.md`,
  `plan.md`, `impl.md`, `review.md`.
- El CLI `rei` es la interfaz común: sesión, estados, validación, adaptadores.
- La revisión humana ocurre antes de implementar: el Work Item queda en `ready`
  esperando aprobación.

## Capabilities and Constraints

- `rei view` sirve un Work Item en el navegador: metadatos, documentos de
  planificación y progreso, validación y sesión; render de Markdown y recarga
  en vivo por SSE.
- **Restricción confirmada:** sin JavaScript ni CSS de terceros; solo assets
  propios embebidos.
- **Restricción confirmada:** runtime-agnóstico — nada atado a hooks de un
  runtime concreto como mecanismo central.
- **Restricción confirmada:** la interfaz habla en español, como el resto de la
  documentación de REI.
- Base técnica: Go (`net/http`, `html/template`, `embed`); render de Markdown
  con `goldmark`.
- Decisión abierta: la dirección visual del visor (esta sesión). Aún no hay
  tokens ni world definidos.

## Brand Commitments

- Nombre: **REI Harness**.
- Voz y documentación en español, directas y sin relleno.

## Evidence on Hand

- Work Items reales bajo `.rei/specs/**` con contenido representativo, p. ej.
  `2026-10-09_23-45__spec-browser-viewer/` (requirements, design, tasks, impl,
  review).
- No hay testimonios, clientes, métricas ni assets de marca. El trabajo futuro
  no debe inventarlos.

## Product Principles

1. La aprobación humana es obligatoria y explícita; nunca implícita.
2. La documentación es la fuente de verdad; el chat no lo es.
3. Runtime-agnóstico: el valor vive en el contrato de rol y el CLI, no en el
   runtime.
4. Código propio y dependencias mínimas por delante de librerías de terceros.
5. Baja fricción de revisión: el material tiene que poder revisarse de verdad,
   no solo existir.

## Accessibility & Inclusion

- Sin estándar formal confirmado. Requisito práctico: buena legibilidad en
  lecturas largas de documentación.
