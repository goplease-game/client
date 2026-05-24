package arena

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
)

// DropZoneCell tracks the state of a safe-zone hex cell that can accept unit drops.
// It wraps the underlying HexCellWidget and manages the drop-arrow animation,
// highlight color, and occupied state independently of the board data.
type DropZoneCell struct {
	cell          *ui.HexCellWidget
	activeGraphic *ebiten.Image // current animation frame; nil when not highlighted
	occupied      bool          // true once a unit has been placed on this cell
	coord         ds.HexCoord
	baseColor     color.Color // restore color used after highlight is cleared
}

// SetHighlight toggles the drop-zone highlight on this cell.
// When active, the cell is tinted and the drop-arrow animation starts.
// When inactive, the cell color is restored and the animation is cleared.
func (sc *DropZoneCell) SetHighlight(active bool) {
	if !active {
		if sc.activeGraphic != nil {
			sc.cell.RemoveChildren()
			sc.activeGraphic = nil
		}
		if sc.occupied {
			sc.cell.SetColor(sc.baseColor)
		} else {
			sc.cell.SetColor(boardCellBgColor)
		}
		return
	}

	// Occupied cells do not show a drop highlight.
	if sc.occupied {
		return
	}

	sc.cell.SetColor(unitDropZoneColor)
	if sc.activeGraphic == nil {
		sc.activeGraphic = animDropArrow.CurrentFrame
	}
}

// SetHover tints the cell when the dragged card hovers directly over it.
// Has no effect if the cell is already occupied.
func (sc *DropZoneCell) SetHover(hover bool) {
	if sc.occupied {
		return
	}
	if hover {
		sc.cell.SetColor(unitDropZoneHoverColor)
	} else {
		sc.cell.SetColor(unitDropZoneColor)
	}
}

// RenderAnim draws the drop-arrow animation frame centered on the hex cell.
// Called from Screen.PostRenderHook between the unit layer and HUD layer.
// No-ops if there is no active animation frame.
func (sc *DropZoneCell) RenderAnim(screen *ebiten.Image) {
	if sc.activeGraphic == nil {
		return
	}
	rect := sc.cell.GetWidget().Rect
	cx := float64(rect.Min.X+rect.Dx()/2) - float64(sc.activeGraphic.Bounds().Dx()/2)
	cy := float64(rect.Min.Y+rect.Dy()/2) - float64(sc.activeGraphic.Bounds().Dy()/2)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(cx, cy)
	screen.DrawImage(sc.activeGraphic, op)
}
