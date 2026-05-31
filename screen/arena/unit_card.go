package arena

import (
	"fmt"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ability/effect"
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
func buildHandCard(c *widget.Container, u *ds.Unit) UnitCardRefs {
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
func buildBoardCard(c ChildAdder, u *ds.Unit, canMove bool) UnitCardRefs {
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
	u.Graphic = icon

	c.AddToHUDLayer(hpBadge(u.CurrentHP, 40, -6))

	if u.CurrentShield > 0 {
		c.AddToHUDLayer(shieldBadge(u.CurrentShield, 11, -6))
	}

	if canMove {
		c.AddToHUDLayer(walkBadge())
	}

	return UnitCardRefs{Icon: icon}
}

// buildQueueUnitCard adds a unit portrait and HP badge to a queue card container.
// Queue cards don't show the walk badge — that's board-only.
func buildQueueUnitCard(c ChildAdder, u *ds.Unit) {
	var img *ebiten.Image
	if u.IsStunned() {
		img = asset.Image(unitStunnedPic, unitIconSize)
	} else {
		img = unitImage(u.TemplateID, unitIconSize)
	}

	icon := widget.NewGraphic(
		widget.GraphicOpts.Image(img),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)
	c.AddToUnitLayer(icon)

	iconTop := -6
	iconLeft := -6
	if u.CurrentShield > 0 {
		c.AddToHUDLayer(shieldBadge(u.CurrentShield, iconTop, iconLeft))
		// move HP badge under shield badge
		iconTop = 23
	}

	c.AddToHUDLayer(hpBadge(u.CurrentHP, iconTop, iconLeft))
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
				Padding:            &widget.Insets{Top: 35, Left: 48},
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
func hpBadge(hp, top, left int) *widget.Container {
	const iconSize = 30

	badge := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(iconSize, iconSize),
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            &widget.Insets{Top: top, Left: left},
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

// shieldBadge returns a small container displaying a shield icon with the shield
// value overlaid, anchored next to the HP badge at the top of the hex cell.
func shieldBadge(value, top, left int) *widget.Container {
	const iconSize = 30

	badge := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(iconSize, iconSize),
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            &widget.Insets{Top: top, Left: left},
			}),
		),
	)

	badge.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(asset.Image("shield_o.png", iconSize)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	tf := ui.TextFaceBold(14)
	badge.AddChild(widget.NewText(
		widget.TextOpts.Text(fmt.Sprintf("%d", value), &tf, colornames.White),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	return badge
}

// buildStatusTooltip builds a tooltip container listing all active status effects on the unit.
func buildStatusTooltip(u *ds.Unit) *widget.Container {
	if len(u.Statuses) == 0 {
		return widget.NewContainer() // empty — tooltip won't show visually
	}

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(ttBgColor)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(8)),
		)),
	)

	for _, us := range u.Statuses {
		if us.Status == nil {
			continue
		}

		// Status name colored by alignment.
		nameColor := ttTextColor
		switch us.Status.Alignment {
		case effect.Positive:
			nameColor = colornames.Palegreen
		case effect.Negative:
			nameColor = colornames.Tomato
		}

		tf := ui.TextFaceBold(14)
		c.AddChild(widget.NewText(
			widget.TextOpts.Text(us.Status.Name, &tf, nameColor),
		))

		// Description.
		descTF := ui.TextFace(12)
		c.AddChild(widget.NewText(
			widget.TextOpts.Text(us.Status.Description, &descTF, ttTextColor),
			widget.TextOpts.MaxWidth(250),
		))

		// Duration if not permanent.
		if us.Duration > 0 {
			durTF := ui.TextFace(11)
			c.AddChild(widget.NewText(
				widget.TextOpts.Text(fmt.Sprintf("Duration: %d turns", us.Duration), &durTF, colornames.Skyblue),
			))
		}
	}

	return c
}
