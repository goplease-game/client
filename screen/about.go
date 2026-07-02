// Package screen ...
package screen

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

// AboutScreen shows information about the game: version, credits, etc.
type AboutScreen struct {
	previous   game.Screen
	ui         *ebitenui.UI
	nextScreen game.Screen
}

// NewAboutScreen creates the about screen. previous is the screen
// to return to when the player presses Back.
func NewAboutScreen(previous game.Screen) *AboutScreen {
	s := &AboutScreen{previous: previous}

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	panel := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(15),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(400, 0),
		),
	)

	titleTF := ui.TextFace(30)
	title := widget.NewText(
		widget.TextOpts.Text("About", &titleTF, nameColor),
	)

	bodyTF := ui.TextFace(16)
	contents := []string{
		"go, please: a turn-based tactical hex-grid game.",
		"--",
		"This is an open-source and community-driven project in early development. ",
		"Anyone is welcome to contribute.",
		"",
		"Community hub: [link=discord]Discord[/link]",
		"Dev hub: [link=source]Github[/link]",
		"--",
		"Built using [link=golang]Go[/link], [link=ebitengine]Ebitengine[/link] and [link=ebitenui]EbitenUI[/link].",
	}

	body := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(ui.RGBFromHex("3e4c51"))),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(25)),
		)),
	)
	for _, line := range contents {
		text := widget.NewText(
			widget.TextOpts.ProcessBBCode(true),
			widget.TextOpts.LinkColor(&widget.TextLinkColor{
				Idle:  nameColor,
				Hover: colornames.Gold,
			}),
			widget.TextOpts.Text(line, &bodyTF, color.White),
			widget.TextOpts.LinkClickedHandler(func(args *widget.LinkEventArgs) {
				err := game.OpenLink(args.Id)
				if err != nil {
					fmt.Printf("open URL error: %v\n", err)
				}
			}),
		)

		game.SetLinksTheme(text.GetWidget())

		body.AddChild(text)
	}

	backButton := secondaryButton("Back", 14, func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = s.previous
	})

	panel.AddChild(title)
	panel.AddChild(body)
	panel.AddChild(backButton)

	root.AddChild(panel)

	s.ui = &ebitenui.UI{Container: root}

	return s
}

// Update advances the about screen UI and returns the previous screen
// once Back is pressed.
func (s *AboutScreen) Update(_ *game.Game) (game.Screen, error) {
	s.ui.Update()

	if s.nextScreen != nil {
		next := s.nextScreen
		s.nextScreen = nil
		return next, nil
	}

	return s, nil
}

// Draw renders the about screen UI.
func (s *AboutScreen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}
