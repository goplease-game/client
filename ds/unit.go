package ds

type Unit struct {
	ID         string `json:"id"`
	TemplateID int    `json:"template_id"`
	OwnerID    string `json:"owner_id"`
	Name       string `json:"name"`

	MaxHP         int `json:"max_hp"`
	CurrentHP     int `json:"current_hp"`
	CurrentShield int `json:"current_shield"`

	AP int `json:"ap"` // Action Points
	MP int `json:"mp"` // Move Points

	// board position, -1 - in hand
	Row int `json:"row"`
	Col int `json:"col"`

	Abilities []AbilityID       `json:"abilities"`
	Cooldowns map[AbilityID]int `json:"cooldowns"`

	IsOpponent bool
}

type PlaceUnitPayload struct {
	Row  int   `json:"row"`
	Col  int   `json:"col"`
	Unit *Unit `json:"unit"`
}

type PlayUnitPayload struct {
	UnitID string `json:"unit_id"`
}

type UnitPlacedPayload struct {
	Row        int `json:"row"`
	Col        int `json:"col"`
	TemplateID int `json:"template_id"`
}
