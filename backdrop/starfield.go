package backdrop

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// StarState defines the phase of the star's lifecycle.
type StarState int

// StarState represents the lifecycle phases of an animated star.
const (
	StarStateFadeIn StarState = iota
	StarStateAlive
	StarStateFadeOut
)

// starLayer groups stars that share a depth, giving a cheap parallax effect:
// far layers move slowly and render small and dim, near layers move fast
// and render larger and brighter.
type starLayer struct {
	speed      float64
	minSize    float32
	maxSize    float32
	brightness float32 // Base alpha configuration, 0..1
}

var starLayers = []starLayer{
	{speed: 0.5, minSize: 0.3, maxSize: 0.6, brightness: 0.10}, // Far layer: tiny, almost invisible dots
	{speed: 1, minSize: 3, maxSize: 5, brightness: 0.15},       // Mid layer: standard stars
	{speed: 2.0, minSize: 6, maxSize: 12, brightness: 0.25},    // Near layer: huge, bright, fast-flying stars
}

const starsPerLayer = 50

// Space palette for diverse star colors.
var starPalette = []color.RGBA{
	{255, 255, 255, 255}, // Pure White
	{230, 240, 255, 255}, // Soft Blue
	{200, 225, 255, 255}, // Deep Blue-ish
	{255, 245, 220, 255}, // Warm Yellowish
	{255, 230, 230, 255}, // Soft Pink/Red tint
}

type star struct {
	x, y         float64
	vx, vy       float64 // Direction vectors for randomized movement
	size         float32
	baseColor    color.RGBA // Randomized color from the palette
	brightness   float32    // Individual randomness factor for transparency
	twinklePhase float64
	twinkleSpeed float64

	// New fields for the lifecycle fade system
	state        StarState
	lifeTimer    float64 // Countdown timer for the current state in seconds
	fadeAlpha    float32 // Local opacity modifier (0.0 to 1.0) during fade phases
	fadeInSpeed  float32 // How fast this specific star fades in (per second)
	fadeOutSpeed float32 // How fast this specific star fades out (per second)
}

// Starfield is a Backdrop of drifting, twinkling stars across several
// parallax layers.
type Starfield struct {
	width, height int
	layers        [][]star
	time          float64
}

// NewStarfield creates a Starfield sized for width x height pixels.
func NewStarfield(width, height int) *Starfield {
	sf := &Starfield{width: width, height: height}
	sf.layers = make([][]star, len(starLayers))
	for li, layer := range starLayers {
		stars := make([]star, starsPerLayer)
		for i := range stars {
			stars[i] = newStar(width, height, layer)
			// Stagger initial stars so they don't all fade in at the exact same moment.
			// Some will start already alive, some fading out, some fading in.
			stars[i].state = StarState(rand.IntN(3))
			switch stars[i].state {
			case StarStateAlive:
				stars[i].fadeAlpha = 1.0
				stars[i].lifeTimer = 5.0 + rand.Float64()*8.0
			default:
				stars[i].fadeAlpha = rand.Float32()
			}
		}
		sf.layers[li] = stars
	}
	return sf
}

func newStar(width, height int, layer starLayer) star {
	angle := rand.Float64() * math.Pi * 2

	// Step 1: Determine the exact size within this layer's bounds
	size := layer.minSize + rand.Float32()*(layer.maxSize-layer.minSize)

	// Step 2: Calculate a normalization factor (0.0 to 1.0) based on size inside this specific layer
	sizeFactor := float64((size - layer.minSize) / (layer.maxSize - layer.minSize + 0.001))

	// Step 3: Strictly tie speed and brightness to the size factor
	actualSpeed := layer.speed * (0.7 + sizeFactor*0.6)
	actualAlpha := layer.brightness * (0.3 + float32(sizeFactor)*0.7)

	// Random fade durations (e.g. takes between 1.0 and 2.5 seconds to fade in/out)
	fadeInDuration := 1.0 + rand.Float32()*1.5
	fadeOutDuration := 1.0 + rand.Float32()*1.5

	return star{
		x:            rand.Float64() * float64(width),
		y:            rand.Float64() * float64(height),
		vx:           math.Cos(angle) * actualSpeed,
		vy:           math.Sin(angle) * actualSpeed,
		size:         size,
		baseColor:    starPalette[rand.IntN(len(starPalette))],
		brightness:   actualAlpha,
		twinklePhase: rand.Float64() * math.Pi * 2,
		twinkleSpeed: 0.5 + rand.Float64()*0.8,

		// Lifecycle Initialization
		state:        StarStateFadeIn,
		fadeAlpha:    0.0, // Starts completely transparent
		fadeInSpeed:  1.0 / fadeInDuration,
		fadeOutSpeed: 1.0 / fadeOutDuration,
		lifeTimer:    0, // Will be set to a random lifetime once it finishes fading in
	}
}

// Update advances star positions, twinkle phase, and lifecycle states by one frame.
func (sf *Starfield) Update() {
	const dt = 1.0 / 60.0
	sf.time += dt
	w, h := float64(sf.width), float64(sf.height)

	for li, layer := range starLayers {
		stars := sf.layers[li]
		for i := range stars {
			s := &stars[i]

			// 1. Move according to the random vector
			s.x += s.vx * dt
			s.y += s.vy * dt

			// 2. Lifecycle State Machine Updates
			switch s.state {
			case StarStateFadeIn:
				s.fadeAlpha += s.fadeInSpeed * dt
				if s.fadeAlpha >= 1.0 {
					s.fadeAlpha = 1.0
					s.state = StarStateAlive
					// RANDOM LIFETIME: Star stays fully alive for a random duration between 5 and 14 seconds
					s.lifeTimer = 5.0 + rand.Float64()*10.0
				}

			case StarStateAlive:
				s.lifeTimer -= dt
				if s.lifeTimer <= 0 {
					s.state = StarStateFadeOut
				}

			case StarStateFadeOut:
				s.fadeAlpha -= s.fadeOutSpeed * dt
				if s.fadeAlpha <= 0 {
					// Star has fully deceased. Respawn it as a brand new star elsewhere!
					*s = newStar(sf.width, sf.height, layer)
				}
			}

			// 3. Screen wrapping bounds check for omnidirectional drift
			margin := float64(s.size)
			if s.x < -margin {
				s.x = w + margin
			} else if s.x > w+margin {
				s.x = -margin
			}

			if s.y < -margin {
				s.y = h + margin
			} else if s.y > h+margin {
				s.y = -margin
			}
		}
	}
}

// Draw renders all star layers back-to-front onto screen.
func (sf *Starfield) Draw(screen *ebiten.Image) {
	for li := range starLayers {
		for _, s := range sf.layers[li] {
			// Skip rendering completely invisible stars to save draw calls
			if s.fadeAlpha <= 0 {
				continue
			}

			twinkle := 0.9 + 0.1*float32(math.Sin(sf.time*s.twinkleSpeed+s.twinklePhase))

			// Multiply the layer & size brightness by the current lifecycle fade alpha
			alpha := clamp01(s.brightness * twinkle * s.fadeAlpha)

			// Calculate scaling factor based on the global high-res hexTexture (from hexfield.go)
			scale := float64(s.size) / hexTexRadius

			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterLinear

			// Center the texture, scale it down to star size, and position on screen
			op.GeoM.Translate(-float64(hexTextureSize)/2, -float64(hexTextureSize)/2)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(s.x, s.y)

			op.ColorScale.Scale(
				float32(s.baseColor.R)/255,
				float32(s.baseColor.G)/255,
				float32(s.baseColor.B)/255,
				1.0,
			)
			op.ColorScale.ScaleAlpha(alpha)

			screen.DrawImage(hexTexture, op)
		}
	}
}

// Resize rescales existing stars proportionally to the new dimensions so
// they stay spread across the whole screen instead of clustering in the
// old bounds, then updates spawn bounds for future respawns.
func (sf *Starfield) Resize(width, height int) {
	if width == sf.width && height == sf.height {
		return
	}
	scaleX := float64(width) / float64(sf.width)
	scaleY := float64(height) / float64(sf.height)
	for li := range sf.layers {
		stars := sf.layers[li]
		for i := range stars {
			stars[i].x *= scaleX
			stars[i].y *= scaleY
		}
	}
	sf.width, sf.height = width, height
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
