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
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
	"golang.org/x/image/colornames"
)

// ── Layout constants ──────────────────────────────────────────────────────────

const (
	cellSize = 50
)

// ── Colors ──────────────────────────────────────────────────────────

var (
	boardCellColor     = colornames.Darkgray
	dropZoneColor      = colornames.Limegreen
	dropZoneHoverColor = colornames.Palegreen
)

// ── dndUnit ───────────────────────────────────────────────────────────────────

type dndUnit struct {
	unit         ds.Unit
	dragWidget   *widget.Container
	sourceWidget *widget.Container
	current      widget.HasWidget
}

func (d *dndUnit) Create(_ widget.HasWidget) (*widget.Container, interface{}) {
	if d.dragWidget == nil {
		unitImg := unitImage(d.unit.TemplateID)
		d.dragWidget = widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.BackgroundImage(
				image.NewNineSliceColor(colornames.Ghostwhite),
			),
		)
		d.dragWidget.AddChild(widget.NewGraphic(
			widget.GraphicOpts.Image(unitImg),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		))
	}

	return d.dragWidget, d.unit
}

func (d *dndUnit) Update(canDrop bool, targetWidget widget.HasWidget, _ interface{}) {
	if d.current != nil && d.current != targetWidget {
		d.current.(*widget.Container).SetBackgroundImage(
			image.NewNineSliceColor(dropZoneColor),
		)

		d.current = nil
	}
	if canDrop && targetWidget != nil {
		targetWidget.(*widget.Container).SetBackgroundImage(
			image.NewNineSliceColor(dropZoneHoverColor),
		)
		d.current = targetWidget
	}
}

func (d *dndUnit) EndDrag(_ bool, _ widget.HasWidget, _ interface{}) {
	if d.current != nil {
		d.current.(*widget.Container).SetBackgroundImage(
			image.NewNineSliceColor(dropZoneColor),
		)

		d.current = nil
	}
}

// ── safeZoneCell ──────────────────────────────────────────────────────────────

type safeZoneCell struct {
	container     *widget.Container
	activeGraphic *widget.Graphic
	occupied      *bool
}

// ── dndUnitWithGlobalHighlight ────────────────────────────────────────────────
type dndUnitWithGlobalHighlight struct {
	*dndUnit
	safeZoneCells []*safeZoneCell
	dragActive    bool
}

func (d *dndUnitWithGlobalHighlight) Create(parent widget.HasWidget) (*widget.Container, interface{}) {
	if !d.dragActive {
		d.dragActive = true
		dropImg := ImageAsset("drop_point.png", ImageSize{W: 52, H: 52})

		for _, sc := range d.safeZoneCells {
			if !*sc.occupied {
				sc.container.SetBackgroundImage(image.NewNineSliceColor(dropZoneColor))

				g := widget.NewGraphic(
					widget.GraphicOpts.Image(dropImg),
					widget.GraphicOpts.WidgetOpts(
						widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
							HorizontalPosition: widget.AnchorLayoutPositionCenter,
							VerticalPosition:   widget.AnchorLayoutPositionCenter,
						}),
					),
				)
				sc.container.AddChild(g)
				sc.activeGraphic = g
			}
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
	for _, sc := range d.safeZoneCells {
		if !*sc.occupied {
			sc.container.SetBackgroundImage(image.NewNineSliceColor(boardCellColor))
			if sc.activeGraphic != nil {
				sc.container.RemoveChild(sc.activeGraphic)
				sc.activeGraphic = nil
			}
		}
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

	unitsQueue []string // unit.ID

	unitCards        map[string]*widget.Container
	unitPanelRef     *widget.Container
	queuePanelRef    *widget.Container
	headerRef        *widget.Container
	footerRef        *widget.Container
	queuePanelIn     bool
	unitPanelIn      bool
	boardCellWidgets [][]*widget.Container // [row][col]
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
		unitsQueue:       []string{},
		unitCards:        make(map[string]*widget.Container),
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

	case "unit_queued":
		var payload struct {
			UnitID string `json:"unit_id"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err == nil {
			s.addUnitToQueue(payload.UnitID)
		}

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

func (s *RoomScreen) unitByID(id string) (ds.Unit, bool) {
	for _, u := range s.myPlayer.Units {
		if u.ID == id {
			return u, true
		}
	}
	return ds.Unit{}, false
}

func (s *RoomScreen) boardCellWidget(u ds.Unit) *widget.Container {
	if u.Row < 0 || u.Col < 0 {
		return nil
	}
	if u.Row >= len(s.boardCellWidgets) {
		return nil
	}
	if u.Col >= len(s.boardCellWidgets[u.Row]) {
		return nil
	}
	return s.boardCellWidgets[u.Row][u.Col]
}

func (s *RoomScreen) addUnitToQueue(unitID string) {
	for _, id := range s.unitsQueue {
		if id == unitID {
			return
		}
	}
	s.unitsQueue = append(s.unitsQueue, unitID)
	s.rebuildQueuePanel()
}

func (s *RoomScreen) rebuildQueuePanel() {
	if s.queuePanelRef == nil || s.headerRef == nil {
		return
	}

	s.queuePanelRef.RemoveChildren()

	if len(s.unitsQueue) == 0 {
		if s.queuePanelIn {
			s.headerRef.RemoveChild(s.queuePanelRef)
			s.queuePanelIn = false
		}
		return
	}

	if !s.queuePanelIn {
		s.headerRef.AddChild(s.queuePanelRef)
		s.queuePanelIn = true
	}

	for i := len(s.unitsQueue) - 1; i >= 0; i-- {
		u, ok := s.unitByID(s.unitsQueue[i])
		if !ok {
			continue
		}

		unitID := u.ID
		queueCard := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(boardCellColor)),
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.MinSize(54, 54),
			),
		)

		queueCard.AddChild(widget.NewGraphic(
			widget.GraphicOpts.Image(unitImage(u.TemplateID)),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
				widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
					queueCard.SetBackgroundImage(image.NewNineSliceColor(colornames.Gold))
					if u, ok := s.unitByID(unitID); ok {
						if bc := s.boardCellWidget(u); bc != nil {
							bc.SetBackgroundImage(image.NewNineSliceColor(colornames.Gold))
						}
					}
				}),
				widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
					queueCard.SetBackgroundImage(image.NewNineSliceColor(boardCellColor))
					if u, ok := s.unitByID(unitID); ok {
						if bc := s.boardCellWidget(u); bc != nil {
							bc.SetBackgroundImage(image.NewNineSliceColor(boardCellColor))
						}
					}
				}),
			),
		))

		s.queuePanelRef.AddChild(queueCard)
	}
}

// ── drawUI ────────────────────────────────────────────────────────────────────

func (s *RoomScreen) drawUI(data ds.NewGamePayload) {
	// ── root ──────────────────────────────────────────────────────────────
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	const headerH, footerH = 80, 90

	// ── header ────────────────────────────────────────────────────────────
	header := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Steelblue)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
				StretchHorizontal:  true,
			}),
			widget.WidgetOpts.MinSize(0, headerH),
		),
	)

	queuePanel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Midnightblue)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(4)),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	s.queuePanelRef = queuePanel
	s.headerRef = header

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
	var safeZoneCells []*safeZoneCell

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

	s.boardCellWidgets = make([][]*widget.Container, len(data.Board))
	for i := range s.boardCellWidgets {
		if len(data.Board[i]) > 0 {
			s.boardCellWidgets[i] = make([]*widget.Container, len(data.Board[i]))
		}
	}

	for rowIdx, row := range data.Board {
		for colIdx, cellData := range row {
			isDroppable := cellData != nil && cellData.IsSafeZone && cellData.Unit == nil

			var cell *widget.Container
			if isDroppable {
				var c *widget.Container
				occupied := false
				dropRow, dropCol := rowIdx, colIdx

				sc := &safeZoneCell{
					container: c,
					occupied:  &occupied,
				}

				c = widget.NewContainer(
					widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(boardCellColor)),
					widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
					widget.ContainerOpts.WidgetOpts(
						widget.WidgetOpts.LayoutData(widget.GridLayoutData{}),
						widget.WidgetOpts.MinSize(64, 64),
						widget.WidgetOpts.CanDrop(func(args *widget.DragAndDropDroppedEventArgs) bool {
							_, ok := args.Data.(ds.Unit)
							return ok && !occupied
						}),
						widget.WidgetOpts.Dropped(func(args *widget.DragAndDropDroppedEventArgs) {
							droppedUnit := args.Data.(ds.Unit)
							occupied = true

							// TODO Server
							// g.Server.Send(ws.Message{Action: "place_unit", Data: ...})

							if sc.activeGraphic != nil {
								sc.container.RemoveChild(sc.activeGraphic)
								sc.activeGraphic = nil
							}

							c.SetBackgroundImage(image.NewNineSliceColor(boardCellColor))

							unitImg := unitImage(droppedUnit.TemplateID)
							c.AddChild(widget.NewGraphic(
								widget.GraphicOpts.Image(unitImg),
								widget.GraphicOpts.WidgetOpts(
									widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
										HorizontalPosition: widget.AnchorLayoutPositionCenter,
										VerticalPosition:   widget.AnchorLayoutPositionCenter,
									}),
								),
							))

							for i, u := range s.myPlayer.Units {
								if u.ID == droppedUnit.ID {
									s.myPlayer.Units[i].Row = dropRow
									s.myPlayer.Units[i].Col = dropCol
									break
								}
							}

							s.onUnitPlaced(droppedUnit.ID)
						}),
					),
				)

				sc.container = c
				safeZoneCells = append(safeZoneCells, sc)
				s.boardCellWidgets[rowIdx][colIdx] = c
				cell = c
			} else {
				cell = widget.NewContainer(
					widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(boardCellColor)),
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

	s.unitPanelRef = unitPanel
	s.footerRef = footer

	for _, unit := range data.Player.Units {
		u := unit
		unitCard := s.buildUnitCard(u, safeZoneCells)
		unitPanel.AddChild(unitCard)
	}

	if len(data.Player.Units) > 0 {
		footer.AddChild(unitPanel)
		s.unitPanelIn = true
	}

	// ── assemble ──────────────────────────────────────────────────────────
	root.AddChild(center)
	root.AddChild(header)
	root.AddChild(footer)

	s.ui = &ebitenui.UI{Container: root}
}

func (s *RoomScreen) buildUnitCard(u ds.Unit, safeZoneCells []*safeZoneCell) *widget.Container {
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
					widget.DragAndDropOpts.ContentsOriginVertical(widget.DND_ANCHOR_START),
					widget.DragAndDropOpts.ContentsOriginHorizontal(widget.DND_ANCHOR_START),
					widget.DragAndDropOpts.Offset(img.Point{X: -10, Y: -10}),
				),
			),
		),
	)

	dnd.dndUnit.sourceWidget = unitCard

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

	s.unitCards[u.ID] = unitCard

	return unitCard
}

func (s *RoomScreen) onUnitPlaced(unitID string) {
	if card, ok := s.unitCards[unitID]; ok {
		s.unitPanelRef.RemoveChild(card)
		delete(s.unitCards, unitID)
	}

	if len(s.unitCards) == 0 && s.unitPanelRef != nil && s.footerRef != nil && s.unitPanelIn {
		s.footerRef.RemoveChild(s.unitPanelRef)
		s.unitPanelIn = false
	}

	s.addUnitToQueue(unitID)
}

func unitImage(templateID int) *ebiten.Image {
	up := path.Join("units", fmt.Sprintf("unit_%d_pic.png", templateID))

	return ImageAsset(up, ImageSize{
		W: 64, H: 64,
	})
}

func imageToNineSlice(img *ebiten.Image) *image.NineSlice {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	return image.NewNineSlice(img,
		[3]int{0, w, 0},
		[3]int{0, h, 0},
	)
}
