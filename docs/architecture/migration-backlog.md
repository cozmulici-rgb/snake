# Migration Backlog

## Status Legend

- `[ ]` not started
- `[~]` in progress
- `[x]` complete

## Epic 1: Architecture Baseline

- [ ] Create `internal/app/session` service facade around current state.
- [ ] Add port interfaces for profile repository, clock, and RNG.
- [ ] Wire console and graphic modes through facade without behavior changes.

Acceptance criteria:
- Both UIs run from the facade.
- Existing gameplay tests continue to pass.

## Epic 2: Domain Refactor

- [ ] Introduce `internal/domain/gameplay` aggregate and value objects.
- [ ] Move collision, scoring, level, and win/lose rules into domain package.
- [ ] Keep persistence representation outside domain entities.

Acceptance criteria:
- Domain package has no infra/UI imports.
- Equivalent behavior verified with regression tests.

## Epic 3: Application Use Cases

- [ ] Implement command/query API on session service.
- [ ] Add read models for HUD/menu/game-over screen.
- [ ] Implement profile sync use case with repository port.

Acceptance criteria:
- UI reads only from application read models.
- No direct `game.State` mutation in UI packages.

## Epic 4: Infrastructure Adapters

- [ ] Move file profile logic to `internal/infra/profile`.
- [ ] Add adapter tests for load/save and migration safety.
- [ ] Add `clock` and `rng` adapters in `internal/infra/system`.

Acceptance criteria:
- App layer depends only on interfaces.
- Infrastructure tests pass on Windows CI.

## Epic 5: Cleanup and Hardening

- [ ] Remove obsolete compatibility paths.
- [ ] Raise domain+app coverage target to at least 85%.
- [ ] Document final architecture and update README developer notes.

Acceptance criteria:
- No duplicate gameplay rule implementation remains.
- CI verifies new package coverage and build integrity.

## Suggested Delivery Sequence

1. Epic 1
2. Epic 2
3. Epic 3
4. Epic 4
5. Epic 5
