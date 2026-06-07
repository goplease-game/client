package scenario

import (
	"github.com/google/uuid"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const safeZoneSize = 2
const arenaRows = 10
const arenaCols = 10
const boardSize = 5

const ClassicFairArena = "Classic Fair Arena"

func init() {
	addScenario(ClassicFairArena, classicFairArena)
}

func classicFairArena() *Scenario {
	p1 := &ds.Player{
		ID:                   "p1",
		Name:                 "P1",
		IsBot:                false,
		PlayerIndex:          0,
		UnitsPlacedThisRound: 0,
	}
	p1.Units = loadUnits(p1.ID, false)

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Cheat A-lot",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true)

	s := &Scenario{
		ID:           uuid.New().String(),
		P1:           p1,
		P2:           p2,
		Queue:        []*ds.Unit{},
		Board:        newHexBoard(boardSize),
		ActiveUnitID: "",
	}

	return s
}

func newHexBoard(size int) ds.Board {
	b := ds.Board{
		Cells: make(map[ds.HexCoord]*ds.BoardCell),
	}

	for q := -size; q <= size; q++ {
		for r := -size; r <= size; r++ {
			if s := -q - r; s < -size || s > size {
				continue
			}
			coord := ds.HexCoord{Q: q, R: r}
			b.Cells[coord] = &ds.BoardCell{Coord: coord}
		}
	}

	for coord, cell := range b.Cells {
		cell.IsSafeZone = coord.Q <= -size+safeZoneSize-1
		b.Cells[coord] = cell
	}

	return b
}
