package mock

import (
	"fmt"
	"log"

	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ability/effect"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/hex"
)

const StatusPermaDuration = -1

var abilityHandlers = map[ability.ID]func(ds.UseAbilityPayload) ([]ds.ApplyState, error){
	ability.BasicMeleeAttack: basicMeleeAttackHandler,
	ability.BasicRangeAttack: basicRangeAttackHandler,
	ability.BasicMagicAttack: basicMagicAttackHandler,

	ability.Fortify:   fortifyHandler,
	ability.Provoke:   provokeHandler,
	ability.PowerPush: powerPushHandler,
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
	caster, target := mustAbilityActors(load)

	st := dealDamageToUnit(target, caster.CurrentAtk)
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

	st := []ds.ApplyState{}

	cells := hex.CellsInRange(caster.Pos, ab.AreaRadius, gameState.Board)

	for _, c := range cells {
		if u := GetUnitAt(c); u != nil && !u.IsOpponent {
			u.CurrentShield += ef.InitialValue
			st = append(st, ds.ApplyState{ChangeShield: new(ef.InitialValue), ToUnitID: u.ID})
			st = append(st, ds.ApplyState{SetShield: new(u.CurrentShield), ToUnitID: u.ID})
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

func powerPushHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster, target := mustAbilityActors(load)

	dealDmg := 3
	var canMove bool
	// skipping pos validation of caster & target, hacking is welcome
	pos := hex.OppositeHex(caster.Pos, target.Pos)
	cell, _ := gameState.Board.Cells[pos]
	if cell != nil && cell.Unit == nil {
		canMove = true
	}

	st := ds.NewUnitStates()
	if canMove {
		st.Add(ds.ApplyState{MoveTo: new(pos)})
	} else {
		dealDmg = 5
	}

	st.Add(dealDamageToUnit(target, dealDmg)...)
	st.ToUnitID(target.ID)
	return st, nil
}

func mustAbilityActors(load ds.UseAbilityPayload) (from *ds.Unit, to *ds.Unit) {
	from = GetUnitByID(load.UnitID)
	to = GetUnitAt(load.Target)

	if from == nil {
		log.Fatalf("invalid ability caster: %s", load.UnitID)
	}
	if to == nil {
		log.Fatalf("invalid ability target: %s", load.Target)
	}

	return
}

func dealDamageToUnit(u *ds.Unit, val int) []ds.ApplyState {
	st := ds.NewUnitStates()
	if u.CurrentShield > 0 {
		var shieldRemoved int
		if u.CurrentShield < val {
			shieldRemoved = u.CurrentShield
			u.CurrentShield = 0
			val = val - shieldRemoved
			st.Add(
				ds.ApplyState{ChangeShield: new(-shieldRemoved)},
				ds.ApplyState{SetShield: new(0)},
			)
		}
	}

	if u.CurrentHP < val {
		val = u.CurrentHP
	}

	u.CurrentHP -= val
	st.Add(
		ds.ApplyState{ChangeHP: new(-val)},
		ds.ApplyState{SetHP: new(u.CurrentHP)},
	)

	// TODO onDeath / onHPReduced triggers
	if u.CurrentHP == 0 {
		u.IsDead = true
		RemoveUnitFromQueue(u.ID)
		st.Add(ds.ApplyState{IsDead: true})
	}

	st.ToUnitID(u.ID)

	return st
}
