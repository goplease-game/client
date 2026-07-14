package asset

import (
	"path"
	"sync"

	"github.com/ebitenui/ebitenui/image"
)

// nineSliceCache stores decoded nine-slice shadow images to avoid re-decoding
// the same PNG file on repeated calls.
var nineSliceCache sync.Map

// NineSlice loads the image at name and wraps it as an *image.NineSlice,
// using corner as the size in pixels of each non-stretching edge/corner.
// The result is cached per filename+corner pair.
func NineSlice(name string, corner int) *image.NineSlice {
	type key struct {
		name   string
		corner int
	}
	k := key{name: name, corner: corner}

	if cached, ok := nineSliceCache.Load(k); ok {
		return cached.(*image.NineSlice) //nolint:forcetypeassert
	}

	img, err := loadEbitenImageFromAssets(path.Join(dataDir, name))
	if err != nil {
		img = lostImagePlaceHolder()
	}

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	ns := image.NewNineSlice(
		img,
		[3]int{corner, w - corner*2, corner},
		[3]int{corner, h - corner*2, corner},
	)

	nineSliceCache.Store(k, ns)
	return ns
}
