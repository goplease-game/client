package mock

import (
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/goplease-game/client/ability"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/grid"
)

type SimAction struct {
	Action string
	JSON   json.RawMessage
	data   any
}

const (
	UnitMovedAction = "unit_moved"
	UseAbility      = "use_ability"
	ApplyState      = "apply_state"
)

// SimulateUnitTurn determines and executes the best available action for the given unit.
// It evaluates class-specific priority scenarios first, then falls back to
// attacking the priority target or moving toward it.
func SimulateUnitTurn(unit *ds.Unit) (acts []SimAction) {
	defer func() {
		for i, a := range acts {
			var err error
			a.JSON, err = json.Marshal(a.data)
			if err != nil {
				log.Fatalf("[SimulateUnitTurn] json.Marshal: %s", err.Error())
			}
			acts[i] = a
		}
	}()

	if unit.CurrentAP < 1 {
		fmt.Printf("[mock] [SimulateUnitTurn] unit %s have no AP", unit.ID)
		return
	}

	if scenarios, ok := unitScenarios[unit.TemplateID]; ok {
		for _, scenario := range scenarios {
			if acts = scenario(unit); len(acts) > 0 {
				return
			}
		}
	}

	// Default: attack priority target or move toward it.
	acts = scenarioAttackPriorityTarget(unit)
	if len(acts) > 0 {
		return
	}

	acts = append(acts, scenarioMoveTowardsPriorityTarget(unit))
	return
}

func simulateSimpleAttack(u *ds.Unit) (data []SimAction) {
	var ab *ability.Ability
	for _, id := range u.Abilities {
		if a := ability.ByID(id); a.IsBasicAttack() {
			ab = &a
		}
	}

	if ab == nil {
		return
	}

	enemy := findClosestEnemy(u)
	if enemy == nil {
		return
	}

	moveTo, ok := findAttackPosition(u, enemy, ab.Range)
	if !ok {
		return
	}

	if moveTo != u.Pos {
		PlaceUnitAt(u, moveTo)
		data = append(data, SimAction{
			Action: UnitMovedAction,
			data: ds.UnitMovedPayload{
				UnitID: u.ID,
				Coord:  moveTo,
			},
		})
	}

	states, err := HandleAbility(ds.UseAbilityPayload{
		UnitID:    u.ID,
		AbilityID: ab.ID,
		Target:    &enemy.Pos,
	})
	if err != nil {
		log.Fatalf("[SimulateUnitTurn] HandleAbility: %s", err.Error())
	}

	data = append(data, SimAction{
		Action: UseAbility,
		data: ds.UseAbilityPayload{
			UnitID:    u.ID,
			AbilityID: ab.ID,
			Target:    &enemy.Pos,
		},
	}, SimAction{
		Action: ApplyState,
		data:   states,
	})

	return data
}

func findAttackPosition(u *ds.Unit, target *ds.Unit, attackRange int) (ds.HexCoord, bool) {
	if grid.Distance(u.Pos, target.Pos) <= attackRange {
		return u.Pos, true
	}

	var bestPos ds.HexCoord
	bestMove := math.MaxInt

	for coord, cell := range gameState.Board.Cells {
		if cell.Unit != nil {
			continue
		}

		moveDist := grid.Distance(u.Pos, coord)
		if moveDist > u.CurrentMP {
			continue
		}

		if grid.Distance(coord, target.Pos) > attackRange {
			continue
		}

		if moveDist < bestMove {
			bestMove = moveDist
			bestPos = coord
		}
	}

	if bestMove == math.MaxInt {
		return ds.HexCoord{}, false
	}

	return bestPos, true
}

func simulateMove(u *ds.Unit) SimAction {
	pos := RandomReachableCell(*u)
	PlaceUnitAt(u, pos)

	return SimAction{
		Action: UnitMovedAction,
		data: ds.UnitMovedPayload{
			UnitID: u.ID,
			Coord:  pos,
		},
	}
}

func findAllEnemies(of *ds.Unit) []*ds.Unit {
	enemies := []*ds.Unit{}
	for _, u := range gameState.UnitsQueue {
		if u.IsEnemy(of) {
			enemies = append(enemies, u)
		}
	}

	return enemies
}

// findClosestEnemy returns the nearest enemy to the given unit,
// breaking ties by lowest HP. Returns nil if no enemies exist.
func findClosestEnemy(of *ds.Unit) *ds.Unit {
	enemies := findAllEnemies(of)
	if len(enemies) == 0 {
		return nil
	}

	var best *ds.Unit
	for _, e := range enemies {
		if best == nil {
			best = e
			continue
		}
		distBest := grid.Distance(of.Pos, best.Pos)
		distE := grid.Distance(of.Pos, e.Pos)
		if distE < distBest || (distE == distBest && e.CurrentHP < best.CurrentHP) {
			best = e
		}
	}
	return best
}

// findClosestReachableEnemy returns the nearest enemy reachable
// within movement + abilityRange, or nil if none can be reached.
func findClosestReachableEnemy(u *ds.Unit, abilityRange int) *ds.Unit {
	enemies := findAllEnemies(u)

	var best *ds.Unit
	for _, e := range enemies {
		_, ok := findAttackPosition(u, e, abilityRange)
		if !ok {
			continue
		}
		if best == nil {
			best = e
			continue
		}
		if grid.Distance(u.Pos, e.Pos) < grid.Distance(u.Pos, best.Pos) {
			best = e
		}
	}
	return best
}

// findClosestEnemyWithBuffs returns the nearest reachable enemy
// that has at least one positive status effect, or nil.
func findClosestEnemyWithBuffs(u *ds.Unit, abilityRange int) *ds.Unit {
	enemies := findAllEnemies(u)

	var best *ds.Unit
	for _, e := range enemies {
		if !hasPositiveStatus(e) {
			continue
		}
		_, ok := findAttackPosition(u, e, abilityRange)
		if !ok {
			continue
		}
		if best == nil {
			best = e
			continue
		}
		if grid.Distance(u.Pos, e.Pos) < grid.Distance(u.Pos, best.Pos) {
			best = e
		}
	}
	return best
}

// findClosestAlly returns the nearest living ally, excluding self. Returns nil if none.
func findClosestAlly(u *ds.Unit) *ds.Unit {
	var best *ds.Unit
	for _, other := range gameState.UnitsQueue {
		if other.ID == u.ID || !other.IsAlly(u) {
			continue
		}
		if best == nil {
			best = other
			continue
		}
		if grid.Distance(u.Pos, other.Pos) < grid.Distance(u.Pos, best.Pos) {
			best = other
		}
	}
	return best
}

// findMostWoundedAllyInRange returns the ally (or self) within healRange
// with the lowest CurrentHP relative to BaseHP.
// Returns nil if all units in range are at full HP.
func findMostWoundedAllyInRange(u *ds.Unit, healRange int) *ds.Unit {
	candidates := append(findAlliesInRange(u, healRange), u)

	var best *ds.Unit
	for _, ally := range candidates {
		if ally.CurrentHP >= ally.BaseHP {
			continue
		}
		if best == nil {
			best = ally
			continue
		}
		if ally.CurrentHP < best.CurrentHP {
			best = ally
		}
	}
	return best
}

// findAllyWithDebuffInRange returns the first ally (or self) within range
// that has at least one negative status effect, or nil.
func findAllyWithDebuffInRange(u *ds.Unit, abilityRange int) *ds.Unit {
	candidates := append(findAlliesInRange(u, abilityRange), u)

	for _, ally := range candidates {
		if hasNegativeStatus(ally) {
			return ally
		}
	}
	return nil
}

// findAdjacentPosition returns the closest free cell adjacent to target
// that u can reach within its movement range, or false if none exists.
func findAdjacentPosition(u *ds.Unit, target *ds.Unit) (ds.HexCoord, bool) {
	neighbors := grid.Neighbors(target.Pos)

	var best ds.HexCoord
	bestDist := -1

	for _, pos := range neighbors {
		cell, ok := gameState.Board.Cells[pos]
		if !ok || cell.Unit != nil {
			continue
		}
		dist := grid.Distance(u.Pos, pos)
		if dist > u.CurrentMP {
			continue
		}
		if bestDist < 0 || dist < bestDist {
			best = pos
			bestDist = dist
		}
	}

	return best, bestDist >= 0
}

// canReachFrom reports whether a unit standing at fromPos could reach
// the target given abilityRange (ignores movement — assumes unit is already at fromPos).
func canReachFrom(fromPos ds.HexCoord, target *ds.Unit, abilityRange int) bool {
	return grid.Distance(fromPos, target.Pos) <= abilityRange
}

// findBestPositionForAOE finds the cell reachable by u (within MovePoints)
// that maximises the score returned by scoreFn(center, radius).
// Returns the best position and its score.
func findBestPositionForAOE(
	u *ds.Unit,
	radius int,
	scoreFn func(u *ds.Unit, center ds.HexCoord, radius int) int,
) (ds.HexCoord, int) {
	reachable := grid.CellsInRange(u.Pos, u.CurrentMP, gameState.Board)

	bestPos := u.Pos
	bestScore := scoreFn(u, u.Pos, radius)

	for _, pos := range reachable {
		cell, ok := gameState.Board.Cells[pos]
		if !ok || (cell.Unit != nil && cell.Unit.ID != u.ID) {
			continue
		}
		score := scoreFn(u, pos, radius)
		if score > bestScore {
			bestScore = score
			bestPos = pos
		}
	}

	return bestPos, bestScore
}

// countAlliesInRadius counts allies of u within radius of center,
// matching the signature expected by findBestPositionForAOE.
func countAlliesInRadius(u *ds.Unit, center ds.HexCoord, radius int) int {
	count := 0
	cells := grid.CellsInRange(center, radius, gameState.Board)
	for _, pos := range cells {
		unit := GetUnitAt(pos)
		if unit != nil && unit.IsAlly(u) && unit.ID != u.ID {
			count++
		}
	}
	return count
}

// simulateMoveTowards moves u one step in the direction of targetPos,
// choosing the reachable cell closest to the target.
func simulateMoveTowards(u *ds.Unit, targetPos ds.HexCoord) SimAction {
	reachable := grid.CellsInRange(u.Pos, u.CurrentMP, gameState.Board)

	bestPos := u.Pos
	bestDist := grid.Distance(u.Pos, targetPos)

	for _, pos := range reachable {
		cell, ok := gameState.Board.Cells[pos]
		if !ok || (cell.Unit != nil && cell.Unit.ID != u.ID) {
			continue
		}
		d := grid.Distance(pos, targetPos)
		if d < bestDist {
			bestDist = d
			bestPos = pos
		}
	}

	PlaceUnitAt(u, bestPos)
	return SimAction{
		Action: UnitMovedAction,
		data: ds.UnitMovedPayload{
			UnitID: u.ID,
			Coord:  bestPos,
		},
	}
}

// --- Status helpers ---

// hasPositiveStatus reports whether the unit has any active positive status effect.
func hasPositiveStatus(u *ds.Unit) bool {
	for _, v := range u.Statuses {
		if v.IsPositive() {
			return true
		}
	}
	return false
}

// hasNegativeStatus reports whether the unit has any active negative status effect.
func hasNegativeStatus(u *ds.Unit) bool {
	for _, v := range u.Statuses {
		if v.IsNegative() {
			return true
		}
	}
	return false
}

// findFreeCellAdjacentTo returns a free board cell adjacent to target
// reachable within stepRange hex steps from u's current position.
func findFreeCellAdjacentTo(u *ds.Unit, target *ds.Unit, stepRange int) (ds.HexCoord, bool) {
	for _, pos := range grid.Neighbors(target.Pos) {
		cell, ok := gameState.Board.Cells[pos]
		if !ok || cell.Unit != nil {
			continue
		}
		if grid.Distance(u.Pos, pos) <= stepRange {
			return pos, true
		}
	}
	return ds.HexCoord{}, false
}

// findAbilityTarget returns the position to pass as Target to HandleAbility,
// and the position to move to before using the ability.
// Returns false if the ability cannot be used against the given target from current position.
func findAbilityTarget(u *ds.Unit, target *ds.Unit, abilityID ability.ID) (moveTo ds.HexCoord, targetPos ds.HexCoord, ok bool) {
	a := ability.ByID(abilityID)

	switch a.Activation {
	case ability.SelectFreeCell:
		// Target is a free cell within range, not a unit.
		// Find a free cell adjacent to the priority target within ability range.
		targetPos, ok = findFreeCellAdjacentTo(u, target, a.Range)
		if !ok {
			return
		}
		moveTo = u.Pos // no walking — ability itself handles repositioning
		return

	default:
		// Target is a unit — find a position from which u can hit it.
		moveTo, ok = findAttackPosition(u, target, a.Range)
		if !ok {
			return
		}
		targetPos = target.Pos
		return
	}
}

// countEnemiesInRangeFrom counts enemies of u within radius of a given position.
// Used to evaluate AoE value before committing to a move.
func countEnemiesInRangeFrom(center ds.HexCoord, u *ds.Unit, radius int) int {
	count := 0
	cells := grid.CellsInRange(center, radius, gameState.Board)
	for _, pos := range cells {
		unit := GetUnitAt(pos)
		if unit != nil && unit.IsEnemy(u) {
			count++
		}
	}
	return count
}
