package backdrop

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/goplease-game/client/asset"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// Configuration constants for step 4.
	hexBranch1Min = 50
	hexBranch1Max = 100

	// Animation durations (in seconds).
	hexFadeInDuration  = 0.6
	hexFadeOutDuration = 0.4
	hexDelayBetweenPx  = 0.2
)

func init() {
	hexTexture = bakeHexTexture(hexTextureSize, hexApothem)
}

// bakeHexTexture rasterizes a filled regular hexagon with analytical antialiasing.
// It uses a distance field approach combined with supersampling to ensure perfectly smooth edges.
func bakeHexTexture(texSize int, apothem float64) *ebiten.Image {
	img := image.NewNRGBA(image.Rect(0, 0, texSize, texSize))
	cx, cy := float64(texSize)/2, float64(texSize)/2

	// Precompute normals for the 6 half-planes
	var nx, ny [6]float64
	for i := range 6 {
		theta := float64(i) * math.Pi / 3
		nx[i], ny[i] = math.Cos(theta), math.Sin(theta)
	}

	for y := range texSize {
		for x := range texSize {
			// 4x Supersampling (SSAA) for smooth edge integration
			subSamples := [4][2]float64{
				{-0.25, -0.25}, {0.25, -0.25},
				{-0.25, 0.25}, {0.25, 0.25},
			}
			insideCount := 0.0

			for _, ss := range subSamples {
				px := float64(x) + 0.5 + ss[0] - cx
				py := float64(y) + 0.5 + ss[1] - cy

				inHex := true
				for i := range 6 {
					dist := px*nx[i] + py*ny[i]
					// Soft edge transition zone (0.5px) for analytical AA
					if dist > apothem+0.25 {
						inHex = false
						break
					}
				}
				if inHex {
					insideCount += 0.25
				}
			}

			if insideCount > 0 {
				alpha := uint8(insideCount * 255)
				img.Set(x, y, color.NRGBA{255, 255, 255, alpha})
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}

// HexState represents the current state of the backdrop animation sequence.
type HexState int

// HexState defines the visual lifecycle stages of a floating hexagon.
const (
	StateGrowing HexState = iota
	StateFadingOut
)

// AnimatedHex stores data for a single hexagon inside the growing structure.
type AnimatedHex struct {
	q, r           int
	localX, localY float64
	img            *ebiten.Image
	tint           color.RGBA
	targetAlpha    float32
	currentAlpha   float32
	isFadingIn     bool
}

// CrystalHexField implements a programmable backdrop that grows procedural
// clusters of hexagons and fades them out sequentially.
type CrystalHexField struct {
	width, height int
	rnd           *rand.Rand

	state            HexState
	centerX, centerY float64
	sharedRadius     float64

	hexes    []*AnimatedHex
	occupied map[[2]int]*AnimatedHex

	limitBranch1 int
	limitBranch2 int
	limitTotal   int

	spawnTimer     float64
	activeHexIndex int
}

// NewCrystalHexField creates and initializes a new CrystalHexField.
func NewCrystalHexField(width, height int) *CrystalHexField {
	phf := &CrystalHexField{
		width:  width,
		height: height,
		rnd:    rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())), //nolint:gosec
	}
	phf.resetStructure()
	return phf
}

// Update advances the animation timelines for fading in/out hex elements.
func (phf *CrystalHexField) Update() {
	const dt = 1.0 / 60.0

	switch phf.state {
	case StateGrowing:
		if phf.activeHexIndex < len(phf.hexes) {
			curr := phf.hexes[phf.activeHexIndex]
			if curr.isFadingIn {
				curr.currentAlpha += (curr.targetAlpha / hexFadeInDuration) * dt
				if curr.currentAlpha >= curr.targetAlpha {
					curr.currentAlpha = curr.targetAlpha
					curr.isFadingIn = false
					phf.spawnTimer = 0
				}
			} else {
				phf.spawnTimer += dt
				if phf.spawnTimer >= hexDelayBetweenPx {
					phf.growNextStep()
				}
			}
		}
	case StateFadingOut:
		// Step 6: Sequentially fade out elements in the order they were spawned
		if phf.activeHexIndex < len(phf.hexes) {
			curr := phf.hexes[phf.activeHexIndex]
			curr.currentAlpha -= (curr.targetAlpha / hexFadeOutDuration) * dt
			if curr.currentAlpha <= 0 {
				curr.currentAlpha = 0
				phf.activeHexIndex++
			}
		} else {
			phf.resetStructure()
		}
	}
}

// Draw renders all visible hex layers onto the screen surface destination.
func (phf *CrystalHexField) Draw(screen *ebiten.Image) {
	scale := phf.sharedRadius / hexTexRadius

	for _, h := range phf.hexes {
		if h.currentAlpha <= 0 {
			continue
		}

		worldX := phf.centerX + h.localX
		worldY := phf.centerY + h.localY

		// Draw background hexagon pattern
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterLinear
		op.GeoM.Translate(-float64(hexTextureSize)/2, -float64(hexTextureSize)/2)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(worldX, worldY)

		op.ColorScale.Scale(
			float32(h.tint.R)/255,
			float32(h.tint.G)/255,
			float32(h.tint.B)/255,
			1,
		)
		op.ColorScale.ScaleAlpha(h.currentAlpha)
		screen.DrawImage(hexTexture, op)

		// Draw inner portrait image asset
		if h.img != nil {
			pw, ph := h.img.Bounds().Dx(), h.img.Bounds().Dy()
			pop := &ebiten.DrawImageOptions{}
			pop.Filter = ebiten.FilterLinear
			pop.GeoM.Translate(-float64(pw)/2, -float64(ph)/2)
			pop.GeoM.Translate(worldX, worldY)

			imageAlphaFactor := (h.currentAlpha / h.targetAlpha) * hexImageAlpha
			pop.ColorScale.ScaleAlpha(imageAlphaFactor)
			screen.DrawImage(h.img, pop)
		}
	}
}

// Resize recalculates internal boundary dimensions proportionally on screen modifications.
func (phf *CrystalHexField) Resize(width, height int) {
	if width == phf.width && height == phf.height {
		return
	}
	scaleX := float64(width) / float64(phf.width)
	scaleY := float64(height) / float64(phf.height)
	phf.centerX *= scaleX
	phf.centerY *= scaleY
	phf.width, phf.height = width, height
}

// resetStructure clears the field and starts a brand new growth cycle from Step 1.
func (phf *CrystalHexField) resetStructure() {
	phf.state = StateGrowing
	phf.hexes = nil
	phf.occupied = make(map[[2]int]*AnimatedHex)
	phf.activeHexIndex = 0
	phf.spawnTimer = 0

	phf.sharedRadius = hexMinRadius + phf.rnd.Float64()*(hexMaxRadius-hexMinRadius)

	// Step 4: Position the first hexagon closer to the screen center (within the middle 50% area)
	phf.centerX = float64(phf.width)*0.25 + phf.rnd.Float64()*(float64(phf.width)*0.5)
	phf.centerY = float64(phf.height)*0.25 + phf.rnd.Float64()*(float64(phf.height)*0.5)

	// Step 4, 5, 6: Dynamic generation of limits based on the branch configuration constants
	phf.limitBranch1 = hexBranch1Min + phf.rnd.IntN(hexBranch1Max-hexBranch1Min+1)
	phf.limitBranch2 = phf.limitBranch1 * 2
	phf.limitTotal = phf.limitBranch1 * 3

	// Create X1 at origin
	firstHex := phf.spawnHexAt(0, 0)
	phf.hexes = append(phf.hexes, firstHex)
	phf.occupied[[2]int{0, 0}] = firstHex
}

// spawnHexAt instantiates a new hex data structure at the given axial coordinates.
func (phf *CrystalHexField) spawnHexAt(q, r int) *AnimatedHex {
	lx, ly := axialToPixel(q, r, phf.sharedRadius)

	imgSize := max(int(phf.sharedRadius)-hexImageMargin, hexImageMinPx)
	idx := phf.rnd.IntN(hexUnitPicCount) + 1
	img := asset.Image(fmt.Sprintf("units/unit_%d_pic.png", idx), imgSize)

	h := &AnimatedHex{
		q:            q,
		r:            r,
		localX:       lx,
		localY:       ly,
		img:          img,
		tint:         hexColors[phf.rnd.IntN(len(hexColors))],
		targetAlpha:  float32(0.1 + phf.rnd.Float64()*0.2),
		currentAlpha: 0,
		isFadingIn:   true,
	}
	return h
}

// isOutOfBounds checks if any of the 6 outer vertices of a hex exceed the screen dimensions.
func (phf *CrystalHexField) isOutOfBounds(h *AnimatedHex) bool {
	worldX := phf.centerX + h.localX
	worldY := phf.centerY + h.localY

	// Check all 6 corner vertices of the pointy-top hexagon
	for i := range 6 {
		angle := float64(i)*math.Pi/3 + math.Pi/6
		vx := worldX + phf.sharedRadius*math.Cos(angle)
		vy := worldY + phf.sharedRadius*math.Sin(angle)

		if vx < 0 || vx > float64(phf.width) || vy < 0 || vy > float64(phf.height) {
			return true
		}
	}
	return false
}

// growNextStep handles branching logic and boundary safety validations.
func (phf *CrystalHexField) growNextStep() {
	count := len(phf.hexes)

	if count >= phf.limitTotal {
		phf.state = StateFadingOut
		phf.activeHexIndex = 0
		return
	}

	var parent *AnimatedHex
	if count == phf.limitBranch1 || count == phf.limitBranch2 {
		// Step 4 & 5: Force branch off from X1 (index 0)
		parent = phf.hexes[0]
	} else {
		// Step 2 & 3: Standard growth from the last spawned hexagon
		parent = phf.hexes[count-1]
	}

	shuffledDirs := phf.rnd.Perm(6)
	found := false

	for _, dirIdx := range shuffledDirs {
		dir := axialNeighborDirs[dirIdx]
		candQ := parent.q + dir[0]
		candR := parent.r + dir[1]

		if _, exists := phf.occupied[[2]int{candQ, candR}]; !exists {
			candHex := phf.spawnHexAt(candQ, candR)

			// FIX: Instead of abruptly resetting, trigger a graceful fade-out
			if phf.isOutOfBounds(candHex) {
				phf.state = StateFadingOut
				phf.activeHexIndex = 0
				return
			}

			phf.hexes = append(phf.hexes, candHex)
			phf.occupied[[2]int{candQ, candR}] = candHex
			phf.activeHexIndex = len(phf.hexes) - 1
			found = true
			break
		}
	}

	// Fallback strategy if the chosen parent is fully trapped by neighbors
	if !found {
		for i := len(phf.hexes) - 1; i >= 0; i-- {
			p := phf.hexes[i]
			for _, dir := range axialNeighborDirs {
				candQ := p.q + dir[0]
				candR := p.r + dir[1]
				if _, exists := phf.occupied[[2]int{candQ, candR}]; !exists {
					candHex := phf.spawnHexAt(candQ, candR)

					// FIX: Apply the same graceful fade-out to the fallback logic
					if phf.isOutOfBounds(candHex) {
						phf.state = StateFadingOut
						phf.activeHexIndex = 0
						return
					}

					phf.hexes = append(phf.hexes, candHex)
					phf.occupied[[2]int{candQ, candR}] = candHex
					phf.activeHexIndex = len(phf.hexes) - 1
					return
				}
			}
		}
		phf.state = StateFadingOut
		phf.activeHexIndex = 0
	}
}
