package ui

import (
	"image"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui/input"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
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

	// bgColor is the current fill color of the hex.
	// Change it via SetColor — no image allocation needed.
	bgColor color.RGBA

	// cachedRect is the last rect assigned by HexLayout.
	// All layers are kept in sync with this rect via SetLocation.
	cachedRect image.Rectangle
}

// NewHexCellWidget creates a hex cell widget at the given axial coordinate.
// opts are passed to the underlying widget.Widget for input handling.
func NewHexCellWidget(coord ds.HexCoord, opts ...widget.WidgetOpt) *HexCellWidget {
	h := &HexCellWidget{
		Coord:   coord,
		bgColor: color.RGBA{60, 60, 80, 255},
		layers:  make(map[int]*widget.Container),
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

// buildHexPath appends a pointy-top hexagon centered in cachedRect to path.
// This is the single source of hex geometry — all render methods use it.
func (h *HexCellWidget) buildHexPath(path *vector.Path) {
	cx := float32(h.cachedRect.Min.X + h.cachedRect.Dx()/2)
	cy := float32(h.cachedRect.Min.Y + h.cachedRect.Dy()/2)
	r := float32(HexRadius)

	for i := 0; i < 6; i++ {
		// Pointy-top orientation: first vertex points straight up (-π/2 offset).
		angle := float64(i)*math.Pi/3 - math.Pi/2
		x := cx + r*float32(math.Cos(angle))
		y := cy + r*float32(math.Sin(angle))
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
	}
	path.Close()
}

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

// renderLayer renders the container for the given z-index if it exists.
// Called only via the named public methods above.
func (h *HexCellWidget) renderLayer(screen *ebiten.Image, z int) {
	if layer, ok := h.layers[z]; ok {
		layer.Render(screen)
	}
}

// --- Color ---

// SetColor updates the hex fill color for the next RenderFill call.
// Accepts any color.Color implementation (color.RGBA, color.NRGBA, etc).
// No image allocation occurs — only a color.RGBA field is written.
func (h *HexCellWidget) SetColor(c color.Color) {
	r, g, b, a := c.RGBA()
	h.bgColor = color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
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
	ds_ := math.Abs(rs - s)

	if dq > dr && dq > ds_ {
		rq = -rr - rs
	} else if dr > ds_ {
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
		if il, ok := interface{}(layer).(input.Layerer); ok {
			il.SetupInputLayer(def)
		}
	}
}
