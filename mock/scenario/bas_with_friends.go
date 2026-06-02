package scenario

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

func basWithFriends() *Scenario {
	s := NewSimpleScenario()

	s.placeUnitAt(s.P1, ds.HexCoord{1, 3}, BasID)
	s.placeUnitAt(s.P1, ds.HexCoord{1, 2}, GritID)
	s.placeUnitAt(s.P1, ds.HexCoord{2, 2}, FletchID)
	s.placeUnitAt(s.P1, ds.HexCoord{0, 3}, SilverID)
	s.placeUnitAt(s.P1, ds.HexCoord{0, 4}, MistID)
	s.placeUnitAt(s.P1, ds.HexCoord{1, 4}, JulyID)

	return s
}
