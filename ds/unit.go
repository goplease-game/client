package ds

import (
	"slices"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/server/ability"
	"github.com/goplease-game/server/ability/status"
)

// Unit represents a single combat unit on the board, including its stats,
// abilities, cooldowns, and statuses.
type Unit struct {
	ID          string `json:"id"`
	TemplateID  int    `json:"template_id"`
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	BaseAtk       int `json:"base_atk"`
	CurrentAtk    int `json:"current_atk"`
	BaseHP        int `json:"base_hp"`
	CurrentHP     int `json:"current_hp"`
	CurrentShield int `json:"current_shield"`
	BaseAP        int `json:"base_ap"` // Action Points
	CurrentAP     int `json:"current_ap"`
	BaseMP        int `json:"base_mp"` // Move Points
	CurrentMP     int `json:"current_mp"`

	Pos HexCoord `json:"pos"`

	Abilities []ability.ID                 `json:"abilities"`
	Cooldowns map[ability.ID]int           `json:"cooldowns"`
	Statuses  map[status.Type]status.Value `json:"statuses"`

	IsOpponent bool `json:"is_opponent"`
	IsDead     bool `json:"is_dead"`

	PhantomAPUsedThisTurn int `json:"phantom_ap_used_this_turn"`

	// Graphic reference to the board portrait widget, nil if not on board
	Graphic *widget.Graphic `json:"-"`
}

// HasAbility reports whether the unit has the given ability.
func (u *Unit) HasAbility(id ability.ID) bool {
	return slices.Contains(u.Abilities, id)
}

// AbilityReady reports whether the given ability is off cooldown.
func (u *Unit) AbilityReady(id ability.ID) bool {
	return u.Cooldowns[id] <= 0
}

// SetCooldown sets the cooldown for the given ability, removing the entry
// entirely when cd is zero.
func (u *Unit) SetCooldown(id ability.ID, cd int) {
	if u.Cooldowns == nil {
		u.Cooldowns = make(map[ability.ID]int)
	}

	if cd == 0 {
		delete(u.Cooldowns, id)
		return
	}

	u.Cooldowns[id] = cd
}

// HasStatus reports whether the unit currently has the given status type.
func (u *Unit) HasStatus(t status.Type) bool {
	_, ok := u.Statuses[t]
	return ok
}

// AddStatus applies the given status value to the unit, replacing any
// existing value for the same status type.
func (u *Unit) AddStatus(value status.Value) {
	if u.Statuses == nil {
		u.Statuses = make(map[status.Type]status.Value)
	}

	u.Statuses[value.Status.Type] = value
}

// RemoveStatus removes the given status type from the unit.
func (u *Unit) RemoveStatus(t status.Type) {
	delete(u.Statuses, t)
}

// IsEnemy reports whether the given unit belongs to a different owner.
func (u *Unit) IsEnemy(to *Unit) bool {
	return u.OwnerID != to.OwnerID
}

// IsAlly reports whether the given unit belongs to the same owner.
func (u *Unit) IsAlly(to *Unit) bool {
	return !u.IsEnemy(to)
}

// CanMove reports whether the unit can move this turn.
func (u *Unit) CanMove() bool {
	return u.CurrentMP > 0
}

// ReachableCells returns all hex cells the unit can reach within its movement points (CurrentMP).
// Movement is calculated using a breadth-first search over the hex grid, where each step
// to a neighboring cell costs 1 CurrentMP.
func (u *Unit) ReachableCells(board Board) []HexCoord {
	type node struct {
		pos  HexCoord
		cost int
	}

	visited := make(map[HexCoord]int)
	visited[u.Pos] = 0

	queue := []node{{u.Pos, 0}}
	result := make([]HexCoord, 0)

	dirs := []HexCoord{
		{+1, 0}, {+1, -1}, {0, -1},
		{-1, 0}, {-1, +1}, {0, +1},
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, d := range dirs {
			next := HexCoord{
				Q: cur.pos.Q + d.Q,
				R: cur.pos.R + d.R,
			}

			cell, ok := board.Cells[next]
			if !ok {
				continue
			}

			if cell.Unit != nil && next != u.Pos {
				continue
			}

			newCost := cur.cost + 1
			if newCost > u.CurrentMP {
				continue
			}

			prev, seen := visited[next]
			if seen && prev <= newCost {
				continue
			}

			visited[next] = newCost
			queue = append(queue, node{next, newCost})

			if !seen {
				result = append(result, next)
			}
		}
	}

	return result
}

// PlaceUnitPayload is the payload for placing a unit on the board at a specific coordinate.
type PlaceUnitPayload struct {
	Coord HexCoord `json:"coord"`
	Unit  *Unit    `json:"unit"`
}

// UnitMovedPayload is the payload broadcast when a unit moves to a new coordinate.
type UnitMovedPayload struct {
	Coord  HexCoord `json:"coord"`
	UnitID string   `json:"unit_id"`
}

// PlayUnitPayload is the payload for playing a unit from hand onto the board.
type PlayUnitPayload struct {
	UnitID string `json:"unit_id"`
}

// UnitPlacedPayload is the payload broadcast when a unit is placed on the board.
type UnitPlacedPayload struct {
	Coord      HexCoord `json:"coord"`
	TemplateID int      `json:"template_id"`
}

// ActiveUnitChangedPayload is the payload broadcast when the active unit changes.
type ActiveUnitChangedPayload struct {
	UnitID string `json:"unit_id"`
}
