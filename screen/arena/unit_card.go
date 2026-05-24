package arena

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/asset"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"golang.org/x/image/colornames"
)

// UnitCardRefs holds widget references returned by card builders.
// Used by callers that need to update the card after creation
// (e.g. swapping the icon on hover).
type UnitCardRefs struct {
	Icon      *widget.Graphic
	HoverIcon *ebiten.Image // pre-tinted hover variant of the unit portrait
	NormIcon  *ebiten.Image // original unit portrait
}

// buildHandCard adds a draggable unit portrait to c.
// Used for cards in the player's hand panel.
// Returns refs so the caller can swap the icon image on cursor enter/exit.
func buildHandCard(c *widget.Container, u ds.Unit) UnitCardRefs {
	normalImg := unitImage(u.TemplateID, unitCardSize)
	hoverImg := ui.TintImage(normalImg, unitCardHoverFgColor)

	icon := widget.NewGraphic(
		widget.GraphicOpts.Image(normalImg),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)
	c.AddChild(icon)

	return UnitCardRefs{
		Icon:      icon,
		HoverIcon: hoverImg,
		NormIcon:  normalImg,
	}
}

// buildBoardCard adds a unit portrait and HUD badges to a ChildAdder (hex cell or container).
// The portrait goes to the unit layer; the HP badge goes to the HUD layer.
// If canMove is true, a walk indicator badge is also added.
func buildBoardCard(c ChildAdder, u ds.Unit, canMove bool) UnitCardRefs {
	icon := widget.NewGraphic(
		widget.GraphicOpts.Image(unitImage(u.TemplateID, unitIconSize)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	c.AddToUnitLayer(icon)
	c.AddToHUDLayer(hpBadge(u.CurrentHP))

	if canMove {
		c.AddToHUDLayer(walkBadge())
	}

	return UnitCardRefs{Icon: icon}
}

// walkBadge returns a small container with a walk icon, anchored to the
// bottom-left corner of the hex cell to indicate the unit can still move.
func walkBadge() *widget.Container {
	const iconSize = 30

	badge := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(iconSize, iconSize),
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            &widget.Insets{Top: 35, Left: -5},
			}),
		),
	)

	badge.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(asset.Image("walk_o.png", iconSize)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	return badge
}

// hpBadge returns a small container that displays a heart icon with the HP
// value overlaid, anchored slightly outside the top-left corner of the hex cell.
func hpBadge(hp int) *widget.Container {
	const iconSize = 30

	badge := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(iconSize, iconSize),
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            &widget.Insets{Top: -6, Left: -6},
			}),
		),
	)

	badge.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(asset.Image("heart_o.png", iconSize)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	tf := ui.TextFaceBold(14)
	badge.AddChild(widget.NewText(
		widget.TextOpts.Text(fmt.Sprintf("%d", hp), &tf, colornames.White),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	return badge
}
