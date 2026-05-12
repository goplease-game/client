package ws

import (
	"encoding/json"

	"github.com/ognev-dev/goplease-ebitengine-client/config"
)

type Client interface {
	Inbox() <-chan Message
	Status() ConnStatus
	Send(v any)
	Connect(playerID string)
	Disconnect()
}

func NewClient() Client {
	if config.Get().MockClient {
		return NewMockClient()
	}

	return NewWSClient()
}

// Message mirrors the server's OutgoingMsg.
type Message struct {
	Action Action          `json:"action"`
	Data   json.RawMessage `json:"data"`
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
	ConnectedAction       Action = "connected"
	NewGameAction         Action = "new_game"
	SearchingOppAction    Action = "searching_opp"
	PlaceUnitAction       Action = "place_unit"
	UnitPlacedAction      Action = "unit_placed"
	OppDisconnectedAction Action = "opp_disconnected"
	CancelMatchAction     Action = "cancel_match"
	MatchCancelledAction  Action = "match_canceled"
	ErrorAction           Action = "error"
)
