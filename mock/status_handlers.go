package mock

import (
	"log"

	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

type statusHandler struct {
	onApply              func(from, to *ds.Unit, v status.Value) ds.ApplyStates
	onRemove             func(u *ds.Unit, v status.Value) ds.ApplyStates
	onUnitAttacked       func(dmg *int, v status.Value)
	onTurnStart          func(u *ds.Unit, v status.Value) ds.ApplyStates
	onTurnEnd            func(u *ds.Unit, v status.Value) ds.ApplyStates
	onOtherStatusApplied func(from, to *ds.Unit, applied *status.Value, v status.Value) ds.ApplyStates
	mutate               func(v *status.Value, from, to *ds.Unit)
}

var statusHandlers = map[status.Type]*statusHandler{
	status.Provoked:       provokedSH,
	status.Provoking:      nil, // this is just decorative status
	status.Stunned:        stunnedSH,
	status.Rallied:        ralliedSH,
	status.Exposed:        exposedSH,
	status.Hamstrung:      hamstrungSH,
	status.Sharpened:      sharpenedSH,
	status.DebuffWard:     debuffWardSH,
	status.TemporalAnchor: temporalAnchorSH,
	status.Frenzied:       frenziedSH,
}

var provokedSH = &statusHandler{
	mutate: func(v *status.Value, from, to *ds.Unit) {
		if v.Meta == nil {
			v.Meta = map[string]any{}
		}
		v.Meta["provoker"] = from.ID
	},
}

var simpleAttackModifierSH = &statusHandler{
	onApply: func(from, to *ds.Unit, sv status.Value) (st ds.ApplyStates) {
		to.CurrentAtk += sv.Value
		st.Add(
			ds.ApplyState{ChangeAtk: new(sv.Value), ToUnitID: to.ID},
			ds.ApplyState{SetAtk: new(to.CurrentAtk), ToUnitID: to.ID},
		)

		return st
	},
	onRemove: func(u *ds.Unit, sv status.Value) (st ds.ApplyStates) {
		u.CurrentAtk -= sv.Value
		st.Add(
			ds.ApplyState{ChangeAtk: new(-sv.Value), ToUnitID: u.ID},
			ds.ApplyState{SetAtk: new(u.CurrentAtk), ToUnitID: u.ID},
		)

		return st
	},
}

var ralliedSH = simpleAttackModifierSH
var sharpenedSH = simpleAttackModifierSH
var frenziedSH = simpleAttackModifierSH

var exposedSH = &statusHandler{
	onUnitAttacked: func(dmg *int, st status.Value) {
		*dmg += st.Value
	},
}

var stunnedSH = &statusHandler{
	onTurnStart: func(u *ds.Unit, v status.Value) ds.ApplyStates {
		return ds.NewUnitStates(ds.ApplyState{
			SkipTurn: true,
			ToUnitID: u.ID,
		})
	},
}

var hamstrungSH = &statusHandler{
	onTurnStart: func(u *ds.Unit, st status.Value) (sts ds.ApplyStates) {
		u.CurrentMP = st.Value

		sts.Add(ds.ApplyState{
			SetMP:    new(st.Value),
			ToUnitID: u.ID,
		})

		return
	},
	onRemove: func(u *ds.Unit, v status.Value) (sts ds.ApplyStates) {
		u.CurrentMP = u.BaseHP

		sts.Add(ds.ApplyState{
			SetMP:    new(u.CurrentMP),
			ToUnitID: u.ID,
		})

		return
	},
}

var temporalAnchorSH = &statusHandler{
	onTurnStart: func(u *ds.Unit, sv status.Value) (sts ds.ApplyStates) {
		u.CurrentAP += sv.Value

		sts.Add(
			ds.ApplyState{ChangeAP: new(sv.Value), ToUnitID: u.ID},
			ds.ApplyState{SetAP: new(u.CurrentAP), ToUnitID: u.ID},
		)

		sv.Meta = map[string]any{
			"hp":     u.CurrentHP,
			"shield": u.CurrentShield,
			"pos":    u.Pos,
		}

		u.Statuses[sv.Status.Type] = sv
		return
	},
	onTurnEnd: func(u *ds.Unit, sv status.Value) (sts ds.ApplyStates) {
		if sv.Meta != nil {
			prevHP := sv.Meta["hp"].(int)
			prevShield := sv.Meta["shield"].(int)
			hpDiff := prevHP - u.CurrentHP
			shDiff := prevShield - u.CurrentShield

			prevPos := sv.Meta["pos"].(ds.HexCoord)

			if hpDiff != 0 {
				u.CurrentHP = prevHP
				sts.Add(
					ds.ApplyState{ChangeHP: new(hpDiff), ToUnitID: u.ID},
					ds.ApplyState{SetHP: new(u.CurrentHP), ToUnitID: u.ID},
				)
			}
			if shDiff != 0 {
				u.CurrentShield = prevShield
				sts.Add(
					ds.ApplyState{ChangeHP: new(shDiff), ToUnitID: u.ID},
					ds.ApplyState{SetHP: new(u.CurrentShield), ToUnitID: u.ID},
				)
			}
			if prevPos != u.Pos {
				u.Pos = prevPos
				sts.Add(
					ds.ApplyState{MoveTo: new(prevPos), ToUnitID: u.ID},
				)
			}

		}

		return
	},
}

var debuffWardSH = &statusHandler{
	onOtherStatusApplied: func(from, to *ds.Unit, applied *status.Value, v status.Value) (sts ds.ApplyStates) {
		if !applied.IsNegative() {
			return
		}

		applied.Duration = 0
		sts.Add(ds.ApplyState{ShowText: new("Debuff Ward!"), ToUnitID: to.ID})
		return
	},
}

func applyStatusToUnit(st status.Type, from, to *ds.Unit) (sts ds.ApplyStates) {
	inst := status.ByType(st)
	if inst == nil {
		log.Printf("applyStatusToUnit: unknown status type %s", st)
		return
	}

	sv := status.Value{
		UnitID:   to.ID,
		Duration: inst.Duration,
		Value:    inst.InitialValue,
		Status:   inst,
	}

	statusH := statusHandlers[st]
	if statusH != nil && statusH.mutate != nil {
		statusH.mutate(&sv, from, to)
	}

	for t, v := range to.Statuses {
		if t == st {
			continue
		}
		h := statusHandlers[t]
		if h == nil || h.onOtherStatusApplied == nil {
			continue
		}
		sts.Add(h.onOtherStatusApplied(from, to, &sv, v)...)
		if sv.Duration == 0 {
			return
		}
	}

	to.AddStatus(sv)

	sts.Add(ds.ApplyState{
		AddStatus:     new(st),
		AddStatusMeta: sv.Meta,
		ToUnitID:      to.ID,
	})

	if statusH != nil && statusH.onApply != nil {
		sts.Add(statusH.onApply(from, to, sv)...)
	}

	return sts
}

func removeStatusFromUnit(st status.Type, u *ds.Unit) (sts ds.ApplyStates) {
	sv, ok := u.Statuses[st]
	if !ok {
		log.Printf("removeStatusFromUnit: unit missing status: %s", st)
		return
	}

	u.RemoveStatus(st)

	h := statusHandlers[st]
	if h != nil && h.onRemove != nil {
		sts.Add(h.onRemove(u, sv)...)
	}

	sts.Add(ds.ApplyState{
		RemoveStatus: new(st),
		ToUnitID:     u.ID,
	})

	return
}

func handleOnTurnStartStatuses(unit *ds.Unit) (sts ds.ApplyStates) {
	for t, v := range unit.Statuses {
		h, ok := statusHandlers[t]
		if !ok || h == nil {
			continue
		}

		if h.onTurnStart == nil {
			continue
		}

		sts.Add(h.onTurnStart(unit, v)...)
	}

	return
}
