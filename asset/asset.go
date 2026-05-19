package asset

import (
	"bytes"
	"embed"
	"fmt"
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

func Image(name string, sizeOpt ...int) *ebiten.Image {
	name = path.Join(dataDir, name)

	var w, h int
	if len(sizeOpt) == 2 {
		w, h = sizeOpt[0], sizeOpt[1]
	}
	if len(sizeOpt) == 1 {
		w, h = sizeOpt[0], sizeOpt[0]
	}

	img, err := loadEbitenImageFromAssets(name)
	if err == nil {
		if w > 0 {
			img = resizeImage(img, w, h)
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
	f, err := dataFS.Open(path)
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
	opts.Filter = ebiten.FilterLinear
	dst.DrawImage(src, opts)
	return dst
}
