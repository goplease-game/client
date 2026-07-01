package arena

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/ui"
	"github.com/goplease-game/server/ability"
	"github.com/goplease-game/server/ability/status"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/colornames"
)

const leftPanelW = 300

// createLeftPanel creates a vertical container holding the info panel and log panel.
func (s *Screen) createLeftPanel() *widget.Container {
	left := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch(
				[]bool{true},        // Stretch the only column.
				[]bool{false, true}, // Info is fixed, log fills remaining height.
			),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				StretchVertical:    true,
				Padding: &widget.Insets{
					Top:    headerH,
					Bottom: footerH + statusH,
				},
			}),
			widget.WidgetOpts.MinSize(leftPanelW, 0),
		),
	)

	s.infoPanelRef = s.createInfoPanel()
	s.logPanelRef = s.createLogPanel()

	left.AddChild(s.infoPanelRef)
	left.AddChild(s.logPanelRef)

	s.leftPanelRef = left

	return left
}

// createInfoPanel creates the panel for the unit details card.
func (s *Screen) createInfoPanel() *widget.Container {
	panel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(ui.DarkenRGB(logPanelBgColor, 10))),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(6)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{}),
			widget.WidgetOpts.MinSize(leftPanelW, 0),
		),
	)

	return panel
}

// rebuildInfoPanel clears and repopulates the info panel with current unit details.
func (s *Screen) rebuildInfoPanel(u *ds.Unit) {
	s.infoPanelRef.RemoveChildren()

	if u == nil {
		return
	}

	tf := ui.TextFace(14)
	nameTF := ui.TextFace(16)

	// --- Header: icon · name · close ---
	header := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch(
				[]bool{true, false}, // Left column stretches, right stays fixed.
				[]bool{false},
			),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)

	left := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(6),
		)),
	)

	left.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(unitImage(u.TemplateID, 24)),
	))
	left.AddChild(widget.NewText(
		widget.TextOpts.Text(u.Name, &nameTF, colornames.Gold),
	))

	header.AddChild(left)

	header.AddChild(widget.NewButton(
		widget.ButtonOpts.Text("×", &tf, &widget.ButtonTextColor{
			Idle:    logTextColor,
			Hover:   color.White,
			Pressed: color.White,
		}),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    image.NewNineSliceColor(color.NRGBA{}),
			Hover:   image.NewNineSliceColor(color.NRGBA{R: 80, G: 80, B: 80, A: 180}),
			Pressed: image.NewNineSliceColor(color.NRGBA{R: 60, G: 60, B: 60, A: 180}),
		}),
		widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
			s.hideInfoPanel()
		}),
	))

	s.infoPanelRef.AddChild(header)
	s.infoPanelRef.AddChild(infoPanelDivider())

	// --- Stats ---
	s.infoPanelRef.AddChild(infoPanelStat(
		"heart.png", "HP",
		fmt.Sprintf("%d of %d", u.CurrentHP, u.BaseHP),
		hpColor, 0, &tf,
	))
	if u.CurrentShield > 0 {
		s.infoPanelRef.AddChild(infoPanelStat(
			"shield.png", "Shield",
			fmt.Sprintf("%d", u.CurrentShield),
			shieldColor, 0, &tf,
		))
	}
	if !u.IsOpponent {
		s.infoPanelRef.AddChild(infoPanelStat(
			"hit.png", "ATK",
			fmt.Sprintf("%d", u.CurrentAtk),
			atkColor, u.CurrentAtk-u.BaseAtk, &tf,
		))
		s.infoPanelRef.AddChild(infoPanelStat(
			"walk.png", "MP",
			fmt.Sprintf("%d of %d", u.CurrentMP, u.BaseMP),
			mpColor, 0, &tf,
		))
	}

	// --- Abilities (own units only) ---
	if !u.IsOpponent && len(u.Abilities) > 0 {
		s.infoPanelRef.AddChild(infoPanelDivider())
		for _, id := range u.Abilities {
			s.infoPanelRef.AddChild(infoPanelAbility(id, u.Cooldowns[id], &tf))
		}
	}

	// --- Statuses ---
	if len(u.Statuses) > 0 {
		s.infoPanelRef.AddChild(infoPanelDivider())
		for st := range u.Statuses {
			s.infoPanelRef.AddChild(infoPanelStatus(status.ByType(st), &tf))
		}
	}
}

// infoPanelDivider returns a 1px horizontal separator.
func infoPanelDivider() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{R: 80, G: 80, B: 80, A: 120})),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
			widget.WidgetOpts.MinSize(0, 1),
		),
	)
}

// infoPanelStat returns a row: icon · label · value [· +N ▲ / -N ▼].
func infoPanelStat(icon, label, value string, valueColor color.Color, bonus int, tf *text.Face) *widget.Container {
	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	img := asset.Image(icon, 16)
	row.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(img)))
	row.AddChild(widget.NewText(widget.TextOpts.Text(label, tf, infoDimColor)))
	row.AddChild(widget.NewText(widget.TextOpts.Text(value, tf, valueColor)))
	if bonus > 0 {
		row.AddChild(widget.NewText(
			widget.TextOpts.Text(fmt.Sprintf("+%d ▲", bonus), tf, infoBonusPositiveColor),
		))
	} else if bonus < 0 {
		row.AddChild(widget.NewText(
			widget.TextOpts.Text(fmt.Sprintf("%d ▼", bonus), tf, infoBonusNegativeColor),
		))
	}
	return row
}

// infoPanelAbility returns a row: name · cooldown state.
func infoPanelAbility(id ability.ID, cooldown int, tf *text.Face) *widget.Container {
	ab := ability.ByID(id)
	name := ab.Name
	isPassive := ab.IsPassive
	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(6),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	img := abilityImage(string(id), 16)
	row.AddChild(widget.NewGraphic(widget.GraphicOpts.Image(img)))
	row.AddChild(widget.NewText(widget.TextOpts.Text(name, tf, logTextColor)))

	var cdText string
	var cdColor color.Color
	switch {
	case isPassive && cooldown > 0:
		cdText, cdColor = fmt.Sprintf("passive · %d turns", cooldown), infoCooldownColor
	case isPassive:
		cdText, cdColor = "passive · ready", infoDimColor
	case cooldown > 0:
		cdText, cdColor = fmt.Sprintf("%d turns", cooldown), infoCooldownColor
	default:
		cdText, cdColor = "ready", infoReadyColor
	}
	row.AddChild(widget.NewText(widget.TextOpts.Text(cdText, tf, cdColor)))
	return row
}

// infoPanelStatus returns a row: ● · status name, coloured by alignment.
func infoPanelStatus(def *status.Status, tf *text.Face) *widget.Container {
	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	dotColor := infoStatusPositiveColor
	switch def.Alignment {
	case status.Negative:
		dotColor = infoStatusNegativeColor
	case status.Neutral:
		dotColor = infoStatusNeutralColor
	}
	row.AddChild(widget.NewText(widget.TextOpts.Text("●", tf, dotColor)))
	row.AddChild(widget.NewText(widget.TextOpts.Text(def.Name, tf, logTextColor)))
	return row
}

// refreshBoardPadding updates the board container left padding based on
// whether the log panel or info panel is currently visible.
func (s *Screen) refreshBoardPadding() {
	leftPad := 0
	if s.infoPanelRef.GetWidget().GetVisibility() == widget.Visibility_Show || s.logPanelRef.GetWidget().GetVisibility() == widget.Visibility_Show {
		leftPad = leftPanelW
	}
	s.boardContainerRef.GetWidget().LayoutData = widget.AnchorLayoutData{
		StretchHorizontal: true,
		StretchVertical:   true,
		Padding: &widget.Insets{
			Top:    headerH,
			Bottom: footerH + statusH,
			Left:   leftPad,
		},
	}
}

// showInfoPanel opens the unit details card for the given unit.
func (s *Screen) showInfoPanel(u *ds.Unit) {
	s.infoPanelUnit = u
	s.rebuildInfoPanel(u)
	s.infoPanelRef.GetWidget().SetVisibility(widget.Visibility_Show)

	s.refreshBoardPadding()
}

// hideInfoPanel clears the content of unit details card.
func (s *Screen) hideInfoPanel() {
	s.infoPanelUnit = nil
	s.infoPanelDirty = false
	s.rebuildInfoPanel(nil)
	s.infoPanelRef.GetWidget().SetVisibility(widget.Visibility_Hide)

	s.refreshBoardPadding()
}

// markInfoPanelDirty schedules a panel rebuild if the given unit is currently displayed.
func (s *Screen) markInfoPanelDirty() {
	if s.infoPanelUnit != nil {
		s.infoPanelDirty = true
	}
}
