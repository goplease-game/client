package scenario

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const SilverGanging = "Silver ganging"

func init() {
	addScenario(SilverGanging, silverGanging)
}

func silverGanging() *Scenario {
	s := NewSimpleScenario()

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Missed Again",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true, BasID, FletchID)
	s.P2 = p2

	s.placeUnitAt(s.P1, SilverID, 3, 2)
	s.placeUnitAt(s.P1, GritID, 1, 2)

	s.placeUnitAt(s.P2, FletchID, 2, 2)
	s.placeUnitAt(s.P2, BasID, 2, 3)

	s.P1.Units = nil

	return s
}
