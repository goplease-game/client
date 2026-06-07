package scenario

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const WinOrLose = "Win or Lose"

func init() {
	addScenario(WinOrLose, winOrLose)
}

func winOrLose() *Scenario {
	s := NewSimpleScenario()

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Lose-A-Lot",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true, MistID)
	s.P2 = p2

	fletch := s.placeUnitAt(s.P1, FletchID, 1, 3)
	fletch.CurrentHP = 1

	mist2 := s.placeUnitAt(s.P2, MistID, 3, 1)
	mist2.CurrentHP = 1

	s.P1.Units = nil

	return s
}
