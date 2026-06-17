package ds

// Player represents a participant in a match, tracking their hand of units,
// placement progress, and the Phantom AP carried over from unequal unit counts.
type Player struct {
	ID          string `json:"id"` // uuid
	Name        string `json:"name"`
	IsBot       bool   `json:"is_bot"`
	PlayerIndex int    `json:"-"`     // 0 or 1
	Units       []Unit `json:"units"` // units at hand

	PhantomAP int `json:"phantom_ap"`

	UnitsPlacedThisRound int `json:"-"`
}
