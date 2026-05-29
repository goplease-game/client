package effect

var statusByType = map[StatusType]*Status{
	DecayingShield: decayingShieldStatus,
	DecayingAttack: decayingAttackStatus,
	Provoked:       provokedStatus,
	Provoking:      provokingStatus,
	Stun:           stunStatus,
	Hamstrung:      hamstrungStatus,
	Exposed:        exposedStatus,
	Sharpened:      sharpenedStatus,
	DebuffWard:     debuffWardStatus,
}

func NewStatus(t StatusType) *Status {
	return statusByType[t]
}

// StatusByType returns the Status definition for the given StatusType, or nil if not found.
func StatusByType(t StatusType) *Status {
	return statusByType[t]
}
