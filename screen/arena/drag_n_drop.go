package arena

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

// dndUnit builds and caches the drag widget for a unit card.
// It implements the EbitenUI DragAndDrop creator interface.
type dndUnit struct {
	unit       *ds.Unit
	dragWidget *widget.Container // lazily created and reused across drags
}

// Create returns the drag widget and the unit as the drop payload.
// The widget is created once and reused on subsequent drags.
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

// dndHandler manages drag-and-drop interactions for placing units onto safe-zone cells.
// It highlights valid drop targets when a drag starts and tracks the hovered cell.
type dndHandler struct {
	*dndUnit
	safeCells   []*DropZoneCell // all safe-zone cells that can receive a drop
	currentCell *DropZoneCell   // cell currently under the dragged card; nil if none
	canDrag     func() bool     // returns false when dragging is not allowed (e.g. not the player's turn)
}

// Create is called by EbitenUI when a drag begins.
// Returns nil to cancel the drag if canDrag reports false.
// Otherwise highlights all safe-zone cells and delegates to dndUnit.Create.
func (d *dndHandler) Create(parent widget.HasWidget) (*widget.Container, interface{}) {
	if !d.canDrag() {
		return nil, nil
	}

	for _, sc := range d.safeCells {
		sc.SetHighlight(true)
	}
	return d.dndUnit.Create(parent)
}

// Update is called each frame while a drag is in progress.
// It tracks which safe-zone cell is currently hovered and updates its tint.
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

// EndDrag is called by EbitenUI when the drag ends (drop or cancel).
// Clears the highlight on all safe-zone cells regardless of outcome.
func (d *dndHandler) EndDrag(_ bool, _ widget.HasWidget, _ interface{}) {
	for _, sc := range d.safeCells {
		sc.SetHighlight(false)
	}
}
