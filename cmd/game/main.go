package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client"
	"github.com/ognev-dev/goplease-ebitengine-client/config"
	"github.com/ognev-dev/goplease-ebitengine-client/screen"
	"github.com/ognev-dev/goplease-ebitengine-client/sfx"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

func main() {
	conf := config.Get()
	ebiten.SetWindowSize(conf.WindowW, conf.WindowH)
	ebiten.SetWindowTitle("go, please")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	sfx.SetVolume(conf.Volume)

	server := ws.NewClient()
	err := ebiten.RunGame(game.New(server, screen.NewMainScreen(server)))
	if err != nil {
		log.Fatal(err)
	}
}
