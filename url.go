package game

import (
	"fmt"

	"github.com/goplease-game/client/config"
)

// ServerWS returns the WebSocket endpoint URL for the given server path.
// The URL uses either the ws or wss scheme depending on the current configuration.
func ServerWS(path string) string {
	conf := config.Get()
	scheme := "ws"
	if conf.Secure {
		scheme = "wss"
	}

	return fmt.Sprintf("%s://%s/%s", scheme, conf.ServerAddr, path)
}

// ServerAPI returns the HTTP API endpoint URL for the given server path.
// The URL uses either the http or https scheme depending on the current configuration.
func ServerAPI(path string) string {
	conf := config.Get()
	scheme := "http"
	if conf.Secure {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/api/%s", scheme, conf.ServerAddr, path)
}
