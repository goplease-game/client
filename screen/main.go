package screen

import (
	"encoding/json"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/config"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/scenario"
	"github.com/goplease-game/client/sfx"
	"github.com/goplease-game/client/ui"
	"github.com/goplease-game/client/ws"
	server "github.com/goplease-game/server"
	"github.com/goplease-game/server/bot"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/colornames"
)

// Shared color palette for the main menu UI.
var (
	nameColor                = ui.RGBFromHex("#00a8e8")
	menuButtonBgColor        = ui.RGBFromHex("#73A5CA")
	menuButtonHoverBgColor   = ui.LightenRGB(menuButtonBgColor, 35)
	menuButtonTextColor      = ui.RGBFromHex("FFF8DE")
	menuButtonHoverTextColor = ui.DarkenRGB(menuButtonBgColor, 45)
)

// MainScreen is the entry screen with the "Play" button.
type MainScreen struct {
	serverCl   *ws.ClientProvider
	ui         *ebitenui.UI
	nextScreen game.Screen
	exit       bool
	descText   *widget.Text
}

// NewMainScreen creates the main menu screen with Play, Practice, Settings,
// About, and (on non-WASM builds) Exit buttons.
func NewMainScreen(serverCl *ws.ClientProvider) *MainScreen {
	serverCl.SwitchToReal()

	s := &MainScreen{
		serverCl: serverCl,
	}

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	footer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				StretchHorizontal:  true,
			}),
		),
	)

	versionTF := ui.TextFace(12)
	versionText := widget.NewText(
		// TODO fetch server's status & version
		widget.TextOpts.Text("client v0.0.1\nserver v0.0.1", &versionTF, color.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	footer.AddChild(versionText)

	root.AddChild(s.mainMenu())
	root.AddChild(footer)

	s.ui = &ebitenui.UI{Container: root}

	return s
}

// Update implements game.Screen. It drives the main menu UI and exits or
// transitions to the next screen when requested.
func (s *MainScreen) Update(_ *game.Game) (game.Screen, error) {
	if s.exit {
		return nil, ebiten.Termination
	}

	s.ui.Update()

	if s.nextScreen != nil {
		next := s.nextScreen
		s.nextScreen = nil
		return next, nil
	}

	return s, nil
}

// Draw implements game.Screen. It renders the main menu UI.
func (s *MainScreen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

// mainMenu builds the title and menu button column shown on the main screen.
func (s *MainScreen) mainMenu() *widget.Container {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchHorizontal:  true,
			}),
		),
	)

	rowC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(30),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Left: 50,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	menuC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(300, 0),
		),
	)

	titleTF := ui.TextFace(40)
	titleText := widget.NewText(
		widget.TextOpts.Text("go, please", &titleTF, nameColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	descTF := ui.TextFace(16)
	s.descText = widget.NewText(
		widget.TextOpts.Text("", &descTF, menuButtonTextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.MaxWidth(300),
	)

	descContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(&widget.Insets{
				Top: 55,
			}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(250, 100),
		),
	)
	descContainer.AddChild(s.descText)

	playButton := s.mainMenuButtonWithDesc("PLAY", 30, "Challenge other players online",
		func(_ *widget.ButtonClickedEventArgs) {
			s.nextScreen = NewSearchScreen(s.serverCl)
		})

	pwfButton := s.mainMenuButtonWithDesc("Play with friend", 16,
		"Challenge a friend using a join code.",
		func(_ *widget.ButtonClickedEventArgs) {
			s.nextScreen = NewPlayWithFriendScreen(s.serverCl)
		})

	practiceButton := s.mainMenuButtonWithDesc("Practice", 16,
		"Learn the basics and play local matches against Richard the Bot. No internet connection required.",
		func(_ *widget.ButtonClickedEventArgs) {
			s.nextScreen = newScenarioScreen(s.serverCl)
		})

	settButton := s.mainMenuButtonWithDesc("Settings", 16, "", func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = NewSettingsScreen(s)
	})

	aboutButton := s.mainMenuButtonWithDesc("About", 16, "", func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = NewAboutScreen(s)
	})

	var exitButton *widget.Button
	if !config.IsWASM() {
		exitButton = s.mainMenuButtonWithDesc("Exit", 14, "", func(_ *widget.ButtonClickedEventArgs) {
			s.exit = true
		})
	}

	menuC.AddChild(titleText)
	menuC.AddChild(playButton)
	menuC.AddChild(pwfButton)
	menuC.AddChild(practiceButton)
	menuC.AddChild(settButton)
	menuC.AddChild(aboutButton)

	if exitButton != nil {
		menuC.AddChild(exitButton)
	}

	rowC.AddChild(menuC)
	rowC.AddChild(descContainer)

	c.AddChild(rowC)

	return c
}

// mainMenuButtonWithDesc creates a primary menu button with hover sound, hover
// text size change, a press-down text shift, and dynamic description text logic.
func (s *MainScreen) mainMenuButtonWithDesc(text string, size float64, desc string, clickHandler widget.ButtonClickedHandlerFunc) *widget.Button {
	tf := ui.TextFace(size)
	tfHover := ui.TextFace(size + 5)
	var button *widget.Button
	button = widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
		widget.ButtonOpts.Image(mainMenuButtonImage()),
		widget.ButtonOpts.Text(text, &tf, &widget.ButtonTextColor{
			Idle:    menuButtonTextColor,
			Hover:   menuButtonHoverTextColor,
			Pressed: menuButtonTextColor,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   45,
			Right:  45,
			Top:    15,
			Bottom: 15,
		}),
		widget.ButtonOpts.PressedHandler(func(_ *widget.ButtonPressedEventArgs) {
			button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			button.GetWidget().CustomData = true
		}),
		widget.ButtonOpts.ReleasedHandler(func(_ *widget.ButtonReleasedEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
			button.GetWidget().CustomData = false
		}),
		widget.ButtonOpts.ClickedHandler(clickHandler),
		widget.ButtonOpts.CursorEnteredHandler(func(_ *widget.ButtonHoverEventArgs) {
			sfx.Play("button_hover.ogg")
			button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			button.Text().SetFace(&tfHover)

			if desc != "" {
				s.descText.Label = desc
			}

			button.GetWidget().Render(nil)
		}),
		widget.ButtonOpts.CursorExitedHandler(func(_ *widget.ButtonHoverEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
			button.Text().SetFace(&tf)

			s.descText.Label = ""
		}),
	)

	return button
}

// mainMenuButtonImage returns the nine-slice background images for
// primary menu buttons.
func mainMenuButtonImage() *widget.ButtonImage {
	idle := image.NewNineSliceColor(menuButtonBgColor)
	hover := image.NewNineSliceColor(menuButtonHoverBgColor)
	pressed := image.NewNineSliceColor(colornames.Gold)

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}
}

// newScenarioScreen loads the default scenario and starts an arena screen
// backed by a connected mock client.
func newScenarioScreen(serverCl *ws.ClientProvider) game.Screen {
	sc := scenario.Load(scenario.Default)
	session := server.NewSessionFromSnapshot(sc.Arena())

	if !sc.DisableBot {
		b := bot.NewWithSession(sc.P2.ID, session)
		go b.Run()
	}

	mockCl := serverCl.SwitchToMock(session, sc.P1.ID)
	session.Start()

	for msg := range mockCl.Inbox() {
		if msg.Action == ws.NewGameAction {
			var data ds.NewGamePayload
			err := json.Unmarshal(msg.Data, &data)
			if err != nil {
				log.Fatalf("scenario: failed to unmarshal new game: %v", err)
			}
			snap := ds.GameSnapshot{
				ArenaID:                    data.ArenaID,
				Board:                      data.Board,
				UnitsQueue:                 data.Queue,
				Player:                     *data.Player,
				OpponentName:               data.Opponent,
				Round:                      data.Round,
				TurnTimeSeconds:            data.TurnTimeSeconds,
				MaxPhantomAPPerUnitPerTurn: data.MaxPhantomAPPerUnitPerTurn,
				Tutorial:                   sc.Tutorial,
			}
			return NewArenaScreen(snap, serverCl, true)
		}
	}

	log.Fatal("scenario: server closed without sending new game")
	return nil
}

// buttonProps holds optional overrides for secondaryButton.
type buttonProps struct {
	layoutData any // widget.AnchorLayoutData or widget.RowLayoutData
}

// secondaryButton creates a smaller menu button styled like
// mainMenuButton, used for secondary actions like Back.
func secondaryButton(text string, size float64, clickHandler widget.ButtonClickedHandlerFunc, props ...buttonProps) *widget.Button {
	tf := ui.TextFace(size)
	tfHover := ui.TextFace(size + 5)

	var layoutData any = widget.RowLayoutData{
		Position: widget.RowLayoutPositionCenter,
	}
	if len(props) > 0 && props[0].layoutData != nil {
		layoutData = props[0].layoutData
	}

	var button *widget.Button
	button = widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.LayoutData(layoutData),
		),
		widget.ButtonOpts.Image(mainMenuButtonImage()),
		widget.ButtonOpts.Text(text, &tf, &widget.ButtonTextColor{
			Idle:    menuButtonTextColor,
			Hover:   menuButtonHoverTextColor,
			Pressed: menuButtonTextColor,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   25,
			Right:  25,
			Top:    10,
			Bottom: 10,
		}),
		widget.ButtonOpts.PressedHandler(func(_ *widget.ButtonPressedEventArgs) {
			button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			button.GetWidget().CustomData = true
		}),
		widget.ButtonOpts.ReleasedHandler(func(_ *widget.ButtonReleasedEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
			button.GetWidget().CustomData = false
		}),
		widget.ButtonOpts.ClickedHandler(clickHandler),
		widget.ButtonOpts.CursorEnteredHandler(func(_ *widget.ButtonHoverEventArgs) {
			sfx.Play("button_hover.ogg")
			button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			button.Text().SetFace(&tfHover)
			button.GetWidget().Render(nil)
		}),
		widget.ButtonOpts.CursorExitedHandler(func(_ *widget.ButtonHoverEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
			button.Text().SetFace(&tf)
		}),
	)
	return button
}

func rowButton(text string, size float64, clickHandler widget.ButtonClickedHandlerFunc) *widget.Button {
	return secondaryButton(text, size, clickHandler, buttonProps{
		layoutData: widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		},
	})
}
