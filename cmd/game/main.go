// Package main ...
package main

import (
	"image"
	"log"

	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/config"
	"github.com/goplease-game/client/screen"
	"github.com/goplease-game/client/sfx"
	"github.com/goplease-game/client/ws"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	conf := config.Get()
	ebiten.SetWindowSize(conf.WindowW, conf.WindowH)
	ebiten.SetWindowTitle("go, please")

	ebiten.SetWindowIcon([]image.Image{asset.Image("window-icon.png")})
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if conf.Fullscreen {
		ebiten.SetFullscreen(true)
	}

	sfx.SetVolume(conf.Volume)

	serverCl := ws.NewClientProvider()
	err := ebiten.RunGame(game.New(screen.NewMainScreen(serverCl)))
	if err != nil {
		log.Fatal(err)
	}
}
