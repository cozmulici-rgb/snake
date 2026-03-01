# Target Architecture

## Package Map

```text
cmd/
  console/
  graphic/

internal/
  app/
    session/
      service.go
      ports.go
      types.go
  domain/
    gameplay/
      aggregate.go
      value_objects.go
  infra/
    profile/
      file_repository.go
    system/
      clock.go
      rng.go
  ui/
    console/
      adapter.go
    graphic/
      adapter.go
```

## Dependency Direction

- `cmd -> ui -> app -> domain`
- `app -> domain`
- `infra -> app ports` and `infra -> domain types` only when required for serialization mapping.
- Domain must not import `app`, `infra`, `ui`, or third-party libraries.

## Core Interfaces (Ports)

## Application Input Ports

- `SessionService`
  - `Start(config PresetConfig)`
  - `ApplyDirection(dir Direction)`
  - `Tick(now time.Time)`
  - `PauseToggle()`
  - `Restart()`
  - `Quit()`

## Application Output Ports

- `ProfileRepository`
  - `Load(ctx context.Context) (Profile, error)`
  - `Save(ctx context.Context, profile Profile) error`
- `Clock`
  - `Now() time.Time`
- `Random`
  - `Intn(n int) int`

## Use-Case Boundaries

- Use cases return immutable view models for UI.
- Domain mutation only occurs through aggregate methods.
- UI cannot mutate state structs directly.

## Testing Strategy by Layer

- Domain: deterministic unit tests for invariants and rules.
- App: unit tests with fake repository/clock/RNG.
- Infra: integration tests against temp file system.
- UI: lightweight smoke tests and mapping tests.
