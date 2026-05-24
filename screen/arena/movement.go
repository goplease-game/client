package arena

import "github.com/ognev-dev/goplease-ebitengine-client/ds"

// isReachableHex reports whether coord is present in the precomputed reachable list.
func isReachableHex(cells []ds.HexCoord, coord ds.HexCoord) bool {
	for _, c := range cells {
		if c == coord {
			return true
		}
	}
	return false
}

// moveUnit updates the local board state to reflect a unit moving to to.
// It clears the unit from its current cell and places it on the destination cell.
// The unit's position fields are updated on the copy stored in the destination cell.
func (s *Screen) moveUnit(u ds.Unit, to ds.HexCoord) {
	from := ds.HexCoord{Q: u.Pos.Q, R: u.Pos.R}

	if cell := s.board.Cells[from]; cell != nil {
		cell.Unit = nil
	}

	u.Pos.Q = to.Q
	u.Pos.R = to.R

	if cell := s.board.Cells[to]; cell != nil {
		cell.Unit = &u
	}
}
