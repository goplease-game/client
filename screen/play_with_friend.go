package screen

import (
	"encoding/json"
	"image/color"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/google/uuid"
	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/backdrop"
	"github.com/goplease-game/client/clipboard"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/ui"
	"github.com/goplease-game/client/ws"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/colornames"
)

// PlayWithFriendScreen lets the player choose between creating or joining a friend game.
type PlayWithFriendScreen struct {
	provider   *ws.ClientProvider
	ui         *ebitenui.UI
	bg         backdrop.Backdrop
	prevScreen game.Screen
	nextScreen game.Screen
}

// NewPlayWithFriendScreen creates the Play With Friend mode selection screen.
func NewPlayWithFriendScreen(provider *ws.ClientProvider, prevScreen *MainScreen) *PlayWithFriendScreen {
	s := &PlayWithFriendScreen{
		provider:   provider,
		bg:         prevScreen.bg,
		prevScreen: prevScreen,
	}

	panel := ui.NewPanel("Play with friend")

	createBtn := ui.SecondaryButton("Create game", 16, func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = NewWaitForFriendScreen(s.provider, s)
	})
	joinBtn := ui.SecondaryButton("Join game", 16, func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = NewJoinFriendScreen(s.provider, s)
	})

	ui.StretchButton(createBtn)
	ui.StretchButton(joinBtn)

	panel.AddContent(createBtn, joinBtn)

	backBtn := ui.SecondaryButton("Back", 14, func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = s.prevScreen
	})
	panel.AddControl(backBtn)

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	root.AddChild(panel.Build())

	s.ui = &ebitenui.UI{Container: root}
	return s
}

// Update implements game.Screen.
func (s *PlayWithFriendScreen) Update(_ *game.Game) (game.Screen, error) {
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

// Draw implements game.Screen.
func (s *PlayWithFriendScreen) Draw(screen *ebiten.Image) {
	s.bg.Draw(screen)
	s.ui.Draw(screen)
}

// Resize updates the backdrop dimensions when the screen or window is resized.
func (s *PlayWithFriendScreen) Resize(width, height int) {
	s.bg.Resize(width, height)
}

// WaitForFriendScreen connects to the server, requests a friend room, displays
// the join code, and waits for the opponent to connect.
type WaitForFriendScreen struct {
	provider   *ws.ClientProvider
	ui         *ebitenui.UI
	bg         backdrop.Backdrop
	nextScreen game.Screen
	prevScreen game.Screen

	panel     *ui.Panel
	codeLbl   *widget.Text
	connected bool
}

// NewWaitForFriendScreen creates the waiting screen and initiates a server connection.
func NewWaitForFriendScreen(provider *ws.ClientProvider, prevScreen *PlayWithFriendScreen) *WaitForFriendScreen {
	s := &WaitForFriendScreen{
		provider:   provider,
		bg:         prevScreen.bg,
		prevScreen: prevScreen,
	}

	provider.Get().Connect(uuid.New().String())

	panel := ui.NewPanel("Waiting for a friend…")
	s.panel = panel

	codeTF := ui.TextFace(40)
	codeRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(10),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	codeLbl := widget.NewText(
		widget.TextOpts.Text("", &codeTF, colornames.Gold),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
			widget.WidgetOpts.MinSize(200, 0),
		),
	)

	copyBtn := ui.SecondaryButton("Copy", 16, func(_ *widget.ButtonClickedEventArgs) {
		clipboard.Write(s.codeLbl.Label)
	})

	codeRow.AddChild(codeLbl)
	codeRow.AddChild(copyBtn)

	hintTF := ui.TextFace(16)
	hintLbl := widget.NewText(
		widget.TextOpts.Text("\nSend this code to your friend.\nThey can join via\nPlay with friend → Join game.", &hintTF, colornames.Whitesmoke),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	)

	panel.AddContent(codeRow, hintLbl)

	backBtn := ui.SecondaryButton("Cancel", 16, func(_ *widget.ButtonClickedEventArgs) {
		s.provider.Get().Send(ws.OutMessage{Action: ws.CancelFriendRoomAction})
		s.nextScreen = prevScreen
	})
	panel.AddControl(backBtn)

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	root.AddChild(panel.Build())

	s.ui = &ebitenui.UI{Container: root}
	s.codeLbl = codeLbl
	return s
}

// Update implements game.Screen.
func (s *WaitForFriendScreen) Update(_ *game.Game) (game.Screen, error) {
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

	server := s.provider.Get()

	if !s.connected && server.Status() == ws.StatusConnected {
		s.connected = true
		server.Send(ws.OutMessage{Action: ws.CreateFriendGameAction})
		s.panel.Title("Waiting for friend to join…")
	}
	if server.Status() == ws.StatusError {
		s.panel.Title(ConnErrorLabel)
	}

	for {
		select {
		case msg := <-server.Inbox():
			if next := s.handleMessage(msg); next != nil {
				return next, nil
			}
		default:
			return s, nil
		}
	}
}

// Draw implements game.Screen.
func (s *WaitForFriendScreen) Draw(screen *ebiten.Image) {
	s.bg.Draw(screen)
	s.ui.Draw(screen)
}

// handleMessage processes server responses on the wait screen.
func (s *WaitForFriendScreen) handleMessage(msg ws.InMessage) game.Screen {
	switch msg.Action {
	case ws.FriendRoomCreatedAction:
		var data ds.FriendRoomCreatedPayload
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			log.Printf("[wait_friend] unmarshal: %v", err)
			return nil
		}
		s.codeLbl.Label = data.JoinCode
		s.panel.Title("Waiting for a friend…")

	case ws.NewGameAction:
		var data ds.NewGamePayload
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			log.Fatalf("[wait_friend] unmarshal new game: %v", err)
		}
		snap := ds.GameSnapshot{
			ArenaID:                    data.ArenaID,
			Board:                      data.Board,
			Player:                     *data.Player,
			OpponentName:               data.Opponent,
			Round:                      1,
			TurnTimeSeconds:            data.TurnTimeSeconds,
			MaxPhantomAPPerUnitPerTurn: data.MaxPhantomAPPerUnitPerTurn,
		}
		return NewArenaScreen(snap, s.provider, false)

	case ws.ErrorAction:
		var e ds.ErrorResponse
		_ = json.Unmarshal(msg.Data, &e)
		s.panel.Title("Error: " + e.Message)
	}
	return nil
}

// JoinFriendScreen lets the player enter a join code to connect to a friend's game.
type JoinFriendScreen struct {
	provider   *ws.ClientProvider
	ui         *ebitenui.UI
	bg         backdrop.Backdrop
	nextScreen game.Screen
	prevScreen game.Screen
	statusLbl  *widget.Text
	codeInput  *widget.TextInput
	connected  bool
}

// NewJoinFriendScreen creates the join-by-code screen and initiates a server connection.
func NewJoinFriendScreen(provider *ws.ClientProvider, prevScreen *PlayWithFriendScreen) *JoinFriendScreen {
	s := &JoinFriendScreen{
		provider:   provider,
		bg:         prevScreen.bg,
		prevScreen: prevScreen,
	}

	provider.Get().Connect(uuid.New().String())

	panel := ui.NewPanel("Join friend's game")

	statusTF := ui.TextFace(15)
	statusLbl := widget.NewText(
		widget.TextOpts.Text("", &statusTF, colornames.Lightgray),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	inputRow := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(10),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	inputTF := ui.TextFace(24)
	codeInput := widget.NewTextInput(
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
			widget.WidgetOpts.MinSize(150, 0),
		),
		widget.TextInputOpts.Image(&widget.TextInputImage{
			Idle:     image.NewNineSliceColor(color.NRGBA{R: 60, G: 80, B: 100, A: 255}),
			Disabled: image.NewNineSliceColor(color.NRGBA{R: 60, G: 80, B: 100, A: 255}),
		}),
		widget.TextInputOpts.Face(&inputTF),
		widget.TextInputOpts.Color(&widget.TextInputColor{
			Idle:          colornames.Gold,
			Disabled:      colornames.Gray,
			Caret:         colornames.Gray,
			DisabledCaret: colornames.Gray,
		}),
		widget.TextInputOpts.Padding(widget.NewInsetsSimple(10)),
		widget.TextInputOpts.Placeholder("Enter Code"),
		widget.TextInputOpts.Validation(func(s string) (bool, *string) {
			upper := strings.ToUpper(s)
			if utf8.RuneCountInString(upper) > 6 { //nolint:mnd
				return false, nil
			}
			return false, &upper
		}),
		widget.TextInputOpts.SubmitHandler(func(_ *widget.TextInputChangedEventArgs) {
			s.sendJoin()
		}),
	)

	pasteBtn := ui.SecondaryButton("Paste", 16, func(_ *widget.ButtonClickedEventArgs) {
		s.statusLbl.Label = ""
		clipboard.Read(func(text string) {
			text = strings.ToUpper(strings.TrimSpace(text))
			if utf8.RuneCountInString(text) == 6 { //nolint:mnd
				s.codeInput.SetText(text)
			} else {
				s.statusLbl.Label = "Nothing valid in clipboard."
			}
		})
	})

	inputRow.AddChild(codeInput)
	inputRow.AddChild(pasteBtn)

	panel.AddContent(inputRow, statusLbl)

	joinBtn := ui.SecondaryButton("Join", 16, func(_ *widget.ButtonClickedEventArgs) {
		s.sendJoin()
	})
	backBtn := ui.SecondaryButton("Back", 16, func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = s.prevScreen
	})
	panel.AddControl(joinBtn, backBtn)

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	root.AddChild(panel.Build())

	s.ui = &ebitenui.UI{Container: root}
	s.statusLbl = statusLbl
	s.codeInput = codeInput
	return s
}

// Update implements game.Screen.
func (s *JoinFriendScreen) Update(_ *game.Game) (game.Screen, error) {
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

	server := s.provider.Get()

	if !s.connected && server.Status() == ws.StatusConnected {
		s.connected = true
		s.statusLbl.Label = "Enter the code your friend shared with you."
	}
	if server.Status() == ws.StatusError {
		s.statusLbl.Label = ConnErrorLabel
	}

	for {
		select {
		case msg := <-server.Inbox():
			if next := s.handleMessage(msg); next != nil {
				return next, nil
			}
		default:
			return s, nil
		}
	}
}

// Draw implements game.Screen.
func (s *JoinFriendScreen) Draw(screen *ebiten.Image) {
	s.bg.Draw(screen)
	s.ui.Draw(screen)
}

// sendJoin reads the current input value and sends a join request to the server.
func (s *JoinFriendScreen) sendJoin() {
	code := s.codeInput.GetText()
	if len(code) == 0 {
		s.statusLbl.Label = "Please enter a code."
		return
	}
	s.provider.Get().Send(ws.OutMessage{
		Action: ws.JoinFriendGameAction,
		Data:   map[string]string{"join_code": code},
	})
	s.statusLbl.Label = "Joining…"
}

// handleMessage processes server responses on the join screen.
func (s *JoinFriendScreen) handleMessage(msg ws.InMessage) game.Screen {
	switch msg.Action {
	case ws.NewGameAction:
		var data ds.NewGamePayload
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			log.Fatalf("[join_friend] unmarshal new game: %v", err)
		}
		snap := ds.GameSnapshot{
			ArenaID:                    data.ArenaID,
			Board:                      data.Board,
			Player:                     *data.Player,
			OpponentName:               data.Opponent,
			Round:                      1,
			TurnTimeSeconds:            data.TurnTimeSeconds,
			MaxPhantomAPPerUnitPerTurn: data.MaxPhantomAPPerUnitPerTurn,
		}
		return NewArenaScreen(snap, s.provider, false)

	case ws.FriendRoomNotFoundAction:
		s.statusLbl.Label = "Code not found. Check the code and try again."

	case ws.FriendRoomExpiredAction:
		s.statusLbl.Label = "This code has expired. Ask your friend to create a new game."

	case ws.ErrorAction:
		var e ds.ErrorResponse
		_ = json.Unmarshal(msg.Data, &e)
		s.statusLbl.Label = "Error: " + e.Message
	}
	return nil
}
