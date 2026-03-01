# ADR 0001: Adopt DDD Boundaries and SOLID Layering

## Status

Accepted

## Context

The project evolved quickly and combines domain, orchestration, and IO concerns in a way that makes extension and testing harder.

## Decision

Adopt a layered architecture aligned with DDD and SOLID:

- Domain as a pure gameplay model with protected invariants.
- Application services orchestrating use cases through interfaces.
- Infrastructure adapters implementing persistence/time/random ports.
- UI adapters (console/graphic) consuming application services only.

## Consequences

Positive:

- Better testability and deterministic behavior validation.
- Cleaner separation between gameplay rules and platform-specific code.
- Safer feature development (new modes, AI, networking) on stable boundaries.

Trade-offs:

- Temporary complexity during migration.
- Additional interface and mapping code.

## Follow-up

- Execute `docs/architecture/migration-backlog.md`.
- Reassess package boundaries after Epic 3 completion.
