package mock

import (
	"embed"
	"log"
	"math/rand/v2"
	"path"

	"github.com/google/uuid"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

const (
	MockedPlayerID = "p2"
	NoActiveUnit   = -1
)

//go:embed *
var data embed.FS

var gameState *GameState

type GameState struct {
	RoomID     string
	Board      ds.Board
	Players    [2]*ds.Player
	UnitsQueue []*ds.Unit
	ActiveUnit int // counting from 1

	CurrentRound int
	ActivePlayer int // 0 or 1 whose turn is
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
	gameState.ActiveUnit = len(gameState.UnitsQueue) + 1
}
