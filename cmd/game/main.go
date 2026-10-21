// Package main ...
package main

import (
	"image"
	"log"

	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/config"
	contentclient "github.com/goplease-game/client/content-client"
	"github.com/goplease-game/client/screen"
	"github.com/goplease-game/client/sfx"
	"github.com/goplease-game/client/user"
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
	contentCl := contentclient.New()
	usr := user.Load(contentCl)

	ctx := &game.Context{
		Server:  serverCl,
		Content: contentCl,
		User:    usr,
	}

	err := ebiten.RunGame(game.New(screen.NewMainScreen(ctx)))
	if err != nil {
		log.Fatal(err)
	}
}
