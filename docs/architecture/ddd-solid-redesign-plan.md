# DDD + SOLID Redesign Plan

## Status

Completed on March 1, 2026. This document is retained as historical transition planning context.
For current-state architecture, see `docs/architecture/target-architecture.md`.

## 1. Objective

Redesign the application so game rules are modeled as a domain first, with clear boundaries and minimal coupling between:

- Domain logic
- Application orchestration
- Infrastructure (profile storage, settings, logging)
- Presentation layers (console and graphic UI)

## 2. Problems in Current Design

- `internal/game` contains domain logic but also runtime concerns and persistence-facing structures.
- UI layers interact directly with concrete game state and timing details.
- Shared behavior across console/graphic modes is duplicated.
- Testing focuses mostly on entity behavior, with limited use-case/service tests.

## 3. Design Principles

- DDD:
  - Explicit bounded context for gameplay.
  - Ubiquitous language for concepts like run, level, obstacle, profile, and preset.
  - Aggregates protect invariants.
- SOLID:
  - SRP: separate domain decisions from IO/render/input concerns.
  - OCP: add features by extending use-cases, not changing core entities broadly.
  - LSP: interfaces remain behavior-compatible across adapters.
  - ISP: small interfaces for clock, RNG, storage, input, and renderer.
  - DIP: app services depend on ports/interfaces, not concrete packages.

## 4. Target Layers

- Domain layer (`internal/domain/*`)
  - Pure gameplay model and invariants.
- Application layer (`internal/app/*`)
  - Use cases: start game, apply input, tick, pause/resume, finalize run, save profile.
- Infrastructure layer (`internal/infra/*`)
  - File profile store, system clock, RNG, config loading.
- Interface layer (`internal/ui/*`, `cmd/*`)
  - Console and Ebiten adapters; map input/render to app services.

## 5. Migration Strategy

All phases below were completed during the redesign effort.

## Phase A: Foundation

- Introduce interfaces (ports) for clock, RNG, profile repository.
- Isolate existing `game.State` usage behind application service facade.
- Keep behavior unchanged.

## Phase B: Domain Extraction

- Move core entities/value objects into `internal/domain/gameplay`.
- Convert mutable free-form state to aggregate with explicit methods.
- Keep deterministic rules and invariants in domain only.

## Phase C: Application Services

- Implement `GameSessionService` orchestrating domain and ports.
- Add command/query style methods used by both UI modes.
- Add DTO/view models for HUD and menu screens.

## Phase D: UI Adapter Cleanup

- Refactor console and graphic packages to use service interfaces only.
- Remove direct dependency on domain internals from UI.
- Reuse one shared game loop coordinator where possible.

## Phase E: Hardening

- Expand unit tests for domain and use-case services.
- Add integration tests for profile persistence adapters.
- Add contract tests for UI adapter-service interaction.

## 6. Definition of Done

- Domain has no dependency on UI, file system, Ebiten, or keyboard library.
- App layer has no dependency on concrete storage/input/render implementations.
- Console and graphic modes compile against app service interfaces only.
- Current gameplay behavior remains equivalent (movement, collisions, levels, stats, persistence).
- Coverage of domain + app packages is at least 85%.

## 7. Risks and Mitigations

- Risk: behavior regressions during extraction.
  - Mitigation: snapshot-style regression tests and incremental migration.
- Risk: over-engineering for small project scope.
  - Mitigation: keep interfaces minimal; avoid speculative abstractions.
- Risk: duplicate logic during transition.
  - Mitigation: short-lived compatibility adapter and explicit cutover milestone.
