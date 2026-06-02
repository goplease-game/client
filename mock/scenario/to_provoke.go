package scenario

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

func toProvoke() *Scenario {
	s := NewSimpleScenario()

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Asking For Trouble",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true, BasID, JulyID)
	s.P2 = p2

	s.placeUnitAt(s.P1, ds.HexCoord{1, 3}, BasID)
	s.placeUnitAt(s.P1, ds.HexCoord{1, 4}, JulyID)

	s.placeUnitAt(s.P2, ds.HexCoord{3, 2}, BasID)
	s.placeUnitAt(s.P2, ds.HexCoord{3, 3}, JulyID)

	s.P1.Units = nil

	return s
}
