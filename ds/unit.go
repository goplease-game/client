package ds

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
)

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

	BaseAP    int `json:"base_ap"` // Action Points
	CurrentAP int `json:"current_ap"`
	MP        int `json:"mp"` // Move Points

	Pos HexCoord `json:"pos"`

	Abilities []ability.ID       `json:"abilities"`
	Cooldowns map[ability.ID]int `json:"cooldowns"`
	Statuses  map[status.Type]status.Value

	IsOpponent bool
	IsDead     bool

	// Graphic reference to the board portrait widget, nil if not on board
	Graphic *widget.Graphic
}

func (u Unit) HasAbility(id ability.ID) bool {
	for _, abID := range u.Abilities {
		if abID == id {
			return true
		}
	}

	return false
}

func (u *Unit) HasStatus(t status.Type) bool {
	_, ok := u.Statuses[t]
	return ok
}

func (u *Unit) AddStatus(value status.Value) {
	if u.Statuses == nil {
		u.Statuses = make(map[status.Type]status.Value)
	}

	u.Statuses[value.Status.Type] = value
}

func (u *Unit) RemoveStatus(t status.Type) {
	delete(u.Statuses, t)
}

func (u *Unit) IsStunned() bool {
	_, ok := u.Statuses[status.Stun]

	return ok
}

// ReachableCells returns all hex cells the unit can reach within its movement points (MP).
// Movement is calculated using a breadth-first search over the hex grid, where each step
// to a neighboring cell costs 1 MP.
func (u Unit) ReachableCells(board Board) []HexCoord {
	type node struct {
		pos  HexCoord
		cost int
	}

	visited := make(map[HexCoord]int)
	queue := []node{{u.Pos, 0}}

	dirs := []HexCoord{
		{+1, 0}, {+1, -1}, {0, -1},
		{-1, 0}, {-1, +1}, {0, +1},
	}

	result := make([]HexCoord, 0)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, d := range dirs {
			next := HexCoord{
				Q: cur.pos.Q + d.Q,
				R: cur.pos.R + d.R,
			}

			// exists?
			cell, ok := board.Cells[next]
			if !ok {
				continue
			}

			// blocked by unit
			if cell.Unit != nil && next != u.Pos {
				continue
			}

			newCost := cur.cost + 1
			if newCost > u.MP {
				continue
			}

			prev, seen := visited[next]
			if seen && prev <= newCost {
				continue
			}

			visited[next] = newCost
			queue = append(queue, node{next, newCost})

			result = append(result, next)
		}
	}

	return result
}

type PlaceUnitPayload struct {
	Coord HexCoord `json:"coord"`
	Unit  *Unit    `json:"unit"`
}

type UnitMovedPayload struct {
	Coord  HexCoord `json:"coord"`
	UnitID string   `json:"unit_id"`
}

type PlayUnitPayload struct {
	UnitID string `json:"unit_id"`
}

type UnitPlacedPayload struct {
	Coord      HexCoord `json:"coord"`
	TemplateID int      `json:"template_id"`
}
