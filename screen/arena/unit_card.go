package arena

import (
	"fmt"
	"math"
	"strconv"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/ui"
	"github.com/goplease-game/server/ability/status"
	"github.com/hajimehoshi/ebiten/v2"
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

// buildHandCard adds a draggable unit portrait.
// Used for cards in the player's hand panel.
// Returns refs so the caller can swap the icon image on cursor enter/exit.
func buildHandCard(c *widget.Container, u *ds.Unit) UnitCardRefs {
	normalImg := asset.Image(unitImagePath(u.TemplateID), unitCardSize)
	hoverImg := asset.TintedImage(unitImagePath(u.TemplateID), unitCardHoverFgColor, unitCardSize)

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
func (s *Screen) buildBoardCard(c ChildAdder, u *ds.Unit) {
	var img *ebiten.Image
	if u.HasStatus(status.Stunned) {
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
	u.Graphic = icon

	c.AddToHUDLayer(hpBadge(u.CurrentHP, 36, -6))

	if u.CurrentShield > 0 {
		c.AddToHUDLayer(shieldBadge(u.CurrentShield, 16, -6))
	}
}

// buildQueueUnitCard adds a unit portrait and HP badge to a queue card container.
// Queue cards don't show the walk badge — that's board-only.
func (s *Screen) buildQueueUnitCard(c ChildAdder, u *ds.Unit) {
	img := unitImage(u.TemplateID, unitIconSize)
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

	hpTop := -6
	shieldTop := -6
	hpLeft := -6
	if u.CurrentShield > 0 {
		hpTop = 15
	}

	c.AddToHUDLayer(hpBadge(u.CurrentHP, hpTop, hpLeft))

	if u.CurrentShield > 0 {
		c.AddToHUDLayer(shieldBadge(u.CurrentShield, shieldTop, hpLeft))
	}

	statusIcons(c, u)
}

func (s *Screen) drawAPMarkers(screen *ebiten.Image, u *ds.Unit, cell *ui.HexCellWidget) {
	const iconSize = 10

	total := u.CurrentAP
	phantomAvail := 0
	if s.player.PhantomAP > 0 {
		phantomAvail = s.maxPhantomAPPerUnitPerTurn - u.PhantomAPUsedThisTurn
		phantomAvail = min(phantomAvail, s.player.PhantomAP)
	}
	total += phantomAvail
	if total == 0 {
		return
	}

	rect := cell.CachedRect()
	cx := float64(rect.Min.X + rect.Dx()/2)
	cy := float64(rect.Min.Y + rect.Dy()/2)
	r := float64(ui.HexRadius) - float64(iconSize)/2 - 1

	spreadPerMarker := 15.0 * math.Pi / 180.0 // 15 degrees between markers
	var angles []float64
	if total == 1 {
		angles = []float64{270 * math.Pi / 180.0}
	} else {
		totalSpread := float64(total-1) * spreadPerMarker
		startAngle := 270*math.Pi/180.0 - totalSpread/2
		for i := range total {
			angles = append(angles, startAngle+float64(i)*spreadPerMarker)
		}
	}

	apLeft := u.CurrentAP
	for i, angle := range angles {
		var img *ebiten.Image
		if i < apLeft {
			img = asset.Image("ap_marker.png", iconSize)
		} else {
			img = asset.Image("phantom_ap_marker.png", iconSize)
		}

		x := cx + r*math.Cos(angle) - float64(iconSize)/2
		y := cy + r*math.Sin(angle) - float64(iconSize)/2

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y+5)
		screen.DrawImage(img, op)
	}
}

// hpBadge returns a small container that displays a heart icon with the HP
// value overlaid, anchored slightly outside the top-left corner of the hex cell.
func hpBadge(hp, top, left int) *widget.Container {
	return badgeContainer(hp, top, left, "heart_o.png")
}

// shieldBadge returns a small container displaying a shield icon with the shield
// value overlaid, anchored next to the HP badge at the top of the hex cell.
func shieldBadge(value, top, left int) *widget.Container {
	return badgeContainer(value, top, left, "shield_o.png")
}

// badgeContainer returns a small container displaying a shield or heart.
func badgeContainer(value, top, left int, img string) *widget.Container {
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
		widget.GraphicOpts.Image(asset.Image(img, iconSize)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	tf := ui.TextFaceBold(14)
	badge.AddChild(widget.NewText(
		widget.TextOpts.Text(strconv.Itoa(value), &tf, colornames.White),
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
		return widget.NewContainer()
	}

	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(ttBgColor)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(8)),
		)),
	)

	for _, st := range status.Order {
		us, ok := u.Statuses[st]
		if !ok || us.Status == nil {
			continue
		}

		// Horizontal row: icon + text
		row := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
				widget.RowLayoutOpts.Spacing(6),
			)),
		)

		// Icon
		const iconSize = 32
		img := asset.Image(fmt.Sprintf("statuses/%s.png", st), iconSize)
		row.AddChild(widget.NewGraphic(
			widget.GraphicOpts.Image(img),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionStart,
				}),
			),
		))

		// Text
		nameColor := neutralStatusNameColor
		if us.IsPositive() {
			nameColor = positiveStatusNameColor
		} else if us.IsNegative() {
			nameColor = negativeStatusNameColor
		}

		var durText string
		if us.Duration > 0 {
			durText = fmt.Sprintf("\n[color=%s]Duration: %d turns[/color]", statusDurationColor, us.Duration)
		}

		fullText := fmt.Sprintf("[color=%s]%s[/color]: %s%s", nameColor, us.Status.Name, us.Status.Description, durText)
		descTF := ui.TextFace(16)
		row.AddChild(widget.NewText(
			widget.TextOpts.Text(fullText, &descTF, ttTextColor),
			widget.TextOpts.MaxWidth(300),
			widget.TextOpts.ProcessBBCode(true),
		))

		c.AddChild(row)
	}

	return c
}

// statusIcons adds status icons to the HUD layer, laid out horizontally
// starting to the right of the HP badge position.
func statusIcons(c ChildAdder, u *ds.Unit) {
	if len(u.Statuses) == 0 {
		return
	}

	const (
		iconSize      = 20
		startTop      = 42
		startLeft     = 42
		spacing       = iconSize + 1
		columnSize    = 3
		iconsMaxCount = 6
	)

	i := 0
	for _, st := range status.Order {
		if i == iconsMaxCount {
			break
		}
		sv, ok := u.Statuses[st]
		if !ok {
			continue
		}
		path := fmt.Sprintf("statuses/%s.png", st)
		iconColor := positiveStatusIconColor
		if sv.IsNegative() {
			iconColor = negativeStatusIconColor
		}

		img := asset.NewImage(path, iconSize).
			Tint(iconColor).
			Shadow(1, 1, 0.3).
			Render()

		col := i / columnSize
		row := i % columnSize

		top := startTop - row*spacing
		left := startLeft - col*spacing

		icon := widget.NewContainer(
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
		icon.AddChild(widget.NewGraphic(
			widget.GraphicOpts.Image(img),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		))

		c.AddToHUDLayer(icon)
		i++
	}
}
