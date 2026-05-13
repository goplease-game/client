package game

import (
	"encoding/json"
	"fmt"
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

const (
	cellSize = 64
	headerH  = 80
	statusH  = 32
	footerH  = 90
)

var (
	boardCellColor     = colornames.Darkgray
	dropZoneColor      = colornames.Limegreen
	dropZoneHoverColor = colornames.Palegreen
	highlightColor     = colornames.Gold
)

// ── SafeZoneCell ─────────────────────────────────────────────────────────────

type SafeZoneCell struct {
	container     *widget.Container
	activeGraphic *widget.Graphic
	occupied      bool
	row, col      int
}

func (sc *SafeZoneCell) SetHighlight(active bool) {
	// Если мы выключаем подсветку, нам неважно, занята клетка или нет —
	// графику нужно убрать в любом случае.
	if !active {
		sc.container.SetBackgroundImage(image.NewNineSliceColor(boardCellColor))
		if sc.activeGraphic != nil {
			sc.container.RemoveChild(sc.activeGraphic)
			sc.activeGraphic = nil
		}
		return
	}

	// А вот если включаем (active == true), тогда проверяем на занятость
	if sc.occupied {
		return
	}

	sc.container.SetBackgroundImage(image.NewNineSliceColor(dropZoneColor))
	if sc.activeGraphic == nil {
		dropImg := ImageAsset("drop_point.png", ImageSize{W: 52, H: 52})
		sc.activeGraphic = widget.NewGraphic(
			widget.GraphicOpts.Image(dropImg),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		sc.container.AddChild(sc.activeGraphic)
	}
}

func (sc *SafeZoneCell) SetHover(hover bool) {
	if !sc.occupied {
		c := dropZoneColor
		if hover {
			c = dropZoneHoverColor
		}
		sc.container.SetBackgroundImage(image.NewNineSliceColor(c))
	}
}

// ── Drag-and-Drop Logic ──────────────────────────────────────────────────────

type dndUnit struct {
	unit       ds.Unit
	dragWidget *widget.Container
}

func (d *dndUnit) Create(_ widget.HasWidget) (*widget.Container, interface{}) {
	if d.dragWidget == nil {
		unitImg := unitImage(d.unit.TemplateID)
		d.dragWidget = widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Ghostwhite)),
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

type dndHandler struct {
	*dndUnit
	safeCells   []*SafeZoneCell
	currentCell *SafeZoneCell
	canDrag     func() bool
}

func (d *dndHandler) Create(parent widget.HasWidget) (*widget.Container, interface{}) {
	if !d.canDrag() {
		return nil, nil
	}

	for _, sc := range d.safeCells {
		sc.SetHighlight(true)
	}
	return d.dndUnit.Create(parent)
}

func (d *dndHandler) Update(canDrop bool, target widget.HasWidget, _ interface{}) {
	if d.currentCell != nil {
		d.currentCell.SetHover(false)
		d.currentCell = nil
	}
	if canDrop && target != nil {
		for _, sc := range d.safeCells {
			if sc.container == target {
				sc.SetHover(true)
				d.currentCell = sc
				break
			}
		}
	}
}

func (d *dndHandler) EndDrag(_ bool, _ widget.HasWidget, _ interface{}) {
	for _, sc := range d.safeCells {
		sc.SetHighlight(false)
	}
}

// ── RoomScreen ───────────────────────────────────────────────────────────────

type RoomScreen struct {
	server       ws.Client
	ui           *ebitenui.UI
	roomID       string
	player       ds.Player
	opponentName string
	isMyTurn     bool

	safeZoneCells    []*SafeZoneCell
	boardCellWidgets [][]*widget.Container
	unitCards        map[string]*widget.Container
	headerRef        *widget.Container
	footerRef        *widget.Container
	queuePanelRef    *widget.Container
	unitPanelRef     *widget.Container
	nextActionBtn    *widget.Button // next & end turn
	statusLabel      *widget.Text

	unitsQueue         []string
	activeUnitIndex    int
	turnNumber         int
	unitPlacedThisTurn bool
	queueIn            bool
	unitPanelIn        bool

	// ready is set to true when the server responds with phase unit_placement,
	// meaning the match has started and the local player may interact.
	ready bool
	// firstDrawn tracks whether we have completed at least one Draw call so
	// that we send ready_to_play exactly once after the UI is fully rendered.
	firstDrawn bool
}

func NewRoomScreen(payload json.RawMessage, server ws.Client) *RoomScreen {
	var data ds.NewGamePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		log.Fatalf("failed to unmarshal: %v", err)
	}

	s := &RoomScreen{
		server:       server,
		roomID:       data.RoomID,
		player:       *data.Player,
		unitCards:    make(map[string]*widget.Container),
		turnNumber:   1,
		opponentName: data.Opponent,
	}

	s.setupUI(data)
	return s
}

func (s *RoomScreen) Update(g *Game) (Screen, error) {
	for {
		select {
		case msg := <-g.Server.Inbox():
			s.handleMessage(g, msg)
		default:
			goto done
		}
	}
done:
	s.ui.Update()
	return s, nil
}

func (s *RoomScreen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)

	// After the very first completed Draw the scene is fully rendered.
	// Send ready_to_play once so the server knows we are displaying the board.
	if !s.firstDrawn {
		s.firstDrawn = true
		s.server.Send(ws.OutMessage{
			Action: ws.ReadyToPlay,
		})
	}
}

func (s *RoomScreen) setupUI(data ds.NewGamePayload) {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	s.headerRef = s.createHeader()
	s.footerRef = s.createFooter()
	center := s.createBoardContainer(data.Board)
	statusBar := s.createStatusBar()

	root.AddChild(center)
	root.AddChild(s.headerRef)
	root.AddChild(statusBar)
	root.AddChild(s.footerRef)

	s.ui = &ebitenui.UI{Container: root}
}

func (s *RoomScreen) createHeader() *widget.Container {
	h := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Steelblue)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true}),
			widget.WidgetOpts.MinSize(0, headerH),
		),
	)

	s.queuePanelRef = widget.NewContainer(
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

	return h
}

func (s *RoomScreen) createFooter() *widget.Container {
	footer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Steelblue)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				VerticalPosition:  widget.AnchorLayoutPositionEnd,
				StretchHorizontal: true,
			}),
			widget.WidgetOpts.MinSize(0, footerH),
		),
	)

	btn := s.buildNextMoveButton()
	btn.GetWidget().LayoutData = widget.AnchorLayoutData{
		HorizontalPosition: widget.AnchorLayoutPositionEnd,
		VerticalPosition:   widget.AnchorLayoutPositionCenter,
		Padding:            &widget.Insets{Right: 12},
	}
	footer.AddChild(btn)

	return footer
}

func (s *RoomScreen) createBoardContainer(boardData ds.Board) *widget.Container {
	cols := 0
	if len(boardData) > 0 {
		cols = len(boardData[0])
	}

	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
				Padding:           &widget.Insets{Top: headerH, Bottom: footerH + statusH},
			}),
		),
	)

	boardWidget := widget.NewContainer(
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

	s.boardCellWidgets = make([][]*widget.Container, len(boardData))
	for r, row := range boardData {
		s.boardCellWidgets[r] = make([]*widget.Container, len(row))
		for c, cellData := range row {
			s.boardCellWidgets[r][c] = s.createCell(r, c, cellData)
			boardWidget.AddChild(s.boardCellWidgets[r][c])
		}
	}

	container.AddChild(boardWidget)
	return container
}

func (s *RoomScreen) createStatusBar() *widget.Container {
	bar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x1e, 0x26, 0x30, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				VerticalPosition:  widget.AnchorLayoutPositionEnd,
				StretchHorizontal: true,
				Padding:           &widget.Insets{Bottom: footerH},
			}),
			widget.WidgetOpts.MinSize(0, statusH),
		),
	)

	tf := ui.TextFace(16)
	s.statusLabel = widget.NewText(
		widget.TextOpts.Text("Waiting for opponent...", &tf, colornames.Lightgray),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	bar.AddChild(s.statusLabel)
	return bar
}

func (s *RoomScreen) createCell(r, c int, data *ds.BoardCell) *widget.Container {
	isDroppable := data != nil && data.IsSafeZone && data.Unit == nil
	sc := &SafeZoneCell{row: r, col: c}

	opts := []widget.ContainerOpt{
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(boardCellColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(cellSize, cellSize),
		),
	}

	if isDroppable {
		opts = append(opts, widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.CanDrop(func(args *widget.DragAndDropDroppedEventArgs) bool {
				_, ok := args.Data.(ds.Unit)
				return ok && s.ready && !sc.occupied && !s.unitPlacedThisTurn
			}),
			widget.WidgetOpts.Dropped(func(args *widget.DragAndDropDroppedEventArgs) {
				unit, ok := args.Data.(ds.Unit)
				if !ok {
					return
				}
				sc.occupied = true
				s.unitPlacedThisTurn = true
				sc.SetHighlight(false)
				sc.container.AddChild(widget.NewGraphic(
					widget.GraphicOpts.Image(unitImage(unit.TemplateID)),
					widget.GraphicOpts.WidgetOpts(
						widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
							HorizontalPosition: widget.AnchorLayoutPositionCenter,
							VerticalPosition:   widget.AnchorLayoutPositionCenter,
						}),
					),
				))
				s.onUnitPlaced(unit, r, c)
			}),
		))
	}

	cell := widget.NewContainer(opts...)
	sc.container = cell
	if isDroppable {
		s.safeZoneCells = append(s.safeZoneCells, sc)
	}

	return cell
}

func (s *RoomScreen) setupUnitPanel() {
	if len(s.player.Units) == 0 {
		return
	}

	s.unitPanelRef = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Slategray)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(len(s.player.Units)),
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

	for _, u := range s.player.Units {
		card := s.buildUnitCard(u)
		s.unitPanelRef.AddChild(card)
		s.unitCards[u.ID] = card
	}

	s.footerRef.AddChild(s.unitPanelRef)
	s.unitPanelIn = true
}

func (s *RoomScreen) buildUnitCard(u ds.Unit) *widget.Container {
	dnd := &dndHandler{
		dndUnit:   &dndUnit{unit: u},
		safeCells: s.safeZoneCells,
		// Drag is only allowed once the server has confirmed placement phase
		// AND no unit has been placed in the current turn yet.
		canDrag: func() bool { return s.ready && !s.unitPlacedThisTurn },
	}

	card := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Ghostwhite)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(cellSize, cellSize),
			widget.WidgetOpts.EnableDragAndDrop(widget.NewDragAndDrop(
				widget.DragAndDropOpts.ContentsCreater(dnd),
				widget.DragAndDropOpts.MinDragStartDistance(10),
			)),
		),
	)

	card.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(unitImage(u.TemplateID)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(highlightColor))
			}),
			widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(colornames.Ghostwhite))
			}),
		),
	))

	return card
}

func (s *RoomScreen) onUnitPlaced(u ds.Unit, r, c int) {
	if card, ok := s.unitCards[u.ID]; ok {
		s.unitPanelRef.RemoveChild(card)
		delete(s.unitCards, u.ID)
	}

	if len(s.unitCards) == 0 && s.unitPanelIn {
		s.footerRef.RemoveChild(s.unitPanelRef)
		s.unitPanelIn = false
	}

	for i := range s.player.Units {
		if s.player.Units[i].ID == u.ID {
			s.player.Units[i].Row = r
			s.player.Units[i].Col = c
			break
		}
	}
	s.addUnitToQueue(u.ID)

	s.server.Send(ws.OutMessage{
		Action: ws.UnitPlacedAction,
		Data: ds.UnitPlacedPayload{
			TemplateID: u.TemplateID,
			Row:        r,
			Col:        c,
		},
	})
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
		if s.queueIn {
			s.headerRef.RemoveChild(s.queuePanelRef)
			s.queueIn = false
		}
		return
	}

	if !s.queueIn {
		s.headerRef.AddChild(s.queuePanelRef)
		s.queueIn = true
	}

	for i := len(s.unitsQueue) - 1; i >= 0; i-- {
		uID := s.unitsQueue[i]
		u, ok := s.unitByID(uID)
		if !ok {
			continue
		}

		card := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(boardCellColor)),
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(54, 54)),
		)

		card.AddChild(widget.NewGraphic(
			widget.GraphicOpts.Image(unitImage(u.TemplateID)),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
				widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
					card.SetBackgroundImage(image.NewNineSliceColor(highlightColor))
					if bc := s.boardCellWidget(u); bc != nil {
						bc.SetBackgroundImage(image.NewNineSliceColor(highlightColor))
					}
				}),
				widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
					card.SetBackgroundImage(image.NewNineSliceColor(boardCellColor))
					if bc := s.boardCellWidget(u); bc != nil {
						bc.SetBackgroundImage(image.NewNineSliceColor(boardCellColor))
					}
				}),
			),
		))
		s.queuePanelRef.AddChild(card)
	}
}

func (s *RoomScreen) handleMessage(g *Game, msg ws.InMessage) {
	fmt.Printf("received: %v\n", msg)

	switch msg.Action {
	case ws.PlaceUnitAction:
		s.ready = true
		s.setupUnitPanel()
		s.setStatus("Place a unit on the board")

	case ws.EndTurnAction:
		s.nextActionBtn.Text().Label = "END\nTURN"
		s.nextActionBtn.GetWidget().Disabled = false
		s.setStatus("You can end your turn")

	case ws.NextAction:
		s.nextActionBtn.Text().Label = "Next"
		s.nextActionBtn.GetWidget().Disabled = false

	case ws.UnitPlacedAction:
		var data ds.PlaceUnitPayload
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			log.Fatal(err)
		}

		// Place the opponent's unit graphic on the board at the given coordinates.
		if data.Row >= 0 && data.Row < len(s.boardCellWidgets) &&
			data.Col >= 0 && data.Col < len(s.boardCellWidgets[data.Row]) {

			cell := s.boardCellWidgets[data.Row][data.Col]
			cell.AddChild(widget.NewGraphic(
				widget.GraphicOpts.Image(unitImage(data.Unit.TemplateID)),
				widget.GraphicOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
						HorizontalPosition: widget.AnchorLayoutPositionCenter,
						VerticalPosition:   widget.AnchorLayoutPositionCenter,
					}),
				),
			))
		}

		// Build a stable synthetic ID and register the unit so the queue
		// panel can look it up via unitByID and render its icon correctly.
		opponentUnitID := fmt.Sprintf("opp_%d_%d_%d", data.Unit.TemplateID, data.Row, data.Col)
		s.player.Units = append(s.player.Units, ds.Unit{
			ID:         opponentUnitID,
			TemplateID: data.Unit.TemplateID,
			Row:        data.Row,
			Col:        data.Col,
		})

		s.addUnitToQueue(opponentUnitID)
	}
}

func (s *RoomScreen) unitByID(id string) (ds.Unit, bool) {
	for _, u := range s.player.Units {
		if u.ID == id {
			return u, true
		}
	}

	return ds.Unit{}, false
}

func (s *RoomScreen) boardCellWidget(u ds.Unit) *widget.Container {
	if u.Row < 0 || u.Row >= len(s.boardCellWidgets) || u.Col < 0 || u.Col >= len(s.boardCellWidgets[u.Row]) {
		return nil
	}
	return s.boardCellWidgets[u.Row][u.Col]
}

func (s *RoomScreen) setStatus(text string) {
	if s.statusLabel != nil {
		s.statusLabel.Label = text
	}
}

func (s *RoomScreen) buildNextMoveButton() *widget.Button {
	size := 80

	tf := ui.TextFace(18)

	// TODO colornames
	idle := image.NewNineSliceSimple(ui.CreateCircleImage(size, color.NRGBA{0x22, 0x8B, 0x22, 0xff}), 0, size)
	hover := image.NewNineSliceSimple(ui.CreateCircleImage(size, color.NRGBA{0x32, 0xAB, 0x32, 0xff}), 0, size)
	pressed := image.NewNineSliceSimple(ui.CreateCircleImage(size, color.NRGBA{0x12, 0x6B, 0x12, 0xff}), 0, size)
	disabled := image.NewNineSliceSimple(ui.CreateCircleImage(size, color.NRGBA{0x88, 0x88, 0x88, 0xff}), 0, size)

	btn := widget.NewButton(
		widget.ButtonOpts.Image(&widget.ButtonImage{
			Idle:     idle,
			Hover:    hover,
			Pressed:  pressed,
			Disabled: disabled,
		}),
		widget.ButtonOpts.Text("Next", &tf, &widget.ButtonTextColor{
			// TODO colornames
			Idle:     color.NRGBA{0xff, 0xff, 0xff, 0xff},
			Disabled: color.NRGBA{0xaa, 0xaa, 0xaa, 0xff},
		}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(size, size),
		),

		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if !s.ready {
				return
			}

			s.server.Send(ws.OutMessage{
				Action: ws.EndTurnAction,
			})
		}),
	)

	btn.GetWidget().Disabled = true // disabled until server unlocks it
	s.nextActionBtn = btn

	return btn
}

func unitImage(templateID int) *ebiten.Image {
	up := path.Join("units", fmt.Sprintf("unit_%d_pic.png", templateID))

	return ImageAsset(up, ImageSize{W: 64, H: 64})
}
