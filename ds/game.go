package ds

type GameSnapshot struct {
	ArenaID         string  `json:"arena_id"`
	Board           Board   `json:"board"`
	Player          Player  `json:"player"`
	OpponentName    string  `json:"opponent_name"`
	UnitsQueue      []*Unit `json:"units_queue"`
	ActiveUnitID    string  `json:"active_unit_id"`
	Round           int     `json:"round"`
	TurnTimeSeconds int     `json:"turn_time_seconds"`

	MaxPhantomAPPerUnitPerTurn int `json:"max_phantom_ap_per_unit_per_turn"`
}

type NewGamePayload struct {
	ArenaID         string  `json:"arena_id"`
	Board           Board   `json:"board"`
	Player          *Player `json:"player"`
	Opponent        string  `json:"opponent"`
	TurnTimeSeconds int     `json:"turn_time_seconds"` // 0 = no timer
}

type ErrorResponse struct {
	Message string `json:"message"`
}
