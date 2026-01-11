package main

import (
	"container/heap"
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// GameState represents the current state of the game
type GameState int

const (
	StatePlaying GameState = iota
	StateWon
	StateLost
)

// Point represents a grid coordinate
type Point struct {
	X, Y int
}

// Priority queue for A*
type pqItem struct {
	point    Point
	priority int // f = g + h
	index    int
}

type priorityQueue []*pqItem

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].priority < pq[j].priority }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*pqItem)
	item.index = n
	*pq = append(*pq, item)
}
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

const (
	// Grid dimensions
	GridWidth  = 20
	GridHeight = 15

	// Cell size in pixels
	CellSize = 40

	// Window dimensions
	ScreenWidth  = GridWidth * CellSize
	ScreenHeight = GridHeight * CellSize
)

// TileType represents what's in a cell
type TileType int

const (
	TileEmpty TileType = iota
	TileGround
	TileWall
	TileBase
	TileSpawn
	TileTower
)

// Colors for each tile type
var tileColors = map[TileType]color.RGBA{
	TileEmpty:  {R: 30, G: 30, B: 30, A: 255},    // Dark gray
	TileGround: {R: 80, G: 60, B: 40, A: 255},    // Brown
	TileWall:   {R: 100, G: 100, B: 100, A: 255}, // Gray
	TileBase:   {R: 50, G: 100, B: 200, A: 255},  // Blue
	TileSpawn:  {R: 200, G: 50, B: 50, A: 255},   // Red
	TileTower:  {R: 50, G: 200, B: 50, A: 255},   // Green
}

var gridLineColor = color.RGBA{R: 60, G: 60, B: 60, A: 255}
var highlightColor = color.RGBA{R: 255, G: 255, B: 255, A: 80}
var pathColor = color.RGBA{R: 255, G: 200, B: 50, A: 180}
var noPathColor = color.RGBA{R: 255, G: 0, B: 0, A: 100}
var enemyColor = color.RGBA{R: 255, G: 100, B: 100, A: 255}
var laserColor = color.RGBA{R: 255, G: 255, B: 0, A: 255}

const (
	EnemySpeed     = 2.0   // Pixels per tick
	SpawnInterval  = 40    // Ticks between spawns within a wave
	EnemyRadius    = 12.0  // Visual radius
	EnemyMaxHP     = 100.0 // Starting HP

	TowerRange    = 120.0 // Pixels
	TowerDamage   = 10.0  // Damage per shot
	TowerCooldown = 30    // Ticks between shots
	LaserDuration = 5     // Ticks to show laser

	// Game balance
	TotalWaves       = 5   // Waves to survive to win
	EnemiesPerWave   = 5   // Base enemies per wave (scales with wave number)
	WaveDelay        = 180 // Ticks between waves
	StartingResource = 100 // Resources at game start
	TowerCost        = 25  // Cost to place a tower
	KillReward       = 10  // Resources earned per kill
)

// Enemy represents a moving enemy
type Enemy struct {
	X, Y      float64 // Position in pixels
	PathIndex int     // Current target waypoint in path
	Path      []Point // Enemy's own copy of the path
	HP        float64 // Current health
}

// Tower represents a placed tower
type Tower struct {
	X, Y     int // Grid position
	Cooldown int // Ticks until can fire again
}

// Laser represents a visual shot effect
type Laser struct {
	FromX, FromY float64
	ToX, ToY     float64
	TTL          int // Ticks remaining to display
}

// Game holds the game state
type Game struct {
	grid [GridHeight][GridWidth]TileType

	// Mouse state
	hoverX, hoverY int  // Grid cell under cursor (-1 if none)
	hoverValid     bool // Is cursor over a valid cell?

	// Pathfinding
	spawn, base Point   // Start and end points
	path        []Point // Current path from spawn to base
	pathBlocked bool    // True if no valid path exists

	// Enemies
	enemies    []*Enemy
	spawnTimer int // Ticks until next spawn

	// Towers
	towers []*Tower
	lasers []*Laser // Visual effects for shots

	// Game state
	state     GameState
	resources int

	// Wave system
	currentWave     int  // Current wave number (1-indexed)
	enemiesThisWave int  // Enemies remaining to spawn this wave
	waveDelay       int  // Ticks until next wave starts
	totalKills      int  // Total enemies killed
}

// NewGame creates a new game with an initial grid layout
func NewGame() *Game {
	g := &Game{}

	// Fill with ground
	for y := 0; y < GridHeight; y++ {
		for x := 0; x < GridWidth; x++ {
			g.grid[y][x] = TileGround
		}
	}

	// Add some walls around edges
	for x := 0; x < GridWidth; x++ {
		g.grid[0][x] = TileWall
		g.grid[GridHeight-1][x] = TileWall
	}
	for y := 0; y < GridHeight; y++ {
		g.grid[y][0] = TileWall
		g.grid[y][GridWidth-1] = TileWall
	}

	// Place base (what we're defending) - bottom center
	g.base = Point{X: GridWidth / 2, Y: GridHeight - 2}
	g.grid[g.base.Y][g.base.X] = TileBase

	// Place spawn point - top center
	g.spawn = Point{X: GridWidth / 2, Y: 1}
	g.grid[g.spawn.Y][g.spawn.X] = TileSpawn

	// Add some interior walls for interest
	for y := 3; y < 8; y++ {
		g.grid[y][5] = TileWall
		g.grid[y][14] = TileWall
	}

	// Calculate initial path
	g.recalculatePath()

	// Initialize game state
	g.state = StatePlaying
	g.resources = StartingResource
	g.currentWave = 1
	g.enemiesThisWave = EnemiesPerWave
	g.waveDelay = 300 // 5 seconds to place initial towers

	return g
}

// isWalkable returns true if a tile can be walked through
func (g *Game) isWalkable(x, y int) bool {
	if x < 0 || x >= GridWidth || y < 0 || y >= GridHeight {
		return false
	}
	tile := g.grid[y][x]
	return tile == TileGround || tile == TileSpawn || tile == TileBase
}

// heuristic calculates Manhattan distance
func heuristic(a, b Point) int {
	dx := a.X - b.X
	dy := a.Y - b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// findPath uses A* to find path from start to goal
func (g *Game) findPath(start, goal Point) []Point {
	openSet := &priorityQueue{}
	heap.Init(openSet)
	heap.Push(openSet, &pqItem{point: start, priority: 0})

	cameFrom := make(map[Point]Point)
	gScore := make(map[Point]int)
	gScore[start] = 0

	// 4-directional movement
	dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*pqItem).point

		if current == goal {
			// Reconstruct path
			path := []Point{current}
			for current != start {
				current = cameFrom[current]
				path = append([]Point{current}, path...)
			}
			return path
		}

		for _, d := range dirs {
			neighbor := Point{X: current.X + d.X, Y: current.Y + d.Y}

			if !g.isWalkable(neighbor.X, neighbor.Y) {
				continue
			}

			tentativeG := gScore[current] + 1

			if oldG, exists := gScore[neighbor]; !exists || tentativeG < oldG {
				cameFrom[neighbor] = current
				gScore[neighbor] = tentativeG
				f := tentativeG + heuristic(neighbor, goal)
				heap.Push(openSet, &pqItem{point: neighbor, priority: f})
			}
		}
	}

	// No path found
	return nil
}

// recalculatePath updates the global path from spawn to base
func (g *Game) recalculatePath() {
	g.path = g.findPath(g.spawn, g.base)
	g.pathBlocked = g.path == nil
}

// recalculateEnemyPaths updates paths for all existing enemies from their current position
func (g *Game) recalculateEnemyPaths() {
	for _, e := range g.enemies {
		// Get enemy's current grid cell
		gridX := int(e.X) / CellSize
		gridY := int(e.Y) / CellSize
		currentCell := Point{X: gridX, Y: gridY}

		// Find new path from current position to base
		newPath := g.findPath(currentCell, g.base)
		if newPath != nil {
			e.Path = newPath
			e.PathIndex = 1 // Start moving toward second waypoint
		}
		// If no path, enemy keeps current path (will walk through obstacle)
		// This matches the "enemies can break through" design
	}
}

// spawnEnemy creates a new enemy at the spawn point
func (g *Game) spawnEnemy() {
	if g.pathBlocked || len(g.path) == 0 {
		return
	}
	// Copy the current path for this enemy
	pathCopy := make([]Point, len(g.path))
	copy(pathCopy, g.path)

	e := &Enemy{
		X:         float64(g.spawn.X*CellSize) + CellSize/2,
		Y:         float64(g.spawn.Y*CellSize) + CellSize/2,
		PathIndex: 1, // Start moving toward second waypoint (first is spawn)
		Path:      pathCopy,
		HP:        EnemyMaxHP,
	}
	g.enemies = append(g.enemies, e)
}

// addTower places a tower and tracks it (returns false if can't afford)
func (g *Game) addTower(x, y int) bool {
	if g.resources < TowerCost {
		return false
	}
	g.resources -= TowerCost
	g.grid[y][x] = TileTower
	g.towers = append(g.towers, &Tower{X: x, Y: y, Cooldown: 0})
	return true
}

// removeTower removes a tower (refunds half cost)
func (g *Game) removeTower(x, y int) {
	g.grid[y][x] = TileGround
	g.resources += TowerCost / 2 // Refund half
	// Remove from tower list
	for i, t := range g.towers {
		if t.X == x && t.Y == y {
			g.towers = append(g.towers[:i], g.towers[i+1:]...)
			break
		}
	}
}

// updateEnemies moves all enemies along the path
func (g *Game) updateEnemies() {
	alive := make([]*Enemy, 0, len(g.enemies))

	for _, e := range g.enemies {
		// Remove dead enemies and grant reward
		if e.HP <= 0 {
			g.resources += KillReward
			g.totalKills++
			continue
		}

		if e.PathIndex >= len(e.Path) {
			// Enemy reached the base - GAME OVER
			g.state = StateLost
			continue
		}

		// Get target waypoint center (from enemy's own path)
		target := e.Path[e.PathIndex]
		targetX := float64(target.X*CellSize) + CellSize/2
		targetY := float64(target.Y*CellSize) + CellSize/2

		// Calculate direction
		dx := targetX - e.X
		dy := targetY - e.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < EnemySpeed {
			// Reached waypoint, move to next
			e.X = targetX
			e.Y = targetY
			e.PathIndex++
		} else {
			// Move toward waypoint
			e.X += (dx / dist) * EnemySpeed
			e.Y += (dy / dist) * EnemySpeed
		}

		alive = append(alive, e)
	}

	g.enemies = alive
}

// updateTowers handles tower targeting and shooting
func (g *Game) updateTowers() {
	for _, t := range g.towers {
		// Decrease cooldown
		if t.Cooldown > 0 {
			t.Cooldown--
			continue
		}

		// Find target: enemy in range that is furthest along its path (closest to base)
		towerX := float64(t.X*CellSize) + CellSize/2
		towerY := float64(t.Y*CellSize) + CellSize/2

		var target *Enemy
		bestProgress := -1

		for _, e := range g.enemies {
			if e.HP <= 0 {
				continue
			}

			// Check range
			dx := e.X - towerX
			dy := e.Y - towerY
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > TowerRange {
				continue
			}

			// "First in path" = highest PathIndex (closest to base)
			if e.PathIndex > bestProgress {
				bestProgress = e.PathIndex
				target = e
			}
		}

		// Fire at target
		if target != nil {
			target.HP -= TowerDamage
			t.Cooldown = TowerCooldown

			// Create laser visual
			g.lasers = append(g.lasers, &Laser{
				FromX: towerX,
				FromY: towerY,
				ToX:   target.X,
				ToY:   target.Y,
				TTL:   LaserDuration,
			})
		}
	}
}

// updateLasers decrements laser TTL and removes expired ones
func (g *Game) updateLasers() {
	alive := make([]*Laser, 0, len(g.lasers))
	for _, l := range g.lasers {
		l.TTL--
		if l.TTL > 0 {
			alive = append(alive, l)
		}
	}
	g.lasers = alive
}

// Update handles game logic
func (g *Game) Update() error {
	// Handle restart on R key when game is over
	if g.state != StatePlaying {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			*g = *NewGame()
		}
		return nil
	}

	// Wave spawning logic
	if g.waveDelay > 0 {
		g.waveDelay--
	} else if g.enemiesThisWave > 0 {
		// Spawn enemies for current wave
		g.spawnTimer--
		if g.spawnTimer <= 0 {
			g.spawnEnemy()
			g.enemiesThisWave--
			g.spawnTimer = SpawnInterval
		}
	} else if len(g.enemies) == 0 {
		// Wave complete, all enemies dead
		if g.currentWave >= TotalWaves {
			// All waves complete - WIN!
			g.state = StateWon
		} else {
			// Start next wave
			g.currentWave++
			g.enemiesThisWave = EnemiesPerWave + g.currentWave // More enemies each wave
			g.waveDelay = WaveDelay
		}
	}

	// Move enemies
	g.updateEnemies()

	// Tower targeting and shooting
	g.updateTowers()

	// Update laser visuals
	g.updateLasers()

	// Get mouse position and convert to grid coordinates
	mx, my := ebiten.CursorPosition()
	gx, gy := mx/CellSize, my/CellSize

	// Check if cursor is within grid bounds
	g.hoverValid = gx >= 0 && gx < GridWidth && gy >= 0 && gy < GridHeight
	if g.hoverValid {
		g.hoverX, g.hoverY = gx, gy
	}

	// Handle clicks (only when playing)
	if g.hoverValid && g.state == StatePlaying {
		tile := g.grid[g.hoverY][g.hoverX]
		gridChanged := false

		// Left click: place tower (only on ground, if can afford)
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			if tile == TileGround {
				if g.addTower(g.hoverX, g.hoverY) {
					gridChanged = true
				}
			}
		}

		// Right click: remove tower (back to ground)
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			if tile == TileTower {
				g.removeTower(g.hoverX, g.hoverY)
				gridChanged = true
			}
		}

		// Recalculate paths if grid changed
		if gridChanged {
			g.recalculatePath()
			g.recalculateEnemyPaths()
		}
	}

	return nil
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Layer 1: Tiles
	for y := 0; y < GridHeight; y++ {
		for x := 0; x < GridWidth; x++ {
			tile := g.grid[y][x]
			c := tileColors[tile]

			px := float32(x * CellSize)
			py := float32(y * CellSize)
			vector.DrawFilledRect(screen, px, py, CellSize, CellSize, c, false)
		}
	}

	// Layer 2: Grid lines
	for x := 0; x <= GridWidth; x++ {
		px := float32(x * CellSize)
		vector.StrokeLine(screen, px, 0, px, ScreenHeight, 1, gridLineColor, false)
	}
	for y := 0; y <= GridHeight; y++ {
		py := float32(y * CellSize)
		vector.StrokeLine(screen, 0, py, ScreenWidth, py, 1, gridLineColor, false)
	}

	// Layer 3: Path indicator
	if g.pathBlocked {
		px := float32(g.spawn.X * CellSize)
		py := float32(g.spawn.Y * CellSize)
		vector.DrawFilledRect(screen, px, py, CellSize, CellSize, noPathColor, false)
	} else {
		for _, p := range g.path {
			px := float32(p.X*CellSize) + CellSize/4
			py := float32(p.Y*CellSize) + CellSize/4
			vector.DrawFilledRect(screen, px, py, CellSize/2, CellSize/2, pathColor, false)
		}
	}

	// Layer 4: Hover highlight
	if g.hoverValid {
		px := float32(g.hoverX * CellSize)
		py := float32(g.hoverY * CellSize)
		vector.DrawFilledRect(screen, px, py, CellSize, CellSize, highlightColor, false)
	}

	// Layer 5: Enemies with HP bars
	for _, e := range g.enemies {
		vector.DrawFilledCircle(screen, float32(e.X), float32(e.Y), EnemyRadius, enemyColor, true)

		// HP bar
		hpRatio := e.HP / EnemyMaxHP
		barWidth := float32(EnemyRadius * 2)
		barHeight := float32(4)
		barX := float32(e.X) - barWidth/2
		barY := float32(e.Y) - EnemyRadius - 6

		vector.DrawFilledRect(screen, barX, barY, barWidth, barHeight, color.RGBA{60, 60, 60, 255}, false)
		hpColor := color.RGBA{uint8(255 * (1 - hpRatio)), uint8(255 * hpRatio), 0, 255}
		vector.DrawFilledRect(screen, barX, barY, barWidth*float32(hpRatio), barHeight, hpColor, false)
	}

	// Layer 6: Lasers (topmost)
	for _, l := range g.lasers {
		vector.StrokeLine(screen, float32(l.FromX), float32(l.FromY), float32(l.ToX), float32(l.ToY), 2, laserColor, false)
	}

	// Layer 7: UI Text
	var statusText string
	switch g.state {
	case StatePlaying:
		waveStatus := fmt.Sprintf("Wave %d/%d", g.currentWave, TotalWaves)
		if g.waveDelay > 0 {
			waveStatus += fmt.Sprintf(" (next in %ds)", g.waveDelay/60+1)
		}
		statusText = fmt.Sprintf("%s | Resources: %d | Kills: %d | Tower cost: %d",
			waveStatus, g.resources, g.totalKills, TowerCost)
	case StateWon:
		statusText = fmt.Sprintf("YOU WIN! Survived all %d waves! Kills: %d | Press R to restart", TotalWaves, g.totalKills)
	case StateLost:
		statusText = fmt.Sprintf("GAME OVER - Enemy reached base! Wave %d | Kills: %d | Press R to restart",
			g.currentWave, g.totalKills)
	}
	ebitenutil.DebugPrint(screen, statusText)
}

// Layout returns the game's screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Claude TD - Demo 0.6")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
