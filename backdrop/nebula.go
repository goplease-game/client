package backdrop

import (
	"image"
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// nebulaBakeWidth and nebulaBakeHeight control the resolution the noise is
// rendered at. Nebulae are inherently soft and blurry, so a low resolution
// upscaled with bilinear filtering looks correct and costs very little CPU.
const (
	nebulaBakeWidth  = 160
	nebulaBakeHeight = 90
	nebulaStateCount = 3     // 3 states is optimal for endless cycle buffer
	nebulaHoldTime   = 10.0  // seconds a state is fully visible before crossfading
	nebulaBlendTime  = 6.0   // seconds a crossfade takes
	nebulaDriftSpeed = 0.015 // slow parallax drift, in bake-pixels/sec
)

// nebulaStop is a single color/position pair in a gradient used to map a
// scalar noise value to a color.
type nebulaStop struct {
	pos float64
	c   color.RGBA
}

// nebulaPalettes are candidate gradients; one is picked per generated state.
var nebulaPalettes = [][]nebulaStop{
	{ // #B4E1EB accent — dark cyan-blue nebula
		{0.0, color.RGBA{6, 12, 15, 255}},
		{0.2, color.RGBA{18, 34, 39, 255}},
		{0.42, color.RGBA{0x51, 0x65, 0x6A, 255}},
		{0.68, color.RGBA{0x51, 0x65, 0x6A, 255}},
		{0.87, color.RGBA{0xB4, 0xE1, 0xEB, 255}},
		{1.0, color.RGBA{200, 232, 238, 255}},
	},
	{ // #FFF78D accent — dark olive-yellow nebula
		{0.0, color.RGBA{14, 12, 4, 255}},
		{0.2, color.RGBA{40, 35, 12, 255}},
		{0.42, color.RGBA{0x73, 0x6F, 0x3F, 255}},
		{0.68, color.RGBA{0x73, 0x6F, 0x3F, 255}},
		{0.87, color.RGBA{0xFF, 0xF7, 0x8D, 255}},
		{1.0, color.RGBA{255, 250, 200, 255}},
	},
	{ // #A5CF83 accent — dark green nebula
		{0.0, color.RGBA{7, 13, 4, 255}},
		{0.2, color.RGBA{23, 38, 14, 255}},
		{0.42, color.RGBA{0x4A, 0x5D, 0x3B, 255}},
		{0.68, color.RGBA{0x4A, 0x5D, 0x3B, 255}},
		{0.87, color.RGBA{0xA5, 0xCF, 0x83, 255}},
		{1.0, color.RGBA{200, 227, 180, 255}},
	},
	{ // #C5A3FF accent (Deep Violet)
		{0.0, color.RGBA{10, 6, 15, 255}},
		{0.2, color.RGBA{28, 18, 40, 255}},
		{0.42, color.RGBA{0x5B, 0x47, 0x75, 255}},
		{0.68, color.RGBA{0x5B, 0x47, 0x75, 255}},
		{0.87, color.RGBA{0xC5, 0xA3, 0xFF, 255}},
		{1.0, color.RGBA{230, 215, 255, 255}},
	},
	{ // #FF9EE2 accent (Cosmic Magenta)
		{0.0, color.RGBA{14, 5, 12, 255}},
		{0.2, color.RGBA{38, 14, 32, 255}},
		{0.42, color.RGBA{0x70, 0x35, 0x62, 255}},
		{0.68, color.RGBA{0x70, 0x35, 0x62, 255}},
		{0.87, color.RGBA{0xFF, 0x9E, 0xE2, 255}},
		{1.0, color.RGBA{255, 210, 243, 255}},
	},
	{ // #8096FF accent (Deep Indigo)
		{0.0, color.RGBA{4, 6, 16, 255}},
		{0.2, color.RGBA{12, 18, 42, 255}},
		{0.42, color.RGBA{0x3F, 0x4D, 0x7A, 255}},
		{0.68, color.RGBA{0x3F, 0x4D, 0x7A, 255}},
		{0.87, color.RGBA{0x80, 0x96, 0xFF, 255}},
		{1.0, color.RGBA{210, 220, 255, 255}},
	},
}

func gradientAt(stops []nebulaStop, t float64) color.RGBA {
	if t <= stops[0].pos {
		return stops[0].c
	}
	last := stops[len(stops)-1]
	if t >= last.pos {
		return last.c
	}
	for i := range len(stops) - 1 {
		a, b := stops[i], stops[i+1]
		if t >= a.pos && t <= b.pos {
			span := b.pos - a.pos
			f := 0.0
			if span > 0 {
				f = (t - a.pos) / span
			}
			return lerpRGBA(a.c, b.c, f)
		}
	}
	return last.c
}

func lerpRGBA(a, b color.RGBA, f float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*f),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*f),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*f),
		A: 255,
	}
}

// noiseGrid is a small wrapping grid of random values used as one octave of
// value noise. Sampling bilinearly interpolates between grid points with a
// smoothstep easing curve, which is cheap and looks organically cloudy.
type noiseGrid struct {
	size   int
	values []float64
}

func newNoiseGrid(size int, rnd *rand.Rand) *noiseGrid {
	values := make([]float64, size*size)
	for i := range values {
		values[i] = rnd.Float64()
	}
	return &noiseGrid{size: size, values: values}
}

func (g *noiseGrid) at(ix, iy int) float64 {
	ix = ((ix % g.size) + g.size) % g.size
	iy = ((iy % g.size) + g.size) % g.size
	return g.values[iy*g.size+ix]
}

func smoothstep(t float64) float64 {
	return t * t * (3 - 2*t)
}

// sample returns a value-noise sample in [0, 1] at normalized coordinates
// (u, v), each expected to range over [0, size).
func (g *noiseGrid) sample(u, v float64) float64 {
	x0, y0 := int(math.Floor(u)), int(math.Floor(v))
	fx, fy := smoothstep(u-float64(x0)), smoothstep(v-float64(y0))

	v00 := g.at(x0, y0)
	v10 := g.at(x0+1, y0)
	v01 := g.at(x0, y0+1)
	v11 := g.at(x0+1, y0+1)

	top := v00 + (v10-v00)*fx
	bottom := v01 + (v11-v01)*fx
	return top + (bottom-top)*fy
}

// fbm sums several octaves of noiseGrid at increasing frequency and
// decreasing amplitude (fractal Brownian motion), which is what gives the
// result its cloud/nebula-like structure instead of flat blob noise.
func fbm(grids []*noiseGrid, x, y float64) float64 {
	amplitude := 0.5
	total, maxAmp := 0.0, 0.0
	for i, g := range grids {
		freq := math.Pow(2, float64(i))
		total += g.sample(x*freq, y*freq) * amplitude
		maxAmp += amplitude
		amplitude *= 0.5
	}

	return total / maxAmp
}

// bakeNebulaTexture renders one nebula state to a low-resolution image using
// fbm noise mapped through a randomly chosen color gradient.
func bakeNebulaTexture(rnd *rand.Rand) *ebiten.Image {
	const octaves = 4
	grids := make([]*noiseGrid, octaves)
	for i := range grids {
		grids[i] = newNoiseGrid(3+i*2, rnd)
	}
	palette := nebulaPalettes[rnd.IntN(len(nebulaPalettes))]

	img := image.NewNRGBA(image.Rect(0, 0, nebulaBakeWidth, nebulaBakeHeight))
	const baseFreq = 3.0
	for y := range nebulaBakeHeight {
		v := float64(y) / float64(nebulaBakeHeight) * baseFreq
		for x := range nebulaBakeWidth {
			u := float64(x) / float64(nebulaBakeWidth) * baseFreq
			n := fbm(grids, u, v)
			// Push mid-tones down so the nebula reads as mostly dark space
			// with brighter wisps, rather than a uniform haze.
			n = math.Pow(n, 1.6)
			c := gradientAt(palette, n)
			img.Set(x, y, c)
		}
	}
	return ebiten.NewImageFromImage(img)
}

// nebulaState binds the compiled image texture with its unique independent drift physics.
type nebulaState struct {
	img         *ebiten.Image
	driftPhaseX float64
	driftPhaseY float64
}

// Nebula is a Backdrop that crossfades between a small set of pre-baked,
// softly drifting noise textures.
type Nebula struct {
	width, height int
	rnd           *rand.Rand
	states        []*nebulaState
	current, next int
	blendT        float64 // 0..1 progress of the current crossfade
	holdT         float64 // seconds remaining before the next crossfade starts
	blending      bool
	driftTime     float64
}

// NewNebula creates a Nebula sized for width x height pixels.
func NewNebula(width, height int) *Nebula {
	rnd := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	states := make([]*nebulaState, nebulaStateCount)
	for i := range states {
		states[i] = &nebulaState{
			img:         bakeNebulaTexture(rnd),
			driftPhaseX: rnd.Float64() * math.Pi * 2,
			driftPhaseY: rnd.Float64() * math.Pi * 2,
		}
	}
	return &Nebula{
		width:   width,
		height:  height,
		rnd:     rnd,
		states:  states,
		current: 0,
		next:    1,
		holdT:   nebulaHoldTime,
	}
}

// Resize updates the target dimensions the baked textures are stretched to.
// The bake resolution itself is independent of screen size, so nothing
// needs to be regenerated.
func (n *Nebula) Resize(width, height int) {
	n.width, n.height = width, height
}

// Update advances the crossfade timer and drift animation by one frame.
func (n *Nebula) Update() {
	const dt = 1.0 / 60.0
	n.driftTime += dt

	if n.blending {
		n.blendT += dt / nebulaBlendTime
		if n.blendT >= 1 {
			n.blendT = 0
			n.blending = false

			// The old current is completely hidden now, we can safe-recycle its index
			oldCurrent := n.current
			n.current = n.next

			// VARIANCE UPGRADE: Generate a brand new unique nebula structure on the fly
			// to guarantee endless variety instead of repeating old textures.
			n.states[oldCurrent] = &nebulaState{
				img:         bakeNebulaTexture(n.rnd),
				driftPhaseX: n.rnd.Float64() * math.Pi * 2,
				driftPhaseY: n.rnd.Float64() * math.Pi * 2,
			}

			// Next state points to the newly re-baked slot
			n.next = oldCurrent
			n.holdT = nebulaHoldTime
		}
		return
	}

	n.holdT -= dt
	if n.holdT <= 0 {
		n.blending = true
		n.blendT = 0
	}
}

// Draw renders the current nebula state, crossfading into the next state
// when a transition is in progress, with a slow drift applied to each layer.
func (n *Nebula) Draw(screen *ebiten.Image) {
	n.drawState(screen, n.states[n.current], 1)
	if n.blending {
		n.drawState(screen, n.states[n.next], float32(n.blendT))
	}
}

func (n *Nebula) drawState(screen *ebiten.Image, state *nebulaState, alpha float32) {
	sx := float64(n.width) / float64(nebulaBakeWidth)
	sy := float64(n.height) / float64(nebulaBakeHeight)

	// FIX: Use the state's persistent drift phases.
	// No matter if the index swaps from next to current, the coordinate physics remain perfectly uniform.
	const oversample = 1.08
	driftX := math.Sin(n.driftTime*nebulaDriftSpeed+state.driftPhaseX) * float64(nebulaBakeWidth) * 0.03
	driftY := math.Cos(n.driftTime*nebulaDriftSpeed*0.8+state.driftPhaseY) * float64(nebulaBakeHeight) * 0.03

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	op.GeoM.Translate(driftX-float64(nebulaBakeWidth)*(oversample-1)/2, driftY-float64(nebulaBakeHeight)*(oversample-1)/2)
	op.GeoM.Scale(sx*oversample, sy*oversample)
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(state.img, op)
}
