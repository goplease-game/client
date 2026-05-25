package ds

type GameSnapshot struct {
	RoomID       string  `json:"room_id"`
	Board        Board   `json:"board"`
	Player       Player  `json:"player"`
	OpponentName string  `json:"opponent_name"`
	UnitsQueue   []*Unit `json:"units_queue"`
	ActiveUnitID string  `json:"active_unit_id"`
	Round        int     `json:"round"`
}

type NewGamePayload struct {
	RoomID   string  `json:"room_id"`
	Board    Board   `json:"board"`
	Player   *Player `json:"player"`
	Opponent string  `json:"opponent"`
}
