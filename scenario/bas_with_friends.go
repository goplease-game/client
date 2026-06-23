package scenario

// BasWithFriends defines the name of the scenario featuring Bas alongside friendly units.
const BasWithFriends = "Bas with friends"

// init registers the "Bas with friends" scenario in the global Scenarios map.
func init() {
	addScenario(BasWithFriends, basWithFriends)
}

// basWithFriends builds and returns a scenario where Player 1 starts with
// a full squad on the board, game over conditions are disabled, and Player 2 has no units.
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
