package scenario

// YouLose ...
const YouLose = "You lose"

func init() {
	addScenario(YouLose, youLoseScenario)
}

func youLoseScenario() *Scenario {
	s := NewSimpleScenario()

	grit := s.placeUnitAt(s.P1, GritID, 0, -1)
	grit.CurrentHP = 1

	// enemy
	s.placeEnemyAt(s.P2, SilverID, 0, 0)
	s.placeEnemyAt(s.P2, FletchID, 1, 1)
	s.placeEnemyAt(s.P2, GritID, -1, 0)
	s.placeEnemyAt(s.P2, MistID, -2, 0)

	s.P1.Units = nil
	s.P2.Units = nil

	return s
}
