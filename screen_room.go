package game

import (
	"encoding/json"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
	"golang.org/x/image/colornames"
)

// ── Layout constants ──────────────────────────────────────────────────────────

const (
	cellSize = 50 // pixels per board cell
)

// ── RoomScreen ────────────────────────────────────────────────────────────────

type RoomScreen struct {
	ui *ebitenui.UI

	roomID       string
	phase        ds.Phase
	isMyTurn     bool
	board        ds.Board
	myPlayer     ds.Player
	opponentName string

	// Interaction
	selectedBoardRow int // board cell selected (unit on board), -1 = none
	selectedBoardCol int
	hoveredRow       int
	hoveredCol       int

	// Status
	statusLine string
}

func NewRoomScreen(newGamePayload json.RawMessage) *RoomScreen {
	var data ds.NewGamePayload
	if err := json.Unmarshal(newGamePayload, &data); err != nil {
		log.Fatalf("new game payload: %v", err)
	}

	s := &RoomScreen{
		roomID:           data.RoomID,
		phase:            data.Phase,
		isMyTurn:         data.IsMyTurn,
		board:            data.Board,
		opponentName:     data.Opponent,
		selectedBoardRow: -1,
		selectedBoardCol: -1,
		hoveredRow:       -1,
		hoveredCol:       -1,
		statusLine:       "Place a unit",
	}
	if data.Player != nil {
		s.myPlayer = *data.Player
	}

	s.drawUI(data)
	return s
}

// ── Update ────────────────────────────────────────────────────────────────────

func (s *RoomScreen) Update(g *Game) (Screen, error) {
	// Drain server inbox
	for {
		select {
		case msg := <-g.Server.Inbox:
			s.handleMessage(g, msg)
		default:
			goto doneInbox
		}
	}
doneInbox:

	return s, nil
}

func (s *RoomScreen) handleMessage(g *Game, msg ws.Message) {
	switch msg.Action {
	case "game_over":
		// TODO
		s.statusLine = "Game over"

	case "error":
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(msg.Data, &e)
		s.statusLine = "Error: " + e.Message
	}
}

// ── Draw ──────────────────────────────────────────────────────────────────────

func (s *RoomScreen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

func (s *RoomScreen) drawUI(data ds.NewGamePayload) {
	textFace, err := ui.TextFace(40)
	if err != nil {
		log.Fatal(err)
	}

	newText := func(content string) *widget.Text {
		return widget.NewText(
			widget.TextOpts.Text(content, &textFace, colornames.White),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		)
	}
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	const headerH = 80
	const footerH = 80

	header := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Steelblue)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				StretchHorizontal:  true,
			}),
			widget.WidgetOpts.MinSize(0, headerH),
		),
	)

	footer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Steelblue)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				StretchHorizontal:  true,
			}),
			widget.WidgetOpts.MinSize(0, footerH),
		),
	)

	center := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchHorizontal:  true,
				StretchVertical:    true,
			}),
			widget.WidgetOpts.MinSize(0, 0),
		),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchHorizontal:  true,
				StretchVertical:    true,
				Padding: &widget.Insets{
					Top:    headerH,
					Bottom: footerH,
				},
			}),
		),
	)

	// --------------------------------------------------------------
	// BOARD --------------------------------------------------------
	// --------------------------------------------------------------

	cellColor := colornames.Darkgray
	boardColor := colornames.Slategray

	board := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(boardColor),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(len(data.Board[0])),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(25)),
			widget.GridLayoutOpts.Spacing(1, 1),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	for _, row := range data.Board {
		for range row {
			cell := widget.NewContainer(
				widget.ContainerOpts.BackgroundImage(
					image.NewNineSliceColor(cellColor),
				),
				widget.ContainerOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.GridLayoutData{}),
					widget.WidgetOpts.MinSize(64, 64),
				),
				widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			)
			board.AddChild(cell)
		}
	}

	center.AddChild(board)

	// --------------------------------------------------------------
	// UNITS --------------------------------------------------------
	// --------------------------------------------------------------

	unitCellColor := colornames.Darkgray
	unitPanelColor := colornames.Slategray

	unitPanel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewNineSliceColor(unitPanelColor),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(len(data.Player.Units)),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(5)),
			widget.GridLayoutOpts.Spacing(1, 1),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	for _, unit := range data.Player.Units {
		cell := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(
				image.NewNineSliceColor(unitCellColor),
			),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{}),
				widget.WidgetOpts.MinSize(64, 64),
			),
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)

		cell.AddChild(newText(unit.Name))
		unitPanel.AddChild(cell)
	}

	center.AddChild(board)

	header.AddChild(newText("im header"))
	footer.AddChild(unitPanel)

	root.AddChild(center)
	root.AddChild(header)
	root.AddChild(footer)

	s.ui = &ebitenui.UI{Container: root}
}
