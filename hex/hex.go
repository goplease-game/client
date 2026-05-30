package hex

import "github.com/ognev-dev/goplease-ebitengine-client/ds"

// CellsInRange returns all board positions within rangeN hex steps of from.
// Uses hex cube distance, not square-grid distance, so diagonals are correct.
// Does not filter by occupancy — use for ability targeting, not movement.
func CellsInRange(from ds.HexCoord, rangeN int, board ds.Board) []ds.HexCoord {
	var result []ds.HexCoord
	for coord := range board.Cells {
		if Distance(from, coord) <= rangeN {
			result = append(result, coord)
		}
	}
	return result
}

// OppositeHex returns the hex coordinate directly opposite to the origin relative to the center.
func OppositeHex(origin, center ds.HexCoord) ds.HexCoord {
	return ds.HexCoord{
		Q: 2*center.Q - origin.Q,
		R: 2*center.R - origin.R,
	}
}

// Distance returns the hex cube distance between two axial coordinates.
// Equivalent to max(|dq|, |dr|, |dq+dr|).
func Distance(a, b ds.HexCoord) int {
	dq := a.Q - b.Q
	dr := a.R - b.R
	return max3(abs(dq), abs(dr), abs(dq+dr))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max3(a, b, c int) int {
	return max(max(a, b), c)
}
