package scenario

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const SilverWithCompany = "Silver with company"

func init() {
	addScenario(SilverWithCompany, silverWithBadCompany)
}

func silverWithBadCompany() *Scenario {
	s := NewSimpleScenario()

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Chicken Heart",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true, BasID, JulyID, FletchID)
	s.P2 = p2

	s.placeUnitAt(s.P1, SilverID, 1, 3)
	s.placeUnitAt(s.P1, GritID, 1, 2)
	s.placeUnitAt(s.P1, JulyID, 1, 4)

	s.placeUnitAt(s.P2, FletchID, 3, 1)
	s.placeUnitAt(s.P2, BasID, 3, 2)
	s.placeUnitAt(s.P2, JulyID, 3, 3)

	s.P1.Units = nil

	return s
}
