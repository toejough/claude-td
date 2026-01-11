# Issues

## Current Work

### Demo 0.1: Window + Grid Rendering

**Goal**: Basic Ebitengine window displaying a 20x15 grid with colored rectangles.

**Acceptance Criteria**:
- Window opens at appropriate size
- 20x15 grid visible with cell borders
- Colored rectangles distinguish cell types
- Clean exit on window close

**Timeline**:
- (not started)

---

## Backlog

### Phase 0: Concept Exploration

- [ ] Demo 0.1: Window + grid rendering
- [ ] Demo 0.2: Click to place/remove tiles
- [ ] Demo 0.3: A* pathfinding visualization
- [ ] Demo 0.4: Entity movement along path
- [ ] Demo 0.5: Tower targeting + hitscan
- [ ] Demo 0.6: Minimal game loop

### Phase 1: Core Engine

- [ ] World system (Grid, Tile, Terrain)
- [ ] Cost-weighted A* pathfinding
- [ ] Entity system (Tower, Enemy, Projectile)
- [ ] Movement system
- [ ] Targeting system
- [ ] Damage system
- [ ] Spawning system
- [ ] Economy system
- [ ] Deterministic simulation loop
- [ ] Headless simulation CLI

### Phase 2: Minimal Playable Game

- [ ] Ebitengine rendering layer
- [ ] Input handling
- [ ] Basic UI (resources, wave counter, win/lose)
- [ ] Wave system
- [ ] Level definition format
- [ ] Save/load state

### Phase 3: Content & Depth

- [ ] Multiple tower types
- [ ] Multiple enemy types
- [ ] Upgrade system
- [ ] Environmental variety
- [ ] Level editor or definitions
- [ ] Balance tuning

### Phase 4: Polish & Distribution

- [ ] Art assets
- [ ] Sound
- [ ] UI polish
- [ ] Multi-platform builds
- [ ] Performance optimization

---

## Completed

(none yet)

---

## Notes

### Design Decisions Log

Captured in README.md. Key decisions:

- Maze-building with cost-weighted pathfinding
- Blocking allowed; enemies break through
- Destructible everything
- Single base tile, one breach = failure
- Hitscan v0.1, travel-time later
- First-in-path targeting v0.1, configurable later
- Hard cutover from demos to production code

### Property Tests to Implement (Phase 1+)

- Every wave beatable with optimal placement
- No instant-loss scenarios with starting resources
- Resource economy net-positive at >80% kill rate
- Pathfinding always finds a route (including destruction)
