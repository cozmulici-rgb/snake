# Migration Backlog

## Status Legend

- `[ ]` not started
- `[~]` in progress
- `[x]` complete

## Epic 1: Architecture Baseline

- [x] Create `internal/app/session` service facade around current state.
- [x] Add port interfaces for profile repository, clock, and RNG.
- [x] Wire console and graphic modes through facade without behavior changes.

Acceptance criteria:
- Both UIs run from the facade.
- Existing gameplay tests continue to pass.

## Epic 2: Domain Refactor

- [x] Introduce `internal/domain/gameplay` aggregate and value objects.
- [x] Move collision, scoring, level, and win/lose rules into domain package.
- [x] Keep persistence representation outside domain entities.

Acceptance criteria:
- Domain package has no infra/UI imports.
- Equivalent behavior verified with regression tests.

## Epic 3: Application Use Cases

- [x] Implement command/query API on session service.
- [x] Add read models for HUD/menu/game-over screen.
- [x] Implement profile sync use case with repository port.

Acceptance criteria:
- UI reads only from application read models.
- No direct `game.State` mutation in UI packages.

## Epic 4: Infrastructure Adapters

- [x] Move file profile logic to `internal/infra/profile`.
- [x] Add adapter tests for load/save and migration safety.
- [x] Add `clock` and `rng` adapters in `internal/infra/system`.

Acceptance criteria:
- App layer depends only on interfaces.
- Infrastructure tests pass on Windows CI.

## Epic 5: Cleanup and Hardening

- [x] Remove obsolete compatibility paths.
- [x] Raise domain+app coverage target to at least 85%.
- [x] Document final architecture and update README developer notes.
- [x] Add automated dependency-direction checks for domain/app/ui boundaries.

Acceptance criteria:
- No duplicate gameplay rule implementation remains.
- CI verifies new package coverage and build integrity.

## Suggested Delivery Sequence

1. Epic 1 (completed)
2. Epic 2 (completed)
3. Epic 3 (completed)
4. Epic 4 (completed)
5. Epic 5 (completed)
