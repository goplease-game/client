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

	// board position, -1 - in hand
	Row int `json:"row"`
	Col int `json:"col"`

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

// ReachableCells returns all board positions the unit can move to this turn
// using the D&D style diagonal rule (1st diagonal costs 1, 2nd costs 2, etc.).
func (u Unit) ReachableCells(board Board) [][2]int {
	var result [][2]int

	rows := len(board)
	if rows == 0 {
		return result
	}
	cols := len(board[0])

	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	max := func(a, b int) int {
		if a > b {
			return a
		}
		return b
	}

	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}

	for dr := -u.MP; dr <= u.MP; dr++ {
		for dc := -u.MP; dc <= u.MP; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}

			adr := abs(dr)
			adc := abs(dc)

			cost := max(adr, adc) + (min(adr, adc) / 2)
			if cost > u.MP {
				continue
			}

			r, c := u.Row+dr, u.Col+dc
			if r < 0 || r >= rows || c < 0 || c >= cols {
				continue
			}
			if board[r][c] != nil && board[r][c].Unit != nil {
				continue
			}

			result = append(result, [2]int{r, c})
		}
	}

	return result
}

type PlaceUnitPayload struct {
	Row  int   `json:"row"`
	Col  int   `json:"col"`
	Unit *Unit `json:"unit"`
}

type PlayUnitPayload struct {
	UnitID string `json:"unit_id"`
}

type UnitPlacedPayload struct {
	Row        int `json:"row"`
	Col        int `json:"col"`
	TemplateID int `json:"template_id"`
}

type UnitMovedPayload struct {
	ToRow  int    `json:"to_row"`
	ToCol  int    `json:"to_col"`
	UnitID string `json:"unit_id"`
}
