package screen

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net/http"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/backdrop"
	"github.com/goplease-game/client/config"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/scenario"
	"github.com/goplease-game/client/sfx"
	"github.com/goplease-game/client/ui"
	"github.com/goplease-game/client/ws"
	server "github.com/goplease-game/game-server"
	"github.com/goplease-game/game-server/bot"
	"github.com/hajimehoshi/ebiten/v2"
)

const mainMusicTheme = "its-time-for-an-adventure.ogg"

var (
	nameColor = ui.RGBFromHex("#00a8e8")
)

var (
	badColor = "#EA7B7B"
)

// MainScreen is the entry screen with the "Play" button.
type MainScreen struct {
	serverCl   *ws.ClientProvider
	ui         *ebitenui.UI
	bg         backdrop.Backdrop
	nextScreen game.Screen
	exit       bool
	descText   *widget.Text
	music      *sfx.MusicTrack
}

// NewMainScreen creates the main menu screen with Play, Practice, Settings,
// About, and (on non-WASM builds) Exit buttons.
func NewMainScreen(serverCl *ws.ClientProvider) *MainScreen {
	serverCl.SwitchToReal()

	sfx.StopAll()
	mainTune := sfx.PlayMusic(mainMusicTheme, true, time.Second*5)

	conf := config.Get()
	s := &MainScreen{
		serverCl: serverCl,
		bg:       backdrop.RandomOf(backdrop.MainScreen, conf.WindowW, conf.WindowH),
		music:    mainTune,
	}

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	footer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				StretchHorizontal:  true,
			}),
		),
	)

	versions := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
	)

	versionTF := ui.TextFace(14)
	var date string
	date, err := game.DisplayDateFromRFC(game.BuildDate)
	if err != nil {
		date = "invalid date"
	}
	clientV := widget.NewText(
		widget.TextOpts.ProcessBBCode(true),
		widget.TextOpts.LinkClickedHandler(func(_ *widget.LinkEventArgs) {
			err := game.OpenLink("source-client")
			if err != nil {
				fmt.Printf("open URL error: %v\n", err)
			}
		}),
		widget.TextOpts.Text(
			fmt.Sprintf("Client: [link=source-client]%s @ %s[/link]", game.Commit, date),
			&versionTF, color.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)
	serverV := widget.NewText(
		widget.TextOpts.ProcessBBCode(true),
		widget.TextOpts.LinkClickedHandler(func(_ *widget.LinkEventArgs) {
			err := game.OpenLink("source-server")
			if err != nil {
				fmt.Printf("open URL error: %v\n", err)
			}
		}),
		widget.TextOpts.Text("Server: connecting...", &versionTF, color.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	game.SetLinksTheme(clientV.GetWidget())
	game.SetLinksTheme(serverV.GetWidget())

	go fetchVersion(game.ServerAPI("version/"), serverV)

	versions.AddChild(clientV)
	versions.AddChild(serverV)

	discordLink := widget.NewText(
		widget.TextOpts.ProcessBBCode(true),
		widget.TextOpts.LinkClickedHandler(func(_ *widget.LinkEventArgs) {
			err := game.OpenLink("discord")
			if err != nil {
				fmt.Printf("open URL error: %v\n", err)
			}
		}),
		widget.TextOpts.Text("[link=discord]Community Discord[/link]", &versionTF, color.White),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionEnd),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
			}),
		),
	)
	game.SetLinksTheme(discordLink.GetWidget())

	footer.AddChild(versions)
	footer.AddChild(discordLink)

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

	s.bg.Update()
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
	s.bg.Draw(screen)
	s.ui.Draw(screen)
}

// Resize updates the backdrop dimensions when the screen or window is resized.
func (s *MainScreen) Resize(width, height int) {
	s.bg.Resize(width, height)
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

	namePic := asset.Image("name.png")

	titleImage := widget.NewGraphic(
		widget.GraphicOpts.Image(namePic),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	descTF := ui.TextFace(16)
	s.descText = widget.NewText(
		widget.TextOpts.Text("", &descTF, ui.MenuButtonTextColor),
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
			s.music.FadeOut(5 * time.Second)
			s.nextScreen = NewSearchScreen(s.serverCl, s)
		})

	pwfButton := s.mainMenuButtonWithDesc("Play with friend", 16,
		"Challenge a friend using a join code.",
		func(_ *widget.ButtonClickedEventArgs) {
			s.nextScreen = NewPlayWithFriendScreen(s.serverCl, s)
		})

	practiceButton := s.mainMenuButtonWithDesc("Practice", 16,
		"Learn the basics and play local matches against Richard the Bot. No internet connection required.",
		func(_ *widget.ButtonClickedEventArgs) {
			s.music.FadeOut(3 * time.Second)
			s.nextScreen = newScenarioScreen(s.serverCl)
		})

	settButton := s.mainMenuButtonWithDesc("Settings", 16, "", func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = NewSettingsScreen(s)
	})

	aboutButton := s.mainMenuButtonWithDesc("About", 16, "", func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = NewAboutScreen(s)
	})

	var lastBtn *widget.Button
	if !config.IsWASM() {
		lastBtn = s.mainMenuButtonWithDesc("Exit", 14, "Snap back to reality", func(_ *widget.ButtonClickedEventArgs) {
			s.exit = true
		})
	} else {
		lastBtn = s.mainMenuButtonWithDesc("Download", 14, "Native binary for Windows, macOS and Linux", func(_ *widget.ButtonClickedEventArgs) {
			err := game.OpenLink("latest-releases")
			if err != nil {
				fmt.Printf("open download link: %v\n", err)
			}
		})
	}

	menuC.AddChild(titleImage)
	menuC.AddChild(playButton)
	menuC.AddChild(pwfButton)
	menuC.AddChild(practiceButton)
	menuC.AddChild(settButton)
	menuC.AddChild(aboutButton)

	if lastBtn != nil {
		menuC.AddChild(lastBtn)
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
	tfHover := ui.TextFace(size + 2)
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
		widget.ButtonOpts.Image(ui.ButtonImage()),
		widget.ButtonOpts.Text(text, &tf, &widget.ButtonTextColor{
			Idle:    ui.MenuButtonTextColor,
			Hover:   ui.MenuButtonHoverTextColor,
			Pressed: ui.MenuButtonTextColor,
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

func fetchVersion(url string, out *widget.Text) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
	}
	resp, err := http.DefaultClient.Do(req) //nolint:gosec
	if err != nil {
		out.Label = "Server: [color=" + badColor + "]unavailable[/color]"
		return
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	var v server.Version
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		out.Label = "Server: [color=" + badColor + "]unavailable[/color]"
		return
	}

	var date string
	date, err = game.DisplayDateFromRFC(v.BuildDate)
	if err != nil {
		date = "invalid date"
	}

	out.Label = "Server: [link=source-server]" + v.Commit + " @ " + date + "[/link]"
}
