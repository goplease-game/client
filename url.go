package game

import (
	"fmt"

	"github.com/goplease-game/client/config"
)

// GameServerAPI returns the HTTP API endpoint URL for the given server path.
// The URL uses either the http or https scheme depending on the current configuration.
func GameServerAPI(path string) string {
	conf := config.Get()
	scheme := "http"
	if conf.GameServer.Secure {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/api/%s", scheme, conf.GameServer.Addr, path)
}
