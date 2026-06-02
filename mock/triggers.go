package mock

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/hex"
)

func init() {
	onDamageReceivedHandlers = []onDamageReceivedHandler{
		useCoverFireAbility,
		useOpportunityAbility,
		useBottomlessVialAbility,
	}
}

type onDeathHandler func(u *ds.Unit) ds.ApplyStates
type onMoveHandler func(u *ds.Unit) ds.ApplyStates
type onDamageReceivedHandler func(source, target *ds.Unit) ds.ApplyStates

var onDeathHandlers = []onDeathHandler{
	useUndyingWillAbility,
}

var onMoveHandlers = []onMoveHandler{
	recalculateFrenzyAbility,
}

var onDamageReceivedHandlers []onDamageReceivedHandler

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

func applyOnDamageReceivedHandlers(source, target *ds.Unit) (st ds.ApplyStates) {
	for _, handler := range onDamageReceivedHandlers {
		st.Add(handler(source, target)...)
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
		ds.ApplyState{UseAbility: new(ds.UseAbilityPayload{UnitID: u.ID, AbilityID: id}), ToUnitID: u.ID},
		ds.ApplyState{ChangeHP: new(u.CurrentHP), ToUnitID: u.ID},
		ds.ApplyState{SetHP: new(u.CurrentHP), ToUnitID: u.ID},
		ds.ApplyState{ChangeShield: new(u.CurrentShield), ToUnitID: u.ID},
		ds.ApplyState{SetShield: new(u.CurrentShield), ToUnitID: u.ID},
	)

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

func useCoverFireAbility(source, _ *ds.Unit) (st ds.ApplyStates) {
	id := ability.CoverFire
	ab := ability.ByID(id)

	unitsWithCoverFire := findEnemiesInRangeWithAbility(source, ab.Range, id)
	for _, u := range unitsWithCoverFire {
		if !u.AbilityReady(id) {
			continue
		}

		u.Cooldowns[id] = ab.Cooldown
		st.Add(ds.ApplyState{UseAbility: new(ds.UseAbilityPayload{
			UnitID:    u.ID,
			AbilityID: id,
			Target:    source.Pos,
		}), ToUnitID: u.ID})
		st.Add(dealDamageToUnit(u, source, 3)...)
	}

	return
}

func useOpportunityAbility(source, target *ds.Unit) (st ds.ApplyStates) {
	if hex.Distance(source.Pos, target.Pos) > 1 { // only melee attacks
		return
	}

	id := ability.Opportunity
	ab := ability.ByID(id)

	unitsWithOpportunity := findEnemiesInRangeWithAbility(target, ab.Range, id)
	for _, u := range unitsWithOpportunity {
		if u.ID == source.ID { // cannot have opportunity for your own attack
			continue
		}
		if !u.AbilityReady(id) {
			continue
		}

		u.Cooldowns[id] = ab.Cooldown
		st.Add(ds.ApplyState{UseAbility: new(ds.UseAbilityPayload{
			UnitID:    u.ID,
			AbilityID: id,
			Target:    target.Pos,
		}), ToUnitID: u.ID})
		st.Add(dealDamageToUnit(u, source, u.CurrentAtk)...)
	}

	return
}

// TODO apply status to display how much max HP increased
func useBottomlessVialAbility(_, target *ds.Unit) (st ds.ApplyStates) {
	id := ability.BottomlessVial
	if !target.HasAbility(id) {
		return
	}
	if !target.AbilityReady(id) {
		return
	}

	target.Cooldowns[id] = ability.ByID(id).Cooldown
	target.BaseHP++

	st.Add(ds.ApplyState{UseAbility: new(ds.UseAbilityPayload{
		UnitID:    target.ID,
		AbilityID: id,
		Target:    target.Pos,
	}), ToUnitID: target.ID})
	st.Add(ds.ApplyState{SetBaseHP: new(target.BaseHP), ToUnitID: target.ID})
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

func findEnemiesInRangeWithAbility(u *ds.Unit, radius int, abID ability.ID) []*ds.Unit {
	enemies := []*ds.Unit{}

	cells := hex.CellsInRange(u.Pos, radius, gameState.Board)
	for _, c := range cells {
		if unit := GetUnitAt(c); unit != nil && unit.OwnerID != u.OwnerID && unit.HasAbility(abID) {
			enemies = append(enemies, unit)
		}
	}

	return enemies
}
