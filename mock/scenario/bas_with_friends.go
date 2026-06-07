package scenario

const BasWithFriends = "Bas with friends"

func init() {
	addScenario(BasWithFriends, basWithFriends)
}

func basWithFriends() *Scenario {
	s := NewSimpleScenario()
	s.DisableGameOver = true

	s.placeUnitAt(s.P1, BasID, 1, 3)
	s.placeUnitAt(s.P1, GritID, 1, 2)
	s.placeUnitAt(s.P1, FletchID, 2, 2)
	s.placeUnitAt(s.P1, SilverID, 0, 3)
	s.placeUnitAt(s.P1, MistID, 0, 4)
	s.placeUnitAt(s.P1, JulyID, 1, 4)

	return s
}
