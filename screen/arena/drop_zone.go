package arena

import (
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
)

type DropZoneCell struct {
	container     *widget.Container
	activeGraphic *widget.Graphic
	occupied      bool
	row, col      int
	baseColor     color.Color
}

func (sc *DropZoneCell) SetHighlight(active bool) {
	if !active {
		if sc.activeGraphic != nil {
			sc.container.RemoveChild(sc.activeGraphic)
			sc.activeGraphic = nil
		}
		if sc.occupied {
			sc.container.SetBackgroundImage(image.NewNineSliceColor(sc.baseColor))
		} else {
			sc.container.SetBackgroundImage(image.NewNineSliceColor(boardCellBgColor))
		}
		return
	}

	if sc.occupied {
		return
	}

	sc.container.SetBackgroundImage(image.NewNineSliceColor(unitDropZoneColor))
	if sc.activeGraphic == nil {
		sc.activeGraphic = widget.NewGraphic(
			widget.GraphicOpts.Image(animDropArrow.CurrentFrame),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		sc.container.AddChild(sc.activeGraphic)
	}
}

func (sc *DropZoneCell) SetHover(hover bool) {
	if sc.occupied {
		return
	}
	if hover {
		sc.container.SetBackgroundImage(image.NewNineSliceColor(unitDropZoneHoverColor))
	} else {
		sc.container.SetBackgroundImage(image.NewNineSliceColor(unitDropZoneColor))
	}
}
