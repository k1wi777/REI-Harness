# Arquitectura

> **Documento del proyecto** — personaliza este archivo al adaptar REI Harness a tu repositorio.
>
> Este documento define las decisiones arquitectónicas del proyecto.
> El Implementer debe respetarlas durante la implementación y el Reviewer las utilizará como criterio de validación.

<!-- REI:PENDIENTE -->

---

# Objetivo

Describe cómo está organizado el proyecto y las reglas arquitectónicas que deben mantenerse durante su evolución.

No documenta requisitos funcionales ni decisiones específicas de una Feature.

---

# Principios

_Enumera aquí las reglas arquitectónicas que siempre deben cumplirse._

Ejemplos:

- Separación clara de responsabilidades.
- Componentes desacoplados.
- Bajo acoplamiento y alta cohesión.
- Simplicidad antes que complejidad.
- Reutilizar antes que duplicar.

---

# Organización del proyecto

_Describe brevemente cómo se divide el proyecto._

Ejemplo:

```text
src/
    components/
    hooks/
    services/
    utils/
```

Indica únicamente la responsabilidad de cada módulo principal.

---

# Flujo general

_Explica cómo fluye la información dentro del sistema._

Ejemplo:

```text
Usuario
    ↓
Interfaz
    ↓
Lógica de aplicación
    ↓
Servicios
    ↓
Persistencia
```

No es necesario representar todos los detalles, únicamente el flujo principal.

---

# Restricciones arquitectónicas

_Documenta aquellas reglas que nunca deben romperse._

Ejemplos:

- No acceder directamente a la base de datos desde la interfaz.
- No importar módulos entre capas que no deban conocerse.
- No introducir dependencias circulares.
- No duplicar lógica de negocio.
- Mantener los componentes reutilizables.

---

# Decisiones importantes

_Documenta únicamente las decisiones globales del proyecto._

Ejemplos:

- Patrón de arquitectura utilizado.
- Gestión del estado.
- Estrategia de comunicación con APIs.
- Organización del routing.
- Sistema de autenticación.

No documentes aquí decisiones específicas de una Feature; esas pertenecen a `design.md`.

---

# Qué NO documentar

Este archivo NO debe contener:

- requisitos funcionales;
- decisiones temporales;
- planificación de Features;
- detalles de implementación concretos;
- cambios específicos de un Work Item.
