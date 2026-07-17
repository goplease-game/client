// Package asset ...
package asset

import (
	"bytes"
	"embed"
	"image"
	"log"
	"path"

	"github.com/hajimehoshi/ebiten/v2"
)

// dataFS holds the embedded static game assets from the 'data' directory.
//
//go:embed data
var dataFS embed.FS

// assetErrPng holds the raw bytes of the fallback placeholder image.
//
//go:embed data/asset_err.png
var assetErrPng []byte

// dataDir defines the root directory name for embedded game assets.
const dataDir = "data"

// Load reads and returns the raw byte content of the specified asset file
// from the embedded filesystem. Logs an error and returns nil if the file cannot be read.
func Load(name string) []byte {
	name = path.Join(dataDir, name)
	data, err := dataFS.ReadFile(name)
	if err != nil {
		log.Printf("failed to load asset '%s': %s\n", name, err)
	}

	return data
}

// decodeEbitenImageFromBytes decodes a slice of raw bytes into an *ebiten.Image instance.
func decodeEbitenImageFromBytes(data []byte) (*ebiten.Image, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(src), nil
}
