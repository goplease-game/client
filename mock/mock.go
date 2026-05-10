package mock

import (
	"embed"
	"encoding/json"
	"errors"

	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

//go:embed *
var data embed.FS

func Load(filename string) ([]byte, error) {
	return data.ReadFile(filename)
}

func GetActionPayload(action ws.Action) (json.RawMessage, error) {
	switch action {
	case ws.NewGameAction:
		return Load("data/new_game.json")
	}

	err := errors.New("mock: invalid action: " + string(action))
	return nil, err
}
