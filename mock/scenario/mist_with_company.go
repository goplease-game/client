package scenario

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const MistWithCompany = "Mist with company"

func init() {
	addScenario(MistWithCompany, mistWithBadCompany)
}

func mistWithBadCompany() *Scenario {
	s := NewSimpleScenario()

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Alt F4",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true, BasID, JulyID, FletchID)
	s.P2 = p2

	s.placeUnitAt(s.P1, MistID, 1, 3)
	s.placeUnitAt(s.P1, GritID, 1, 2)
	s.placeUnitAt(s.P1, JulyID, 1, 4)

	fletch := s.placeUnitAt(s.P2, FletchID, 3, 1)
	s.placeUnitAt(s.P2, BasID, 3, 2)
	s.placeUnitAt(s.P2, JulyID, 2, 3)

	sv := status.Value{
		UnitID:   fletch.ID,
		Duration: 1,
		Value:    1,
		Status:   status.ByType(status.Rallied),
	}
	fletch.AddStatus(sv)

	s.P1.Units = nil

	return s
}
