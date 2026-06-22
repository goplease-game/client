// Package arena ...
package arena

import (
	"fmt"
	stdImg "image"
	"image/color"
	"path"
	"strconv"
	"strings"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/sfx"
	"github.com/goplease-game/client/ui"
	"github.com/goplease-game/server/ability"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

// showAbilityPanel builds and attaches an ability card row to the footer
// for the given unit. Any previously shown panel is removed first.
// Does nothing if the unit has no abilities.
func (s *Screen) showAbilityPanel(unit *ds.Unit) {
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
		ab := ability.ByID(abilityID)
		s.abilityPanelRef.AddChild(s.buildAbilityCard(ab))
	}

	s.footerRef.AddChild(s.abilityPanelRef)
	s.abilityPanelIn = true
}

// hideAbilityPanel removes the ability panel from the footer and resets panel state.
// Safe to call when no panel is currently shown.
func (s *Screen) hideAbilityPanel() {
	if !s.abilityPanelIn || s.abilityPanelRef == nil {
		return
	}
	s.footerRef.RemoveChild(s.abilityPanelRef)
	s.abilityPanelRef = nil
	s.abilityPanelIn = false
}

// buildAbilityCard builds a single ability card widget with hover highlight,
// tooltip, ability icon, and an optional cooldown badge.
func (s *Screen) buildAbilityCard(ab ability.Ability) *widget.Container {
	bgColor := abilityCardBgColor(ab)
	u := s.unitByID(s.activeUnitID)
	onCooldown := u != nil && !u.AbilityReady(ab.ID)
	disabled := !onCooldown && u != nil && !ab.IsPassive && !s.unitCanAct(u)
	blocked := onCooldown || disabled

	iconGraphic := s.buildAbilityIcon(ab)
	card := s.buildAbilityCardContainer(ab, bgColor, blocked, iconGraphic)
	card.AddChild(iconGraphic)

	switch {
	case onCooldown:
		card.AddChild(s.buildCooldownOverlay(u.Cooldowns[ab.ID]))
	case disabled:
		card.AddChild(s.buildDisabledOverlay())
	}

	return card
}

// unitCanAct reports whether the unit has AP available to use an ability.
// A unit can act if it has base AP, or if the team has Phantom AP remaining
// and the unit hasn't already spent its phantom AP allowance this turn.
func (s *Screen) unitCanAct(u *ds.Unit) bool {
	if u.CurrentAP > 0 {
		return true
	}
	if s.player.PhantomAP < 1 {
		return false
	}

	return u.PhantomAPUsedThisTurn < s.maxPhantomAPPerUnitPerTurn
}

func (s *Screen) buildAbilityCardContainer(ab ability.Ability, bgColor color.Color, blocked bool, iconGraphic *widget.Graphic) *widget.Container {
	var card *widget.Container
	card = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(bgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(abilityCardSize, abilityCardSize),
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
				if blocked || s.selectedAbility != nil {
					return
				}
				s.clearAbilityHighlight()
				sfx.Play(unitHoverSound)
				card.SetBackgroundImage(image.NewNineSliceColor(ui.DarkenRGB(bgColor, 30)))
				s.highlightAbilityRange(ab)
			}),
			widget.WidgetOpts.CursorExitHandler(func(_ *widget.WidgetCursorExitEventArgs) {
				if blocked || s.selectedAbility != nil {
					return
				}
				if s.selectedAbility == nil || s.selectedAbility.ID != ab.ID {
					card.SetBackgroundImage(image.NewNineSliceColor(bgColor))
					s.clearAbilityHighlight()
				}
			}),
			widget.WidgetOpts.MouseButtonReleasedHandler(func(args *widget.WidgetMouseButtonReleasedEventArgs) {
				if args.Button == ebiten.MouseButtonLeft && args.Inside {
					if blocked {
						return
					}
					s.onAbilityCardClicked(ab, card, bgColor)
					if s.selectedAbility != nil && s.selectedAbility.ID == ab.ID {
						iconGraphic.Image = asset.TintedImage(abilityImagePath(string(ab.ID)), activeAbilityFgColor, abilityCardSize)
						s.selectedAbilityIcon = iconGraphic
					}
				}
			}),
		),
	)

	return card
}

func (s *Screen) buildAbilityIcon(ab ability.Ability) *widget.Graphic {
	return widget.NewGraphic(
		widget.GraphicOpts.Image(abilityImage(string(ab.ID))),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)
}

func (s *Screen) buildCooldownOverlay(remaining int) *widget.Container {
	overlay := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(
			color.NRGBA{0x00, 0x00, 0x00, 0x99},
		)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
	)
	overlay.AddChild(widget.NewText(
		widget.TextOpts.Text(strconv.Itoa(remaining), &abilityCooldownCounterTF, abilityCooldownCounterColor),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	return overlay
}

func (s *Screen) buildDisabledOverlay() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(
			color.NRGBA{0x00, 0x00, 0x00, 0xbb},
		)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
	)
}

// abilityCardBgColor returns the background colour for an ability card
// based on its type: basic attack, passive, or regular ability.
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

// abilityImage loads the icon for the given ability ID at the specified size.
// Size defaults to 64px if not provided.
func abilityImage(abilityID string, sizeOpt ...int) *ebiten.Image {
	size := 64
	if len(sizeOpt) > 0 {
		size = sizeOpt[0]
	}
	return asset.Image(abilityImagePath(abilityID), size)
}

func abilityImagePath(abilityID string) string {
	return path.Join("abilities", abilityID+".png")
}

// buildAbilityToolTip constructs the tooltip content for an ability card,
// including icon, name, description, and optional stat rows (cooldown, range, passive).
func (s *Screen) buildAbilityToolTip(ab ability.Ability) *widget.Container {
	c := buildToolTipBase(abilityImage(string(ab.ID), 28), ab.Name)

	c.AddChild(widget.NewText(
		widget.TextOpts.Text(ab.Description, &toolTipTextTF, ttTextColor),
		widget.TextOpts.MaxWidth(350),
	))

	if ab.DamageHint != "" {
		if u := s.unitByID(s.activeUnitID); u != nil {
			val := strings.ReplaceAll(ab.DamageHint, ability.HintCurrentATK, strconv.Itoa(u.CurrentAtk))
			c.AddChild(buildToolTipRow("Damage: "+val, colornames.Orange))
		}
	}
	if ab.Cooldown > 0 {
		c.AddChild(buildToolTipRow(fmt.Sprintf("Cooldown: %d", ab.Cooldown), colornames.Skyblue))
	}
	if ab.Range > 0 {
		label := "Melee"
		if ab.Range > 1 {
			label = fmt.Sprintf("Range: %d", ab.Range)
		}
		c.AddChild(buildToolTipRow(label, colornames.Palegreen))
	}
	if ab.IsPassive {
		c.AddChild(buildToolTipRow("Passive", colornames.Plum))
	}

	return c
}

// buildToolTipRow returns a horizontal row container with a single coloured text label.
// Used to display ability stats (cooldown, range, passive) in tooltips.
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
