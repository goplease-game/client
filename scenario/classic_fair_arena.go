package scenario

import (
	"github.com/goplease-game/client/tutorial"
)

// ClassicFairArena defines the name of the standard fair arena scenario.
const ClassicFairArena = "Classic Fair Arena"

// init registers the classic fair arena scenario in the global Scenarios map.
func init() {
	addScenario(ClassicFairArena, classicFairArena)
}

// classicFairArena builds and returns a simple scenario configured
// with the basic tutorial chapter.
func classicFairArena() *Scenario {
	s := NewSimpleScenario()
	s.Tutorial = tutorial.Basics

	return s
}
