package scenario

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const PhantomAP = "Phantom AP"

func init() {
	addScenario(PhantomAP, phantomAP)
}

func phantomAP() *Scenario {
	s := NewSimpleScenario()

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Fantomas",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true)
	s.P2 = p2

	s.placeUnitAt(s.P2, FletchID, 1, 0)
	s.placeUnitAt(s.P1, FletchID, 1, 3)
	s.placeUnitAt(s.P2, MistID, 2, 0)
	s.placeUnitAt(s.P1, MistID, 2, 3)
	s.placeUnitAt(s.P2, BasID, 3, 0)
	s.placeUnitAt(s.P1, BasID, 3, 3)
	s.placeUnitAt(s.P2, SilverID, 4, 0)
	s.placeUnitAt(s.P1, SilverID, 1, 2)
	s.placeUnitAt(s.P2, GritID, 5, 0)
	s.placeUnitAt(s.P1, GritID, 1, 1)
	s.placeUnitAt(s.P2, JulyID, 5, 1)
	s.placeUnitAt(s.P1, JulyID, 1, 4)

	for _, u := range s.Queue {
		if u.OwnerID == s.P1.ID {
			u.CurrentHP = 2
			//u.BaseAP = i + 1
		}
	}

	return s
}
