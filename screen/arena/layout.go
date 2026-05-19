package arena

import (
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

// ---------------------------------------------------------------------------
// Header
// ---------------------------------------------------------------------------

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

	return h
}

// ---------------------------------------------------------------------------
// Footer
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Status bar
// ---------------------------------------------------------------------------

func (s *Screen) createStatusBar() *widget.Container {
	bar := widget.NewContainer(
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

	tf := ui.TextFace(18)
	s.statusLabel = widget.NewText(
		widget.TextOpts.Text("Waiting for opponent...", &tf, statusBarTextColor),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	bar.AddChild(s.statusLabel)
	return bar
}

// ---------------------------------------------------------------------------
// Next-move button
// ---------------------------------------------------------------------------

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
			if !s.ready {
				return
			}
			s.stopEndTurnPulse()
			s.server.Send(ws.OutMessage{Action: ws.EndTurnAction})
		}),
	)

	btn.GetWidget().Disabled = true
	s.nextActionBtn = btn
	return btn
}

// ---------------------------------------------------------------------------
// End-turn button pulse helpers
// ---------------------------------------------------------------------------

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

func (s *Screen) stopEndTurnPulse() {
	s.endTurnBtnPulseActive = false
	s.nextActionBtn.Image().Idle = endTurnBtnIdle()
}

// ---------------------------------------------------------------------------
// Button image constructors (single source of truth for button colours)
// ---------------------------------------------------------------------------

func endTurnBtnIdle() *image.NineSlice {
	return image.NewBorderedNineSliceColor(
		color.NRGBA{0x22, 0x8B, 0x22, 0xff},
		color.NRGBA{0x11, 0x55, 0x11, 0xff},
		3,
	)
}

func endTurnBtnHover() *image.NineSlice {
	return image.NewBorderedNineSliceColor(
		color.NRGBA{0x32, 0xAB, 0x32, 0xff},
		color.NRGBA{0x11, 0x55, 0x11, 0xff},
		3,
	)
}

func endTurnBtnPressed() *image.NineSlice {
	return image.NewBorderedNineSliceColor(
		color.NRGBA{0x12, 0x6B, 0x12, 0xff},
		color.NRGBA{0x11, 0x55, 0x11, 0xff},
		3,
	)
}

func endTurnBtnDisabled() *image.NineSlice {
	return image.NewBorderedNineSliceColor(
		color.NRGBA{0x88, 0x88, 0x88, 0xff},
		color.NRGBA{0x55, 0x55, 0x55, 0xff},
		3,
	)
}

// ---------------------------------------------------------------------------
// Button helpers
// ---------------------------------------------------------------------------

func (s *Screen) setNextActionLabel(label string) {
	if s.nextActionBtn != nil {
		s.nextActionBtn.Text().Label = label
	}
}

func (s *Screen) enableNextActionBtn() {
	if s.nextActionBtn != nil {
		s.nextActionBtn.GetWidget().Disabled = false
	}
}
