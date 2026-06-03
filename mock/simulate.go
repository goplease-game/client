package mock

import (
	"encoding/json"
	"log"
	"math"

	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/hex"
)

type SimAction struct {
	Action string
	JSON   json.RawMessage
	data   any
}

const (
	UnitMovedAction = "unit_moved"
	UseAbility      = "use_ability"
	ApplyState      = "apply_state"
)

// SimulateUnitTurn ...
func SimulateUnitTurn(unit *ds.Unit) (acts []SimAction) {
	defer func() {
		for i, a := range acts {
			var err error
			a.JSON, err = json.Marshal(a.data)
			if err != nil {
				log.Fatalf("[SimulateUnitTurn] json.Marshal: %s", err.Error())
			}
			acts[i] = a
		}
	}()

	acts = simulateSimpleAttack(unit)
	if len(acts) > 0 {
		return
	}

	// can't attack, try to move
	act := simulateMove(unit)
	acts = append(acts, act)
	return
}

func simulateSimpleAttack(u *ds.Unit) (data []SimAction) {
	var ab *ability.Ability
	for _, id := range u.Abilities {
		if a := ability.ByID(id); a.IsBasicAttack() {
			ab = &a
		}
	}

	if ab == nil {
		return
	}

	enemy := findClosestEnemy(u)
	if enemy == nil {
		return
	}

	moveTo, ok := findAttackPosition(u, enemy, ab.Range)
	if !ok {
		return
	}

	if moveTo != u.Pos {
		PlaceUnitAt(u, moveTo)
		data = append(data, SimAction{
			Action: UnitMovedAction,
			data: ds.UnitMovedPayload{
				UnitID: u.ID,
				Coord:  moveTo,
			},
		})
	}

	states, err := HandleAbility(ds.UseAbilityPayload{
		UnitID:    u.ID,
		AbilityID: ab.ID,
		Target:    enemy.Pos,
	})
	if err != nil {
		log.Fatalf("[SimulateUnitTurn] HandleAbility: %s", err.Error())
	}

	data = append(data, SimAction{
		Action: UseAbility,
		data: ds.UseAbilityPayload{
			UnitID:    u.ID,
			AbilityID: ab.ID,
			Target:    enemy.Pos,
		},
	}, SimAction{
		Action: ApplyState,
		data:   states,
	})

	return data
}

func findAttackPosition(u *ds.Unit, target *ds.Unit, attackRange int) (ds.HexCoord, bool) {
	if hex.Distance(u.Pos, target.Pos) <= attackRange {
		return u.Pos, true
	}

	var bestPos ds.HexCoord
	bestMove := math.MaxInt

	for coord, cell := range gameState.Board.Cells {
		if cell.Unit != nil {
			continue
		}

		moveDist := hex.Distance(u.Pos, coord)
		if moveDist > u.CurrentMP {
			continue
		}

		if hex.Distance(coord, target.Pos) > attackRange {
			continue
		}

		if moveDist < bestMove {
			bestMove = moveDist
			bestPos = coord
		}
	}

	if bestMove == math.MaxInt {
		return ds.HexCoord{}, false
	}

	return bestPos, true
}

func findClosestEnemy(u *ds.Unit) *ds.Unit {
	var closest *ds.Unit
	bestDist := math.MaxInt

	for _, cell := range gameState.Board.Cells {
		if cell.Unit == nil {
			continue
		}

		enemy := cell.Unit
		// mocked unit have enemy.IsOpponent: true
		if enemy.IsOpponent || enemy.IsDead {
			continue
		}

		dist := hex.Distance(u.Pos, enemy.Pos)
		if dist < bestDist {
			bestDist = dist
			closest = enemy
		}
	}

	return closest
}

func simulateMove(u *ds.Unit) SimAction {
	pos := RandomReachableCell(*u)
	PlaceUnitAt(u, pos)

	return SimAction{
		Action: UnitMovedAction,
		data: ds.UnitMovedPayload{
			UnitID: u.ID,
			Coord:  pos,
		},
	}
}
