package game

import (
	"encoding/json"
	"fmt"
	img "image"
	"image/color"
	"log"
	"path"

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
	cellSize = 50
)

// ── Highlight colors ──────────────────────────────────────────────────────────

var (
	safeZoneIdle  = color.NRGBA{60, 130, 60, 255}
	safeZoneHover = color.NRGBA{120, 220, 120, 255}
)

// ── dndUnit ───────────────────────────────────────────────────────────────────
type dndUnit struct {
	unit    ds.Unit
	dndObj  *widget.Container
	text    *widget.Text
	current widget.HasWidget
}

func (d *dndUnit) Create(parent widget.HasWidget) (*widget.Container, interface{}) {
	if d.dndObj == nil {
		face, _ := ui.TextFace(30)
		d.dndObj = widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.BackgroundImage(
				image.NewNineSliceColor(color.NRGBA{80, 180, 255, 220}),
			),
		)
		d.text = widget.NewText(
			widget.TextOpts.Text(d.unit.Name, &face, color.Black),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			})),
		)

		d.dndObj.AddChild(d.text)
	}
	return d.dndObj, d.unit
}

func (d *dndUnit) Update(canDrop bool, targetWidget widget.HasWidget, _ interface{}) {
	if d.current != nil && d.current != targetWidget {
		d.current.(*widget.Container).SetBackgroundImage(
			image.NewNineSliceColor(safeZoneIdle),
		)
		d.current = nil
	}
	if canDrop && targetWidget != nil {
		targetWidget.(*widget.Container).SetBackgroundImage(
			image.NewNineSliceColor(safeZoneHover),
		)
		d.current = targetWidget
	}
}

func (d *dndUnit) EndDrag(_ bool, _ widget.HasWidget, _ interface{}) {
	if d.current != nil {
		d.current.(*widget.Container).SetBackgroundImage(
			image.NewNineSliceColor(safeZoneIdle),
		)
		d.current = nil
	}
}

// ── dndUnitWithGlobalHighlight ────────────────────────────────────────────────
type dndUnitWithGlobalHighlight struct {
	*dndUnit
	safeZoneCells []*widget.Container
	dragActive    bool
}

func (d *dndUnitWithGlobalHighlight) Create(parent widget.HasWidget) (*widget.Container, interface{}) {
	if !d.dragActive {
		d.dragActive = true
		for _, c := range d.safeZoneCells {
			c.SetBackgroundImage(image.NewNineSliceColor(safeZoneIdle))
		}
	}
	return d.dndUnit.Create(parent)
}

func (d *dndUnitWithGlobalHighlight) Update(canDrop bool, targetWidget widget.HasWidget, data interface{}) {
	d.dndUnit.Update(canDrop, targetWidget, data)
}

func (d *dndUnitWithGlobalHighlight) EndDrag(dropped bool, sourceWidget widget.HasWidget, data interface{}) {
	d.dragActive = false
	d.dndUnit.EndDrag(dropped, sourceWidget, data)
	for _, c := range d.safeZoneCells {
		c.SetBackgroundImage(image.NewNineSliceColor(color.NRGBA{80, 80, 80, 255}))
	}
}

// ── RoomScreen ────────────────────────────────────────────────────────────────

type RoomScreen struct {
	ui *ebitenui.UI

	roomID       string
	phase        ds.Phase
	isMyTurn     bool
	board        ds.Board
	myPlayer     ds.Player
	opponentName string

	selectedBoardRow int
	selectedBoardCol int
	hoveredRow       int
	hoveredCol       int

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
	for {
		select {
		case msg := <-g.Server.Inbox:
			s.handleMessage(g, msg)
		default:
			goto doneInbox
		}
	}
doneInbox:

	s.ui.Update()
	return s, nil
}

func (s *RoomScreen) handleMessage(g *Game, msg ws.Message) {
	switch msg.Action {
	case "game_over":
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

// ── drawUI ────────────────────────────────────────────────────────────────────

func (s *RoomScreen) drawUI(data ds.NewGamePayload) {
	face, err := ui.TextFace(28)
	if err != nil {
		log.Fatal(err)
	}

	newText := func(content string) *widget.Text {
		return widget.NewText(
			widget.TextOpts.Text(content, &face, colornames.White),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			})),
		)
	}

	// ── root ──────────────────────────────────────────────────────────────
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	const headerH, footerH = 80, 90

	// ── header ────────────────────────────────────────────────────────────
	header := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Steelblue)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
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
	header.AddChild(newText("GoPlease"))

	// ── footer ────────────────────────────────────────────────────────────
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

	// ── center ────────────────────────────────────────────────────────────
	center := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
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

	// ── board ─────────────────────────────────────────────────────────────
	var safeZoneCells []*widget.Container

	cols := 0
	if len(data.Board) > 0 {
		cols = len(data.Board[0])
	}

	board := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Slategray)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(cols),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(25)),
			widget.GridLayoutOpts.Spacing(2, 2),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	normalCellBg := colornames.Darkgray

	for _, row := range data.Board {
		for _, cellData := range row {
			isDroppable := cellData != nil && cellData.IsSafeZone && cellData.Unit == nil

			var cell *widget.Container
			if isDroppable {
				var c *widget.Container
				c = widget.NewContainer(
					widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(normalCellBg)),
					widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
					widget.ContainerOpts.WidgetOpts(
						widget.WidgetOpts.LayoutData(widget.GridLayoutData{}),
						widget.WidgetOpts.MinSize(64, 64),
						widget.WidgetOpts.CanDrop(func(args *widget.DragAndDropDroppedEventArgs) bool {
							_, ok := args.Data.(ds.Unit)
							return ok
						}),
						widget.WidgetOpts.Dropped(func(args *widget.DragAndDropDroppedEventArgs) {
							droppedUnit := args.Data.(ds.Unit)

							// TODO Server
							// g.Server.Send(ws.Message{Action: "place_unit", Data: ...})

							c.SetBackgroundImage(image.NewNineSliceColor(normalCellBg))

							f, _ := ui.TextFace(18)
							c.AddChild(widget.NewText(
								widget.TextOpts.Text(droppedUnit.Name, &f, colornames.White),
								widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
									HorizontalPosition: widget.AnchorLayoutPositionCenter,
									VerticalPosition:   widget.AnchorLayoutPositionCenter,
								})),
							))
						}),
					),
				)
				cell = c
				safeZoneCells = append(safeZoneCells, cell)
			} else {
				cell = widget.NewContainer(
					widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(normalCellBg)),
					widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
					widget.ContainerOpts.WidgetOpts(
						widget.WidgetOpts.LayoutData(widget.GridLayoutData{}),
						widget.WidgetOpts.MinSize(64, 64),
					),
				)
			}

			board.AddChild(cell)
		}
	}

	center.AddChild(board)

	// ── unit panel ────────────────────────────────────────────────────────
	unitCols := len(data.Player.Units)
	if unitCols == 0 {
		unitCols = 1
	}

	unitPanel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Slategray)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(unitCols),
			widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(5)),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	for _, unit := range data.Player.Units {
		u := unit

		dnd := &dndUnitWithGlobalHighlight{
			dndUnit:       &dndUnit{unit: u},
			safeZoneCells: safeZoneCells,
		}

		unitCard := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Ghostwhite)),
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{}),
				widget.WidgetOpts.MinSize(64, 64),
				widget.WidgetOpts.EnableDragAndDrop(
					widget.NewDragAndDrop(
						widget.DragAndDropOpts.ContentsCreater(dnd),
						widget.DragAndDropOpts.MinDragStartDistance(10),
						widget.DragAndDropOpts.ContentsOriginVertical(widget.DND_ANCHOR_END),
						widget.DragAndDropOpts.ContentsOriginHorizontal(widget.DND_ANCHOR_END),
						widget.DragAndDropOpts.Offset(img.Point{X: -5, Y: -5}),
					),
				),
			),
		)

		unitImg := unitImage(u.TemplateID)

		unitCard.AddChild(widget.NewGraphic(
			widget.GraphicOpts.Image(unitImg),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
				widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
					unitCard.SetBackgroundImage(image.NewNineSliceColor(colornames.Gold))
				}),
				widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
					unitCard.SetBackgroundImage(image.NewNineSliceColor(colornames.Ghostwhite))
				}),
			),
		))

		unitPanel.AddChild(unitCard)
	}

	footer.AddChild(unitPanel)

	// ── assemble ──────────────────────────────────────────────────────────
	root.AddChild(center)
	root.AddChild(header)
	root.AddChild(footer)

	s.ui = &ebitenui.UI{Container: root}
}

func unitImage(templateID int) *ebiten.Image {
	up := path.Join("units", fmt.Sprintf("unit_%d_pic.png", templateID))

	return ImageAsset(up, ImageSize{
		W: 64, H: 64,
	})
}

// 149 kb
