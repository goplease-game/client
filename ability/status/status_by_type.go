package status

var statuses = map[Type]*Status{
	DecayingShield: decayingShieldStatus,
	Rallied:        ralliedStatus,
	Provoked:       provokedStatus,
	Provoking:      provokingStatus,
	Stun:           stunStatus,
	Hamstrung:      hamstrungStatus,
	Exposed:        exposedStatus,
	Sharpened:      sharpenedStatus,
	DebuffWard:     debuffWardStatus,
	TemporalAnchor: temporalAnchorStatus,
}

// ByType returns the Status definition for the given Type, or nil if not found.
func ByType(t Type) *Status {
	return statuses[t]
}
