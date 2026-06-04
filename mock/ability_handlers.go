package mock

import (
	"fmt"
	"log"

	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/hex"
)

var abilityHandlers = map[ability.ID]func(abilityUsed) (ds.ApplyStates, error){
	ability.BasicMeleeAttack: basicMeleeAttackHandler,
	ability.BasicRangeAttack: basicRangeAttackHandler,
	ability.BasicMagicAttack: basicMagicAttackHandler,

	ability.Fortify:    fortifyHandler,
	ability.Provoke:    provokeHandler,
	ability.ShieldBash: shieldBashHandler,

	ability.BattleCry:   battleCryHandler,
	ability.IdolihuSpin: idolihuSpinHandler,
	ability.PowerPush:   powerPushHandler,

	ability.PiercingShot:  piercingShotHandler,
	ability.HuntersMark:   huntersMarkHandler,
	ability.HamstringShot: hamstringShotHandler,

	ability.ShadowStep: shadowStepHandler,
	ability.GangUp:     gangUpHandler,
	ability.Eliminate:  eliminateHandler,

	ability.Translocation: translocationHandler,
	ability.TimeWarp:      timeWarpHandler,
	ability.Purge:         purgeHandler,

	ability.Heal:     healHandler,
	ability.Equalize: equalizeHandler,
	ability.Purify:   purifyHandler,
}

type abilityUsed struct {
	By *ds.Unit
	Ab ability.Ability
	At ds.HexCoord
}

// HandleAbility is cooking a response for specific ability. We don't do validation here,
// because it is just a mock implementation, so you can hack whatever you want.
func HandleAbility(data ds.UseAbilityPayload) (resp ds.ApplyStates, err error) {
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

	unit := GetUnitByID(data.UnitID)
	if unit == nil {
		err = fmt.Errorf("[mock] unit not found: %s", data.AbilityID)
		return
	}
	ab := ability.ByID(data.AbilityID)

	event := abilityUsed{
		By: unit,
		Ab: ab,
		At: data.Target,
	}

	resp, err = handler(event)
	if err != nil {
		return
	}

	unit.SetCooldown(ab.ID, ab.Cooldown)
	return
}

func basicMeleeAttackHandler(e abilityUsed) (ds.ApplyStates, error) {
	target := mustUnitAt(e.At)

	st := dealDamageToUnit(e.By, target, e.By.CurrentAtk)
	return st, nil
}

func basicRangeAttackHandler(e abilityUsed) (ds.ApplyStates, error) {
	return basicMeleeAttackHandler(e)
}

func basicMagicAttackHandler(e abilityUsed) (ds.ApplyStates, error) {
	return basicMeleeAttackHandler(e)
}

func fortifyHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	units := findAlliesInRange(e.By, e.Ab.AreaRadius)
	val := e.Ab.Effect.AddShield
	for _, u := range units {
		u.CurrentShield += val
		sts.Add(
			ds.ApplyState{ChangeShield: new(val), ToUnitID: u.ID},
			ds.ApplyState{SetShield: new(u.CurrentShield), ToUnitID: u.ID},
		)
	}

	return sts, nil
}

func provokeHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	sts.Add(applyStatusToUnit(status.Provoking, e.By, e.By)...)

	units := findEnemiesInRange(e.By, e.Ab.AreaRadius)
	for _, u := range units {
		sts.Add(applyStatusToUnit(status.Provoked, e.By, u)...)
	}

	return sts, nil
}

func shieldBashHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	target := mustUnitAt(e.At)
	sts.Add(applyStatusToUnit(e.Ab.Effect.ApplyStatus, e.By, target)...)

	return
}

func powerPushHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	target := mustUnitAt(e.At)

	dealDmg := e.Ab.Effect.DealDamage

	pos := hex.OppositeHex(e.By.Pos, target.Pos)
	cell, _ := gameState.Board.Cells[pos]
	if cell != nil && cell.Unit == nil {
		sts.Add(ds.ApplyState{MoveTo: new(pos), ToUnitID: target.ID})
		target.Pos = pos
	} else {
		dealDmg = e.Ab.Effect.DealAltDamage
	}

	sts.Add(dealDamageToUnit(e.By, target, dealDmg)...)
	return sts, nil
}

func gangUpHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	target := mustUnitAt(e.At)
	dealDmg := e.By.CurrentAtk
	pos := hex.OppositeHex(e.By.Pos, target.Pos)
	cell, _ := gameState.Board.Cells[pos]
	if cell != nil && cell.Unit != nil && cell.Unit.IsAlly(e.By) {
		dealDmg += e.Ab.Effect.BonusDamage
	}

	sts.Add(dealDamageToUnit(e.By, target, dealDmg)...)
	return
}

func eliminateHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	target := mustUnitAt(e.At)

	sts.Add(dealDamageToUnit(e.By, target, e.Ab.Effect.DealDamage)...)
	if target.IsDead {
		ap := e.Ab.Effect.AddAP
		sts.Add(
			ds.ApplyState{ChangeAP: new(ap), ToUnitID: e.By.ID},
			ds.ApplyState{SetAP: new(ap), ToUnitID: e.By.ID},
		)
	}

	return
}

func translocationHandler(e abilityUsed) (ds.ApplyStates, error) {
	target := mustUnitAt(e.At)

	// Swapping with self is a no-op and likely a bug — abort.
	if target.ID == e.By.ID {
		return nil, fmt.Errorf("translocation: cannot swap unit with itself")
	}

	from := e.By.Pos
	to := target.Pos

	e.By.Pos = to
	target.Pos = from

	return nil, nil
}

func timeWarpHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	target := mustUnitAt(e.At)

	sts.Add(
		applyStatusToUnit(e.Ab.Effect.ApplyStatus, e.By, target)...,
	)

	return
}

func purgeHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	target := mustUnitAt(e.At)

	for st, v := range target.Statuses {
		if v.IsPositive() {
			sts.Add(removeStatusFromUnit(st, target)...)
		}
	}

	return
}

func purifyHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	target := mustUnitAt(e.At)

	for st, v := range target.Statuses {
		if v.IsNegative() {
			sts.Add(removeStatusFromUnit(st, target)...)
		}
	}

	sts.Add(healUnit(target, e.Ab.Effect.AddHP)...)
	sts.Add(applyStatusToUnit(e.Ab.Effect.ApplyStatus, e.By, target)...)

	return
}

func healHandler(e abilityUsed) (ds.ApplyStates, error) {
	target := mustUnitAt(e.At)

	st := healUnit(target, e.Ab.Effect.AddHP)
	return st, nil
}

func equalizeHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	var sumHP int
	units := findAlliesInRange(e.By, e.Ab.AreaRadius)
	for _, u := range units {
		sumHP += u.CurrentHP
	}

	count := len(units)
	if count <= 1 {
		return
	}

	eq := sumHP / count
	remainder := sumHP - eq*count

	for _, u := range units {
		if u.CurrentHP == eq {
			continue
		}

		changeBy := eq - u.CurrentHP
		u.CurrentHP = eq

		sts.Add(
			ds.ApplyState{
				ChangeHP: new(changeBy),
				ToUnitID: u.ID,
			},
			ds.ApplyState{
				SetHP:    new(u.CurrentHP),
				ToUnitID: u.ID,
			},
		)
	}

	if remainder > 0 {
		for i := 0; i < remainder; i++ {
			u := units[i%count]
			u.CurrentHP++

			for j, v := range sts {
				if v.ToUnitID != u.ID {
					continue
				}

				if v.SetHP != nil {
					v.SetHP = new(u.CurrentHP)
				}
				if v.ChangeHP != nil {
					*v.ChangeHP += 1
				}

				sts[j] = v
			}
		}
	}

	return
}

func idolihuSpinHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	units := findEnemiesInRange(e.By, e.Ab.AreaRadius)
	for _, u := range units {
		sts.Add(dealDamageToUnit(e.By, u, e.By.CurrentAtk)...)
	}

	return
}

func piercingShotHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	cells := lineSingleAreaCells(e.By.Pos, e.At, e.Ab.AreaRadius)
	for _, c := range cells {
		unit := GetUnitAt(c.Coord)
		if unit != nil && unit.IsEnemy(e.By) {
			sts.Add(dealDamageToUnit(e.By, unit, e.Ab.Effect.DealDamage)...)
		}
	}

	return
}

func battleCryHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	units := findAlliesInRange(e.By, e.Ab.AreaRadius)
	for _, u := range units {
		sts.Add(applyStatusToUnit(e.Ab.Effect.ApplyStatus, e.By, u)...)
	}

	return
}

func shadowStepHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	if GetUnitAt(e.At) != nil {
		err = fmt.Errorf("shadowStep: target cell %s is occupied", e.At)
		return
	}
	PlaceUnitAt(e.By, e.At)

	sts.Add(
		applyStatusToUnit(e.Ab.Effect.ApplyStatus, e.By, e.By)...,
	)

	return
}

func huntersMarkHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	target := mustUnitAt(e.At)

	sts.Add(
		applyStatusToUnit(e.Ab.Effect.ApplyStatus, e.By, target)...,
	)

	return
}

func hamstringShotHandler(e abilityUsed) (sts ds.ApplyStates, err error) {
	target := mustUnitAt(e.At)

	sts.Add(dealDamageToUnit(e.By, target, e.Ab.Effect.DealDamage)...)
	sts.Add(applyStatusToUnit(e.Ab.Effect.ApplyStatus, e.By, target)...)

	return
}

func mustUnitAt(at ds.HexCoord) *ds.Unit {
	unit := GetUnitAt(at)
	if unit == nil {
		log.Fatalf("unit not found at: %s", at)
	}

	return unit
}

// dealDamageToUnit applies val damage to the unit, accounting for shield absorption.
// Shield absorbs damage first and any excess damage carries over to HP.
// If HP reaches zero, the unit is marked as dead and removed from the queue.
// Returns a slice of ApplyState mutations to be sent to the client for visual feedback.
func dealDamageToUnit(source, target *ds.Unit, val int) (st ds.ApplyStates) {
	defer func() {
		if len(st) > 0 {
			st.Add(applyOnDamageReceivedHandlers(source, target)...)
		}
	}()

	eff, ok := target.Statuses[status.Exposed]
	if ok {
		val += eff.Value
	}

	var shieldRemoved int
	if target.CurrentShield > 0 {
		if target.CurrentShield < val {
			shieldRemoved = target.CurrentShield
			target.CurrentShield = 0
			val = val - shieldRemoved
		} else {
			shieldRemoved = val
			target.CurrentShield -= val
			val = 0
		}
	}

	if shieldRemoved > 0 {
		st.Add(
			ds.ApplyState{ChangeShield: new(-shieldRemoved), ToUnitID: target.ID},
			ds.ApplyState{SetShield: new(target.CurrentShield), ToUnitID: target.ID},
		)
	}

	// shield fully absorbed the damage
	if val == 0 {
		return st
	}

	if target.CurrentHP < val {
		val = target.CurrentHP
	}

	target.CurrentHP -= val
	st.Add(
		ds.ApplyState{ChangeHP: new(-val), ToUnitID: target.ID},
		ds.ApplyState{SetHP: new(target.CurrentHP), ToUnitID: target.ID},
	)

	if target.CurrentHP <= 0 {
		target.IsDead = true
		st.Add(applyOnDeathHandlers(target)...)

		if target.IsDead {
			RemoveUnitFromQueue(target.ID)
			st.Add(ds.ApplyState{IsDead: true, ToUnitID: target.ID})
		}
	}

	return st
}

func healUnit(u *ds.Unit, val int) ds.ApplyStates {
	if u.CurrentHP == u.BaseHP {
		return ds.ApplyStates{}
	}

	u.CurrentHP += val
	if u.CurrentHP > u.BaseHP {
		val = val - (u.CurrentHP - u.BaseHP)
		u.CurrentHP = u.BaseHP
	}

	if val == 0 {
		return ds.ApplyStates{}
	}

	st := ds.NewUnitStates(
		ds.ApplyState{ChangeHP: new(val), ToUnitID: u.ID},
		ds.ApplyState{SetHP: new(u.CurrentHP), ToUnitID: u.ID},
	)

	return st
}

// lineAreaCells returns all board cells in all 6 directions from `from` up to `length` steps.
// Used for AreaLine abilities like PiercingShot.
func lineAreaCells(from ds.HexCoord, radius int) []*ds.BoardCell {
	dirs := []ds.HexCoord{
		{Q: 1, R: 0}, {Q: -1, R: 0},
		{Q: 0, R: 1}, {Q: 0, R: -1},
		{Q: 1, R: -1}, {Q: -1, R: 1},
	}

	var result []*ds.BoardCell
	for _, dir := range dirs {
		cur := from
		for i := 0; i < radius; i++ {
			cur = ds.HexCoord{Q: cur.Q + dir.Q, R: cur.R + dir.R}
			cell, ok := gameState.Board.Cells[cur]
			if !ok {
				break
			}
			result = append(result, cell)
		}
	}
	return result
}

// lineSingleAreaCells returns cells along a ray from [from] strictly in the direction of targetPos,
// up to radius steps. If targetPos does not lie on any of the 6 hex axes from [from],
// it returns an empty slice.
func lineSingleAreaCells(from, targetPos ds.HexCoord, radius int) []*ds.BoardCell {
	if from == targetPos {
		return []*ds.BoardCell{}
	}

	dq := targetPos.Q - from.Q
	dr := targetPos.R - from.R

	// A valid hex axis requires dr==0, dq==0, or dq==-dr.
	// In all valid cases the unit direction is just sign(dq), sign(dr).
	if dr != 0 && dq != 0 && dq != -dr {
		return []*ds.BoardCell{}
	}

	sign := func(x int) int {
		if x > 0 {
			return 1
		}
		if x < 0 {
			return -1
		}
		return 0
	}

	dir := ds.HexCoord{Q: sign(dq), R: sign(dr)}

	var result []*ds.BoardCell
	cur := from
	for range radius {
		cur = ds.HexCoord{Q: cur.Q + dir.Q, R: cur.R + dir.R}
		cell, ok := gameState.Board.Cells[cur]
		if !ok {
			break
		}
		result = append(result, cell)
	}

	return result
}

func findAlliesInRange(u *ds.Unit, radius int) []*ds.Unit {
	units := []*ds.Unit{}

	cells := hex.CellsInRange(u.Pos, radius, gameState.Board)
	for _, c := range cells {
		unit := GetUnitAt(c)
		if unit != nil && unit.IsAlly(u) {
			units = append(units, unit)
		}
	}

	return units
}

func findEnemiesInRange(u *ds.Unit, radius int) []*ds.Unit {
	units := []*ds.Unit{}

	cells := hex.CellsInRange(u.Pos, radius, gameState.Board)
	for _, c := range cells {
		unit := GetUnitAt(c)
		if unit != nil && unit.IsEnemy(u) {
			units = append(units, unit)
		}
	}

	return units
}
