package ui

import (
	"image"
	"math"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/ds"
)

// HexRadius is the center-to-vertex radius of a hex cell in pixels.
// For a pointy-top hex: width = √3 * HexRadius ≈ 64px, height = 2 * HexRadius = 74px.
const HexRadius = 37

// HexChild is implemented by any widget that can be positioned by HexLayout.
// It provides the axial coordinate used to compute the widget's screen position.
type HexChild interface {
	GetHexCoord() ds.HexCoord
	GetWidget() *widget.Widget
}

// HexLayout is an EbitenUI Layouter that positions hex cell widgets
// using axial (Q, R) coordinates. It supports pointy-top hex orientation.
//
// HexSize defaults to HexRadius if not set.
type HexLayout struct {
	HexSize float64
}

// Layout positions each widget that implements HexChild using axial-to-pixel
// conversion. Widgets that do not implement HexChild are skipped.
// SetLocation is called both on the EbitenUI widget (for input handling)
// and on the widget itself if it implements the interface (for geometry rebuild).
func (l *HexLayout) Layout(widgets []widget.PreferredSizeLocateableWidget, rect image.Rectangle) {
	if l.HexSize == 0 {
		l.HexSize = HexRadius
	}

	hexW := int(math.Round(math.Sqrt(3) * l.HexSize))
	hexH := int(math.Round(2 * l.HexSize))

	// Find the minimum x and y across all cells to normalize negative coordinates.
	minX, minY := math.MaxFloat64, math.MaxFloat64
	for _, w := range widgets {
		hc, ok := w.(HexChild)
		if !ok {
			continue
		}
		x, y := axialToPixel(hc.GetHexCoord(), l.HexSize)
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
	}

	for _, w := range widgets {
		hc, ok := w.(HexChild)
		if !ok {
			continue
		}

		coord := hc.GetHexCoord()
		x, y := axialToPixel(coord, l.HexSize)

		// Normalize so the top-left cell starts at (0,0), then offset by container.
		x -= minX
		y -= minY
		x += float64(rect.Min.X)
		y += float64(rect.Min.Y)

		rx := int(math.Round(x))
		ry := int(math.Round(y))

		r := image.Rect(rx, ry, rx+hexW, ry+hexH)

		w.GetWidget().SetLocation(r)

		if h, ok := w.(interface{ SetLocation(img image.Rectangle) }); ok {
			h.SetLocation(r)
		}
	}
}

// PreferredSize returns the minimum bounding box that contains all hex cells.
// Used by EbitenUI to size the board container.
func (l *HexLayout) PreferredSize(widgets []widget.PreferredSizeLocateableWidget) (int, int) {
	if l.HexSize == 0 {
		l.HexSize = HexRadius
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, w := range widgets {
		hc, ok := w.(HexChild)
		if !ok {
			continue
		}

		x, y := axialToPixel(hc.GetHexCoord(), l.HexSize)
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}

	hexW := int(math.Round(math.Sqrt(3) * l.HexSize))
	hexH := int(math.Round(2 * l.HexSize))

	width := int(math.Round(maxX-minX)) + hexW
	height := int(math.Round(maxY-minY)) + hexH

	return width, height
}

// axialToPixel converts axial hex coordinates (Q, R) to pixel coordinates
// using integer-aligned hex dimensions to prevent subpixel gaps between cells.
// Returns the top-left corner of the hex bounding box.
func axialToPixel(h ds.HexCoord, size float64) (float64, float64) {
	hexW := int(math.Round(math.Sqrt(3) * size))
	hexH := int(math.Round(2 * size))

	x := float64(h.Q)*float64(hexW) + float64(h.R)*float64(hexW)/2
	y := float64(h.R) * float64(hexH) * 3 / 4

	return x, y
}
