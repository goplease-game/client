package mock

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"path"

	"github.com/google/uuid"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

var UnitsPerPlacementPhase = 3

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
			newUnit.Abilities = make([]ds.AbilityID, len(unit.Abilities))
			copy(newUnit.Abilities, unit.Abilities)
		}

		if unit.Cooldowns != nil {
			newUnit.Cooldowns = make(map[ds.AbilityID]int)
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
	for _, row := range snap.Board {
		for _, cell := range row {
			if cell != nil && cell.Unit != nil && cell.Unit.IsOpponent {
				onBoard = append(onBoard, cell.Unit)
			}
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
	for i, unitID := range snap.UnitsQueue {
		if unitID == snap.ActiveUnitID {
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
		UnitsQueue:   unitsQueueFromIDs(snap.UnitsQueue, snap.Board),
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

// unitsQueueFromIDs rebuilds the queue as []*ds.Unit from the stored IDs,
// looking up each unit on the board.
func unitsQueueFromIDs(ids []string, board ds.Board) []*ds.Unit {
	index := make(map[string]*ds.Unit)
	for _, row := range board {
		for _, cell := range row {
			if cell != nil && cell.Unit != nil {
				index[cell.Unit.ID] = cell.Unit
			}
		}
	}

	var queue []*ds.Unit
	for _, id := range ids {
		if u, ok := index[id]; ok {
			queue = append(queue, u)
		}
	}
	return queue
}

func GetGameState() *GameState {
	return gameState
}

func LoadData(filename string) ([]byte, error) {
	filename = path.Join("data", filename)

	return data.ReadFile(filename)
}

func GetRandomUnoccupiedSafeZoneCell() (row, col int) {
	rows := len(gameState.Board)
	cols := len(gameState.Board[0])

	var emptyCells []struct{ r, c int }

	for c := cols - 2; c < cols; c++ {
		for r := range rows {
			if gameState.Board[r][c].Unit == nil {
				emptyCells = append(emptyCells, struct{ r, c int }{r, c})
			}
		}
	}

	if len(emptyCells) == 0 {
		log.Fatal("[mock] GetRandomUnoccupiedSafeZoneCell: no empty cells in safe zone")
	}

	cell := emptyCells[rand.IntN(len(emptyCells))]
	return cell.r, cell.c
}

func PlaceUnitAt(u *ds.Unit, row, col int) {
	gameState.Board[row][col].Unit = u
}

func GetUnitByID(id string) *ds.Unit {
	for _, u := range gameState.UnitsQueue {
		if u.ID == id {
			return u
		}
	}

	return nil
}

func RandomReachableCell(u ds.Unit) (row, col int) {
	cells := u.ReachableCells(gameState.Board)

	cell := cells[rand.IntN(len(cells))]
	return cell[0], cell[1]
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
