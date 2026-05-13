package ds

type NewGamePayload struct {
	RoomID   string  `json:"room_id"`
	Board    Board   `json:"board"`
	Player   *Player `json:"player"`
	Opponent string  `json:"opponent"`
}
