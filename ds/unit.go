package ds

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
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

	IsOpponent bool

	// Graphic reference to the board portrait widget, nil if not on board
	Graphic *widget.Graphic
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

// ApplyState represents a single, atomic state mutation applied to a unit.
// Sequential execution of these states forms the visual timeline on the client side.
type ApplyState struct {
	ToUnitID string `json:"to_unit_id"`

	// Movement
	ChangePos *HexCoord `json:"change_pos,omitempty"` // New position on the grid

	// Delta changes used to trigger floating text or combat UI animations
	ChangeHP     *int `json:"change_hp,omitempty"`     // Health delta (e.g., -5, +12)
	ChangeAP     *int `json:"change_ap,omitempty"`     // Action points delta
	ChangeMP     *int `json:"change_mp,omitempty"`     // Movement points delta
	ChangeShield *int `json:"change_shield,omitempty"` // Shield delta
	ChangeAtk    *int `json:"change_atk,omitempty"`    // Attack power delta

	// Absolute values used for hard state synchronization after the animation plays
	SetHP     *int `json:"set_hp,omitempty"`     // Hard set current health
	SetAP     *int `json:"set_ap,omitempty"`     // Hard set current action points
	SetMP     *int `json:"set_mp,omitempty"`     // Hard set current movement points
	SetShield *int `json:"set_shield,omitempty"` // Hard set current shield
	SetAtk    *int `json:"set_atk,omitempty"`    // Hard set current attack power

	// Statuses and effects
	IsDead        bool     `json:"is_dead,omitempty"`
	AddEffects    []string `json:"add_effects,omitempty"`
	RemoveEffects []string `json:"remove_effects,omitempty"`
}
