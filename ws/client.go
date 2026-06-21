// Package ws ...
package ws

import (
	"encoding/json"
)

// Client is the interface implemented by both the real WebSocket client and
// the mock client used in practice mode.
type Client interface {
	Inbox() <-chan InMessage
	Status() ConnStatus
	Send(v OutMessage)
	Connect(playerID string)
	Disconnect()
}

// NewClient returns a WSClient, or a mock client if mock mode is enabled in config.
func NewClient() Client {
	return NewWSClient()
}

// InMessage is a message received from the server, with the payload left
// as raw JSON until the action is known.
type InMessage struct {
	Action Action          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

// OutMessage is a message sent to the server.
type OutMessage struct {
	Action Action `json:"action"`
	Data   any    `json:"data"`
}

// ConnStatus represents the current state of the WebSocket connection.
type ConnStatus int

// Connection status values.
const (
	StatusDisconnected ConnStatus = iota
	StatusConnecting
	StatusConnected
	StatusError
)

// Action identifies the type of message exchanged over the WebSocket connection.
type Action string

// Action identifiers exchanged between client and server.
const (
	ConnectedAction         Action = "connected"
	SearchingOppAction      Action = "searching_opp"
	NewGameAction           Action = "new_game"
	ReadyToPlay             Action = "ready_to_play"
	WaitingForOpponent      Action = "waiting_for_opponent"
	EndTurnAction           Action = "end_turn"
	PlaceUnitAction         Action = "place_unit"
	UnitPlacedAction        Action = "unit_placed"
	PlayUnitAction          Action = "play_unit"
	UnitMovedAction         Action = "unit_moved"
	UseAbility              Action = "use_ability"
	ApplyState              Action = "apply_state"
	NewRound                Action = "new_round"
	OppDisconnectedAction   Action = "opp_disconnected"
	CancelMatchAction       Action = "cancel_match"
	MatchCancelledAction    Action = "match_canceled"
	Surrender               Action = "surrender"
	OpponentSurrendered     Action = "opponent_surrendered"
	YouWin                  Action = "you_win"
	YouLose                 Action = "you_lose"
	ErrorAction             Action = "error"
	ActiveUnitChangedAction Action = "active_unit_changed"
)
