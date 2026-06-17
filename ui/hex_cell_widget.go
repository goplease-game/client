package ui

import (
	"image"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/ds"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// z-index constants define the render order for hex cell layers.
// Higher values are drawn on top.
const (
	zIndexUnit = 0   // unit portraits and base visuals
	zIndexHUD  = 100 // hp badges, status icons, UI overlays
	zIndexFX   = 200 // damage numbers, attack effects, temporary visuals
)

// HexCellWidget is a custom widget that renders a hexagonal cell directly
// via vector.FillPath, bypassing EbitenUI's rectangular background system.
// This gives us: correct hex-shaped hit testing, zero-alloc color updates,
// and full control over the render layer order in Screen.Draw.
//
// Child widgets are organized into named z-index layers (unit, HUD, FX).
// Each layer is a separate EbitenUI container so layout and rendering
// are handled correctly by EbitenUI.
type HexCellWidget struct {
	widget *widget.Widget
	Coord  ds.HexCoord

	// layers maps z-index to an EbitenUI container for that render layer.
	layers map[int]*widget.Container

	// bgColor is the current fill color of the grid.
	// Change it via SetColor — no image allocation needed.
	bgColor color.RGBA

	// cachedRect is the last rect assigned by HexLayout.
	// All layers are kept in sync with this rect via SetLocation.
	cachedRect image.Rectangle

	boardCells map[ds.HexCoord]*ds.BoardCell // reference to board for boundary detection
}

// NewHexCellWidget creates a hex cell widget at the given axial coordinate.
// opts are passed to the underlying widget.Widget for input handling.
func NewHexCellWidget(coord ds.HexCoord, boardCells map[ds.HexCoord]*ds.BoardCell, opts ...widget.WidgetOpt) *HexCellWidget {
	h := &HexCellWidget{
		Coord:      coord,
		bgColor:    color.RGBA{60, 60, 80, 255},
		layers:     make(map[int]*widget.Container),
		boardCells: boardCells,
	}
	h.widget = widget.NewWidget(opts...)
	return h
}

// --- EbitenUI interfaces ---

// GetWidget satisfies widget.PreferredSizeLocateableWidget and HexChild.
func (h *HexCellWidget) GetWidget() *widget.Widget {
	return h.widget
}

// GetHexCoord satisfies the HexChild interface so HexLayout can position this widget.
func (h *HexCellWidget) GetHexCoord() ds.HexCoord {
	return h.Coord
}

// PreferredSize returns the bounding box of a pointy-top hex with radius HexRadius.
// HexLayout calls this to know how much space to reserve.
func (h *HexCellWidget) PreferredSize() (int, int) {
	w := int(math.Sqrt(3) * float64(HexRadius))
	ht := int(2 * float64(HexRadius))
	return w, ht
}

// SetLocation is called by HexLayout after it computes the screen rect for this cell.
// All layers are repositioned so EbitenUI lays out child widgets correctly.
func (h *HexCellWidget) SetLocation(rect image.Rectangle) {
	h.widget.SetLocation(rect)

	for _, layer := range h.layers {
		layer.SetLocation(rect)
	}

	h.cachedRect = rect
}

// Validate satisfies widget.PreferredSizeLocateableWidget.
// EbitenUI calls this during the layout pass to check widget state.
func (h *HexCellWidget) Validate() {
	// No validation needed for a hex cell.
}

// --- Rendering ---

// RenderFill draws the hex polygon fill onto screen using the current bgColor.
// Anti-aliasing is intentionally disabled to prevent boundary flickering
// between adjacent hex cells.
func (h *HexCellWidget) RenderFill(screen *ebiten.Image) {
	if h.cachedRect.Empty() {
		return
	}
	var path vector.Path
	h.buildHexPath(&path)
	var opts vector.DrawPathOptions
	opts.AntiAlias = false
	opts.ColorScale.ScaleWithColor(h.bgColor)
	vector.FillPath(screen, &path, &vector.FillOptions{}, &opts)
}

// AppendHexPath appends this cell's hex polygon to an existing path.
// More efficient than building a new path when batching multiple cells
// (e.g. for grid stroke rendering in Screen.renderGrid).
func (h *HexCellWidget) AppendHexPath(path *vector.Path) {
	if h.cachedRect.Empty() {
		return
	}
	h.buildHexPath(path)
}

// NOTE: this alternative buildHexPath could draw circle cells (set cornerRadius)
// buildHexPath appends a pointy-top hexagon with rounded corners centered in cachedRect to path.
// cornerRadius controls how much each vertex is chamfered — 0 gives sharp corners.
/*
func (h *HexCellWidget) buildHexPath(path *vector.Path) {
	const cornerRadius = 0.0 // pixels

	cx := float32(h.cachedRect.Min.X + h.cachedRect.Dx()/2)
	cy := float32(h.cachedRect.Min.Y + h.cachedRect.Dy()/2)
	r := float32(HexRadius)

	// Compute all 6 vertices.
	vx := [6]float32{}
	vy := [6]float32{}
	for i := 0; i < 6; i++ {
		angle := float64(i)*math.Pi/3 - math.Pi/2
		vx[i] = cx + r*float32(math.Cos(angle))
		vy[i] = cy + r*float32(math.Sin(angle))
	}

	for i := 0; i < 6; i++ {
		curr := i
		next := (i + 1) % 6
		prev := (i + 5) % 6

		// Vectors from current vertex toward neighbors.
		toPrevX := vx[prev] - vx[curr]
		toPrevY := vy[prev] - vy[curr]
		toNextX := vx[next] - vx[curr]
		toNextY := vy[next] - vy[curr]

		lenPrev := float32(math.Sqrt(float64(toPrevX*toPrevX + toPrevY*toPrevY)))
		lenNext := float32(math.Sqrt(float64(toNextX*toNextX + toNextY*toNextY)))

		// Points on edges at cornerRadius distance from vertex.
		p1x := vx[curr] + toPrevX/lenPrev*cornerRadius
		p1y := vy[curr] + toPrevY/lenPrev*cornerRadius
		p2x := vx[curr] + toNextX/lenNext*cornerRadius
		p2y := vy[curr] + toNextY/lenNext*cornerRadius

		if i == 0 {
			path.MoveTo(p1x, p1y)
		} else {
			path.LineTo(p1x, p1y)
		}

		// Quadratic bezier through the vertex rounds the corner.
		path.QuadTo(vx[curr], vy[curr], p2x, p2y)
	}

	path.Close()
}
*/

// RenderUnitLayer renders all widgets on the unit layer (z=0).
// Call from PostRenderHook after RenderFill and grid stroke.
func (h *HexCellWidget) RenderUnitLayer(screen *ebiten.Image) {
	h.renderLayer(screen, zIndexUnit)
}

// RenderHUDLayer renders all widgets on the HUD layer (z=100).
// Call from PostRenderHook after RenderUnitLayer.
func (h *HexCellWidget) RenderHUDLayer(screen *ebiten.Image) {
	h.renderLayer(screen, zIndexHUD)
}

// RenderFXLayer renders all widgets on the FX layer (z=200).
// Call from PostRenderHook after RenderHUDLayer.
func (h *HexCellWidget) RenderFXLayer(screen *ebiten.Image) {
	h.renderLayer(screen, zIndexFX)
}

// --- Color ---

// SetColor updates the hex fill color for the next RenderFill call.
// Accepts any color.Color implementation (color.RGBA, color.NRGBA, etc).
// No image allocation occurs — only a color.RGBA field is written.
func (h *HexCellWidget) SetColor(c color.Color) {
	r, g, b, a := c.RGBA()
	h.bgColor = color.RGBA{
		R: uint8(r >> 8), //nolint:gosec
		G: uint8(g >> 8), //nolint:gosec
		B: uint8(b >> 8), //nolint:gosec
		A: uint8(a >> 8), //nolint:gosec
	}
}

// --- Child widget management ---

// AddChild adds widgets to the unit layer (z=0).
// Implements ChildAdder for compatibility with buildBoardCard.
func (h *HexCellWidget) AddChild(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return h.layerFor(zIndexUnit).AddChild(children...)
}

// AddToUnitLayer adds widgets to the unit layer (z=0), rendered below HUD.
// Use for unit portraits and other base-level visuals.
func (h *HexCellWidget) AddToUnitLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return h.layerFor(zIndexUnit).AddChild(children...)
}

// AddToHUDLayer adds widgets to the HUD layer (z=100), rendered above units.
// Use for hp badges, status icons, and other UI overlays.
func (h *HexCellWidget) AddToHUDLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return h.layerFor(zIndexHUD).AddChild(children...)
}

// AddToFXLayer adds widgets to the FX layer (z=200), rendered above HUD.
// Use for damage numbers, attack effects, and other temporary visuals.
func (h *HexCellWidget) AddToFXLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return h.layerFor(zIndexFX).AddChild(children...)
}

// RemoveChild removes a specific child widget from the unit layer.
func (h *HexCellWidget) RemoveChild(w widget.PreferredSizeLocateableWidget) {
	h.layerFor(zIndexUnit).RemoveChild(w)
}

// RemoveChildren removes all child widgets from all layers.
func (h *HexCellWidget) RemoveChildren() {
	for _, layer := range h.layers {
		layer.RemoveChildren()
	}
}

// --- Hit testing ---

// HitTest returns true if the screen point (mx, my) falls inside the hex polygon.
// Uses cube-coordinate rounding so the result matches the true hex shape,
// not the rectangular bounding box that EbitenUI uses for mouse events.
func (h *HexCellWidget) HitTest(mx, my int) bool {
	rect := h.cachedRect
	cx := float64(rect.Min.X + rect.Dx()/2)
	cy := float64(rect.Min.Y + rect.Dy()/2)

	dx := float64(mx) - cx
	dy := float64(my) - cy

	size := float64(HexRadius)

	// Convert pixel offset to fractional axial coordinates.
	q := (dx*math.Sqrt(3)/3 - dy/3) / size
	r := (dy * 2 / 3) / size

	// Third cube coordinate (q + r + s == 0).
	s := -q - r

	// Round to nearest hex center, then fix the coordinate with the largest
	// rounding error to keep q+r+s == 0.
	rq := math.Round(q)
	rr := math.Round(r)
	rs := math.Round(s)

	dq := math.Abs(rq - q)
	dr := math.Abs(rr - r)
	ds := math.Abs(rs - s)

	if dq > dr && dq > ds {
		rq = -rr - rs
	} else if dr > ds {
		rr = -rq - rs
	}

	// The point is inside this hex only if rounded coords are (0, 0).
	return rq == 0 && rr == 0
}

// CachedRect returns the current screen rectangle for this hex cell.
// Used by Screen.renderGrid and other rendering passes that need the cell bounds.
func (h *HexCellWidget) CachedRect() image.Rectangle {
	return h.cachedRect
}

// SetupInputLayer delegates input layer setup to all z-index layers.
// Required by EbitenUI to propagate input handling to child containers.
func (h *HexCellWidget) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	for _, layer := range h.layers {
		if il, ok := any(layer).(input.Layerer); ok {
			il.SetupInputLayer(def)
		}
	}
}

// buildHexPath appends a pointy-top hexagon to path, rounding only corners
// that are on the board boundary (both adjacent edges have no neighbor).
func (h *HexCellWidget) buildHexPath(path *vector.Path) {
	const cornerRadius = float32(HexRadius)

	cx := float32(h.cachedRect.Min.X + h.cachedRect.Dx()/2)
	cy := float32(h.cachedRect.Min.Y + h.cachedRect.Dy()/2)
	r := float32(HexRadius)

	neighborDirs := [6]ds.HexCoord{
		{Q: 1, R: -1},
		{Q: 1, R: 0},
		{Q: 0, R: 1},
		{Q: -1, R: 1},
		{Q: -1, R: 0},
		{Q: 0, R: -1},
	}

	hasNeighbor := [6]bool{}
	// #nosec G602 -- i comes from ranging over a fixed-size [6] array.
	for i, d := range neighborDirs {
		neighbor := ds.HexCoord{Q: h.Coord.Q + d.Q, R: h.Coord.R + d.R}
		_, hasNeighbor[i] = h.boardCells[neighbor]
	}

	// Vertex i is an outer corner if both adjacent edges have no neighbor.
	isOuterCorner := func(i int) bool {
		prev := (i + 5) % 6
		return !hasNeighbor[prev] && !hasNeighbor[i]
	}

	isBoardVertex := func() bool {
		// Count the maximum run of consecutive missing neighbors,
		// treating the array as circular.
		best := 0
		for start := range 6 {
			if hasNeighbor[start] {
				continue
			}
			count := 0
			for j := range 6 {
				if !hasNeighbor[(start+j)%6] {
					count++
				} else {
					break
				}
			}
			if count > best {
				best = count
			}
		}
		return best == 3
	}()

	vx := [6]float32{}
	vy := [6]float32{}
	for i := range 6 {
		angle := float64(i)*math.Pi/3 - math.Pi/2
		vx[i] = cx + r*float32(math.Cos(angle))
		vy[i] = cy + r*float32(math.Sin(angle))
	}

	started := false
	for i := range 6 {
		next := (i + 1) % 6
		prev := (i + 5) % 6

		toPrevX := vx[prev] - vx[i]
		toPrevY := vy[prev] - vy[i]
		toNextX := vx[next] - vx[i]
		toNextY := vy[next] - vy[i]

		lenPrev := float32(math.Sqrt(float64(toPrevX*toPrevX + toPrevY*toPrevY)))
		lenNext := float32(math.Sqrt(float64(toNextX*toNextX + toNextY*toNextY)))

		cr := cornerRadius
		// Board-vertex hexes get half the corner radius to avoid distortion.
		if isBoardVertex {
			cr = cornerRadius / 2
		}
		// Hard clamp to half edge length as a safety net.
		if safeRadius := lenPrev / 2; cr > safeRadius {
			cr = safeRadius
		}

		p1x := vx[i] + toPrevX/lenPrev*cr
		p1y := vy[i] + toPrevY/lenPrev*cr
		p2x := vx[i] + toNextX/lenNext*cr
		p2y := vy[i] + toNextY/lenNext*cr

		if isOuterCorner(i) {
			if !started {
				path.MoveTo(p1x, p1y)
				started = true
			} else {
				path.LineTo(p1x, p1y)
			}
			path.QuadTo(vx[i], vy[i], p2x, p2y)
		} else {
			if !started {
				path.MoveTo(vx[i], vy[i])
				started = true
			} else {
				path.LineTo(vx[i], vy[i])
			}
		}
	}

	path.Close()
}

// renderLayer renders the container for the given z-index if it exists.
// Called only via the named public methods above.
func (h *HexCellWidget) renderLayer(screen *ebiten.Image, z int) {
	if layer, ok := h.layers[z]; ok {
		layer.Render(screen)
	}
}

// layerFor returns the container for the given z-index, creating it if needed.
// The container is positioned at cachedRect so EbitenUI lays out children correctly.
func (h *HexCellWidget) layerFor(z int) *widget.Container {
	if c, ok := h.layers[z]; ok {
		return c
	}
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	if !h.cachedRect.Empty() {
		c.SetLocation(h.cachedRect)
	}
	h.layers[z] = c
	return c
}
