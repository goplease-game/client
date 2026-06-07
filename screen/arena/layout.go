package arena

import (
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/config"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
	"golang.org/x/image/colornames"
)

const (
	abilityCardSize = 64 // ability card size in the footer panel
	unitCardSize    = 64 // unit card size in the hand and queue panel
	unitIconSize    = 54 // unit portrait size rendered on the board hex

	headerH = 80
	statusH = 32
	footerH = 90
)

// createHeader builds the top bar container that holds the unit queue panel.
// The queue panel is stored in s.queuePanelRef for later population.
func (s *Screen) createHeader() *widget.Container {
	h := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(headerBgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true}),
			widget.WidgetOpts.MinSize(0, headerH),
		),
	)

	s.queuePanelRef = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(unitPanelBgColor)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(4)),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	tf := ui.TextFace(16)
	menuBtn := widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    image.NewNineSliceColor(color.NRGBA{0x44, 0x44, 0x44, 0xff}),
			Hover:   image.NewNineSliceColor(color.NRGBA{0x66, 0x66, 0x66, 0xff}),
			Pressed: image.NewNineSliceColor(color.NRGBA{0x33, 0x33, 0x33, 0xff}),
		}),
		widget.ButtonOpts.Text("≡", &tf, &widget.ButtonTextColor{Idle: colornames.White}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(36, 36),
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				Padding:            &widget.Insets{Right: 8},
			}),
		),
		widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
			s.toggleGameMenu()
		}),
	)
	h.AddChild(s.queuePanelRef)
	h.AddChild(menuBtn)

	return h
}

// createFooter builds the bottom bar container with the Next button anchored
// to the right. The ability panel is added dynamically via showAbilityPanel.
func (s *Screen) createFooter() *widget.Container {
	footer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(footerBgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				VerticalPosition:  widget.AnchorLayoutPositionEnd,
				StretchHorizontal: true,
			}),
			widget.WidgetOpts.MinSize(0, footerH),
		),
	)

	btn := s.buildNextMoveButton()
	btn.GetWidget().LayoutData = widget.AnchorLayoutData{
		HorizontalPosition: widget.AnchorLayoutPositionEnd,
		VerticalPosition:   widget.AnchorLayoutPositionCenter,
		Padding:            &widget.Insets{Right: 12},
	}
	footer.AddChild(btn)

	return footer
}

// createStatusBar builds the thin bar above the footer that shows game status text.
func (s *Screen) createStatusBar() *widget.Container {
	s.statusBarRef = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(statusBarBgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				VerticalPosition:  widget.AnchorLayoutPositionEnd,
				StretchHorizontal: true,
				Padding:           &widget.Insets{Bottom: footerH},
			}),
			widget.WidgetOpts.MinSize(0, statusH),
		),
	)

	s.setStatus("Waiting for opponent...")

	return s.statusBarRef
}

// buildNextMoveButton creates the Next/End Turn button and stores it in s.nextActionBtn.
// The button starts disabled and is enabled when it becomes the player's turn.
func (s *Screen) buildNextMoveButton() *widget.Button {
	const size = 80
	tf := ui.TextFace(18)

	btn := widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:     endTurnBtnIdle(),
			Hover:    endTurnBtnHover(),
			Pressed:  endTurnBtnPressed(),
			Disabled: endTurnBtnDisabled(),
		}),
		widget.ButtonOpts.Text("Next", &tf, &widget.ButtonTextColor{
			Idle:     color.NRGBA{0xff, 0xff, 0xff, 0xff},
			Disabled: color.NRGBA{0xaa, 0xaa, 0xaa, 0xff},
		}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(size, size),
		),
		widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
			if config.Get().DevMode.Enabled {
				printD("NEXT TURN PRESSED")
			}
			if !s.ready {
				printD("NOT READY")

				return
			}
			s.stopEndTurnPulse()

			if u := s.unitByID(s.activeUnitID); u != nil {
				s.activeUnitMoved = true
				if bc := s.boardCellWidget(u); bc != nil {
					bc.RemoveChildren()
					s.buildBoardCard(bc, u, false)
				}
			}
			s.activeUnitID = ""

			s.setPulseHexTargets(nil)
			s.server.Send(ws.OutMessage{Action: ws.EndTurnAction})
		}),
	)

	btn.GetWidget().Disabled = true
	s.nextActionBtn = btn
	return btn
}

// pulseEndTurnBtn updates the Next button border colour for the current pulse frame.
// t is a normalised value in [0, 1] driven by the pulse sine wave.
func (s *Screen) pulseEndTurnBtn(t float64) {
	borderColor := ui.LerpColor(
		color.RGBA{0x11, 0x55, 0x11, 0xff},
		color.RGBA{0x88, 0xFF, 0x88, 0xff},
		t,
	)
	s.nextActionBtn.Image().Idle = image.NewBorderedNineSliceColor(
		color.NRGBA{0x22, 0x8B, 0x22, 0xff},
		borderColor,
		3,
	)
}

// stopEndTurnPulse cancels the pulse animation and restores the button's idle image.
func (s *Screen) stopEndTurnPulse() {
	s.endTurnBtnPulseActive = false
	s.nextActionBtn.Image().Idle = endTurnBtnIdle()
}

// endTurnBtnIdle returns the default nine-slice image for the Next button.
func endTurnBtnIdle() *image.NineSlice {
	return image.NewBorderedNineSliceColor(
		color.NRGBA{0x22, 0x8B, 0x22, 0xff},
		color.NRGBA{0x11, 0x55, 0x11, 0xff},
		3,
	)
}

// endTurnBtnHover returns the hovered nine-slice image for the Next button.
func endTurnBtnHover() *image.NineSlice {
	return image.NewBorderedNineSliceColor(
		color.NRGBA{0x32, 0xAB, 0x32, 0xff},
		color.NRGBA{0x11, 0x55, 0x11, 0xff},
		3,
	)
}

// endTurnBtnPressed returns the pressed nine-slice image for the Next button.
func endTurnBtnPressed() *image.NineSlice {
	return image.NewBorderedNineSliceColor(
		color.NRGBA{0x12, 0x6B, 0x12, 0xff},
		color.NRGBA{0x11, 0x55, 0x11, 0xff},
		3,
	)
}

// endTurnBtnDisabled returns the disabled nine-slice image for the Next button.
func endTurnBtnDisabled() *image.NineSlice {
	return image.NewBorderedNineSliceColor(
		color.NRGBA{0x88, 0x88, 0x88, 0xff},
		color.NRGBA{0x55, 0x55, 0x55, 0xff},
		3,
	)
}

// setNextActionLabel updates the label text on the Next button.
func (s *Screen) setNextActionLabel(label string) {
	if s.nextActionBtn != nil {
		s.nextActionBtn.Text().Label = label
	}
}

// enableNextActionBtn enables the Next button so the player can end their turn.
func (s *Screen) enableNextActionBtn() {
	if s.nextActionBtn != nil {
		s.nextActionBtn.GetWidget().Disabled = false
	}
}

// updateNextActionLabel sets the Next button label based on
// whether the unit has exhausted both movement and AP.
func (s *Screen) updateNextActionLabel() {
	u := s.unitByID(s.activeUnitID)

	if u == nil || (s.activeUnitMoved && s.unitCanAct(u)) {
		s.setNextActionLabel("END\nTURN")
	} else {
		s.setNextActionLabel("SKIP\nTURN")
	}
}

// createGameMenu builds a full-screen semi-transparent overlay with a centered
// menu containing Restart, Surrender, and Exit buttons. Hidden by default.
func (s *Screen) createGameMenu() *widget.Container {
	overlay := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0, 0, 0, 160})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
	)

	menu := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(headerBgColor)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(24)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	if s.OnRestartScreen != nil {
		menu.AddChild(s.menuButton("Restart", func(args *widget.ButtonClickedEventArgs) {
			s.closeGameMenu()
			s.nextScreen = s.OnRestartScreen()
		}))
	}

	menu.AddChild(s.menuButton("Surrender", func(args *widget.ButtonClickedEventArgs) {
		s.closeGameMenu()
		// TODO: s.onSurrender()
	}))
	menu.AddChild(s.menuButton("Exit", func(args *widget.ButtonClickedEventArgs) {
		s.nextScreen = s.OnExitScreen()
	}))

	overlay.AddChild(menu)

	return overlay
}

// menuButton creates a styled button for the game menu.
func (s *Screen) menuButton(label string, onClick widget.ButtonClickedHandlerFunc) *widget.Button {
	tf := ui.TextFaceBold(16)
	return widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    image.NewNineSliceColor(menuButtonBgColor),
			Hover:   image.NewNineSliceColor(menuButtonHoverBgColor),
			Pressed: image.NewNineSliceColor(menuButtonBgColor),
		}),
		widget.ButtonOpts.Text(label, &tf, &widget.ButtonTextColor{
			Idle:  menuButtonTextColor,
			Hover: menuButtonHoverTextColor,
		}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(160, 44),
		),
		widget.ButtonOpts.ClickedHandler(onClick),
	)
}

// openGameMenu shows the game menu overlay.
func (s *Screen) openGameMenu() {
	if s.menuOverlayRef == nil {
		s.menuOverlayRef = s.createGameMenu()
		s.menuUI = &ebitenui.UI{Container: s.menuOverlayRef}
	}
	s.menuVisible = true
}

// closeGameMenu hides the game menu overlay.
func (s *Screen) closeGameMenu() {
	s.menuVisible = false
	s.menuOverlayRef.GetWidget().SetVisibility(widget.Visibility_Hide)
}

// toggleGameMenu opens or closes the game menu.
func (s *Screen) toggleGameMenu() {
	if s.menuVisible {
		s.closeGameMenu()
	} else {
		s.openGameMenu()
	}
}
