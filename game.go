package game

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"log"
	"path"

	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/config"
	"github.com/ognev-dev/goplease-ebitengine-client/mock"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

//go:embed assets
var assets embed.FS

//go:embed assets/asset_err.png
var assetErrPng []byte

const assetsDir = "assets"

// Game is the root ebiten.Game implementation.
// It owns shared resources (Server connection, player identity) and delegates
// Update/Draw to the currently active Screen.
type Game struct {
	screen   Screen // active screen
	Server   *ws.Client
	PlayerID string // stable UUID for this client session
}

func NewGame() *Game {
	g := &Game{
		PlayerID: uuid.NewString(),
		Server:   ws.NewClient(),
	}

	config.Get().UseMockData = true
	data, err := mock.GetActionPayload(ws.NewGameAction)
	if err != nil {
		panic(err)
	}
	g.screen = NewRoomScreen(data)

	//g.screen = NewMainScreen()
	return g
}

// SwitchTo replaces the active screen.
func (g *Game) SwitchTo(s Screen) { g.screen = s }

func (g *Game) Update() error {
	next, err := g.screen.Update(g)
	if err != nil {
		return err
	}
	if next != g.screen {
		g.screen = next
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.screen.Draw(screen)
}

func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	conf := config.Get()
	return conf.WindowW, conf.WindowH
}

func Asset(name string) []byte {
	name = path.Join(assetsDir, name)
	data, err := assets.ReadFile(name)
	if err != nil {
		log.Printf("failed to load asset '%s': %s\n", name, err)
	}

	return data
}

type ImageSize struct {
	W, H int
}

func ImageAsset(name string, sizeOpt ...ImageSize) *ebiten.Image {
	name = path.Join(assetsDir, name)

	img, err := loadEbitenImageFromAssets(name)
	if err == nil {
		if len(sizeOpt) > 0 {
			img = resizeImage(img, sizeOpt[0].W, sizeOpt[0].H)
		}

		return img
	}
	log.Printf("failed to load asset '%s': %v", name, err)

	placeholder, err := decodeEbitenImageFromBytes(assetErrPng)
	if err != nil {
		// well, something went wrong for real
		panic(fmt.Sprintf("failed to load placeholder image: %v", err))
	}

	return placeholder
}

func loadEbitenImageFromAssets(path string) (*ebiten.Image, error) {
	f, err := assets.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(src), nil
}

func decodeEbitenImageFromBytes(data []byte) (*ebiten.Image, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(src), nil
}

func resizeImage(src *ebiten.Image, w, h int) *ebiten.Image {
	dst := ebiten.NewImage(w, h)
	opts := &ebiten.DrawImageOptions{}
	sx := float64(w) / float64(src.Bounds().Dx())
	sy := float64(h) / float64(src.Bounds().Dy())
	opts.GeoM.Scale(sx, sy)
	dst.DrawImage(src, opts)
	return dst
}
