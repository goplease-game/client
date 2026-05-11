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
	activeGraphic *widget.Graphic // nil когда драга нет
	occupied      *bool
}

// ── dndUnitWithGlobalHighlight ────────────────────────────────────────────────
type dndUnitWithGlobalHighlight struct {
	*dndUnit
	safeZoneCells []*safeZoneCell // ← указатели
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

	unitCards     map[string]*widget.Container
	unitPanelRef  *widget.Container
	queuePanelRef *widget.Container // ← панель очереди в хедере
	// карта юнитов по ID для быстрого доступа к ds.Unit при добавлении в очередь
	unitsByID map[string]ds.Unit
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
		unitsByID:        make(map[string]ds.Unit),
	}
	if data.Player != nil {
		s.myPlayer = *data.Player
		for _, u := range data.Player.Units {
			s.unitsByID[u.ID] = u
		}
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
		// Пример: сервер присылает юнита, которого нужно добавить в очередь.
		// Предполагаем payload: { "unit_id": "..." }
		var payload struct {
			UnitID string `json:"unit_id"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err == nil {
			if u, ok := s.unitsByID[payload.UnitID]; ok {
				s.addUnitToQueue(u)
			}
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

// ── addUnitToQueue ────────────────────────────────────────────────────────────
// Добавляет юнита в очередь активации: обновляет список unitsQueue и рендерит
// карточку в queuePanel. Вызывается как локально (после дропа), так и извне
// (например, из handleMessage при получении серверного события).

func (s *RoomScreen) addUnitToQueue(u ds.Unit) {
	// Не добавлять дубликаты
	for _, id := range s.unitsQueue {
		if id == u.ID {
			return
		}
	}
	s.unitsQueue = append(s.unitsQueue, u.ID)

	s.rebuildQueuePanel()
}

// rebuildQueuePanel полностью перерисовывает queuePanel из среза unitsQueue.
// Первый элемент среза отображается первым слева.
// EbitenUI не поддерживает InsertChildAt, поэтому пересоздаём содержимое целиком.
func (s *RoomScreen) rebuildQueuePanel() {
	if s.queuePanelRef == nil {
		return
	}

	s.queuePanelRef.RemoveChildren()

	for i := len(s.unitsQueue) - 1; i >= 0; i-- {
		u, ok := s.unitsByID[s.unitsQueue[i]]
		if !ok {
			continue
		}

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
			),
		))

		s.queuePanelRef.AddChild(queueCard)
	}
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
			widget.RowLayoutOpts.Spacing(12),
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

	// ── queue panel (в хедере) ────────────────────────────────────────────
	queuePanel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Midnightblue)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(4)),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(54, 54),
		),
	)

	s.queuePanelRef = queuePanel
	header.AddChild(queuePanel)

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

	for _, row := range data.Board {
		for _, cellData := range row {
			isDroppable := cellData != nil && cellData.IsSafeZone && cellData.Unit == nil

			var cell *widget.Container
			if isDroppable {
				var c *widget.Container
				occupied := false

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

							s.onUnitPlaced(droppedUnit)
						}),
					),
				)

				sc.container = c
				safeZoneCells = append(safeZoneCells, sc)

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

	for _, unit := range data.Player.Units {
		u := unit
		unitCard := s.buildUnitCard(u, safeZoneCells)
		unitPanel.AddChild(unitCard)
	}

	footer.AddChild(unitPanel)

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

// onUnitPlaced вызывается после того, как игрок задропил юнита на доску.
// Убирает карточку из unitPanel и добавляет юнита в очередь активации.
func (s *RoomScreen) onUnitPlaced(u ds.Unit) {
	if card, ok := s.unitCards[u.ID]; ok {
		s.unitPanelRef.RemoveChild(card)
		delete(s.unitCards, u.ID)
	}

	s.addUnitToQueue(u)
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
		[3]int{0, w, 0}, // горизонталь: левый=0, центр=вся ширина, правый=0
		[3]int{0, h, 0}, // вертикаль:   верх=0,  центр=вся высота, низ=0
	)
}
