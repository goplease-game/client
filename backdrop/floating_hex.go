package backdrop

import (
	"fmt"
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/goplease-game/client/asset"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	hexTextureSize = 128 // px, baked once and reused for every instance
	hexApothem     = 46  // px, half-distance across flat sides in the baked texture

	hexMinRadius = 26.0 // px, on-screen corner-to-center radius range
	hexMaxRadius = 78.0

	hexMinSpeed = 5.0 // px/sec drift speed range
	hexMaxSpeed = 16.0

	hexMinRotSpeed = -0.12 // rad/sec spin range
	hexMaxRotSpeed = 0.12

	hexMinAlpha = 0.08
	hexMaxAlpha = 0.26

	hexAreaPerHex = 42000 // lower = denser field
	hexMinCount   = 50
	hexMaxCount   = 100

	hexImageAlpha  = 0.35 // portrait opacity, independent of the tile's own alpha
	hexImageMargin = 5    // px subtracted from hex radius to size the portrait
	hexImageMinPx  = 8    // px, floor so asset.Image never gets a degenerate size
)

// hexColors are the tint colors instances are drawn with, sharing the same
// palette family used for the nebula backdrop so the two feel related.
var hexColors = []color.RGBA{
	{0xB4, 0xE1, 0xEB, 255},
	{0xFF, 0xF7, 0x8D, 255},
	{0xFF, 0x60, 0x60, 255},
	{0xA5, 0xCF, 0x83, 255},
	{0x73, 0xA5, 0xCA, 255},
}

// hexTexture is a single baked white hexagon mask shared by every instance;
// per-instance color and transparency are applied at draw time via
// DrawImageOptions.ColorScale, so only one shape ever needs rasterizing.
var hexTexture *ebiten.Image

func init() {
	hexTexture = bakeHexTexture(hexTextureSize, hexApothem)
}

// pointInHexagon reports whether (x, y) lies inside a regular hexagon
// centered at the origin with the given apothem (center-to-edge distance).
// It tests the point against six half-planes whose normals are spaced 60
// degrees apart, which is a cheap and exact way to rasterize any regular
// hexagon regardless of orientation.
func pointInHexagon(x, y, apothem float64) bool {
	for i := range 6 {
		theta := float64(i) * math.Pi / 3
		nx, ny := math.Cos(theta), math.Sin(theta)
		if x*nx+y*ny > apothem {
			return false
		}
	}
	return true
}

// hexTexRadius is the corner-to-center radius baked into hexTexture,
// derived from its apothem. Draw scales instances relative to this so a
// requested on-screen radius maps to the correct GeoM scale factor.
var hexTexRadius = hexApothem / math.Cos(math.Pi/6)

// axialNeighborDirs are the six axial hex-grid neighbor offsets for a
// pointy-top layout, matching the orientation baked into hexTexture (see
// pointInHexagon: normals starting at 0deg give vertical flat edges on the
// left/right, i.e. points at top and bottom).
var axialNeighborDirs = [6][2]int{
	{1, 0}, {1, -1}, {0, -1}, {-1, 0}, {-1, 1}, {0, 1},
}

// hexClusterSpacing slightly exceeds 1 so adjacent hexes in a cluster show a
// thin gap at their shared edge instead of perfectly seamless tiling, which
// reads better once alpha blending is involved.
const hexClusterSpacing = 1.08

// genClusterAxial grows a random connected group of the given size on the
// hex grid: starting from a single cell, it repeatedly attaches a random
// unused neighbor of an already-included cell. This always yields a valid
// adjacent cluster regardless of shape.
func genClusterAxial(rnd *rand.Rand, size int) [][2]int {
	cells := [][2]int{{0, 0}}
	seen := map[[2]int]bool{{0, 0}: true}
	for len(cells) < size {
		base := cells[rnd.IntN(len(cells))]
		dir := axialNeighborDirs[rnd.IntN(len(axialNeighborDirs))]
		cand := [2]int{base[0] + dir[0], base[1] + dir[1]}
		if seen[cand] {
			continue
		}
		seen[cand] = true
		cells = append(cells, cand)
	}
	return cells
}

// axialToPixel converts axial hex coordinates to a local pixel offset for a
// pointy-top layout with the given corner radius and spacing multiplier.
func axialToPixel(q, r int, radius float64) (float64, float64) {
	spaced := radius * hexClusterSpacing
	x := spaced * math.Sqrt(3) * (float64(q) + float64(r)/2)
	y := spaced * 1.5 * float64(r)
	return x, y
}

// hexUnitPicCount is how many units/unit_N_pic.png variants exist to pick
// from; keep in sync with the asset set.
const hexUnitPicCount = 6

// hexMember is one hexagon's fixed offset from its group's center, computed
// once at generation time; the group only ever moves and rotates as a
// whole, so members stay adjacent for the group's entire lifetime. Each
// member also carries its own randomly chosen unit portrait, drawn upright
// on top of the tile regardless of the group's rotation.
type hexMember struct {
	localX, localY float64
	img            *ebiten.Image
}

// hexGroup is a small cluster of 2-4 same-sized, adjacent hexagons that
// drift and spin together as a single rigid body.
type hexGroup struct {
	x, y           float64 // center position
	vx, vy         float64
	rotation       float64
	rotSpeed       float64
	radius         float64 // shared corner radius for every member
	boundingRadius float64 // for edge-wrap margin, includes member offsets
	tint           color.RGBA
	alpha          float32
	members        []hexMember
}

func newHexGroup(width, height int, rnd *rand.Rand) hexGroup {
	size := 2 + rnd.IntN(3) // 2, 3, or 4
	radius := hexMinRadius + rnd.Float64()*(hexMaxRadius-hexMinRadius)

	cells := genClusterAxial(rnd, size)
	members := make([]hexMember, len(cells))
	boundingRadius := radius

	imgSize := max(int(radius)-hexImageMargin, hexImageMinPx)

	for i, cell := range cells {
		lx, ly := axialToPixel(cell[0], cell[1], radius)
		idx := rnd.IntN(hexUnitPicCount) + 1
		img := asset.Image(fmt.Sprintf("units/unit_%d_pic.png", idx), imgSize)
		members[i] = hexMember{localX: lx, localY: ly, img: img}
		if d := math.Hypot(lx, ly) + radius; d > boundingRadius {
			boundingRadius = d
		}
	}

	angle := rnd.Float64() * math.Pi * 2
	speed := hexMinSpeed + rnd.Float64()*(hexMaxSpeed-hexMinSpeed)
	return hexGroup{
		x:              rnd.Float64() * float64(width),
		y:              rnd.Float64() * float64(height),
		vx:             math.Cos(angle) * speed,
		vy:             math.Sin(angle) * speed,
		rotation:       rnd.Float64() * math.Pi * 2,
		rotSpeed:       hexMinRotSpeed + rnd.Float64()*(hexMaxRotSpeed-hexMinRotSpeed),
		radius:         radius,
		boundingRadius: boundingRadius,
		tint:           hexColors[rnd.IntN(len(hexColors))],
		alpha:          float32(hexMinAlpha + rnd.Float64()*(hexMaxAlpha-hexMinAlpha)),
		members:        members,
	}
}

// HexField is a Backdrop of small hex-grid clusters that drift in
// independent random directions, gently spinning as rigid groups, wrapping
// around screen edges.
type HexField struct {
	width, height int
	groups        []hexGroup
	rnd           *rand.Rand
}

// NewFloatingHexField creates a HexField sized for width x height pixels. The
// number of clusters scales with screen area so total hex count stays
// within sane bounds regardless of window size.
func NewFloatingHexField(width, height int) *HexField {
	rnd := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	target := (width * height) / hexAreaPerHex
	target = min(max(target, hexMinCount), hexMaxCount)

	var groups []hexGroup
	total := 0
	for total < target {
		g := newHexGroup(width, height, rnd)
		groups = append(groups, g)
		total += len(g.members)
	}
	return &HexField{width: width, height: height, groups: groups, rnd: rnd}
}

// Update advances every group's position and rotation by one frame,
// wrapping groups that drift past a screen edge to the opposite side. The
// wrap margin uses each group's bounding radius so the whole cluster fully
// leaves view before reappearing.
func (hf *HexField) Update() {
	const dt = 1.0 / 60.0
	w, h := float64(hf.width), float64(hf.height)

	for i := range hf.groups {
		g := &hf.groups[i]
		g.x += g.vx * dt
		g.y += g.vy * dt
		g.rotation += g.rotSpeed * dt

		margin := g.boundingRadius
		if g.x < -margin {
			g.x = w + margin
		} else if g.x > w+margin {
			g.x = -margin
		}
		if g.y < -margin {
			g.y = h + margin
		} else if g.y > h+margin {
			g.y = -margin
		}
	}
}

// Draw renders every hexagon in every group at its current rotated world
// position, sharing its group's tint, transparency, size, and rotation.
func (hf *HexField) Draw(screen *ebiten.Image) {
	for _, g := range hf.groups {
		scale := g.radius / hexTexRadius
		sin, cos := math.Sincos(g.rotation)

		for _, m := range g.members {
			worldX := g.x + m.localX*cos - m.localY*sin
			worldY := g.y + m.localX*sin + m.localY*cos

			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterLinear
			op.GeoM.Translate(-float64(hexTextureSize)/2, -float64(hexTextureSize)/2)
			op.GeoM.Rotate(g.rotation)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(worldX, worldY)
			op.ColorScale.Scale(
				float32(g.tint.R)/255,
				float32(g.tint.G)/255,
				float32(g.tint.B)/255,
				1,
			)
			op.ColorScale.ScaleAlpha(g.alpha)
			screen.DrawImage(hexTexture, op)

			if m.img != nil {
				pw, ph := m.img.Bounds().Dx(), m.img.Bounds().Dy()
				pop := &ebiten.DrawImageOptions{}
				pop.Filter = ebiten.FilterLinear
				pop.GeoM.Translate(-float64(pw)/2, -float64(ph)/2)
				pop.GeoM.Translate(worldX, worldY)
				pop.ColorScale.ScaleAlpha(hexImageAlpha)
				screen.DrawImage(m.img, pop)
			}
		}
	}
}

// Resize rescales existing group center positions proportionally to the new
// dimensions so the field stays spread across the whole screen, and updates
// wrap bounds for future frames. Cluster shape and hex size are left
// untouched since they're physical, not screen-relative.
func (hf *HexField) Resize(width, height int) {
	if width == hf.width && height == hf.height {
		return
	}
	scaleX := float64(width) / float64(hf.width)
	scaleY := float64(height) / float64(hf.height)
	for i := range hf.groups {
		hf.groups[i].x *= scaleX
		hf.groups[i].y *= scaleY
	}
	hf.width, hf.height = width, height
}
