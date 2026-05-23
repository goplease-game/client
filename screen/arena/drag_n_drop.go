package arena

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

type dndUnit struct {
	unit       ds.Unit
	dragWidget *widget.Container
}

func (d *dndUnit) Create(_ widget.HasWidget) (*widget.Container, interface{}) {
	if d.dragWidget == nil {
		unitImg := unitImage(d.unit.TemplateID)
		d.dragWidget = widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(unitDragBgColor)),
		)
		d.dragWidget.AddChild(widget.NewGraphic(
			widget.GraphicOpts.Image(unitImg),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		))
	}

	return d.dragWidget, d.unit
}

type dndHandler struct {
	*dndUnit
	safeCells   []*DropZoneCell
	currentCell *DropZoneCell
	canDrag     func() bool
}

func (d *dndHandler) Create(parent widget.HasWidget) (*widget.Container, interface{}) {
	if !d.canDrag() {
		return nil, nil
	}

	for _, sc := range d.safeCells {
		sc.SetHighlight(true)
	}
	return d.dndUnit.Create(parent)
}

func (d *dndHandler) Update(canDrop bool, target widget.HasWidget, _ interface{}) {
	if d.currentCell != nil {
		d.currentCell.SetHover(false)
		d.currentCell = nil
	}
	if canDrop && target != nil {
		for _, sc := range d.safeCells {
			if sc.cell == target {
				sc.SetHover(true)
				d.currentCell = sc
				break
			}
		}
	}
}

func (d *dndHandler) EndDrag(_ bool, _ widget.HasWidget, _ interface{}) {
	for _, sc := range d.safeCells {
		sc.SetHighlight(false)
	}
}
