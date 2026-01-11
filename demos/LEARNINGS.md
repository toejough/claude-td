# Demo Learnings

Lessons, properties, and design decisions discovered during throwaway demos. This feeds directly into Phase 1 production code.

## Properties Discovered

These become simulation tests in Phase 1.

### Pathfinding

- [ ] Path from spawn to base must exist (or be explicitly blocked)
- [ ] Path must not traverse non-walkable tiles (walls, towers)
- [ ] Path must be contiguous (each step is adjacent to the previous)

### Enemy Movement

- [ ] Enemies must follow their assigned path waypoints in order
- [ ] Enemies must not move backward along their path
- [ ] Enemies must not walk through towers/walls (unless "breaking through" mechanic)
- [ ] When grid changes, enemies must recalculate from **current position**, not restart
- [ ] Enemy position must always be within grid bounds

### Grid State

- [ ] Spawn tile must exist and be unique
- [ ] Base tile must exist and be unique
- [ ] Towers can only be placed on ground tiles
- [ ] Towers can only be removed (not walls, spawn, base)

### Tower Targeting

- [ ] Towers only target enemies within range
- [ ] Towers target "first in path" (highest path progress) among in-range enemies
- [ ] Towers respect cooldown between shots
- [ ] Towers deal exact damage amount (no overkill tracking needed for hitscan)
- [ ] Dead enemies (HP <= 0) are not valid targets

### Combat

- [ ] Enemy HP decreases by exact damage amount when hit
- [ ] Enemies with HP <= 0 are removed from play
- [ ] Hitscan damage is instant (no projectile travel)

### Game Flow

- [ ] Player must have setup time before first wave to place initial towers
- [ ] Wave countdown must be visible so player knows when enemies are coming
- [ ] Game must pause updates when in won/lost state (only allow restart)

## Design Decisions (Validated by Demos)

### Path Ownership
**Decision**: Each enemy owns a copy of their path.
**Why**: When global path changes, enemies shouldn't teleport or reverse. They recalculate from current position.
**Demo**: 0.4 - enemies going backwards when tower placed

### Path Recalculation Trigger
**Decision**: Recalculate enemy paths immediately when grid changes.
**Why**: Delayed recalculation causes enemies to walk through newly-placed obstacles.
**Demo**: 0.4 - enemies walking through towers

### Coordinate Systems
**Decision**: Grid coordinates (int) for logic, pixel coordinates (float64) for rendering/movement.
**Why**: Smooth movement requires sub-cell positioning; pathfinding works on discrete cells.
**Demo**: 0.4 - smooth enemy movement along path

## Bugs & Corrections

### Demo 0.4: Enemies Going Backwards
**Symptom**: Placing tower extended path; enemies already past the obstruction went backwards.
**Cause**: Enemies shared global path by index; path change invalidated their index.
**Fix**: Each enemy gets path copy at spawn.
**Lesson**: Mutable shared state + indices = bugs. Entities should own their navigation state.

### Demo 0.4: Enemies Walking Through Towers
**Symptom**: Enemies walked straight through towers placed after they spawned.
**Cause**: Enemies kept old path, didn't recalculate when grid changed.
**Fix**: On grid change, recalculate each enemy's path from their current position.
**Lesson**: Path recalculation must consider entity's current position, not just spawn.

### Demo 0.5: Enemies and Lasers Drawn Under Grid
**Symptom**: Enemies and laser effects appeared behind grid lines.
**Cause**: Grid lines were drawn last, covering everything.
**Fix**: Explicit layer ordering - draw grid lines early, entities and effects later.
**Lesson**: Define draw order explicitly as layers. In 2D games: background → terrain → grid → indicators → entities → effects → UI.

## Architecture Insights

### What Worked Well
- Separating grid state from entity state
- Using A* with clear isWalkable() predicate
- Fixed-size grid with constants
- Explicit draw layer ordering (tiles → grid lines → path → hover → enemies → lasers)

### What Needs Improvement for Phase 1
- Pathfinding should be extracted to pure function (currently method on Game)
- Enemy update logic should be separate from rendering state
- Need clear separation: World (grid) → Pathfinder → Entities → Renderer

## Questions for Phase 1

1. Should enemies store path as []Point or as a Path struct with helper methods?
2. How to handle "no path" case - enemy stops? breaks through? despawns?
3. Path caching - recalculate on every grid change, or cache and invalidate?

---

## Demo Completion Log

### Demo 0.1: Window + Grid Rendering ✓
- Learned: Ebitengine basics, vector drawing API
- No bugs

### Demo 0.2: Click to Place/Remove Tiles ✓
- Learned: Mouse input handling, coordinate conversion
- No bugs

### Demo 0.3: A* Pathfinding Visualization ✓
- Learned: heap-based A*, path reconstruction
- No bugs

### Demo 0.4: Entity Movement Along Path ✓
- Learned: Path ownership matters, recalculation timing matters
- Bugs: 2 (backwards movement, walking through towers)

### Demo 0.5: Tower Targeting + Hitscan ✓
- Learned: Draw order matters, explicit layering prevents visual bugs
- Learned: Tower state (cooldown) must be tracked per-tower, not globally
- Bugs: 1 (draw order - enemies under grid lines)

### Demo 0.6: Minimal Game Loop ✓
- Learned: Players need setup time before first wave
- Learned: Resource economy needs balancing (cost vs reward vs starting amount)
- Learned: Game state machine (playing/won/lost) simplifies update logic
- Bugs: 1 (no setup time before wave 1)
