package arena

import "github.com/ognev-dev/goplease-ebitengine-client/ds"

// isReachable reports whether (r, c) is in the precomputed reachable list.
func isReachable(cells [][2]int, r, c int) bool {
	for _, cell := range cells {
		if cell[0] == r && cell[1] == c {
			return true
		}
	}
	return false
}

// moveUnit updates the board state, moving unit from its current position to (toR, toC).
func (s *Screen) moveUnit(u ds.Unit, toR, toC int) {
	// Clear old cell.
	s.board[u.Row][u.Col].Unit = nil

	// Update unit position.
	u.Row = toR
	u.Col = toC
	s.board[toR][toC].Unit = &u
}
