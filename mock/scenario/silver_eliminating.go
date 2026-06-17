package scenario

import (
	"github.com/goplease-game/client/ds"
)

const SilverEliminating = "Silver eliminating"

func init() {
	addScenario(SilverEliminating, silverEliminating)
}

func silverEliminating() *Scenario {
	s := NewSimpleScenario()

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Sir Lag-A-Lot",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true, MistID, FletchID)
	s.P2 = p2

	s.placeUnitAt(s.P1, SilverID, 3, 2)
	s.placeUnitAt(s.P1, GritID, 1, 2)

	s.placeUnitAt(s.P2, FletchID, 2, 2)
	p2Mist := s.placeUnitAt(s.P2, MistID, 2, 3)

	p2Mist.CurrentHP = 3

	s.P1.Units = nil
	return s
}
