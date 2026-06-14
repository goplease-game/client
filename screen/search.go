package screen

import (
	"encoding/json"
	"fmt"
	stdImage "image"
	"image/color"
	"log"
	"math/rand"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	game "github.com/ognev-dev/goplease-ebitengine-client"
	"github.com/ognev-dev/goplease-ebitengine-client/asset"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
	"golang.org/x/image/colornames"
)

const (
	ConnectingLabel   = "Connecting..."
	SearchingOppLabel = "Searching for opponent…"
	ConnErrorLabel    = "Connection error. Press Esc to go back."

	unitCount    = 6
	fadeDuration = 30
	travelTicks  = 90
	spawnAt      = travelTicks / 2
)

var searchOppUnitColors = []color.Color{
	ui.RGBFromHex("B5BAFF"),
	ui.RGBFromHex("AEE2FF"),
	ui.RGBFromHex("7AE2CF"),
	ui.RGBFromHex("FF6060"),
	ui.RGBFromHex("7288AE"),
	ui.RGBFromHex("4BB8FA"),
	ui.RGBFromHex("9FCBAD"),
	ui.RGBFromHex("FFC94D"),
	ui.RGBFromHex("A2CB8B"),
	ui.RGBFromHex("C0E1D2"),
	ui.RGBFromHex("F0FFC2"),
	ui.RGBFromHex("F08D39"),
	ui.RGBFromHex("FAACBF"),
	ui.RGBFromHex("76D2DB"),
}

// unitAnim holds the state of a single unit travelling across the placeholder container.
type unitAnim struct {
	img         *ebiten.Image
	tick        int
	alpha       float32
	x           float64
	startX      float64
	endX        float64
	centerY     float64
	spawnedNext bool
}

// init loads a random tinted unit image and resets all movement state for a new pass.
func (a *unitAnim) init(containerW, containerH float64) {
	idx := rand.Intn(unitCount) + 1
	col := searchOppUnitColors[rand.Intn(len(searchOppUnitColors))]
	img := asset.TintedImage(fmt.Sprintf("units/unit_%d_pic.png", idx), col, 64)
	a.img = ebiten.NewImageFromImage(img)
	iw := float64(a.img.Bounds().Dx())
	ih := float64(a.img.Bounds().Dy())

	a.tick = 0
	a.alpha = 1
	a.startX = -iw
	a.endX = containerW
	a.x = a.startX
	a.centerY = (containerH - ih) / 2
	a.spawnedNext = false
}

// update advances the unit by one tick. It returns true when the unit has completed its path.
func (a *unitAnim) update() bool {
	if a.img == nil {
		return true
	}
	a.tick++

	progress := float64(a.tick) / float64(travelTicks)
	a.x = a.startX + (a.endX-a.startX)*progress

	fadeStart := travelTicks - fadeDuration
	if a.tick >= fadeStart {
		a.alpha = 1 - float32(a.tick-fadeStart)/float32(fadeDuration)
	}

	return a.tick >= travelTicks
}

// draw renders the unit onto screen offset by containerRect's origin.
func (a *unitAnim) draw(screen *ebiten.Image, containerRect stdImage.Rectangle) {
	if a.img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(a.alpha)
	op.GeoM.Translate(
		float64(containerRect.Min.X)+a.x,
		float64(containerRect.Min.Y)+a.centerY,
	)
	screen.DrawImage(a.img, op)
}

// SearchScreen is the matchmaking screen shown while the client waits for an opponent.
type SearchScreen struct {
	server          ws.Client
	ui              *ebitenui.UI
	statusLbl       *widget.Text
	elapsedLbl      *widget.Text
	animPlaceholder *widget.Container
	tick            int
	units           []*unitAnim
	animReady       bool
}

// NewSearchScreen creates a SearchScreen and initiates a server connection.
func NewSearchScreen(server ws.Client) *SearchScreen {
	s := &SearchScreen{
		server: server,
	}

	s.server.Connect(uuid.New().String())

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	center := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(12),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(300, 0),
		),
	)

	animPlaceholder := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(300, 128),
		),
	)

	statusTF := ui.TextFace(25)
	statusLbl := widget.NewText(
		widget.TextOpts.Text(ConnectingLabel, &statusTF, colornames.Lightgray),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	elapsedTF := ui.TextFace(15)
	elapsedLbl := widget.NewText(
		widget.TextOpts.Text("elapsed: 0s", &elapsedTF, colornames.Gray),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	center.AddChild(animPlaceholder)
	center.AddChild(statusLbl)
	center.AddChild(elapsedLbl)

	footer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
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

	hintTF := ui.TextFace(15)
	hintLbl := widget.NewText(
		widget.TextOpts.Text("[Esc] cancel", &hintTF, colornames.Dimgray),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)
	footer.AddChild(hintLbl)

	root.AddChild(center)
	root.AddChild(footer)

	s.ui = &ebitenui.UI{Container: root}
	s.statusLbl = statusLbl
	s.elapsedLbl = elapsedLbl
	s.animPlaceholder = animPlaceholder
	return s
}

// spawnUnit initialises a new unitAnim from the current placeholder bounds and appends it to the active list.
func (s *SearchScreen) spawnUnit() {
	rect := s.animPlaceholder.GetWidget().Rect
	a := &unitAnim{}
	a.init(float64(rect.Dx()), float64(rect.Dy()))
	s.units = append(s.units, a)
}

// updateUnits advances all active units, spawns the next unit when the leading one reaches
// the midpoint, and removes units that have completed their path.
func (s *SearchScreen) updateUnits() {
	var spawned []*unitAnim

	alive := s.units[:0]
	for _, a := range s.units {
		done := a.update()

		if !a.spawnedNext && a.tick >= spawnAt {
			a.spawnedNext = true
			rect := s.animPlaceholder.GetWidget().Rect
			next := &unitAnim{}
			next.init(float64(rect.Dx()), float64(rect.Dy()))
			spawned = append(spawned, next)
		}

		if !done {
			alive = append(alive, a)
		}
	}

	s.units = append(alive, spawned...)
}

// Update implements game.Screen. It drives the matchmaking loop, unit animation, and input handling.
func (s *SearchScreen) Update(g *game.Game) (game.Screen, error) {
	s.tick++
	s.ui.Update()

	if !s.animReady {
		rect := s.animPlaceholder.GetWidget().Rect
		if rect.Dx() > 0 {
			s.animReady = true
			s.spawnUnit()
		}
	} else {
		s.updateUnits()
	}

	s.elapsedLbl.Label = fmt.Sprintf("elapsed: %ds", s.tick/60)

	if g.Server.Status() == ws.StatusConnected && s.statusLbl.Label == ConnectingLabel {
		g.Server.Send(ws.OutMessage{
			Action: ws.NewGameAction,
		})
		s.statusLbl.Label = SearchingOppLabel
	}
	if g.Server.Status() == ws.StatusError {
		s.statusLbl.Label = ConnErrorLabel
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.Server.Status() == ws.StatusConnected {
			g.Server.Send(ws.OutMessage{
				Action: "cancel_match",
				Data:   nil,
			})
		}
		return NewMainScreen(g.Server), nil
	}

	for {
		select {
		case msg := <-g.Server.Inbox():
			if next := s.handleMessage(msg); next != nil {
				return next, nil
			}
		default:
			return s, nil
		}
	}
}

// handleMessage dispatches an incoming server message and returns the next screen if a transition is required.
func (s *SearchScreen) handleMessage(msg ws.InMessage) game.Screen {
	fmt.Printf("[search] received: %v\n", msg.Action)
	if msg.Data != nil {
		fmt.Printf("JSON: %s\n", string(msg.Data))
	}

	switch msg.Action {
	case ws.SearchingOppAction:
		s.statusLbl.Label = SearchingOppLabel
	case ws.NewGameAction:
		var data ds.NewGamePayload
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Fatalf("new game: failed to unmarshal: %v", err)
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
		return NewArenaScreen(snap, s.server, false)
	case ws.ErrorAction:
		var e ds.ErrorResponse
		_ = json.Unmarshal(msg.Data, &e)
		s.statusLbl.Label = "Error: " + e.Message
	}
	return nil
}

// Draw implements game.Screen. It renders the UI and all active unit animations.
func (s *SearchScreen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
	if s.animReady {
		rect := s.animPlaceholder.GetWidget().Rect
		for _, a := range s.units {
			a.draw(screen, rect)
		}
	}
}
