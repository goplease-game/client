package ds

type Player struct {
	ID          string `json:"id"` // uuid
	Name        string `json:"name"`
	IsBot       bool   `json:"is_bot"`
	PlayerIndex int    `json:"-"`     // 0 or 1
	Units       []Unit `json:"units"` // units at hand

	PhantomAP int `json:"phantom_ap"`

	UnitsPlacedThisRound int
}
