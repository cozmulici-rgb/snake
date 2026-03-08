# Web Redesign Plan

## Objective

Redesign the browser-based Three.js Snake UI so it reads as a deliberate game interface instead of a prototype overlay. The redesign should improve scan speed, gameplay readability, moment-to-moment feedback, and the sense of depth without sacrificing control clarity.

This plan applies to the current web frontend in:

- `internal/ui/web/static/index.html`
- `internal/ui/web/static/app.js`
- `internal/ui/web/server.go`

## Product Goals

1. Make the HUD readable in under one second.
2. Keep gameplay objects more visually important than the grid and chrome.
3. Reserve overlays for state transitions instead of persistent competition with the board.
4. Make "Snake 3D" feel materially more three-dimensional while preserving top-down clarity.
5. Keep desktop and mobile layouts functional with the same core information architecture.

## Design Direction

### Visual language

- Background: deep navy with a lighter atmospheric center glow.
- Panels: dark blue glass with stronger separation from the board.
- Primary accent: electric blue for actions and active UI.
- Gameplay accents:
  - Snake: neon green
  - Food: warm orange-red
  - Obstacles: purple-blue
- Typography:
  - headings: bold, compact sans serif
  - stats: tabular or mono-friendly numerals for alignment and quick scanning

### Experience principles

- The board is the main stage.
- The HUD should summarize, not narrate.
- Motion should reinforce events, not decorate empty states.
- Game-over and pause states should feel intentional and actionable.

## Priority Order

Implement in this order unless blocked:

1. HUD hierarchy and information reduction
2. Board readability and object differentiation
3. State overlays: start, pause, game over
4. Motion and feedback events
5. Camera/depth pass
6. Secondary systems: sound, high score, mobile gesture input

## Phase Plan

## Phase 1: HUD Hierarchy

### Goals

- Make `Score` and `Level` the fastest-read values.
- Reduce the feeling of an eight-metric debug bar.
- Make the top panel feel intentionally grouped.

### Changes

- Split HUD into primary and secondary groups.
- Promote `Score` to a large stacked metric.
- Pair `Level` with explicit progress to next level.
- Demote or hide low-value metrics unless active.
- Replace raw `Next Level` count with a clearer progress model:
  - progress bar
  - or `Food to next level: N`

### Proposed HUD structure

Left cluster:

- Snake 3D
- Score
- Level
- Progress to next level

Right cluster:

- Length
- Speed
- Food eaten
- Time elapsed

Conditional:

- Obstacles count only after obstacles spawn

### Acceptance criteria

- Score is visually dominant.
- The HUD remains readable on desktop and mobile.
- The top panel can be understood without reading every value.

## Phase 2: Board Readability

### Goals

- Ensure the snake, food, and obstacles are easier to identify than the grid.
- Reduce the prototype-like visual noise from the board surface.

### Changes

- Lower grid opacity and thickness.
- Reduce board contrast against gameplay pieces.
- Add stronger material distinction:
  - snake head brighter and slightly larger
  - snake body with softer extrusion and subtle glow
  - food with brighter emissive pulse
  - obstacles with distinct purple-blue tone and softer glow
- Review board fog and lighting so gameplay objects do not blend into the background.

### Acceptance criteria

- The player can identify snake head direction immediately.
- Food is visible at a glance even near grid intersections.
- Obstacles no longer blend into the board.

## Phase 3: Overlay Redesign

### Goals

- Make overlays feel like state screens, not control dumps.
- Improve start, pause, and game-over clarity.

### Changes

- Replace the generic control stack with state-specific overlays.
- Introduce explicit sections for:
  - current state
  - summary
  - primary action
  - secondary action
- Redesign the difficulty selector:
  - label it as `Difficulty`
  - remove numeric indexing
- Game-over screen should emphasize:
  - final score
  - length
  - survival time
  - restart shortcut
- Add a dedicated pause overlay with resume and restart.

### Overlay content targets

Start:

- Difficulty selector
- Start action
- concise keyboard hint

Paused:

- Resume
- Restart
- controls reminder

Game over:

- Final score
- Length
- Time survived
- `Press R to restart`

### Acceptance criteria

- Each overlay exposes one obvious primary action.
- Difficulty selection reads like a game setting, not a list item.
- Game-over state communicates performance in under two seconds.

## Phase 4: Motion and Event Feedback

### Goals

- Increase perceived polish and gameplay responsiveness.
- Add clear reward and state-change feedback.

### Changes

- Food pulse and float animation
- Snake growth pulse on eat
- Small score popup near the snake head
- Level-up flash or banner
- Board or camera micro reaction on level changes

### Acceptance criteria

- Eating food feels different from regular movement.
- Level transitions are visible without reading HUD numbers.

## Phase 5: Depth and Camera Pass

### Goals

- Make `Snake 3D` visually true without harming readability.

### Changes

- Introduce a subtle isometric tilt or camera offset.
- Add deeper shadows and vertical separation.
- Slightly increase mesh volume and edge definition.
- Consider a mild camera follow that preserves the whole board.

### Constraints

- Keep the full board readable.
- Avoid perspective distortion that makes cell alignment ambiguous.

### Acceptance criteria

- The scene reads as 3D in screenshots.
- The board still feels precise to control.

## Phase 6: Secondary Systems

These are worthwhile, but not blockers for the main redesign.

### Add after the core visual pass

- Best score / high score presentation
- Sound effects:
  - eat
  - level up
  - game over
  - movement tick if subtle enough
- Mobile touch gestures
- Collapsible side UI during active gameplay
- Subtle background effects

## Control and UX Updates

Ensure these are visible in the interface where relevant:

- `Space` to start
- `R` to restart
- `P` to pause/resume
- `Esc` for pause or overlay exit if adopted

Prefer state-aware hints instead of showing all controls at all times.

## Technical Notes

### Files likely to change

- `internal/ui/web/static/index.html`
- `internal/ui/web/static/app.js`
- `internal/ui/web/server.go`

### Implementation guidance

- Keep overlays data-driven from snapshot state.
- Add small helper renderers for HUD groups instead of one large string template.
- Avoid making mobile behavior an afterthought; validate narrow layouts at each phase.
- Preserve `window.render_game_to_text` and `window.advanceTime(ms)` for automated verification.

## Validation Checklist

- Desktop start state screenshot
- Desktop gameplay screenshot
- Desktop game-over screenshot
- Mobile start state screenshot
- Mobile ready state screenshot
- Mobile gameplay screenshot
- Verify `render_game_to_text` matches visible state
- Verify pause, restart, and start shortcuts
- Verify no new console errors

## Recommended First Slice

If this redesign is implemented incrementally, start with:

1. Rebuild the HUD into primary and secondary groups.
2. Reduce grid dominance and increase food visibility.
3. Redesign the game-over overlay around performance summary and restart.

This produces the largest quality improvement for the least architectural risk.
