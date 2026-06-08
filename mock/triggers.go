package mock

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
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
type onTurnStarHandler func(unit *ds.Unit) ds.ApplyStates

var onDeathHandlers = []onDeathHandler{
	useUndyingWillAbility,
}

var onMoveHandlers = []onMoveHandler{
	recalculateFrenzyAbility,
}

var onDamageReceivedHandlers []onDamageReceivedHandler

var onTurnStarHandlers = []onMoveHandler{
	useFocusFieldAbility,
	recalculateFrenzyAbility,
	handleOnTurnStartStatuses,
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

func applyOnDamageReceivedHandlers(source, target *ds.Unit) (st ds.ApplyStates) {
	for _, handler := range onDamageReceivedHandlers {
		st.Add(handler(source, target)...)
	}

	return
}

func ApplyOnTurnStartHandlers(unit *ds.Unit) (st ds.ApplyStates) {
	unit.PhantomAPUsedThisTurn = 0

	for _, handler := range onTurnStarHandlers {
		st.Add(handler(unit)...)
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

func recalculateFrenzyAbility(_ *ds.Unit) (sts ds.ApplyStates) {
	id := ability.Frenzy
	ab := ability.ByID(id)

	for _, u := range gameState.UnitsQueue {
		if !u.HasAbility(id) {
			continue
		}

		enemies := countEnemiesInRange(u, ab.AreaRadius, 2)
		isFrenzied := u.HasStatus(ab.Effect.ApplyStatus)

		// Remove
		if enemies < 2 && isFrenzied {
			sts.Add(removeStatusFromUnit(ab.Effect.ApplyStatus, u)...)
			continue
		}

		// Add
		if enemies >= 2 && !isFrenzied {
			sts.Add(
				applyStatusToUnit(ab.Effect.ApplyStatus, u, u)...,
			)
		}
	}

	return
}

func useCoverFireAbility(source, target *ds.Unit) (st ds.ApplyStates) {
	if source.IsAlly(target) {
		return
	}

	id := ability.CoverFire
	ab := ability.ByID(id)

	unitsWithCoverFire := findEnemiesInRangeWithAbility(source, ab.Range, id)
	for _, u := range unitsWithCoverFire {
		if !u.AbilityReady(id) {
			continue
		}

		if target.ID == u.ID {
			continue // cannot apply CF from self
		}

		u.SetCooldown(id, ab.Cooldown)
		st.Add(ds.ApplyState{UseAbility: new(ds.UseAbilityPayload{
			UnitID:    u.ID,
			AbilityID: id,
			Target:    source.Pos,
		}), ToUnitID: u.ID})

		st.Add(dealDamageToUnit(u, source, ab.Effect.DealDamage)...)
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

		u.SetCooldown(id, ab.Cooldown)
		st.Add(ds.ApplyState{UseAbility: new(ds.UseAbilityPayload{
			UnitID:    u.ID,
			AbilityID: id,
			Target:    target.Pos,
		}), ToUnitID: u.ID})
		st.Add(dealDamageToUnit(u, target, u.CurrentAtk)...)
	}

	return
}

func useFocusFieldAbility(unit *ds.Unit) (st ds.ApplyStates) {
	id := ability.FocusField
	ab := ability.ByID(id)

	unitsWithFocusField := findAlliesInRangeWithAbility(unit, ab.Range, id)
	for _, u := range unitsWithFocusField {
		if u.ID == unit.ID { // cannot have focus field to yourself
			continue
		}

		var abUsed bool
		for abID, cd := range unit.Cooldowns {
			if ability.ByID(abID).IsPassive {
				continue
			}

			if cd > 0 {
				cd--
				unit.SetCooldown(abID, cd)
				abUsed = true
				st.Add(ds.ApplyState{SetCooldown: new(map[ability.ID]int{abID: cd}), ToUnitID: unit.ID})
			}
		}

		if abUsed {
			st.Add(ds.ApplyState{UseAbility: new(ds.UseAbilityPayload{
				UnitID:    u.ID,
				AbilityID: id,
				Target:    unit.Pos,
			}), ToUnitID: unit.ID})
		}

		return // trigger only once
	}

	return
}

// TODO apply status to display how much max HP increased
func useBottomlessVialAbility(_, target *ds.Unit) (st ds.ApplyStates) {
	id := ability.BottomlessVial
	ab := ability.ByID(id)

	units := findAlliesInRangeWithAbility(target, ab.AreaRadius, id)
	for _, u := range units {
		if !u.AbilityReady(id) {
			continue
		}

		if u.ID == target.ID {
			continue // cannot use on self
		}

		u.SetCooldown(id, ab.Cooldown)
		target.BaseHP += ab.Effect.AddHP

		st.Add(ds.ApplyState{UseAbility: new(ds.UseAbilityPayload{
			UnitID:    target.ID,
			AbilityID: id,
			Target:    target.Pos,
		}), ToUnitID: target.ID})
		st.Add(ds.ApplyState{SetBaseHP: new(target.BaseHP), ToUnitID: target.ID})
		st.Add(healUnit(target, ab.Effect.HealHP)...)

		return // apply only once
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

func findAlliesInRangeWithAbility(u *ds.Unit, radius int, abID ability.ID) []*ds.Unit {
	units := []*ds.Unit{}

	cells := hex.CellsInRange(u.Pos, radius, gameState.Board)
	for _, c := range cells {
		if unit := GetUnitAt(c); unit != nil && unit.OwnerID == u.OwnerID && unit.HasAbility(abID) {
			units = append(units, unit)
		}
	}

	return units
}
