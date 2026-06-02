package mock

import (
	"fmt"
	"log"

	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/hex"
)

const StatusPermanentDuration = -1

var abilityHandlers = map[ability.ID]func(ds.UseAbilityPayload) ([]ds.ApplyState, error){
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

// HandleAbility is cooking a response for specific ability. We don't validation here,
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

	st := dealDamageToUnit(caster, target, caster.CurrentAtk)
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
	ef := status.ByType(status.DecayingShield)

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

	sts := ds.NewUnitStates()
	sts.Add(applyStatusToUnit(caster, status.Provoking))

	cells := hex.CellsInRange(caster.Pos, ab.AreaRadius, gameState.Board)
	for _, c := range cells {
		if u := GetUnitAt(c); u != nil && u.IsOpponent {
			sts.Add(applyStatusToUnit(u, status.Provoked, map[string]any{"provoker": caster.ID}))
		}
	}

	return sts, nil
}

func shieldBashHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	_, target := mustAbilityActors(load)

	st := ds.NewUnitStates()
	st.Add(applyStatusToUnit(target, status.Stun))

	return st, nil
}

func powerPushHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster, target := mustAbilityActors(load)

	dealDmg := 2
	st := ds.NewUnitStates()

	pos := hex.OppositeHex(caster.Pos, target.Pos)
	cell, _ := gameState.Board.Cells[pos]
	if cell != nil && cell.Unit == nil {
		st.Add(ds.ApplyState{MoveTo: new(pos), ToUnitID: target.ID})
		target.Pos = pos
	} else {
		dealDmg += 1
	}

	st.Add(dealDamageToUnit(caster, target, dealDmg)...)
	return st, nil
}

func gangUpHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster, target := mustAbilityActors(load)

	dealDmg := caster.CurrentAtk
	pos := hex.OppositeHex(caster.Pos, target.Pos)
	cell, _ := gameState.Board.Cells[pos]
	if cell != nil && cell.Unit != nil && !cell.Unit.IsOpponent {
		dealDmg += 2
	}

	st := dealDamageToUnit(caster, target, dealDmg)
	return st, nil
}

func eliminateHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster, target := mustAbilityActors(load)

	dealDmg := 3
	st := dealDamageToUnit(caster, target, dealDmg)
	if target.IsDead {
		st.Add(
			ds.ApplyState{ChangeAP: new(1), ToUnitID: caster.ID},
			ds.ApplyState{SetAP: new(1), ToUnitID: caster.ID},
		)
	}

	return st, nil
}

func translocationHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster, target := mustAbilityActors(load)

	from := caster.Pos
	to := target.Pos

	caster.Pos = to
	target.Pos = from

	return nil, nil
}

func timeWarpHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	_, target := mustAbilityActors(load)

	st := ds.NewUnitStates(applyStatusToUnit(target, status.TemporalAnchor))

	return st, nil
}

func purgeHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	_, target := mustAbilityActors(load)

	st := ds.NewUnitStates()
	for statusType, v := range target.Statuses {
		if v.IsPositive() {
			st.Add(removeStatusFromUnit(target, statusType))
		}
	}

	return st, nil
}

func purifyHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	_, target := mustAbilityActors(load)

	st := ds.NewUnitStates()
	for statusType, v := range target.Statuses {
		if v.IsNegative() {
			st.Add(removeStatusFromUnit(target, statusType))
		}
	}
	st.Add(healUnit(target, 2)...)
	st.Add(applyStatusToUnit(target, status.DebuffWard))

	return st, nil
}

func healHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	_, target := mustAbilityActors(load)

	st := healUnit(target, 5)
	return st, nil
}

func equalizeHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster := GetUnitByID(load.UnitID)
	if caster == nil {
		log.Fatalf("invalid ability caster: %s", load.UnitID)
	}

	ab := ability.ByID(ability.Equalize)

	var sumHP int
	var units []*ds.Unit

	st := ds.NewUnitStates()

	cells := hex.CellsInRange(caster.Pos, ab.AreaRadius, gameState.Board)
	for _, c := range cells {
		if u := GetUnitAt(c); u != nil && !u.IsOpponent {
			units = append(units, u)
			sumHP += u.CurrentHP
		}
	}

	count := len(units)
	if count <= 1 {
		return st, nil
	}

	eq := sumHP / count
	remainder := sumHP - eq*count

	for _, u := range units {
		if u.CurrentHP == eq {
			continue
		}

		changeBy := eq - u.CurrentHP
		u.CurrentHP = eq

		st.Add(
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

			for j, v := range st {
				if v.ToUnitID != u.ID {
					continue
				}

				if v.SetHP != nil {
					v.SetHP = new(u.CurrentHP)
				}
				if v.ChangeHP != nil {
					*v.ChangeHP += 1
				}

				st[j] = v
			}
		}
	}

	return st, nil
}

func idolihuSpinHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster := GetUnitByID(load.UnitID)
	if caster == nil {
		log.Fatalf("invalid ability caster: %s", load.UnitID)
	}

	ab := ability.Abilities[ability.IdolihuSpin]
	st := ds.NewUnitStates()
	cells := hex.CellsInRange(caster.Pos, ab.AreaRadius, gameState.Board)

	for _, c := range cells {
		if u := GetUnitAt(c); u != nil && u.IsOpponent {
			st = append(st, dealDamageToUnit(caster, u, caster.CurrentAtk)...)
		}
	}

	return st, nil
}

func piercingShotHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster := GetUnitByID(load.UnitID)
	if caster == nil {
		log.Fatalf("invalid ability caster: %s", load.UnitID)
	}

	flatAtk := 2
	ab := ability.Abilities[ability.PiercingShot]
	st := ds.NewUnitStates()
	cells := lineAreaCells(caster.Pos, ab.AreaRadius)

	for _, c := range cells {
		if c.Unit != nil && c.Unit.IsOpponent {
			st = append(st, dealDamageToUnit(caster, c.Unit, flatAtk)...)
		}
	}

	return st, nil
}

func battleCryHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster := GetUnitByID(load.UnitID)
	if caster == nil {
		log.Fatalf("invalid ability caster: %s", load.UnitID)
	}

	ab := ability.Abilities[ability.BattleCry]
	ef := status.ByType(status.Rallied)

	st := ds.NewUnitStates()
	cells := hex.CellsInRange(caster.Pos, ab.AreaRadius, gameState.Board)

	for _, c := range cells {
		if u := GetUnitAt(c); u != nil && !u.IsOpponent {
			u.CurrentAtk += ef.InitialValue
			st.Add(
				applyStatusToUnit(u, status.Rallied),
				ds.ApplyState{ChangeAtk: new(ef.InitialValue), ToUnitID: u.ID},
				ds.ApplyState{SetAtk: new(u.CurrentAtk), ToUnitID: u.ID},
			)
		}
	}

	return st, nil
}

func shadowStepHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster := GetUnitByID(load.UnitID)
	if caster == nil {
		log.Fatalf("invalid ability caster: %s", load.UnitID)
	}

	PlaceUnitAt(caster, load.Target)

	ef := status.ByType(status.Sharpened)
	caster.CurrentAtk += ef.InitialValue

	st := ds.NewUnitStates()
	st.Add(
		applyStatusToUnit(caster, status.Sharpened),
		ds.ApplyState{ChangeAtk: new(ef.InitialValue), ToUnitID: caster.ID},
		ds.ApplyState{SetAtk: new(caster.CurrentAtk), ToUnitID: caster.ID},
	)

	return st, nil
}

func huntersMarkHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	_, target := mustAbilityActors(load)

	st := ds.NewUnitStates(applyStatusToUnit(target, status.Exposed))
	return st, nil
}

func hamstringShotHandler(load ds.UseAbilityPayload) ([]ds.ApplyState, error) {
	caster, target := mustAbilityActors(load)

	dmg := 2
	st := dealDamageToUnit(caster, target, dmg)
	st.Add(applyStatusToUnit(target, status.Hamstrung))

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

func applyStatusToUnit(u *ds.Unit, st status.Type, metaOpt ...map[string]any) (state ds.ApplyState) {
	ste := status.ByType(st)
	if ste == nil {
		log.Printf("applyStatusToUnit: unknown status type %s", st)
		return
	}

	var meta map[string]any
	if metaOpt != nil {
		meta = metaOpt[0]
	}

	u.AddStatus(status.Value{
		UnitID:   u.ID,
		Duration: ste.Duration,
		Value:    ste.InitialValue,
		Status:   ste,
		Meta:     meta,
	})

	return ds.ApplyState{
		AddStatus:     new(st),
		AddStatusMeta: &meta,
		ToUnitID:      u.ID,
	}
}

func removeStatusFromUnit(u *ds.Unit, st status.Type) (state ds.ApplyState) {
	ste := status.ByType(st)
	if ste == nil {
		log.Printf("applyStatusToUnit: unknown status type %s", st)
		return
	}

	delete(u.Statuses, st)

	return ds.ApplyState{
		RemoveStatus: new(st),
		ToUnitID:     u.ID,
	}
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
