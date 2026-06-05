package asset

import (
	"bytes"
	"embed"
	"image"
	"log"
	"path"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed data
var dataFS embed.FS

//go:embed data/asset_err.png
var assetErrPng []byte

const dataDir = "data"

func Load(name string) []byte {
	name = path.Join(dataDir, name)
	data, err := dataFS.ReadFile(name)
	if err != nil {
		log.Printf("failed to load asset '%s': %s\n", name, err)
	}

	return data
}

func decodeEbitenImageFromBytes(data []byte) (*ebiten.Image, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(src), nil
}
