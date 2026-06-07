package arena

import (
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
)

// showGameOverOverlay displays the game-over overlay with the given title.
// Creates the overlay lazily on first call.
func (s *Screen) showGameOverOverlay(win bool) {
	title := "You Lose"
	titleColor := gameOverLoseColor
	if win {
		title = "You Win"
		titleColor = gameOverWinColor
	}

	overlay := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0, 0, 0, 180})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
	)

	panel := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(24),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(40)),
		)),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(headerBgColor)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	tf := ui.TextFaceBold(48)
	panel.AddChild(widget.NewText(
		widget.TextOpts.Text(title, &tf, titleColor),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	))

	buttons := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(12),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	buttons.AddChild(s.menuButton("Play Again", func(_ *widget.ButtonClickedEventArgs) {
		if s.OnRestartScreen != nil {
			s.nextScreen = s.OnRestartScreen()
		}
	}))
	buttons.AddChild(s.menuButton("Main Menu", func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = s.OnExitScreen()
	}))

	panel.AddChild(buttons)
	overlay.AddChild(panel)

	s.gameOverUI = &ebitenui.UI{Container: overlay}
	s.gameOverVisible = true
}
