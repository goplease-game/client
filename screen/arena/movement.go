package arena

import "github.com/ognev-dev/goplease-ebitengine-client/ds"

// isReachable reports whether coord is in the precomputed reachable list.
func isReachableHex(cells []ds.HexCoord, coord ds.HexCoord) bool {
	for _, c := range cells {
		if c == coord {
			return true
		}
	}
	return false
}

// moveUnit updates the board state, moving unit from its current position to (toR, toC).
func (s *Screen) moveUnit(u ds.Unit, to ds.HexCoord) {
	from := ds.HexCoord{Q: u.Pos.Q, R: u.Pos.R}

	// Clear old cell.
	if cell := s.board.Cells[from]; cell != nil {
		cell.Unit = nil
	}

	// Update unit position.
	u.Pos.Q = to.Q
	u.Pos.R = to.R

	// Set new cell.
	if cell := s.board.Cells[to]; cell != nil {
		cell.Unit = &u
	}
}
