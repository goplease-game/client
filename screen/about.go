// Package screen ...
package screen

import (
	"fmt"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/backdrop"
	"github.com/goplease-game/client/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

// AboutScreen shows information about the game: version, credits, etc.
type AboutScreen struct {
	ui         *ebitenui.UI
	bg         backdrop.Backdrop
	nextScreen game.Screen
	prevScreen game.Screen
}

// NewAboutScreen creates the about screen. previous is the screen
// to return to when the player presses Back.
func NewAboutScreen(prevScreen *MainScreen) *AboutScreen {
	s := &AboutScreen{
		prevScreen: prevScreen,
		bg:         prevScreen.bg,
	}

	panel := ui.NewPanel("About")

	bodyTF := ui.TextFace(16)
	lines := []string{
		"go, please: a turn-based tactical hex-grid game.",
		" ",
		"This is an open-source and community-driven project in early development. ",
		"Anyone is welcome to contribute.",
		" ",
		"Community hub: [link=discord]Discord[/link]",
		"Dev hub: [link=source]Github[/link]",
		" ",
		"Built using [link=golang]Go[/link], [link=ebitengine]Ebitengine[/link] and [link=ebitenui]EbitenUI[/link].",
	}

	for _, line := range lines {
		text := widget.NewText(
			widget.TextOpts.ProcessBBCode(true),
			widget.TextOpts.LinkColor(&widget.TextLinkColor{
				Idle:  nameColor,
				Hover: colornames.Gold,
			}),
			widget.TextOpts.Text(line, &bodyTF, ui.RGBFromHex("e3e9ef")),
			widget.TextOpts.LinkClickedHandler(func(args *widget.LinkEventArgs) {
				err := game.OpenLink(args.Id)
				if err != nil {
					fmt.Printf("open URL error: %v\n", err)
				}
			}),
		)

		game.SetLinksTheme(text.GetWidget())
		panel.AddContent(text)
	}

	panel.AddControl(ui.SecondaryButton("Back", 14, func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = s.prevScreen
	}))

	s.ui = &ebitenui.UI{
		Container: panel.Build(),
	}

	return s
}

// Update advances the about screen UI and returns the previous screen
// once Back is pressed.
func (s *AboutScreen) Update(_ *game.Game) (game.Screen, error) {
	s.bg.Update()
	s.ui.Update()

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return s.prevScreen, nil
	}

	if s.nextScreen != nil {
		next := s.nextScreen
		s.nextScreen = nil
		return next, nil
	}

	return s, nil
}

// Draw renders the about screen UI.
func (s *AboutScreen) Draw(screen *ebiten.Image) {
	s.bg.Draw(screen)
	s.ui.Draw(screen)
}
