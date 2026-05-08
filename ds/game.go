package ds

type Phase string

const (
	PhaseUnitPlacement Phase = "unit_placement" // current player places units
	PhaseUnitActing    Phase = "unit_acting"    // current player playing with unit
	PhaseGameOver      Phase = "game_over"
)

type NewGamePayload struct {
	RoomID   string  `json:"room_id"`
	Phase    Phase   `json:"phase"`
	IsMyTurn bool    `json:"is_my_turn"`
	Board    Board   `json:"board"`
	Player   *Player `json:"player"`
	Opponent string  `json:"opponent"`
}
