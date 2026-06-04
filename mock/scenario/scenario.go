package scenario

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const Default = BasWithFriends

var Scenarios = map[Name]func() *Scenario{}

func addScenario(name Name, scenario func() *Scenario) {
	Scenarios[name] = scenario
}

const (
	BasID    = 1
	GritID   = 2
	FletchID = 3
	SilverID = 4
	MistID   = 5
	JulyID   = 6
)

//go:embed units.json
var unitsJSON []byte

type Name string

// Load returns a new Scenario instance for the given name.
// Panics if name is not registered in Scenarios.
func Load(name Name) *Scenario {
	return Scenarios[name]()
}

type Scenario struct {
	ID           string
	P1           *ds.Player
	P2           *ds.Player
	Queue        []*ds.Unit
	Board        ds.Board
	ActiveUnitID string
}

// placeUnitAt picks a unit by templateID from player's hand,
// sets its position and places it on the board cell at coord.
// The unit is also appended to the scenario queue.
func (s *Scenario) placeUnitAt(from *ds.Player, unitID, atQ, atR int) *ds.Unit {
	var unit *ds.Unit
	unit, from.Units = pickUnitByTemplateID(from.Units, unitID)
	at := ds.HexCoord{Q: atQ, R: atR}
	unit.Pos = at
	s.Board.Cells[at].Unit = unit
	s.Queue = append(s.Queue, unit)

	return unit
}

// newEmptyRectBoard creates a rectangular board of w columns and h rows
// in offset hex coordinates. All cells are marked as safe zones.
func newEmptyRectBoard(w, h int) ds.Board {
	b := ds.Board{
		Cells: make(map[ds.HexCoord]*ds.BoardCell),
	}

	for r := 0; r < h; r++ {
		qOffset := r / 2
		for q := -qOffset; q < w-qOffset; q++ {
			coord := ds.HexCoord{
				Q: q,
				R: r,
			}

			b.Cells[coord] = &ds.BoardCell{
				Coord:      coord,
				IsSafeZone: true,
			}
		}
	}

	return b
}

// loadUnits loads all units from the embedded units.json,
// assigns ownerID and opponent flag to each.
// If ids are provided, only units matching those template IDs are returned.
func loadUnits(ownerID string, isOpp bool, ids ...int) []ds.Unit {
	var units []ds.Unit
	if err := json.Unmarshal(unitsJSON, &units); err != nil {
		log.Fatalf("Error loading units: %v", err)
		return nil
	}

	for i := range units {
		units[i].ID = ownerID + fmt.Sprintf("-%d", units[i].TemplateID)
		units[i].OwnerID = ownerID
		units[i].IsOpponent = isOpp
	}

	if len(ids) == 0 {
		return units
	}

	idMap := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idMap[id] = struct{}{}
	}

	selected := make([]ds.Unit, 0, len(ids))
	for _, u := range units {
		if _, found := idMap[u.TemplateID]; found {
			selected = append(selected, u)
		}
	}

	return selected
}

// pickUnitByTemplateID finds the first unit with the given template ID,
// removes it from the slice and returns it alongside the updated slice.
// Calls log.Fatalf if no unit with the given ID is found.
func pickUnitByTemplateID(units []ds.Unit, id int) (*ds.Unit, []ds.Unit) {
	for i, u := range units {
		if u.TemplateID == id {
			newUnits := append(units[:i], units[i+1:]...)
			return &u, newUnits
		}
	}

	log.Fatalf("unit with template id %d not found", id)
	return nil, units
}

// Copy returns a deep copy of the scenario via JSON marshal/unmarshal round-trip.
// Calls log.Fatalf if marshaling or unmarshaling fails.
func Copy(src *Scenario) *Scenario {
	bytes, err := json.Marshal(src)
	if err != nil {
		log.Fatalf("failed to marshal scenario: %v", err)
	}

	var dst Scenario
	err = json.Unmarshal(bytes, &dst)
	if err != nil {
		log.Fatalf("failed to unmarshal scenario: %v", err)
	}

	return &dst
}

// NewSimpleScenario creates a minimal 6×6 scenario with a single player (P1)
// and no opponent or queue. Useful as a blank slate for testing placement logic.
func NewSimpleScenario() *Scenario {
	board := newEmptyRectBoard(6, 6)
	q := []*ds.Unit{}

	p1 := &ds.Player{
		ID:                   "p1",
		Name:                 "P1",
		IsBot:                false,
		PlayerIndex:          0,
		UnitsPlacedThisRound: 0,
	}
	p1.Units = loadUnits(p1.ID, false)

	return &Scenario{
		ID:    uuid.NewString(),
		P1:    p1,
		P2:    nil,
		Queue: q,
		Board: board,
	}
}
