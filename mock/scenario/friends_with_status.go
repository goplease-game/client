package scenario

import (
	"math/rand"

	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const FriendsWithStatus = "Friends with Status"

func init() {
	addScenario(FriendsWithStatus, friendsWithStatus)
}

func friendsWithStatus() *Scenario {
	s := NewSimpleScenario()

	bas := s.placeUnitAt(s.P1, BasID, 1, 3)
	grit := s.placeUnitAt(s.P1, GritID, 1, 2)
	fletch := s.placeUnitAt(s.P1, FletchID, 2, 2)
	silver := s.placeUnitAt(s.P1, SilverID, 0, 3)
	mist := s.placeUnitAt(s.P1, MistID, 0, 4)
	july := s.placeUnitAt(s.P1, JulyID, 1, 4)

	allSt := map[status.Type]status.Value{}
	for key, st := range status.Statuses {
		// we'll get an infinite loop of rounds
		// if everyone is stunned (because of auto-end turn)
		if key == status.Stunned {
			continue
		}
		allSt[key] = status.Value{
			UnitID:   "",
			Duration: 10,
			Value:    st.InitialValue,
			Status:   st,
		}
	}

	bas.Statuses = allSt
	grit.Statuses = allSt
	fletch.Statuses = allSt
	silver.Statuses = allSt
	mist.Statuses = allSt
	july.Statuses = allSt

	p2 := &ds.Player{
		ID:          "p2",
		Name:        "Richard Statusless",
		IsBot:       true,
		PlayerIndex: 1,
	}
	p2.Units = loadUnits(p2.ID, true, BasID, JulyID, FletchID)
	s.P2 = p2

	fletch2 := s.placeUnitAt(s.P2, FletchID, 4, 0)
	bas2 := s.placeUnitAt(s.P2, BasID, 4, 1)
	july2 := s.placeUnitAt(s.P2, JulyID, 4, 2)

	randomSt := func(from, to int) map[status.Type]status.Value {
		count := from + rand.Intn(to-from+1)

		indices := rand.Perm(len(status.Order))[:count]
		result := make(map[status.Type]status.Value, count)
		for _, i := range indices {
			t := status.Order[i]
			result[t] = status.Value{Status: status.ByType(t)}
		}
		return result
	}

	bas2.Statuses = randomSt(1, 3)
	july2.Statuses = randomSt(2, 5)
	fletch2.Statuses = randomSt(3, 7)

	return s
}
