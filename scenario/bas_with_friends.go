package scenario

const BasWithFriends = "Bas with friends"

func init() {
	addScenario(BasWithFriends, basWithFriends)
}

func basWithFriends() *Scenario {
	s := NewSimpleScenario()
	s.DisableGameOver = true

	s.placeUnitAt(s.P1, BasID, 0, 0)
	s.placeUnitAt(s.P1, GritID, 0, -1)
	s.placeUnitAt(s.P1, FletchID, 1, -1)
	s.placeUnitAt(s.P1, SilverID, -1, 0)
	s.placeUnitAt(s.P1, MistID, -1, 1)
	s.placeUnitAt(s.P1, JulyID, 0, 1)

	s.P2.Units = nil
	s.DisableBot = true

	return s
}
