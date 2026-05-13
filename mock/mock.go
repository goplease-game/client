package mock

import (
	"embed"
	"math/rand"
	"path"

	"github.com/google/uuid"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

//go:embed *
var data embed.FS

var gameState *GameState

type GameState struct {
	RoomID     string
	Board      ds.Board
	Player1    *ds.Player
	Player2    *ds.Player
	UnitsQueue []*ds.Unit

	CurrentTurn  int
	ActivePlayer int // 0 or 1 whose turn is
	Phase        ds.Phase
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
		newUnit.OwnerID = "p2"

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
		Player1:      p1,
		Player2:      p2,
		UnitsQueue:   []*ds.Unit{},
		CurrentTurn:  1,
		ActivePlayer: 0,
		Phase:        data.Phase,
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

func GetRandomSafeZoneCell(rows int, cols int) (row, col int) {
	lastCol := cols - 1
	preLastCol := cols - 2

	selectedCol := preLastCol
	if rand.Intn(2) == 1 {
		selectedCol = lastCol
	}

	selectedRow := rand.Intn(rows)

	return selectedRow, selectedCol
}

func PickRandomUnit() *ds.Unit {
	units := gameState.Player2.Units
	count := len(gameState.Player2.Units)
	if count == 0 {
		return nil
	}

	idx := rand.Intn(count)
	pickedUnit := units[idx]

	units[idx] = units[len(units)-1]
	units = units[:len(units)-1]

	gameState.Player2.Units = units

	return &pickedUnit
}
