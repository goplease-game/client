package asset

import (
	"fmt"
	"image"
	"image/color"
	"path"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

// imageCache stores pre-processed images to avoid redundant transformations.
var imageCache sync.Map
var solidImageCache sync.Map

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

// shadowParams holds the configuration for rendering an image shadow.
type shadowParams struct {
	color   color.Color
	offsetX int
	offsetY int
	alpha   float32
}

// imageCacheKey is a flat comparable struct used as sync.Map key.
// color.Color is an interface so we store its RGBA components instead.
type imageCacheKey struct {
	filename      string
	width, height int
	r, g, b, a    uint32
	shadow        shadowCacheKey
}

// shadowCacheKey represents a unique comparable key for shadow parameters.
type shadowCacheKey struct {
	offsetX, offsetY int
	alpha            float32
	r, g, b          uint32
}

// NewImage initializes and returns a new ImageBuilder for the specified file.
// Accepts optional size dimensions: one value for a square, two for width and height.
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

// Tint applies a color overlay transformation to the image during rendering.
func (b *ImageBuilder) Tint(c color.Color) *ImageBuilder {
	b.color = c
	return b
}

// Shadow configures shadow parameters for the image, with an optional custom shadow color.
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

// Render processes the pipeline (loading, resizing, tinting, applying shadows)
// and returns the final *ebiten.Image using cached layers where possible.
func (b *ImageBuilder) Render() *ebiten.Image {
	img := loadOrStore(b.baseKey(), func() *ebiten.Image {
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
		img = loadOrStore(b.tintKey(), func() *ebiten.Image {
			return TintImage(img, b.color)
		})
	}

	if b.shadow != nil {
		s := b.shadow
		img = loadOrStore(b.shadowKey(), func() *ebiten.Image {
			return applyShadow(img, s.offsetX, s.offsetY, s.alpha, s.color)
		})
	}

	return img
}

// loadOrStore returns the cached image for key, or calls build(), stores and returns the result.
func loadOrStore(key imageCacheKey, build func() *ebiten.Image) *ebiten.Image {
	if cached, ok := imageCache.Load(key); ok {
		return cached.(*ebiten.Image) //nolint:forcetypeassert
	}
	img := build()
	imageCache.Store(key, img)
	return img
}

// applyShadow generates a new image containing the original image and its offset shadow.
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

// colorKey extracts RGBA components from a color.Color, returning zeroes if nil.
func colorKey(c color.Color) (r, g, b, a uint32) {
	if c == nil {
		return 0, 0, 0, 0
	}
	return c.RGBA()
}

// loadEbitenImageFromAssets loads and decodes an image file from the game assets filesystem.
func loadEbitenImageFromAssets(path string) (*ebiten.Image, error) {
	f, err := dataFS.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(src), nil
}

// resizeImage scales the source image to the specified width and height using linear filtering.
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

// lostImagePlaceHolder loads and returns the emergency fallback image if an asset is missing.
func lostImagePlaceHolder() *ebiten.Image {
	img, err := decodeEbitenImageFromBytes(assetErrPng)
	if err != nil {
		panic(fmt.Sprintf("failed to load placeholder image: %v", err))
	}
	return img
}

// baseKey generates a cache key representing the base resized image layer.
func (b *ImageBuilder) baseKey() imageCacheKey {
	return imageCacheKey{
		filename: b.filename,
		width:    b.width,
		height:   b.height,
	}
}

// tintKey generates a cache key representing the tinted image layer.
func (b *ImageBuilder) tintKey() imageCacheKey {
	r, g, bl, a := colorKey(b.color)
	key := b.baseKey()
	key.r, key.g, key.b, key.a = r, g, bl, a
	return key
}

// shadowKey generates a cache key representing the final layer with shadow effects applied.
func (b *ImageBuilder) shadowKey() imageCacheKey {
	cr, cg, cb, ca := colorKey(b.color)
	sr, sg, sb, _ := colorKey(b.shadow.color)
	key := b.baseKey()
	key.r, key.g, key.b, key.a = cr, cg, cb, ca
	key.shadow = shadowCacheKey{
		offsetX: b.shadow.offsetX,
		offsetY: b.shadow.offsetY,
		alpha:   b.shadow.alpha,
		r:       sr,
		g:       sg,
		b:       sb,
	}
	return key
}

// TintImage returns a copy of src tinted with iconColor, preserving the
// original per-pixel brightness and alpha (e.g. for recoloring status icons).
func TintImage(src *ebiten.Image, iconColor color.Color) *ebiten.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	result := ebiten.NewImage(w, h)

	tr, tg, tb, _ := iconColor.RGBA()

	for y := range h {
		for x := range w {
			cr, cg, cb, ca := src.At(x, y).RGBA()
			if ca == 0 {
				continue
			}

			brightness := (float32(cr) + float32(cg) + float32(cb)) / (3 * float32(ca))

			result.Set(x, y, color.NRGBA{
				R: uint8(float32(tr>>8) * brightness),
				G: uint8(float32(tg>>8) * brightness),
				B: uint8(float32(tb>>8) * brightness),
				A: uint8(ca >> 8),
			})
		}
	}

	return result
}

// SolidImage returns a cached 1x1 *ebiten.Image filled with c.
func SolidImage(c color.Color) *ebiten.Image {
	r, g, b, a := c.RGBA()
	key := struct{ r, g, b, a uint32 }{r, g, b, a}

	if cached, ok := solidImageCache.Load(key); ok {
		return cached.(*ebiten.Image) //nolint:forcetypeassert
	}

	img := ebiten.NewImage(1, 1)
	img.Fill(c)
	solidImageCache.Store(key, img)
	return img
}
