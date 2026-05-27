package mock

import (
	"fmt"

	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

var abilityHandlers = map[ability.ID]func(ds.UseAbilityPayload) (ds.ApplyState, error){
	ability.BasicMeleeAttack: basicMeleeAttackHandler,
}

// HandleAbility is cooking a response for specific ability. We not dont validation here,
// because it is just a mock implementation, so you can hack whatever you want.
func HandleAbility(data ds.UseAbilityPayload) (resp ds.ApplyState, err error) {
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

func basicMeleeAttackHandler(load ds.UseAbilityPayload) (ds.ApplyState, error) {
	caster := GetUnitByID(load.UnitID)
	target := GetUnitAt(load.Target)

	st := ds.ApplyState{}
	st.ChangeHP = new(caster.CurrentAtk)

	_ = target

	return st, nil
}
