package scenario

import (
	"github.com/goplease-game/client/ds"
)

const GritWithCompany = "Grit with company"

func init() {
	addScenario(GritWithCompany, gritWithBadCompany)
}

func gritWithBadCompany() *Scenario {
	s := NewSimpleScenario()

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Thinking Long",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true, BasID, JulyID, FletchID)
	s.P2 = p2

	s.placeUnitAt(s.P1, GritID, 1, 3)
	s.placeUnitAt(s.P1, JulyID, 1, 4)

	s.placeUnitAt(s.P2, FletchID, 3, 1)
	s.placeUnitAt(s.P2, BasID, 3, 2)
	s.placeUnitAt(s.P2, JulyID, 3, 3)

	s.P1.Units = nil

	return s
}
