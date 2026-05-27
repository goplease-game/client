package mock

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"path"

	"github.com/google/uuid"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

var UnitsPerPlacementPhase = 1

const (
	totalUnitsPerPlayer = 6
	MockedPlayerID      = "p2"
	NoActiveUnit        = -1
)

type RoundPhase int

const (
	PlayPhase RoundPhase = iota
	PlacementPhase
)

//go:embed *
var data embed.FS

var gameState *GameState

type GameState struct {
	RoomID     string
	Board      ds.Board
	Players    [2]*ds.Player
	UnitsQueue []*ds.Unit
	ActiveUnit int

	CurrentRound int
	ActivePlayer int // 0 or 1 whose turn is

	Phase RoundPhase
}

func NewGameState(data ds.NewGamePayload) *GameState {
	p1 := data.Player

	p1Units := make([]ds.Unit, len(p1.Units))
	copy(p1Units, p1.Units)
	p1.Units = p1Units

	p2 := &ds.Player{
		ID:          uuid.NewString(),
		Name:        p1.Name,
		IsBot:       false,
		PlayerIndex: 1,
		Units:       make([]ds.Unit, len(p1.Units)),
	}

	for i, unit := range p1.Units {
		newUnit := unit
		newUnit.ID = uuid.NewString()
		newUnit.OwnerID = MockedPlayerID

		if unit.Abilities != nil {
			newUnit.Abilities = make([]ability.ID, len(unit.Abilities))
			copy(newUnit.Abilities, unit.Abilities)
		}

		if unit.Cooldowns != nil {
			newUnit.Cooldowns = make(map[ability.ID]int)
			for k, v := range unit.Cooldowns {
				newUnit.Cooldowns[k] = v
			}
		}

		p2.Units[i] = newUnit
	}

	gameState = &GameState{
		RoomID:       data.RoomID,
		Board:        data.Board,
		Players:      [2]*ds.Player{p1, p2},
		UnitsQueue:   []*ds.Unit{},
		CurrentRound: 1,
		ActivePlayer: 0,
		ActiveUnit:   NoActiveUnit, // when out of bound - start new round
	}

	return gameState
}

func RestoreGameState(name string, snap ds.GameSnapshot) *GameState {
	// Collect opponent units already on the board.
	var onBoard []*ds.Unit

	for _, cell := range snap.Board.Cells {
		if cell != nil && cell.Unit != nil && cell.Unit.IsOpponent {
			onBoard = append(onBoard, cell.Unit)
		}
	}

	// Load the full initial unit list to figure out what's still in hand.
	var inHand []ds.Unit
	if len(onBoard) < totalUnitsPerPlayer {
		initialUnits := loadInitialUnits()
		inHand = unitsNotOnBoard(initialUnits, onBoard)
	}

	p2 := &ds.Player{
		ID:          uuid.NewString(),
		Name:        snap.OpponentName,
		IsBot:       false,
		PlayerIndex: 1,
		Units:       inHand,
	}

	activeUnitIdx := NoActiveUnit
	for i, unit := range snap.UnitsQueue {
		if unit.ID == snap.ActiveUnitID {
			activeUnitIdx = i
			break
		}
	}

	p1 := snap.Player
	p1Units := make([]ds.Unit, len(snap.Player.Units))
	copy(p1Units, snap.Player.Units)
	p1.Units = p1Units

	gameState = &GameState{
		RoomID:       snap.RoomID,
		Board:        snap.Board,
		Players:      [2]*ds.Player{&p1, p2},
		UnitsQueue:   snap.UnitsQueue,
		CurrentRound: snap.Round,
		ActiveUnit:   activeUnitIdx,
		ActivePlayer: 0,
	}

	fmt.Printf("[mock] new game state loaded from %s\n", name)
	return gameState
}

// loadInitialUnits reads new_game.json and returns the full unit list
// that each player starts with.
func loadInitialUnits() []ds.Unit {
	raw, err := LoadData("new_game.json")
	if err != nil {
		log.Fatal("RestoreGameState: failed to load new_game.json:", err)
	}
	var payload ds.NewGamePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		log.Fatal("RestoreGameState: failed to parse new_game.json:", err)
	}
	return payload.Player.Units
}

// unitsNotOnBoard returns units from the initial list whose TemplateID
// is not represented among the opponent's on-board units.
func unitsNotOnBoard(initial []ds.Unit, onBoard []*ds.Unit) []ds.Unit {
	onBoardTemplates := make(map[int]bool, len(onBoard))
	for _, u := range onBoard {
		onBoardTemplates[u.TemplateID] = true
	}

	var result []ds.Unit
	for _, u := range initial {
		if !onBoardTemplates[u.TemplateID] {
			newUnit := u
			newUnit.ID = uuid.NewString()
			newUnit.OwnerID = MockedPlayerID
			result = append(result, newUnit)
		}
	}

	return result
}

func GetGameState() *GameState {
	return gameState
}

func LoadData(filename string) ([]byte, error) {
	filename = path.Join("data", filename)

	return data.ReadFile(filename)
}

func GetRandomUnoccupiedOpponentSafeZoneCell() ds.HexCoord {
	// Find max Q per row.
	maxQPerRow := make(map[int]int)
	for coord := range gameState.Board.Cells {
		if q, ok := maxQPerRow[coord.R]; !ok || coord.Q > q {
			maxQPerRow[coord.R] = coord.Q
		}
	}

	// Collect empty cells from the two rightmost columns per row.
	var empty []ds.HexCoord
	for coord, cell := range gameState.Board.Cells {
		if cell.Unit != nil {
			continue
		}
		maxQ := maxQPerRow[coord.R]
		if coord.Q != maxQ && coord.Q != maxQ-1 {
			continue
		}
		empty = append(empty, coord)
	}

	if len(empty) == 0 {
		log.Fatal("[mock] no empty cells in opponent safe zone")
	}

	return empty[rand.IntN(len(empty))]
}

func PlaceUnitAt(u *ds.Unit, coord ds.HexCoord) {
	if cell := gameState.Board.Cells[coord]; cell != nil {
		cell.Unit = u
	}
}

func GetUnitByID(id string) *ds.Unit {
	for _, u := range gameState.UnitsQueue {
		if u.ID == id {
			return u
		}
	}

	return nil
}

func RandomReachableCell(u ds.Unit) ds.HexCoord {
	cells := u.ReachableCells(gameState.Board)

	cell := cells[rand.IntN(len(cells))]
	return cell
}

func PickRandomUnitOfFromHandP2() *ds.Unit {
	units := gameState.Players[1].Units
	count := len(units)
	if count == 0 {
		return nil
	}

	idx := rand.IntN(count)
	pickedUnit := units[idx]

	units[idx] = units[count-1]
	units = units[:count-1]

	gameState.Players[1].Units = units

	return &pickedUnit
}

func PickUnitFromHandByTemplateP1(templateID int) *ds.Unit {
	units := gameState.Players[0].Units
	var pickedUnit *ds.Unit
	var idx int

	for i, u := range units {
		if u.TemplateID == templateID {
			pickedUnit = &u
			idx = i
			break
		}
	}

	if pickedUnit == nil {
		log.Fatalf("[mock] PickUnitFromHandByTemplateP1: no unit in hand with templateID: %d", templateID)
	}

	units[idx] = units[len(units)-1]
	units = units[:len(units)-1]

	gameState.Players[0].Units = units
	return pickedUnit
}

func AddUnitToQueue(u *ds.Unit) {
	gameState.UnitsQueue = append(gameState.UnitsQueue, u)
}

func GetUnitAt(pos ds.HexCoord) *ds.Unit {
	for _, u := range gameState.UnitsQueue {
		if u.Pos.Q == pos.Q && u.Pos.R == pos.R {
			return u
		}
	}

	return nil
}
