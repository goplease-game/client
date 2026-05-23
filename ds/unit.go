package ds

import "github.com/ognev-dev/goplease-ebitengine-client/ability"

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

	AP int `json:"ap"` // Action Points
	MP int `json:"mp"` // Move Points

	Pos HexCoord `json:"pos"`

	Abilities []ability.ID       `json:"abilities"`
	Cooldowns map[ability.ID]int `json:"cooldowns"`

	IsOpponent bool
}

//// ReachableCells returns all board positions the unit can move to this turn.
//// A cell (r, c) is reachable if:
////   - manhattan distance to the unit <= unit.MP
////   - the cell is within board bounds
////   - the cell is not already occupied by any unit
//func (u Unit) ReachableCells(board Board) [][2]int {
//	var result [][2]int
//
//	rows := len(board)
//	if rows == 0 {
//		return result
//	}
//	cols := len(board[0])
//
//	abs := func(x int) int {
//		if x < 0 {
//			return -x
//		}
//		return x
//	}
//
//	for dr := -u.MP; dr <= u.MP; dr++ {
//		for dc := -u.MP; dc <= u.MP; dc++ {
//			if dr == 0 && dc == 0 {
//				continue
//			}
//			if abs(dr)+abs(dc) > u.MP {
//				continue
//			}
//
//			r, c := u.Row+dr, u.Col+dc
//			if r < 0 || r >= rows || c < 0 || c >= cols {
//				continue
//			}
//			if board[r][c] != nil && board[r][c].Unit != nil {
//				continue
//			}
//
//			result = append(result, [2]int{r, c})
//		}
//	}
//
//	return result
//}

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
