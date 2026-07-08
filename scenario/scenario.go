// Package scenario ...
package scenario

import (
	"fmt"
	"log"

	"github.com/goplease-game/client/tutorial"
	server "github.com/goplease-game/server"
	"github.com/goplease-game/server/bot"
	sds "github.com/goplease-game/server/ds"
)

// Default represents the fallback or standard arena scenario.
const Default = ClassicFairArena

// Scenarios maps scenario names to their respective initialization functions.
var Scenarios = map[Name]func() *Scenario{}

// addScenario registers a new scenario factory function under the given name.
func addScenario(name Name, scenario func() *Scenario) {
	Scenarios[name] = scenario
}

// Unit template IDs used to identify specific units within scenarios.
const (
	BasID    = 1
	GritID   = 2
	FletchID = 3
	SilverID = 4
	MistID   = 5
	JulyID   = 6
)

// Name defines a unique string identifier for a scenario configuration.
type Name string

// Load returns a new Scenario instance for the given name.
// Panics if name is not registered in Scenarios.
func Load(name Name) *Scenario {
	return Scenarios[name]()
}

// Scenario holds the pre-configured initial state of a game arena,
// including players, board layout, queue order, and tutorial steps.
type Scenario struct {
	ID              sds.ID
	P1              *server.Player
	P2              *server.Player
	Queue           []*server.Unit
	Board           *server.Board
	ActiveUnitID    sds.ID
	DisableGameOver bool
	DisableBot      bool
	Tutorial        tutorial.Chapter
}

// NewSimpleScenario initializes and returns a default Scenario with two players,
// their starting units, and an empty board.
func NewSimpleScenario() *Scenario {
	p1ID := sds.NewID()
	p2ID := sds.NewID()

	p1 := server.NewPlayer(p1ID, "Player 1", 0, server.StartingUnits(p1ID))
	p2 := server.NewPlayer(p2ID, bot.PlayerName(), 1, server.StartingUnits(p2ID))

	s := &Scenario{
		ID:    sds.NewID(),
		P1:    p1,
		P2:    p2,
		Queue: []*server.Unit{},
		Board: server.NewBoard(),
	}

	return s
}

// Arena constructs and returns a fully initialized server Arena state
// populated with the data defined in the current Scenario.
func (s *Scenario) Arena() *server.Arena {
	return &server.Arena{
		ID:                     s.ID,
		Board:                  s.Board,
		Players:                [2]*server.Player{s.P1, s.P2},
		UnitsQueue:             s.Queue,
		CurrentRound:           0,
		ActivePlayer:           0,
		ActiveUnitID:           s.ActiveUnitID,
		Phase:                  server.PlacementPhase,
		UnitsPerPlacementPhase: server.UnitsPerPlacementPhase,
		DisableGameOver:        s.DisableGameOver,
		DisableBot:             s.DisableBot,
		DisableTurnTimer:       true,
	}
}

// placeUnitAt picks a unit by templateID from player's hand,
// sets its position and places it on the board cell at coord.
// The unit is also appended to the scenario queue.
func (s *Scenario) placeUnitAt(from *server.Player, unitID, atQ, atR int) *server.Unit {
	var unit *server.Unit
	unit, from.Units = pickUnitByTemplateID(from.Units, unitID)
	at := server.HexCoord{Q: atQ, R: atR}

	_, ok := s.Board.Cells[at]
	if !ok {
		fmt.Printf("[scenario] [placeUnitAt] cell at %s not exists!\n", at)
		return nil
	}

	unit.Pos = &at
	unit.OwnerID = from.ID

	s.Board.Cells[at].Unit = unit
	s.Queue = append(s.Queue, unit)

	return unit
}

// placeEnemyAt is the same as placeUnitAt, but sets IsOpponent to true.
func (s *Scenario) placeEnemyAt(from *server.Player, unitID, atQ, atR int) *server.Unit {
	unit := s.placeUnitAt(from, unitID, atQ, atR)
	if unit == nil {
		return unit
	}

	unit.IsOpponent = true
	return unit
}

// pickUnitByTemplateID finds the first unit with the given template ID,
// removes it from the slice and returns it alongside the updated slice.
// Calls log.Fatalf if no unit with the given ID is found.
func pickUnitByTemplateID(units []*server.Unit, id int) (*server.Unit, []*server.Unit) {
	for i, u := range units {
		if u.TemplateID == id {
			newUnits := make([]*server.Unit, 0, len(units)-1)
			newUnits = append(newUnits, units[:i]...)
			newUnits = append(newUnits, units[i+1:]...)

			return u, newUnits
		}
	}

	log.Fatalf("unit with template id %d not found", id)
	return nil, units
}
