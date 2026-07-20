package game

import (
	"log"

	"github.com/ebitenui/ebitenui/themes"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/ui"
	"golang.org/x/image/colornames"
)

var links = map[string]string{
	"source":          "https://github.com/goplease-game",
	"source-client":   "https://github.com/goplease-game/client",
	"source-server":   "https://github.com/goplease-game/server",
	"golang":          "https://go.dev",
	"ebitengine":      "https://ebitengine.org",
	"ebitenui":        "https://github.com/ebitenui/ebitenui",
	"discord":         "https://discord.gg/8KPNMrFT9v",
	"latest-releases": "https://github.com/goplease-game/client/releases/latest",
}

// OpenLink opens the link by given key.
func OpenLink(key string) error {
	url, ok := links[key]
	if !ok {
		log.Println("No link found for key:", key)
	}

	return OpenURL(url)
}

// SetLinksTheme applies a custom theme to the widget with colors for text links.
func SetLinksTheme(w *widget.Widget) {
	th := themes.GetBasicLightTheme()
	th.TextTheme.LinkColor = &widget.TextLinkColor{
		Idle:  ui.RGBFromHex("#00a8e8"),
		Hover: colornames.Gold,
	}

	w.SetTheme(th)
}
