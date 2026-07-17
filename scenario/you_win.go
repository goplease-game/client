package scenario

// YouWin ...
const YouWin = "You win"

func init() {
	addScenario(YouWin, youWinScenario)
}

func youWinScenario() *Scenario {
	s := NewSimpleScenario()

	s.placeUnitAt(s.P1, BasID, 0, 0)
	s.placeUnitAt(s.P1, GritID, 0, -1)
	s.placeUnitAt(s.P1, FletchID, 1, -1)
	s.placeUnitAt(s.P1, SilverID, -1, 0)
	s.placeUnitAt(s.P1, MistID, -1, 1)
	s.placeUnitAt(s.P1, JulyID, 0, 1)

	silver := s.placeEnemyAt(s.P2, SilverID, 0, 3)
	silver.CurrentHP = 1

	s.P2.Units = nil

	return s
}
