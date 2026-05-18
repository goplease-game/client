package arena

import (
	"fmt"
	stdImg "image"
	"image/color"
	"path"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/asset"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"golang.org/x/image/colornames"
)

// ---------------------------------------------------------------------------
// Panel lifecycle
// ---------------------------------------------------------------------------

func (s *Screen) showAbilityPanel(unit ds.Unit) {
	s.hideAbilityPanel()

	if len(unit.Abilities) == 0 {
		return
	}

	s.abilityPanelRef = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(abilitiesPanelBgColor)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(len(unit.Abilities)),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(5)),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	for _, abilityID := range unit.Abilities {
		ab := ability.ByID(string(abilityID))
		s.abilityPanelRef.AddChild(s.buildAbilityCard(ab))
	}

	s.footerRef.AddChild(s.abilityPanelRef)
	s.abilityPanelIn = true
}

func (s *Screen) hideAbilityPanel() {
	if !s.abilityPanelIn || s.abilityPanelRef == nil {
		return
	}
	s.footerRef.RemoveChild(s.abilityPanelRef)
	s.abilityPanelRef = nil
	s.abilityPanelIn = false
}

// ---------------------------------------------------------------------------
// Card
// ---------------------------------------------------------------------------

func (s *Screen) buildAbilityCard(ab ability.Ability) *widget.Container {
	bgColor := abilityCardBgColor(ab)

	var card *widget.Container
	card = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(bgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(cellSize, cellSize),
			widget.WidgetOpts.ToolTip(
				widget.NewToolTip(
					widget.ToolTipOpts.Content(s.buildAbilityToolTip(ab)),
					widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
					widget.ToolTipOpts.Offset(stdImg.Point{X: 0, Y: -8}),
					widget.ToolTipOpts.AnchorOriginHorizontal(widget.TOOLTIP_ANCHOR_MIDDLE),
					widget.ToolTipOpts.AnchorOriginVertical(widget.TOOLTIP_ANCHOR_START),
					widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_MIDDLE),
					widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_END),
				),
			),
			widget.WidgetOpts.CursorEnterHandler(func(_ *widget.WidgetCursorEnterEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(ui.DarkenRGB(bgColor, 30)))
			}),
			widget.WidgetOpts.CursorExitHandler(func(_ *widget.WidgetCursorExitEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(bgColor))
			}),
		),
	)

	card.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(abilityImage(ab.ID)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	if ab.Cooldown > 0 {
		card.AddChild(buildCooldownBadge(ab.Cooldown))
	}

	return card
}

// abilityCardBgColor returns the background colour for an ability card
// based on its type (basic attack, passive, or regular).
func abilityCardBgColor(ab ability.Ability) color.Color {
	switch {
	case ab.IsBasicAttack():
		return basicAttackBgColor
	case ab.IsPassive:
		return passiveAbilityBgColor
	default:
		return abilityBgColor
	}
}

// buildCooldownBadge returns a small top-left badge showing the cooldown turns.
func buildCooldownBadge(cooldown int) *widget.Container {
	badge := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(2),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            &widget.Insets{Top: 3, Left: 3},
			}),
		),
	)
	badge.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(asset.Image("turn.png", 12)),
	))
	tf := ui.TextFace(11)
	badge.AddChild(widget.NewText(
		widget.TextOpts.Text(fmt.Sprintf("%d", cooldown), &tf, colornames.White),
	))
	return badge
}

// ---------------------------------------------------------------------------
// Assets
// ---------------------------------------------------------------------------

func abilityImage(abilityID string, sizeOpt ...int) *ebiten.Image {
	size := 64
	if len(sizeOpt) > 0 {
		size = sizeOpt[0]
	}
	return asset.Image(path.Join("abilities", abilityID+".png"), size)
}

// ---------------------------------------------------------------------------
// Tooltip
// ---------------------------------------------------------------------------

func (s *Screen) buildAbilityToolTip(ab ability.Ability) *widget.Container {
	c := buildToolTipBase(abilityImage(ab.ID, 28), ab.Name)

	c.AddChild(widget.NewText(
		widget.TextOpts.Text(ab.Description, &toolTipTextTF, ttTextColor),
		widget.TextOpts.MaxWidth(350),
	))

	if ab.Cooldown > 0 {
		c.AddChild(buildToolTipRow(fmt.Sprintf("Cooldown: %d", ab.Cooldown), colornames.Skyblue))
	}
	if ab.Range > 0 {
		c.AddChild(buildToolTipRow(fmt.Sprintf("Range: %d", ab.Range), colornames.Palegreen))
	}
	if ab.IsPassive {
		c.AddChild(buildToolTipRow("Passive", colornames.Plum))
	}

	return c
}

func buildToolTipRow(text string, c color.Color) *widget.Container {
	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(4),
		)),
	)
	row.AddChild(widget.NewText(
		widget.TextOpts.Text(text, &toolTipTextTF, c),
	))
	return row
}
