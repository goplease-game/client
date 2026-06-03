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
	onOtherStatusApplied func(from, to *ds.Unit, st *status.Value, v status.Value)
	mutate               func(v *status.Value, from, to *ds.Unit)
}

var statusHandlers = map[status.Type]*statusHandler{
	status.Provoked:  provokedSH,
	status.Provoking: nil, // this is just decorative status

	status.Stunned:        stunnedSH,
	status.Rallied:        ralliedSH,
	status.Exposed:        exposedSH,
	status.Hamstrung:      hamstrungSH,
	status.Sharpened:      sharpenedSH,
	status.DebuffWard:     nil, // TODO
	status.TemporalAnchor: nil,
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

var debuffWardSH = &statusHandler{
	//onOtherStatusApplied: func(u *ds.Unit, sv status.Value) (st ds.ApplyStates) {
	//	u.CurrentAtk -= sv.Value
	//	st.Add(
	//		ds.ApplyState{ChangeAtk: new(-sv.Value), ToUnitID: u.ID},
	//		ds.ApplyState{SetAtk: new(u.CurrentAtk), ToUnitID: u.ID},
	//	)
	//
	//	return st
	//},
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

	h := statusHandlers[st]
	if h != nil && h.mutate != nil {
		h.mutate(&sv, from, to)
	}

	to.AddStatus(sv)

	sts.Add(ds.ApplyState{
		AddStatus:     new(st),
		AddStatusMeta: sv.Meta,
		ToUnitID:      to.ID,
	})

	if h != nil && h.onApply != nil {
		sts.Add(h.onApply(from, to, sv)...)
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
