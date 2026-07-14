package backdrop

import (
	"math"
	"math/rand/v2"

	"github.com/goplease-game/client/asset"
	"github.com/hajimehoshi/ebiten/v2"
)

// MoteField is a small full-screen ambient particle layer, similar in spirit
// to backdrop.HexField but lighter — used as a subtle overlay above the
// board rather than a full backdrop.
type MoteField struct {
	width, height int
	motes         []mote
}

// mote is a single slow-drifting ambient particle, fixed alpha for its
// lifetime — only position animates, avoiding any flicker risk.
type mote struct {
	x, y   float64
	vx, vy float64
	alpha  float32
	img    *ebiten.Image
}

const (
	moteMinAlpha = 0.05
	moteMaxAlpha = 0.12
	moteMinSpeed = 5.0 // px/sec, same range as backdrop.hexMinSpeed
	moteMaxSpeed = 16.0
	moteCount    = 8

	moteMinSizePx = 8 // px, randomized on-screen size range
	moteMaxSizePx = 20
)

// NewMoteField creates a MoteField sized for width x height pixels.
func NewMoteField(width, height int) *MoteField {
	mf := &MoteField{width: width, height: height}
	mf.motes = make([]mote, moteCount)
	for i := range mf.motes {
		angle := rand.Float64() * math.Pi * 2
		speed := moteMinSpeed + rand.Float64()*(moteMaxSpeed-moteMinSpeed)
		size := moteMinSizePx + rand.IntN(moteMaxSizePx-moteMinSizePx+1)

		mf.motes[i] = mote{
			x:     rand.Float64() * float64(width),
			y:     rand.Float64() * float64(height),
			vx:    math.Cos(angle) * speed,
			vy:    math.Sin(angle) * speed,
			alpha: float32(moteMinAlpha + rand.Float64()*(moteMaxAlpha-moteMinAlpha)),
			img:   asset.Image("board-mote.png", size),
		}
	}
	return mf
}

// Update ...
func (mf *MoteField) Update() {
	const dt = 1.0 / 60.0
	w, h := float64(mf.width), float64(mf.height)

	for i := range mf.motes {
		m := &mf.motes[i]
		m.x += m.vx * dt
		m.y += m.vy * dt

		margin := float64(m.img.Bounds().Dx()) / 2
		if m.x < -margin {
			m.x = w + margin
		} else if m.x > w+margin {
			m.x = -margin
		}
		if m.y < -margin {
			m.y = h + margin
		} else if m.y > h+margin {
			m.y = -margin
		}
	}
}

// Draw renders each mote using its own pre-sized particle image, centered
// on its current position and faded by its fixed alpha.
func (mf *MoteField) Draw(screen *ebiten.Image) {
	for _, m := range mf.motes {
		pw, ph := m.img.Bounds().Dx(), m.img.Bounds().Dy()

		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterLinear
		op.GeoM.Translate(-float64(pw)/2, -float64(ph)/2)
		op.GeoM.Translate(m.x, m.y)
		op.ColorScale.ScaleAlpha(m.alpha)
		screen.DrawImage(m.img, op)
	}
}

// Resize rescales existing mote positions proportionally to the new
// dimensions, same approach as backdrop.HexField.Resize.
func (mf *MoteField) Resize(width, height int) {
	if width == mf.width && height == mf.height {
		return
	}
	scaleX := float64(width) / float64(mf.width)
	scaleY := float64(height) / float64(mf.height)
	for i := range mf.motes {
		mf.motes[i].x *= scaleX
		mf.motes[i].y *= scaleY
	}
	mf.width, mf.height = width, height
}
