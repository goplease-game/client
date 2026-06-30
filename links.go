package game

import (
	"log"

	"github.com/pkg/browser"
)

var links = map[string]string{
	"source":     "https://github.com/goplease-game",
	"golang":     "https://go.dev",
	"ebitengine": "https://ebitengine.org",
	"ebitenui":   "https://github.com/ebitenui/ebitenui",
	"discord":    "https://discord.gg/8KPNMrFT9v",
}

// OpenLink opens the link by given key.
func OpenLink(key string) error {
	url, ok := links[key]
	if !ok {
		log.Println("No link found for key:", key)
	}

	return browser.OpenURL(url)
}
