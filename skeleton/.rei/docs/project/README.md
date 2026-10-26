# Documentación del proyecto para REI Harness

Estos archivos describen **las reglas específicas de tu repositorio** — arquitectura, convenciones y verificación.

**Debes personalizarlos** al implementar REI Harness por primera vez. Los agentes los consultan en cada implementación y revisión.

| Archivo | Propósito | Quién lo lee por defecto |
|---------|-----------|--------------------------|
| `architecture.md` | Decisiones arquitectónicas del proyecto. | Implementer (siempre); Spec Author (al diseñar); Reviewer (solo si aplica) |
| `conventions.md` | Estilo de código y convenciones de desarrollo. | Implementer (siempre); Reviewer (solo si aplica) |
| `verification.md` | Checkpoints y evidencia para validar el trabajo. | Reviewer (siempre); Implementer (solo el mapeo de checkpoints) |

También personaliza las secciones correspondientes de `AGENTS.md` (propósito, stack, comandos) y los checks en `.rei/config.json` (verificación).
