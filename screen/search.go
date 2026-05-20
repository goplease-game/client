package screen

import (
	"bytes"
	"encoding/json"
	"fmt"
	stdImage "image"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	game "github.com/ognev-dev/goplease-ebitengine-client"
	"github.com/ognev-dev/goplease-ebitengine-client/asset"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/screen/arena"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
	"github.com/setanarut/anim"
	"golang.org/x/image/colornames"
)

var animPlayer *anim.AnimationPlayer

const (
	ConnectingLabel   = "Connecting..."
	SearchingOppLabel = "Searching for opponent…"
	ConnErrorLabel    = "Connection error. Press Esc to go back."
)

type SearchScreen struct {
	server          ws.Client
	ui              *ebitenui.UI
	statusLbl       *widget.Text
	elapsedLbl      *widget.Text
	animPlaceholder *widget.Container
	tick            int
}

func NewSearchScreen(server ws.Client) *SearchScreen {
	s := &SearchScreen{
		server: server,
	}

	runner := asset.Load("runner.png")

	img, _, err := stdImage.Decode(bytes.NewReader(runner))
	if err != nil {
		log.Fatal(err)
	}
	spriteSheet := anim.Atlas{
		Name:  "Default",
		Image: ebiten.NewImageFromImage(img),
	}
	animPlayer = anim.NewAnimationPlayer(spriteSheet)
	animPlayer.NewAnim("run", 0, 0, 128, 128, 8, false, false, 12)

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	// Central column: animation placeholder + status + elapsed
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

	// placeholder for a runner animation
	animPlaceholder := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(128, 128),
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

func (s *SearchScreen) Update(g *game.Game) (game.Screen, error) {
	s.tick++
	s.ui.Update()
	animPlayer.Update()

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

func (s *SearchScreen) handleMessage(msg ws.InMessage) game.Screen {
	switch msg.Action {
	case ws.SearchingOppAction:
		s.statusLbl.Label = SearchingOppLabel
	case ws.NewGameAction:
		var data ds.NewGamePayload
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Fatalf("new game: failed to unmarshal: %v", err)
		}
		snap := ds.GameSnapshot{
			RoomID:       data.RoomID,
			Board:        data.Board,
			Player:       *data.Player,
			OpponentName: data.Opponent,
			Round:        1,
		}
		return arena.NewScreen(snap, s.server)
	case ws.ErrorAction:
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(msg.Data, &e)
		s.statusLbl.Label = "Error: " + e.Message
	}
	return nil
}

func (s *SearchScreen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)

	frame := animPlayer.CurrentFrame
	rect := s.animPlaceholder.GetWidget().Rect
	fw := frame.Bounds().Dx()
	fh := frame.Bounds().Dy()

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		float64(rect.Min.X+rect.Dx()/2-fw/2),
		float64(rect.Min.Y+rect.Dy()/2-fh/2),
	)

	screen.DrawImage(frame, op)
}
