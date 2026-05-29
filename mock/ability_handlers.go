package mock

import (
	"fmt"
	"log"

	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ability/effect"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/hex"
)

var abilityHandlers = map[ability.ID]func(ds.UseAbilityPayload) ([]ds.ApplyState, error){
	ability.BasicMeleeAttack: basicMeleeAttackHandler,
	ability.BasicRangeAttack: basicRangeAttackHandler,
	ability.BasicMagicAttack: basicMagicAttackHandler,

	ability.Fortify: fortifyHandler,
	ability.Provoke: provokeHandler,
}

// HandleAbility is cooking a response for specific ability. We not dont validation here,
// because it is just a mock implementation, so you can hack whatever you want.
func HandleAbility(data ds.UseAbilityPayload) (resp []ds.ApplyState, err error) {
	_, ok := ability.Abilities[data.AbilityID]
	if !ok {
		err = fmt.Errorf("[mock] invalid ability: %s", data.AbilityID)
		return
	}

	handler, ok := abilityHandlers[data.AbilityID]
	if !ok {
		err = fmt.Errorf("[mock] no handler for ability: %s", data.AbilityID)
		return
	}

	return handler(data)
}

func basicMeleeAttackHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster := GetUnitByID(load.UnitID)
	target := GetUnitAt(load.Target)

	if caster == nil {
		log.Fatalf("invalid ability caster: %s", load.UnitID)
	}
	if target == nil {
		log.Fatalf("invalid ability target: %s", load.Target)
	}

	target.CurrentHP = target.CurrentHP - caster.CurrentAtk

	st1 := ds.ApplyState{ChangeHP: new(-caster.CurrentAtk), ToUnitID: target.ID}
	st2 := ds.ApplyState{SetHP: new(target.CurrentHP), ToUnitID: target.ID}

	st := []ds.ApplyState{st1, st2}
	if target.CurrentHP <= 0 {
		target.IsDead = true
		for i, u := range gameState.UnitsQueue {
			if u.ID == target.ID {
				gameState.UnitsQueue = append(gameState.UnitsQueue[:i], gameState.UnitsQueue[i+1:]...)
				break
			}
		}
		st = append(st, ds.ApplyState{IsDead: true, ToUnitID: target.ID})
	}

	return st, nil
}

func basicRangeAttackHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	return basicMeleeAttackHandler(load)
}

func basicMagicAttackHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	return basicMeleeAttackHandler(load)
}

func fortifyHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster := GetUnitByID(load.UnitID)
	if caster == nil {
		log.Fatalf("invalid ability caster: %s", load.UnitID)
	}

	ab := ability.Abilities[ability.Fortify]
	ef := effect.StatusByType(effect.DecayingShield)

	st := []ds.ApplyState{
		{ChangeShield: new(ef.InitialValue), ToUnitID: caster.ID},
		{SetShield: new(caster.CurrentShield+ef.InitialValue), ToUnitID: caster.ID},
	}

	cells := hex.CellsInRange(caster.Pos, ab.AreaRadius, gameState.Board)

	for _, c := range cells {
		if u := GetUnitAt(c); u != nil && !u.IsOpponent {
			st = append(st, ds.ApplyState{ChangeShield: new(ef.InitialValue), ToUnitID: u.ID})
			st = append(st, ds.ApplyState{SetShield: new(u.CurrentShield+ef.InitialValue), ToUnitID: u.ID})
		}
	}

	return st, nil
}

func provokeHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster := GetUnitByID(load.UnitID)
	if caster == nil {
		log.Fatalf("invalid ability caster: %s", load.UnitID)
	}

	ab := ability.Abilities[ability.Provoke]

	st := []ds.ApplyState{{
		AddStatus: new(ds.StatusWithMeta{Status: effect.Provoking}),
		ToUnitID:  caster.ID,
	}}

	ste := new(ds.StatusWithMeta{Status: effect.Provoked, Meta: map[string]any{
		"provoker": caster.ID,
	}})

	cells := hex.CellsInRange(caster.Pos, ab.AreaRadius, gameState.Board)

	for _, c := range cells {
		if u := GetUnitAt(c); u != nil && u.IsOpponent {
			st = append(st, ds.ApplyState{AddStatus: ste, ToUnitID: u.ID})
		}
	}

	return st, nil
}
