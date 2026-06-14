package scenario

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const JulyWithCompany = "July with company"

func init() {
	addScenario(JulyWithCompany, julyWithBadCompany)
}

func julyWithBadCompany() *Scenario {
	s := NewSimpleScenario()

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Potato Sack",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true, BasID, JulyID, FletchID)
	s.P2 = p2

	s.placeUnitAt(s.P1, JulyID, 1, 3)
	fletch := s.placeUnitAt(s.P1, FletchID, 1, 2)
	grit := s.placeUnitAt(s.P1, GritID, 1, 4)
	grit.CurrentHP = 1
	grit.CurrentShield = 9

	fletch.AddStatus(status.Value{
		UnitID:   fletch.ID,
		Duration: 1,
		Value:    1,
		Status:   status.ByType(status.Marked),
	})
	fletch.CurrentHP = 8

	fletch2 := s.placeUnitAt(s.P2, FletchID, 3, 1)
	s.placeUnitAt(s.P2, BasID, 3, 2)
	s.placeUnitAt(s.P2, JulyID, 2, 3)

	fletch2.AddStatus(status.Value{
		UnitID:   fletch.ID,
		Duration: 1,
		Value:    1,
		Status:   status.ByType(status.DebuffWard),
	})

	s.P1.Units = nil

	return s
}
