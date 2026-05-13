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
	ConnectedAction Action = "connected"
	NewGameAction   Action = "new_game"

	// Player done with unit placement
	EndUnitPlacement Action = "end_unit_placement"
	// Player done with unit action
	EndUnitActing Action = "end_unit_action"

	SearchingOppAction Action = "searching_opp"
	PlaceUnitAction    Action = "place_unit"

	UnitPlacedAction      Action = "unit_placed"
	OppDisconnectedAction Action = "opp_disconnected"
	CancelMatchAction     Action = "cancel_match"
	MatchCancelledAction  Action = "match_canceled"
	ErrorAction           Action = "error"
)
