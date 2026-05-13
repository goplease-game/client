package ws

import (
	"encoding/json"

	"github.com/ognev-dev/goplease-ebitengine-client/config"
)

type Client interface {
	Inbox() <-chan InMessage
	Status() ConnStatus
	Send(v OutMessage)
	Connect(playerID string)
	Disconnect()
}

func NewClient() Client {
	if config.Get().MockClient {
		return NewMockClient()
	}

	return NewWSClient()
}

type InMessage struct {
	Action Action          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type OutMessage struct {
	Action Action `json:"action"`
	Data   any    `json:"data"`
}

type ConnStatus int

const (
	StatusDisconnected ConnStatus = iota
	StatusConnecting
	StatusConnected
	StatusError
)

type Action string

const (
	ConnectedAction    Action = "connected"
	SearchingOppAction Action = "searching_opp"
	NewGameAction      Action = "new_game"
	ReadyToPlay        Action = "ready_to_play"
	WaitingForOpponent Action = "waiting_for_opponent"
	EndTurnAction      Action = "end_turn"
	NextAction         Action = "next_action"

	PlaceUnitAction Action = "place_unit"

	UnitPlacedAction      Action = "unit_placed"
	OppDisconnectedAction Action = "opp_disconnected"
	CancelMatchAction     Action = "cancel_match"
	MatchCancelledAction  Action = "match_canceled"
	ErrorAction           Action = "error"
)
