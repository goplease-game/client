package arena

import (
	"slices"

	"github.com/goplease-game/client/ds"
)

// isReachableHex reports whether coord is present in the precomputed reachable list.
func isReachableHex(cells []ds.HexCoord, coord ds.HexCoord) bool {
	return slices.Contains(cells, coord)
}

// moveUnit updates the local board state to reflect a unit moving to to.
// It clears the unit from its current cell and places it on the destination cell.
// The unit's position fields are updated on the copy stored in the destination cell.
func (s *Screen) moveUnit(u *ds.Unit, to ds.HexCoord) {
	if cell := s.board.Cells[u.Pos]; cell != nil {
		cell.Unit = nil
	}

	u.Pos = to
	if cell := s.board.Cells[to]; cell != nil {
		cell.Unit = u
	}
}
