package status

var Statuses = map[Type]*Status{
	Rallied:        ralliedStatus,
	Provoked:       provokedStatus,
	Provoking:      provokingStatus,
	Stunned:        stunnedStatus,
	Hamstrung:      hamstrungStatus,
	Marked:         markedStatus,
	Sharpened:      sharpenedStatus,
	DebuffWard:     debuffWardStatus,
	TemporalAnchor: temporalAnchorStatus,
	Frenzied:       frenziedStatus,
	Impatience:     impatienceStatus,
}

// Order defines the display order of statuses on unit cards.
var Order = []Type{
	// negative first
	Provoked,
	Hamstrung,
	Marked,
	Stunned,

	// positive
	Rallied,
	Sharpened,
	DebuffWard,
	Frenzied,
	TemporalAnchor,
	Impatience,

	// neutral
	Provoking,
}

// ByType returns the Status definition for the given Type, or nil if not found.
func ByType(t Type) *Status {
	return Statuses[t]
}
