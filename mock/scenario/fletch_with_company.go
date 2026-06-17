package scenario

import (
	"github.com/goplease-game/client/ds"
)

const FletchWithCompany = "Fletch with company"

func init() {
	addScenario(FletchWithCompany, fletchWithBadCompany)
}

func fletchWithBadCompany() *Scenario {
	s := NewSimpleScenario()

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Standing Still",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true)
	s.P2 = p2

	s.placeUnitAt(s.P1, FletchID, 0, 2)
	s.placeUnitAt(s.P1, BasID, -1, 3)
	s.placeUnitAt(s.P1, JulyID, 0, 1)

	s.placeUnitAt(s.P2, FletchID, 2, 2)
	s.placeUnitAt(s.P2, BasID, 3, 2)
	s.placeUnitAt(s.P2, JulyID, 4, 2)
	s.placeUnitAt(s.P2, SilverID, 2, 1)
	s.placeUnitAt(s.P2, GritID, 1, 3)
	s.placeUnitAt(s.P2, MistID, 2, 0)

	s.P1.Units = nil

	return s
}
