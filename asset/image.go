package asset

import (
	"fmt"
	"image"
	"image/color"
	"path"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"golang.org/x/image/colornames"
)

var imageCache sync.Map

// ImageBuilder builds an *ebiten.Image with optional resize, tint, and shadow.
// Results are cached at each step — resize, tint, and shadow — so shared
// base images are never processed twice.
//
// Usage:
//
//	img := asset.NewImage("icon.png", 16).Tint(colornames.Red).Render()
//	img2 := asset.NewImage("icon.png", 16).Shadow(2, 2, 0.5).Render()
type ImageBuilder struct {
	filename string
	color    color.Color
	width    int
	height   int
	shadow   *shadowParams
}

type shadowParams struct {
	color   color.Color
	offsetX int
	offsetY int
	alpha   float32
}

// cacheKey is a flat comparable struct used as sync.Map key.
// color.Color is an interface so we store its RGBA components instead.
type imageCacheKey struct {
	filename      string
	width, height int
	r, g, b, a    uint32
	shadow        shadowCacheKey
}

type shadowCacheKey struct {
	offsetX, offsetY int
	alpha            float32
	r, g, b          uint32
}

func NewImage(filename string, sizeOpt ...int) *ImageBuilder {
	b := &ImageBuilder{
		filename: path.Join(dataDir, filename),
	}
	if len(sizeOpt) == 2 {
		b.width, b.height = sizeOpt[0], sizeOpt[1]
	} else if len(sizeOpt) == 1 {
		b.width, b.height = sizeOpt[0], sizeOpt[0]
	}
	return b
}

func (b *ImageBuilder) Tint(c color.Color) *ImageBuilder {
	b.color = c
	return b
}

func (b *ImageBuilder) Shadow(offX, offY int, alpha float32, colorOpt ...color.Color) *ImageBuilder {
	sh := &shadowParams{
		offsetX: offX,
		offsetY: offY,
		alpha:   alpha,
		color:   colornames.Black,
	}
	if len(colorOpt) == 1 {
		sh.color = colorOpt[0]
	}

	b.shadow = sh
	return b
}

func (b *ImageBuilder) Render() *ebiten.Image {
	resizeKey := imageCacheKey{
		filename: b.filename,
		width:    b.width,
		height:   b.height,
	}
	img := loadOrStore(resizeKey, func() *ebiten.Image {
		src, err := loadEbitenImageFromAssets(b.filename)
		if err != nil {
			return lostImagePlaceHolder()
		}
		if b.width > 0 {
			return resizeImage(src, b.width, b.height)
		}
		return src
	})

	if b.color != nil {
		r, g, bl, a := b.color.RGBA()
		tintKey := imageCacheKey{
			filename: b.filename,
			width:    b.width,
			height:   b.height,
			r:        r,
			g:        g,
			b:        bl,
			a:        a,
		}
		img = loadOrStore(tintKey, func() *ebiten.Image {
			return ui.TintImage(img, b.color)
		})
	}

	if b.shadow != nil {
		s := b.shadow
		cr, cg, cb, ca := colorKey(b.color)
		sr, sg, sb, _ := colorKey(s.color)
		shadowKey := imageCacheKey{
			filename: b.filename,
			width:    b.width,
			height:   b.height,
			r:        cr,
			g:        cg,
			b:        cb,
			a:        ca,
			shadow: shadowCacheKey{
				offsetX: s.offsetX,
				offsetY: s.offsetY,
				alpha:   s.alpha,
				r:       sr,
				g:       sg,
				b:       sb,
			},
		}
		img = loadOrStore(shadowKey, func() *ebiten.Image {
			return applyShadow(img, s.offsetX, s.offsetY, s.alpha, s.color)
		})
	}

	return img
}

// loadOrStore returns the cached image for key, or calls build(), stores and returns the result.
func loadOrStore(key imageCacheKey, build func() *ebiten.Image) *ebiten.Image {
	if cached, ok := imageCache.Load(key); ok {
		return cached.(*ebiten.Image)
	}
	img := build()
	imageCache.Store(key, img)
	return img
}

func applyShadow(src *ebiten.Image, offsetX, offsetY int, alpha float32, c color.Color) *ebiten.Image {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := ebiten.NewImage(w+offsetX, h+offsetY)

	sr, sg, sb, _ := colorKey(c)
	shadowOpts := &ebiten.DrawImageOptions{}
	shadowOpts.GeoM.Translate(float64(offsetX), float64(offsetY))
	shadowOpts.ColorScale.SetR(float32(sr) / 0xffff)
	shadowOpts.ColorScale.SetG(float32(sg) / 0xffff)
	shadowOpts.ColorScale.SetB(float32(sb) / 0xffff)
	shadowOpts.ColorScale.SetA(alpha)
	dst.DrawImage(src, shadowOpts)

	dst.DrawImage(src, &ebiten.DrawImageOptions{})
	return dst
}

func colorKey(c color.Color) (r, g, b, a uint32) {
	if c == nil {
		return 0, 0, 0, 0
	}
	return c.RGBA()
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

// Image is a convenience wrapper for NewImage(...).Render().
func Image(name string, sizeOpt ...int) *ebiten.Image {
	return NewImage(name, sizeOpt...).Render()
}

// TintedImage is a convenience wrapper for NewImage(...).Tint(...).Render().
func TintedImage(name string, col color.Color, sizeOpt ...int) *ebiten.Image {
	return NewImage(name, sizeOpt...).Tint(col).Render()
}

func lostImagePlaceHolder() *ebiten.Image {
	img, err := decodeEbitenImageFromBytes(assetErrPng)
	if err != nil {
		panic(fmt.Sprintf("failed to load placeholder image: %v", err))
	}
	return img
}
