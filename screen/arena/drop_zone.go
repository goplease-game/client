package arena

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
)

type DropZoneCell struct {
	cell          *ui.HexCellWidget
	activeGraphic *ebiten.Image
	occupied      bool
	coord         ds.HexCoord
	baseColor     color.Color
}

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

	if sc.occupied {
		return
	}

	sc.cell.SetColor(unitDropZoneColor)
	if sc.activeGraphic == nil {
		sc.activeGraphic = animDropArrow.CurrentFrame
	}
}

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
