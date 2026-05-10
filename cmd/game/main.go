package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client"
	"github.com/ognev-dev/goplease-ebitengine-client/config"
)

func main() {
	conf := config.Get()
	ebiten.SetWindowSize(conf.WindowW, conf.WindowH)
	ebiten.SetWindowTitle("go, please")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game.NewGame()); err != nil {
		log.Fatal(err)
	}
}
