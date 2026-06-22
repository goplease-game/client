package scenario

import (
	"github.com/goplease-game/client/tutorial"
)

const ClassicFairArena = "Classic Fair Arena"

func init() {
	addScenario(ClassicFairArena, classicFairArena)
}

func classicFairArena() *Scenario {
	s := NewSimpleScenario()
	s.Tutorial = tutorial.Basics

	return s
}
