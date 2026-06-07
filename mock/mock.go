package mock

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"path"

	"github.com/google/uuid"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/mock/scenario"
)

var UnitsPerPlacementPhase = 3

const (
	totalUnitsPerPlayer = 6
	MockedPlayerID      = "p2"
	NoActiveUnit        = -1
	safeZoneSize        = 2
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

	Phase                  RoundPhase
	UnitsPerPlacementPhase int
	GameOver               bool
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
		newUnit.IsOpponent = true

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
		RoomID:                 data.RoomID,
		Board:                  data.Board,
		Players:                [2]*ds.Player{p1, p2},
		UnitsQueue:             []*ds.Unit{},
		CurrentRound:           1,
		ActivePlayer:           0,
		ActiveUnit:             NoActiveUnit, // when out of bound - start new round
		UnitsPerPlacementPhase: UnitsPerPlacementPhase,
	}

	return gameState
}

func CheckGameOver() (over bool, playerIdx int) {
	if len(gameState.Players[0].Units) > 0 {
		return
	}

	if len(gameState.Players[1].Units) > 0 {
		return
	}

	p1Units, p2Units := 0, 0
	for _, u := range gameState.UnitsQueue {
		if u.OwnerID == gameState.Players[0].ID {
			p1Units++
			continue
		}

		p2Units++
	}

	if p2Units == 0 {
		playerIdx = 1
	}

	if p1Units == 0 || p2Units == 0 {
		gameState.GameOver = true
	}

	return gameState.GameOver, playerIdx
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
		RoomID:                 snap.RoomID,
		Board:                  snap.Board,
		Players:                [2]*ds.Player{&p1, p2},
		UnitsQueue:             snap.UnitsQueue,
		CurrentRound:           snap.Round,
		ActiveUnit:             activeUnitIdx,
		ActivePlayer:           0,
		UnitsPerPlacementPhase: UnitsPerPlacementPhase,
	}

	fmt.Printf("[mock] new game state loaded from %s\n", name)
	return gameState
}

func LoadScenario(name scenario.Name) ds.GameSnapshot {
	sc := scenario.Load(name)

	if sc.P1 == nil {
		sc.P1 = &ds.Player{}
	}
	if sc.P2 == nil {
		sc.P2 = &ds.Player{}
	}

	// find active unit index
	activeUnit := NoActiveUnit
	if sc.ActiveUnitID != "" {
		for i := range sc.Queue {
			if sc.Queue[i].ID == sc.ActiveUnitID {
				activeUnit = i
			}
		}
	}

	// server state - keep as is
	gameState = &GameState{
		RoomID:                 sc.ID,
		Board:                  sc.Board,
		Players:                [2]*ds.Player{sc.P1, sc.P2},
		UnitsQueue:             sc.Queue,
		CurrentRound:           1,
		ActiveUnit:             activeUnit,
		ActivePlayer:           0,
		UnitsPerPlacementPhase: UnitsPerPlacementPhase,
	}

	sc2 := scenario.Copy(sc)

	snap := ds.GameSnapshot{
		RoomID:          sc2.ID,
		Board:           sc2.Board,
		Player:          *sc2.P1,
		OpponentName:    "Richard To Blame",
		UnitsQueue:      sc2.Queue,
		ActiveUnitID:    sc2.ActiveUnitID,
		Round:           1,
		TurnTimeSeconds: 0,
	}

	return snap
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
	// Find the maximum Q on the board — opponent's safe zone mirrors the player's.
	maxQ := math.MinInt
	for coord := range gameState.Board.Cells {
		if coord.Q > maxQ {
			maxQ = coord.Q
		}
	}

	var empty []ds.HexCoord
	for coord, cell := range gameState.Board.Cells {
		if cell.Unit != nil {
			continue
		}
		if coord.Q >= maxQ-safeZoneSize+1 {
			empty = append(empty, coord)
		}
	}

	if len(empty) == 0 {
		log.Fatal("[mock] no empty cells in opponent safe zone")
	}

	return empty[rand.IntN(len(empty))]
}

func PlaceUnitAt(u *ds.Unit, coord ds.HexCoord) {
	if cell := gameState.Board.Cells[coord]; cell != nil {
		u.Pos = coord
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
		if u.Pos == pos {
			return u
		}
	}

	return nil
}

func RemoveUnitFromQueue(unitID string) {
	for i, u := range gameState.UnitsQueue {
		if u.ID == unitID {
			gameState.UnitsQueue = append(gameState.UnitsQueue[:i], gameState.UnitsQueue[i+1:]...)
			break
		}
	}
}

func HandleEndTurn() (st ds.ApplyStates) {
	unit := ActiveUnit()
	if unit == nil {
		return
	}

	// decrease status duration
	for t, sv := range unit.Statuses {
		if sv.Duration == status.Permanent {
			continue
		}

		// before status removed, trigger onTurnEnd
		h, ok := statusHandlers[t]
		if ok && h != nil && h.onTurnEnd != nil {
			st.Add(h.onTurnEnd(unit, sv)...)
		}

		sv.Duration--
		if sv.Duration < 1 {
			st.Add(removeStatusFromUnit(t, unit)...)
			continue
		} else {
			st.Add(ds.ApplyState{
				SetStatusDuration: map[status.Type]int{t: sv.Duration},
				ToUnitID:          unit.ID,
			})
		}

		unit.Statuses[t] = sv
	}

	// reduce ability cooldowns
	for abID, cd := range unit.Cooldowns {
		if cd > 0 {
			cd--
			unit.SetCooldown(abID, cd)
			st.Add(ds.ApplyState{SetCooldown: new(map[ability.ID]int{abID: cd}), ToUnitID: unit.ID})
		}
	}

	// shield always decreased by 1 every turn
	if unit.CurrentShield > 0 {
		unit.CurrentShield--
		st.Add(
			ds.ApplyState{ChangeShield: new(-1), ToUnitID: unit.ID},
			ds.ApplyState{SetShield: new(unit.CurrentShield), ToUnitID: unit.ID},
		)
	}

	return st
}

func ActiveUnit() *ds.Unit {
	if gameState.ActiveUnit < 0 || gameState.ActiveUnit >= len(gameState.UnitsQueue) {
		return nil
	}

	return gameState.UnitsQueue[gameState.ActiveUnit]
}
