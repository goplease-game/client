package mock

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/hex"
)

type onDeathHandler func(u *ds.Unit) ds.ApplyStates
type onMoveHandler func(u *ds.Unit) ds.ApplyStates

var onDeathHandlers = []onDeathHandler{
	useUndyingWillAbility,
}

var onMoveHandlers = []onMoveHandler{
	recalculateFrenzyAbility,
}

func applyOnDeathHandlers(u *ds.Unit) (st ds.ApplyStates) {
	for _, handler := range onDeathHandlers {
		st.Add(handler(u)...)
	}

	return
}

func ApplyOnMoveHandlers(u *ds.Unit) (st ds.ApplyStates) {
	for _, handler := range onMoveHandlers {
		st.Add(handler(u)...)
	}

	return
}

func useUndyingWillAbility(u *ds.Unit) (st ds.ApplyStates) {
	id := ability.UndyingWill
	if !u.HasAbility(id) {
		return
	}

	if u.Cooldowns[id] > 0 {
		return
	}

	u.CurrentHP = 1
	u.CurrentShield = 5
	u.IsDead = false

	ab := ability.ByID(id)
	u.Cooldowns[id] = ab.Cooldown

	st.Add(
		ds.ApplyState{UseAbility: new(ds.UseAbilityPayload{UnitID: u.ID, AbilityID: id})},
		ds.ApplyState{ChangeHP: new(u.CurrentHP)},
		ds.ApplyState{SetHP: new(u.CurrentHP)},
		ds.ApplyState{ChangeShield: new(u.CurrentShield)},
		ds.ApplyState{SetShield: new(u.CurrentShield)},
	)
	st.ToUnitID(u.ID)

	return
}

func recalculateFrenzyAbility(_ *ds.Unit) (st ds.ApplyStates) {
	id := ability.Frenzy
	ab := ability.ByID(id)

	for _, u := range gameState.UnitsQueue {
		if !u.HasAbility(id) {
			continue
		}

		enemies := countEnemiesInRange(u, ab.AreaRadius, 2)
		hasFrenzy := u.HasStatus(status.Frenzied)

		// Remove
		if enemies < 2 && hasFrenzy {
			u.CurrentAtk--

			st.Add(
				removeStatusFromUnit(u, status.Frenzied),
				ds.ApplyState{ChangeAtk: new(-1), ToUnitID: u.ID},
				ds.ApplyState{SetAtk: new(u.CurrentAtk), ToUnitID: u.ID},
			)

			continue
		}

		// ADD
		if enemies >= 2 && !hasFrenzy {
			u.CurrentAtk++

			st.Add(
				applyStatusToUnit(u, status.Frenzied),
				ds.ApplyState{ChangeAtk: new(1), ToUnitID: u.ID},
				ds.ApplyState{SetAtk: new(u.CurrentAtk), ToUnitID: u.ID},
			)
		}
	}

	return
}

func countEnemiesInRange(u *ds.Unit, radius int, atLeastOpt ...int) (count int) {
	var atLeast int
	if len(atLeastOpt) > 0 {
		atLeast = atLeastOpt[0]
	}

	cells := hex.CellsInRange(u.Pos, radius, gameState.Board)
	for _, c := range cells {
		if unit := GetUnitAt(c); unit != nil && unit.OwnerID != u.OwnerID {
			count++
			if atLeast > 0 && count >= atLeast {
				break
			}
		}
	}

	return
}
