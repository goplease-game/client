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
	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/config"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/sfx"
	"github.com/goplease-game/client/ui"
	"github.com/goplease-game/server/ability"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

type abilitySlot struct {
	ability     ability.Ability
	card        *widget.Container
	bgColor     color.Color
	iconGraphic *widget.Graphic
}

// showAbilityPanel builds and attaches the action row (Move button + ability
// cards) to the footer for the given unit. Any previously shown panel is
// removed first.
func (s *Screen) showAbilityPanel(unit *ds.Unit) {
	s.hideAbilityPanel()

	s.abilityPanelRef = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(8),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	s.abilityPanelRef.AddChild(s.buildMoveButton(unit))

	abilitiesContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(abilitiesPanelBgColor)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(len(unit.Abilities)),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(5)),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
	)

	s.abilitySlots = s.abilitySlots[:0] // reset before rebuild

	for i, abilityID := range unit.Abilities {
		ab := ability.ByID(abilityID)
		card, bgColor, iconGraphic := s.buildAbilityCard(ab, i)
		abilitiesContainer.AddChild(card)

		s.abilitySlots = append(s.abilitySlots, abilitySlot{
			ability:     ab,
			card:        card,
			bgColor:     bgColor,
			iconGraphic: iconGraphic,
		})
	}

	s.abilityPanelRef.AddChild(abilitiesContainer)

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
func (s *Screen) buildAbilityCard(ab ability.Ability, slot int) (*widget.Container, color.Color, *widget.Graphic) {
	bgColor := abilityCardBgColor(ab)
	u := s.unitByID(s.activeUnitID)
	onCooldown := u != nil && !u.AbilityReady(ab.ID)
	disabled := !onCooldown && u != nil && !ab.IsPassive && !s.unitCanAct(u)
	blocked := onCooldown || disabled

	iconGraphic := s.buildAbilityIcon(ab)
	card := s.buildAbilityCardContainer(ab, bgColor, blocked, iconGraphic)
	card.AddChild(iconGraphic)
	if key := abilityKeyForSlot(slot); key != nil {
		card.AddChild(s.buildKeyHint(key))
	}

	switch {
	case onCooldown:
		card.AddChild(s.buildCooldownOverlay(u.Cooldowns[ab.ID]))
	case disabled:
		card.AddChild(s.buildDisabledOverlay())
	}

	return card, bgColor, iconGraphic
}

// buildMoveButton builds the Move toggle card shown to the left of the
// ability cards. It mirrors clicking the unit on the board: selects the unit
// for movement, or deselects it if already selected. Disabled (greyed out)
// if the unit has no movement points left.
func (s *Screen) buildMoveButton(u *ds.Unit) *widget.Container {
	disabled := !u.CanMove()
	selected := s.selectedUnitID == u.ID

	bgColor := moveButtonBgColor
	if selected {
		bgColor = moveButtonActiveBgColor
	}

	iconGraphic := widget.NewGraphic(
		widget.GraphicOpts.Image(asset.Image("move.png", moveButtonSize)),
	)

	iconWrapper := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(6)),
		)),
	)
	iconGraphic.GetWidget().LayoutData = widget.AnchorLayoutData{
		HorizontalPosition: widget.AnchorLayoutPositionCenter,
		VerticalPosition:   widget.AnchorLayoutPositionCenter,
	}
	iconWrapper.AddChild(iconGraphic)

	var card *widget.Container
	card = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewBorderedNineSliceColor(bgColor, ui.DarkenRGB(bgColor, 20), 3)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(moveButtonSize, moveButtonSize),
			widget.WidgetOpts.CursorEnterHandler(func(_ *widget.WidgetCursorEnterEventArgs) {
				if disabled {
					return
				}
				sfx.Play(unitHoverSound)
				card.SetBackgroundImage(image.NewBorderedNineSliceColor(ui.DarkenRGB(bgColor, 20), ui.DarkenRGB(bgColor, 40), 3))
			}),
			widget.WidgetOpts.CursorExitHandler(func(_ *widget.WidgetCursorExitEventArgs) {
				if disabled {
					return
				}
				card.SetBackgroundImage(image.NewBorderedNineSliceColor(bgColor, ui.DarkenRGB(bgColor, 20), 3))
			}),
			widget.WidgetOpts.MouseButtonReleasedHandler(func(args *widget.WidgetMouseButtonReleasedEventArgs) {
				if args.Button == ebiten.MouseButtonLeft && args.Inside {
					if disabled {
						return
					}
					s.onMoveButtonClicked(u)
				}
			}),
		),
	)
	card.AddChild(iconWrapper)

	// MP remaining, top-left corner.
	mpFace := ui.TextFaceBold(14)
	card.AddChild(widget.NewText(
		widget.TextOpts.Text(strconv.Itoa(u.CurrentMP), &mpFace, colornames.White),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            &widget.Insets{Left: 4, Top: 2},
			}),
		),
	))

	// Keybind hint, bottom-right corner.
	if moveKey := config.Get().Keybindings.Move; moveKey != nil {
		hintFace := ui.TextFaceBold(14)
		card.AddChild(widget.NewText(
			widget.TextOpts.Text(game.KeyName(moveKey), &hintFace, colornames.White),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionEnd,
					VerticalPosition:   widget.AnchorLayoutPositionEnd,
					Padding:            &widget.Insets{Right: 4, Bottom: 2},
				}),
			),
		))
	}

	if disabled {
		card.AddChild(s.buildDisabledOverlay(75))
	}

	return card
}

// onMoveButtonClicked toggles unit selection for movement, same as clicking
// the unit on the board would, then refreshes the panel to reflect the new
// selection state (highlight on/off).
func (s *Screen) onMoveButtonClicked(u *ds.Unit) {
	if !u.CanMove() {
		sfx.Play(selectError)
		return
	}

	s.selectUnit(u)
	s.showAbilityPanel(u)
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
				if blocked || s.selectedAbility != nil || s.selectedUnitID != "" {
					return
				}
				s.clearAbilityHighlight()
				sfx.Play(unitHoverSound)
				card.SetBackgroundImage(image.NewNineSliceColor(ui.DarkenRGB(bgColor, 30)))
				s.highlightAbilityRange(ab)
			}),
			widget.WidgetOpts.CursorExitHandler(func(_ *widget.WidgetCursorExitEventArgs) {
				if blocked || s.selectedAbility != nil || s.selectedUnitID != "" {
					return
				}
				if s.selectedAbility == nil || s.selectedAbility.ID != ab.ID {
					card.SetBackgroundImage(image.NewNineSliceColor(bgColor))
					s.clearAbilityHighlight()
				}
			}),
			widget.WidgetOpts.MouseButtonReleasedHandler(func(args *widget.WidgetMouseButtonReleasedEventArgs) {
				if args.Button == ebiten.MouseButtonLeft && args.Inside {
					s.activateAbilitySlot(ab, card, bgColor, iconGraphic)
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

// buildDisabledOverlay returns a black overlay container that stretches to
// fill its parent, used to visually grey out a disabled card or button.
// opacityOpt is an optional transparency percent (0 = fully opaque,
// 100 = fully transparent); defaults to 25 if not provided.
func (s *Screen) buildDisabledOverlay(opacityOpt ...int) *widget.Container {
	opacity := 25
	if len(opacityOpt) > 0 {
		opacity = opacityOpt[0]
	}

	alpha := 255 * (100 - opacity) / 100

	return widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(
			color.NRGBA{0x00, 0x00, 0x00, uint8(alpha)},
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

// buildKeyHint returns a small pill-shaped label showing the given key,
// anchored to the bottom-right corner of its parent. Used as a hotkey
// hint on ability/move cards. The semi-transparent dark background keeps
// the text readable regardless of the card's own background color.
func (s *Screen) buildKeyHint(key *ebiten.Key) *widget.Container {
	hintFace := ui.TextFaceBold(14)

	label := widget.NewText(
		widget.TextOpts.Text(game.KeyName(key), &hintFace, colornames.White),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	badge := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(
			color.NRGBA{0x00, 0x00, 0x00, 100},
		)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(&widget.Insets{Left: 4, Right: 4, Top: 1, Bottom: 1}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				Padding:            &widget.Insets{Right: 2, Bottom: 2},
			}),
		),
	)
	badge.AddChild(label)

	return badge
}

// activateAbilitySlot mirrors the mouse-click path: checks whether the
// ability is currently blocked (on cooldown / unit can't act), invokes the
// shared click handler, and applies the "selected" icon tint on success.
// Used by both the mouse handler and hotkey handling in Update().
func (s *Screen) activateAbilitySlot(ab ability.Ability, card *widget.Container, bgColor color.Color, iconGraphic *widget.Graphic) {
	u := s.unitByID(s.activeUnitID)
	onCooldown := u != nil && !u.AbilityReady(ab.ID)
	disabled := !onCooldown && u != nil && !ab.IsPassive && !s.unitCanAct(u)
	if onCooldown || disabled {
		sfx.Play(selectError)
		return
	}

	s.cancelAbilitySelection()
	s.onAbilityCardClicked(ab, card, bgColor)

	if s.selectedAbility != nil && s.selectedAbility.ID == ab.ID {
		iconGraphic.Image = asset.TintedImage(abilityImagePath(string(ab.ID)), activeAbilityFgColor, abilityCardSize)
		s.selectedAbilityIcon = iconGraphic
	}
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

// abilityKeyForSlot returns the configured hotkey for the ability in the
// given slot (0-based: 0 → Ability1, 1 → Ability2, ...), or nil if the
// slot has no keybinding (either out of range or unbound).
func abilityKeyForSlot(slot int) *ebiten.Key {
	kb := config.Get().Keybindings
	switch slot {
	case 0:
		return kb.Ability1
	case 1:
		return kb.Ability2
	case 2:
		return kb.Ability3
	case 3:
		return kb.Ability4
	default:
		return nil
	}
}
