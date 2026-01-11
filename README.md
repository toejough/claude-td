# Claude TD

A tower defense game built in Go with Ebitengine.

## Vision

Maze-building tower defense inspired by Defense Grid, with destructible environments and cost-weighted pathfinding. Players place towers to create mazes; enemies choose optimal paths including breaking through obstacles when that's cheaper than going around.

## Core Design Decisions

### Gameplay

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Map style | Maze-building | Player creates paths via tower placement |
| Pathfinding | Cost-weighted A* | Enemies weigh travel distance vs destruction cost |
| Blocking | Allowed | Towers can fully block paths; enemies break through |
| Environment | Destructible | Everything can be destroyed with sufficient resources |
| Enemy interaction | Full | Enemies can destroy obstacles (towers, terrain) |
| Base | Single tile (expandable) | One enemy entering = failure |
| Grid | 20x15 default | Snap placement, configurable size |
| Projectiles | Hitscan (v0.1) | Room for travel-time + homing/predictive later |
| Targeting | First-in-path (v0.1) | Configurable priority in future |
| Resources | Generic "resources" | Earned from kills, spent on towers |

### Inspirations

- **Defense Grid**: Primary inspiration - maze building, strategic depth, quality pathfinding
- **Bloons TD**: Tower variety, upgrade paths
- **Defend Your Castle**: Action-oriented base defense
- **Revenge of the Titans**: Art style, UI, resource collection during waves

### v0.1 Scope (Minimal)

- One enemy type (walks, has HP)
- One tower type (shoots, deals hitscan damage)
- One resource type
- Basic wave system
- Win: survive N waves
- Lose: one enemy enters base

## Architecture

```
claude-td/
├── core/                 # Pure Go, NO Ebitengine imports
│   ├── world/            # Grid, Tile, Terrain
│   ├── entity/           # Tower, Enemy, Projectile
│   ├── path/             # Cost-weighted A* pathfinding
│   ├── system/           # Movement, Targeting, Damage, Spawn
│   ├── economy/          # Resources, costs
│   ├── sim/              # Game state, deterministic tick loop
│   └── level/            # Level definitions
├── render/               # Ebitengine drawing (reads state, never mutates)
├── input/                # Input → game commands
├── ui/                   # HUD, menus
├── cmd/
│   ├── game/             # Main game binary
│   └── sim/              # Headless simulation CLI
├── demos/                # Phase 0 throwaway prototypes
│   └── prototype/        # Single evolving demo
├── testdata/             # Saved states, replays
├── build/                # targ build targets
├── issues.md             # Work tracking
└── go.mod
```

### Key Architectural Constraint

**`core/` has zero dependencies on Ebitengine or rendering.** All game logic is pure functions on data. This enables:

- Headless simulation (run 1000 ticks in milliseconds)
- Deterministic replay (same inputs = same outputs)
- Full test coverage of game mechanics
- Bug reports as saved state + tick sequence

### Simulation Model

- Fixed timestep: 60 ticks/second
- Game logic independent of framerate
- State transitions are pure functions: `State × Input → State`

## Development Methodology

### Testing Strategy

| Layer | Approach |
|-------|----------|
| Core logic | TDD with imptest (RED → GREEN → REFACTOR) |
| Properties | Simulation-validated invariants |
| Balance | Spreadsheet intent → simulation check → playtest feel |
| Feel | Playtesting with feedback as property assertions |

### Property-Based Testing

Game properties are defined as invariants the simulation validates:

- "Every wave should be beatable with optimal tower placement"
- "No enemy should reach base in under N ticks with starting resources"
- "Resource economy should be net-positive if player kills >80% of enemies"

Feedback from playtesting becomes new properties or tuning changes.

### Simulation CLI

```bash
# Validate a specific scenario
go run ./cmd/sim --wave=3 --towers="2,2:basic;5,5:basic" --validate

# Run randomized testing
go run ./cmd/sim --wave=3 --random-placement=10000 --report
```

### Workflow

1. Define property/behavior in plain language
2. Write failing test (RED)
3. Implement (GREEN)
4. Refactor if needed
5. Playtest, give feedback
6. Feedback becomes new properties or tuning
7. Repeat

### Quality Gates

- Tests pass
- targ check passes
- Simulation properties hold
- Playtested and approved

### Process

| Aspect | Decision |
|--------|----------|
| Branching | Trunk-based, main always green |
| Sessions | Always end with runnable code |
| Pairing | Claude drives, Joe steers |
| Demos | Throwaway, hard cutover to production |

## Technology Stack

| Component | Choice | Why |
|-----------|--------|-----|
| Language | Go | Familiarity, toolchain, cross-platform |
| Game engine | Ebitengine | Mature Go 2D engine, WASM/mobile support |
| Testing | imptest | Interactive mocks, step-through testing |
| Build/CLI | targ | Build targets + CLI framework |
| Platforms | Desktop, Web (WASM), Mobile | Ebitengine supports all |

## Development Phases

### Phase 0: Concept Exploration (Throwaway Demos)

Learn Ebitengine, validate feel, iterate fast. Colored rectangles.

| Demo | Goal |
|------|------|
| 0.1 | Window + grid rendering |
| 0.2 | Click to place/remove tiles |
| 0.3 | A* pathfinding visualization |
| 0.4 | Entity movement along path |
| 0.5 | Tower targeting + hitscan |
| 0.6 | Minimal game loop (spawn → defend → win/lose) |

Each demo is disposable. Learn, then throw away.

### Phase 1: Core Engine

Testable foundation with full coverage. Pure Go, no Ebitengine.

- World: Grid, tiles, terrain types
- Pathfinding: Cost-weighted A* with destruction costs
- Entities: Tower, Enemy, Projectile as data
- Systems: Movement, targeting, damage, spawning
- Simulation: Fixed-timestep, deterministic, headless
- Economy: Resources, costs, rewards

### Phase 2: Minimal Playable Game

- Ebitengine rendering layer
- Input → game commands
- Basic UI (resources, wave, win/lose)
- Wave definitions
- Level format
- Save/load state

### Phase 3: Content & Depth

- Multiple tower/enemy types
- Upgrades
- Environmental variety
- Level editor or definitions
- Balance tuning

### Phase 4: Polish & Distribution

- Art, sound
- UI polish
- Multi-platform builds
- Performance optimization

## Future Considerations

Noted for architecture but not in v0.1:

- Base as multi-tile structure with HP
- Projectiles with travel time (homing vs predictive)
- Configurable tower targeting priority
- Environmental terrain effects (water, mud = slow)
- Explosives for rock destruction
- Variable grid sizes per level
