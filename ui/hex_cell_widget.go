package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

var boardCellBgColor = color.RGBA{0x45, 0x63, 0x7a, 255}

const (
	zIndexUnit = 0
	zIndexHUD  = 100
	zIndexFX   = 200
)

type hexChild struct {
	widget widget.PreferredSizeLocateableWidget
	zIndex int
}

// HexCellWidget is a custom widget that renders a hexagonal cell directly
// via DrawTriangles, bypassing EbitenUI's rectangular background system.
// This gives us: correct hex-shaped hit testing, zero-alloc color updates,
// and full control over the render layer order in Screen.Draw.
type HexCellWidget struct {
	widget *widget.Widget
	Coord  ds.HexCoord

	overlay *widget.Container // EbitenUI container for child widgets (hp badge, unit icon)
	layers  map[int]*widget.Container

	// bgColor is the current fill color of the hex.
	// Change it via SetColor — no image allocation needed.
	bgColor color.RGBA

	// cachedVs / cachedIs hold the triangle mesh for this hex.
	// They are rebuilt only when the widget rect changes (i.e. on layout),
	// not every frame, so no per-frame allocations occur during rendering.
	cachedVs   []ebiten.Vertex
	cachedIs   []uint16
	cachedRect image.Rectangle

	// children are nested widgets drawn on top of the hex (e.g. unit icons).
	children []widget.PreferredSizeLocateableWidget
}

func NewHexCellWidget(coord ds.HexCoord, opts ...widget.WidgetOpt) *HexCellWidget {
	h := &HexCellWidget{
		Coord:   coord,
		bgColor: color.RGBA{60, 60, 80, 255},
		layers:  make(map[int]*widget.Container),
	}
	h.widget = widget.NewWidget(opts...)

	// overlay is an invisible container that covers the hex rect exactly,
	// so EbitenUI can lay out child widgets (hp badge, unit icon) correctly.
	h.overlay = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

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
// We rebuild the triangle mesh here so Render never allocates.
func (h *HexCellWidget) SetLocation(rect image.Rectangle) {
	h.widget.SetLocation(rect)
	h.overlay.GetWidget().SetLocation(rect)

	for _, layer := range h.layers {
		layer.SetLocation(rect)
	}

	if rect == h.cachedRect {
		return
	}
	h.cachedRect = rect
	h.rebuildGeometry(rect)
}

// rebuildGeometry computes the six hex vertices centered in rect and
// triangulates them into cachedVs / cachedIs via vector.Path.
// Called once per layout pass, not per frame.
func (h *HexCellWidget) rebuildGeometry(rect image.Rectangle) {
	cx := float32(rect.Min.X + rect.Dx()/2)
	cy := float32(rect.Min.Y + rect.Dy()/2)
	r := float32(HexRadius)

	var path vector.Path
	for i := 0; i < 6; i++ {
		// pointy-top orientation: first vertex points straight up (-π/2 offset)
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

	h.cachedVs, h.cachedIs = path.AppendVerticesAndIndicesForFilling(
		h.cachedVs[:0], h.cachedIs[:0],
	)
}

// --- Rendering ---

func (h *HexCellWidget) RenderFill(screen *ebiten.Image) {
	if h.cachedRect.Empty() {
		return
	}
	path := h.hexPath()
	var opts vector.DrawPathOptions
	opts.AntiAlias = false
	opts.ColorScale.ScaleWithColor(h.bgColor)
	vector.FillPath(screen, &path, &vector.FillOptions{}, &opts)
}

func (h *HexCellWidget) RenderStroke(screen *ebiten.Image) {
	if h.cachedRect.Empty() {
		return
	}
	path := h.hexPath()
	var opts vector.DrawPathOptions
	opts.AntiAlias = true
	opts.ColorScale.ScaleWithColor(boardCellBgColor)
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: 1}, &opts)
}

// hexPath builds the hex vector.Path for this cell's current rect.
func (h *HexCellWidget) hexPath() vector.Path {
	cx := float32(h.cachedRect.Min.X + h.cachedRect.Dx()/2)
	cy := float32(h.cachedRect.Min.Y + h.cachedRect.Dy()/2)
	r := float32(HexRadius)

	var path vector.Path
	for i := 0; i < 6; i++ {
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
	return path
}

// RenderOverlay draws child widgets (unit icon, hp badge) via EbitenUI container.
func (h *HexCellWidget) RenderOverlay(screen *ebiten.Image) {
	h.overlay.Render(screen)
}

// SetColor updates the fill color for the next Render call.
// Accepts any color.Color implementation (color.RGBA, color.NRGBA, etc).
func (h *HexCellWidget) SetColor(c color.Color) {
	r, g, b, a := c.RGBA()
	h.bgColor = color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
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

	// Round to nearest hex center, then fix rounding errors in the
	// coordinate with the largest deviation to keep q+r+s == 0.
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

// Validate satisfies widget.PreferredSizeLocateableWidget.
// EbitenUI calls this during the layout pass to check widget state.
func (h *HexCellWidget) Validate() {
	// No validation needed for a hex cell.
}

// AddChild adds one or more child widgets to be drawn on top of the hex.
// Returns a RemoveChildFunc for consistency with widget.Container.
func (h *HexCellWidget) AddChild(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return h.overlay.AddChild(children...)
}

// ClearChildren removes all child widgets from the hex cell.
func (h *HexCellWidget) ClearChildren() {
	h.children = h.children[:0]
}

// RemoveChildren removes all child widgets from the hex cell.
func (h *HexCellWidget) RemoveChildren() {
	h.overlay.RemoveChildren()
}

// RemoveChild removes a specific child widget from the hex cell.
func (h *HexCellWidget) RemoveChild(w widget.PreferredSizeLocateableWidget) {
	h.overlay.RemoveChild(w)
}

func (h *HexCellWidget) CachedVs() []ebiten.Vertex {
	return h.cachedVs
}

func (h *HexCellWidget) AddToUnitLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	fmt.Println("AddToUnitLayer", h.Coord, len(children))
	return h.AddChildZ(zIndexUnit, children...)
}

func (h *HexCellWidget) AddToHUDLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return h.AddChildZ(zIndexHUD, children...)
}

func (h *HexCellWidget) AddToFXLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return h.AddChildZ(zIndexFX, children...)
}

func (h *HexCellWidget) RenderUnitLayer(screen *ebiten.Image) {
	h.renderLayer(screen, zIndexUnit)
}

func (h *HexCellWidget) RenderHUDLayer(screen *ebiten.Image) {
	h.renderLayer(screen, zIndexHUD)
}

func (h *HexCellWidget) RenderFXLayer(screen *ebiten.Image) {
	h.renderLayer(screen, zIndexFX)
}

// renderLayer is private — called only via named methods above.
func (h *HexCellWidget) renderLayer(screen *ebiten.Image, z int) {
	if layer, ok := h.layers[z]; ok {
		layer.Render(screen)
	}
}

func (h *HexCellWidget) AddChildZ(z int, children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return h.layerFor(z).AddChild(children...)
}

func (h *HexCellWidget) layerFor(z int) *widget.Container {
	fmt.Println("layerFor", h.Coord, z, "cachedRect:", h.cachedRect)
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

func (h *HexCellWidget) CachedRect() image.Rectangle {
	return h.cachedRect
}
