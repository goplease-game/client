package arena

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/config"
	"github.com/ognev-dev/goplease-ebitengine-client/mock"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"golang.org/x/image/colornames"
)

const (
	devPanelW       = 300
	devPanelHeaderH = 28
)

// setupDevPanel adds the dev panel to the root container if DevMode is enabled.
// Must be called after setupUI so the root container exists.
func (s *Screen) setupDevPanel(root *widget.Container) {
	if !config.Get().DevMode.Enabled {
		return
	}

	s.devPanelRef = s.buildDevPanel()
	root.AddChild(s.devPanelRef)
}

func (s *Screen) buildDevPanel() *widget.Container {
	panel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewBorderedNineSliceColor(
			color.NRGBA{0x1a, 0x1a, 0x2e, 0xee},
			color.NRGBA{0x44, 0x44, 0x88, 0xff},
			1,
		)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(0),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				Padding:            &widget.Insets{Top: 8, Right: 8},
			}),
			widget.WidgetOpts.MinSize(devPanelW, 0),
		),
	)

	panel.AddChild(s.buildDevPanelHeader(panel))
	s.devPanelBody = s.buildDevPanelBody()
	//panel.AddChild(s.devPanelBody)

	s.devPanelMinimized = true

	return panel
}

// buildDevPanelHeader builds the title bar with a minimize toggle.
func (s *Screen) buildDevPanelHeader(panel *widget.Container) *widget.Container {
	tf := ui.TextFaceBold(12)

	header := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x2a, 0x2a, 0x4e, 0xff})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(4)),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(devPanelW, devPanelHeaderH),
		),
	)

	titleLabel := widget.NewText(
		widget.TextOpts.Text("Dev Panel", &tf, colornames.Lightblue),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)
	header.AddChild(titleLabel)

	// Minimize / restore button.
	var toggleBtn *widget.Button
	tfBtn := ui.TextFace(11)
	toggleBtn = widget.NewButton(
		widget.ButtonOpts.Text("+", &tfBtn, &widget.ButtonTextColor{
			Idle:  colornames.White,
			Hover: colornames.Yellow,
		}),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    image.NewNineSliceColor(color.NRGBA{0x33, 0x33, 0x66, 0xff}),
			Hover:   image.NewNineSliceColor(color.NRGBA{0x44, 0x44, 0x88, 0xff}),
			Pressed: image.NewNineSliceColor(color.NRGBA{0x22, 0x22, 0x44, 0xff}),
		}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(20, 20),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
		widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
			s.devPanelMinimized = !s.devPanelMinimized
			if s.devPanelMinimized {
				panel.RemoveChild(s.devPanelBody)
				toggleBtn.Text().Label = "+"
			} else {
				panel.AddChild(s.devPanelBody)
				toggleBtn.Text().Label = "−"
			}
		}),
	)
	header.AddChild(toggleBtn)

	return header
}

// buildDevPanelBody builds the collapsible content area.
func (s *Screen) buildDevPanelBody() *widget.Container {
	body := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(8)),
			widget.RowLayoutOpts.Spacing(6),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(devPanelW, 0),
		),
	)

	body.AddChild(s.buildSaveSection())
	body.AddChild(buildDivider())
	body.AddChild(s.buildLoadSection())

	return body
}

// buildSaveSection builds the "Save current state" button + status label.
func (s *Screen) buildSaveSection() *widget.Container {
	section := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
		)),
	)

	tf := ui.TextFace(11)
	tfSmall := ui.TextFace(10)

	section.AddChild(widget.NewText(
		widget.TextOpts.Text("Save state", &tf, colornames.Lightgray),
	))

	nameInput := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(devPanelW-16, 24),
		),
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     image.NewNineSliceColor(color.NRGBA{0x22, 0x22, 0x44, 0xff}),
			Disabled: image.NewNineSliceColor(color.NRGBA{0x11, 0x11, 0x22, 0xff}),
		}),
		widget.TextInputOpts.Face(&tf),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          colornames.White,
			Disabled:      colornames.Gray,
			Caret:         colornames.White,
			DisabledCaret: colornames.Gray,
		}),
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(4)),
		widget.TextInputOpts.Placeholder("save name..."),
	)
	section.AddChild(nameInput)

	statusOk := widget.NewText(
		widget.TextOpts.Text("", &tfSmall, colornames.Palegreen),
	)
	statusErr := widget.NewText(
		widget.TextOpts.Text("", &tfSmall, colornames.Red),
	)

	saveBtn := widget.NewButton(
		widget.ButtonOpts.Text("💾 Save", &tf, &widget.ButtonTextColor{
			Idle:  colornames.White,
			Hover: colornames.Yellow,
		}),
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:    image.NewNineSliceColor(color.NRGBA{0x22, 0x55, 0x22, 0xff}),
			Hover:   image.NewNineSliceColor(color.NRGBA{0x33, 0x77, 0x33, 0xff}),
			Pressed: image.NewNineSliceColor(color.NRGBA{0x11, 0x33, 0x11, 0xff}),
		}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(devPanelW-16, 24),
		),
		widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
			name := strings.TrimSpace(nameInput.GetText())
			if name == "" {
				name = fmt.Sprintf("save_%d", time.Now().Unix())
			}
			if err := mock.SaveState(name, s.takeSnapshot()); err != nil {
				statusErr.Label = "Error: " + err.Error()
				statusOk.Label = ""
			} else {
				statusOk.Label = "Saved: " + name + ".json"
				statusErr.Label = ""
				s.rebuildLoadList()
			}
		}),
	)

	section.AddChild(saveBtn)
	section.AddChild(statusOk)
	section.AddChild(statusErr)

	return section
}

// buildLoadSection builds the scrollable list of available states.
func (s *Screen) buildLoadSection() *widget.Container {
	section := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
		)),
	)

	tf := ui.TextFace(11)
	section.AddChild(widget.NewText(
		widget.TextOpts.Text("Load state", &tf, colornames.Lightgray),
	))

	s.devLoadList = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(2),
		)),
	)
	s.rebuildLoadList()

	section.AddChild(s.devLoadList)
	return section
}

// rebuildLoadList refreshes the list of loadable states.
func (s *Screen) rebuildLoadList() {
	if s.devLoadList == nil {
		return
	}
	s.devLoadList.RemoveChildren()

	tf := ui.TextFace(10)
	for _, name := range mock.ListStates() {
		n := name // capture for closure
		btn := widget.NewButton(
			widget.ButtonOpts.Text(n, &tf, &widget.ButtonTextColor{
				Idle:  colornames.White,
				Hover: colornames.Yellow,
			}),
			widget.ButtonOpts.Image(&widget.ButtonImage{
				Idle:    image.NewNineSliceColor(color.NRGBA{0x22, 0x22, 0x44, 0xff}),
				Hover:   image.NewNineSliceColor(color.NRGBA{0x33, 0x33, 0x66, 0xff}),
				Pressed: image.NewNineSliceColor(color.NRGBA{0x11, 0x11, 0x22, 0xff}),
			}),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(devPanelW-16, 22),
			),
			widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
				s.loadDevState(n)
			}),
		)
		s.devLoadList.AddChild(btn)
	}
}

// loadDevState loads the selected state and reinitialises the screen.
func (s *Screen) loadDevState(name string) {
	snap, err := mock.LoadState(name)
	if err != nil {
		s.setStatus("Dev: failed to load " + name)
		return
	}
	s.restoreSnapshot(snap)
	s.setStatus("Dev: loaded " + name)

	mock.RestoreGameState(name, snap)
	s.pendingScreen = NewScreen(snap, s.server)
}

func buildDivider() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x44, 0x44, 0x66, 0xff})),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(devPanelW-16, 1),
		),
	)
}
