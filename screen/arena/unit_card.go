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

// UnitCardRefs holds internal widget references for hover effects.
type UnitCardRefs struct {
	Icon      *widget.Graphic
	HoverIcon *ebiten.Image
	NormIcon  *ebiten.Image
}

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

		c.AddChild(badge)
	}

	return UnitCardRefs{Icon: icon}
}

// hpBadge returns a fixed-size container anchored to the top-left corner.
// It layers a heart icon and an HP number on top of each other via AnchorLayout.
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

	// Heart icon fills the badge area.
	badge.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(asset.Image("heart_o.png", iconSize)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	// HP number drawn on top of the heart, centred.
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
