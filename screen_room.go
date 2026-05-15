package game

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math"
	"path"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
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
	opponentCellColor  = colornames.Orangered
	opponentQueueColor = colornames.Crimson
	unitPulseColor1    = colornames.Limegreen
	unitPulseColor2    = colornames.White
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
	board        ds.Board
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

	abilityPanelRef *widget.Container
	abilityPanelIn  bool

	pulseWidgets          []*widget.Container
	pulseTick             float64
	endTurnBtnPulseActive bool

	unitsQueue         []string
	activeUnitID       string
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
		board:        data.Board,
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

	if len(s.pulseWidgets) > 0 || s.endTurnBtnPulseActive {
		s.pulseTick += 0.05
	}

	t := (math.Sin(s.pulseTick) + 1) / 2

	if len(s.pulseWidgets) > 0 {
		c := lerpColor(unitPulseColor1, unitPulseColor2, t)
		for _, w := range s.pulseWidgets {
			w.SetBackgroundImage(image.NewNineSliceColor(c))
		}
	}

	if s.endTurnBtnPulseActive && s.nextActionBtn != nil {
		borderColor := lerpColor(
			color.RGBA{0x11, 0x55, 0x11, 0xff},
			color.RGBA{0x88, 0xFF, 0x88, 0xff},
			t,
		)
		s.nextActionBtn.Image().Idle = image.NewBorderedNineSliceColor(
			color.NRGBA{0x22, 0x8B, 0x22, 0xff},
			borderColor,
			3,
		)
	}

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
	if s.unitPanelIn || len(s.player.Units) == 0 {
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

	for i, pu := range s.player.Units {
		if pu.ID == u.ID {
			s.player.Units = append(s.player.Units[:i], s.player.Units[i+1:]...)
			break
		}
	}

	u.Row = r
	u.Col = c
	u.IsOpponent = false
	s.board[r][c].Unit = &u

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

		bgColor := boardCellColor
		restoreBoardColor := boardCellColor
		if u.IsOpponent {
			bgColor = opponentQueueColor
			restoreBoardColor = opponentCellColor
		}

		isActive := uID == s.activeUnitID

		card := widget.NewContainer(
			widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(bgColor)),
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(54, 54)),
		)

		if isActive {
			s.pulseWidgets = append(s.pulseWidgets, card)
		}

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
					card.SetBackgroundImage(image.NewNineSliceColor(restoreBoardColor))
					if bc := s.boardCellWidget(u); bc != nil {
						bc.SetBackgroundImage(image.NewNineSliceColor(restoreBoardColor))
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
		s.unitPlacedThisTurn = false
		s.hideAbilityPanel()
		s.setupUnitPanel()
		s.setStatus("Place a unit on the board")

	case ws.EndRoundAction:
		s.nextActionBtn.Text().Label = "END\nROUND"
		s.nextActionBtn.GetWidget().Disabled = false
		s.endTurnBtnPulseActive = true
		s.setStatus("You can end your turn")
		s.highlightActiveUnit("")

	case ws.EndTurnAction:
		s.nextActionBtn.Text().Label = "END\nTURN"
		s.nextActionBtn.GetWidget().Disabled = false
		s.endTurnBtnPulseActive = true
		s.highlightActiveUnit("")

	case ws.PlayUnitAction:
		var data ds.PlayUnitPayload
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			log.Fatal(err)
		}

		unit, ok := s.unitByID(data.UnitID)
		if !ok {
			log.Fatal("unit not found: ", data.UnitID)
		}

		if s.unitPanelIn {
			s.footerRef.RemoveChild(s.unitPanelRef)
			s.unitPanelIn = false
		}
		s.showAbilityPanel(unit)

		s.highlightActiveUnit(data.UnitID)
		s.nextActionBtn.Text().Label = "SKIP\nTURN"
		s.nextActionBtn.GetWidget().Disabled = false
		s.setStatus(fmt.Sprintf("Play unit: %s", unit.Name))

	case ws.WaitingForOpponent:
		s.hideAbilityPanel()
		s.setStatus("Waiting for opponent...")

	case ws.UnitPlacedAction:
		var data ds.PlaceUnitPayload
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			log.Fatal(err)
		}

		if data.Row >= 0 && data.Row < len(s.boardCellWidgets) &&
			data.Col >= 0 && data.Col < len(s.boardCellWidgets[data.Row]) {

			cell := s.boardCellWidgets[data.Row][data.Col]
			cell.SetBackgroundImage(image.NewNineSliceColor(opponentCellColor))
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

		opponentUnitID := fmt.Sprintf("opp_%d_%d_%d", data.Unit.TemplateID, data.Row, data.Col)
		s.board[data.Row][data.Col].Unit = &ds.Unit{
			ID:         opponentUnitID,
			TemplateID: data.Unit.TemplateID,
			Row:        data.Row,
			Col:        data.Col,
			IsOpponent: true,
		}

		s.addUnitToQueue(opponentUnitID)
	}
}

func (s *RoomScreen) unitByID(id string) (ds.Unit, bool) {
	for _, row := range s.board {
		for _, cell := range row {
			if cell == nil || cell.Unit == nil {
				continue
			}

			if cell.Unit.ID == id {
				return *cell.Unit, true
			}
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

func (s *RoomScreen) highlightActiveUnit(unitID string) {
	if s.activeUnitID != "" {
		if prev, ok := s.unitByID(s.activeUnitID); ok {
			restoreColor := boardCellColor
			if prev.IsOpponent {
				restoreColor = opponentCellColor
			}
			if bc := s.boardCellWidget(prev); bc != nil {
				bc.SetBackgroundImage(image.NewNineSliceColor(restoreColor))
			}
		}
	}
	s.pulseWidgets = nil
	s.pulseTick = 0
	s.activeUnitID = unitID

	if unitID == "" {
		s.rebuildQueuePanel()
		return
	}

	if u, ok := s.unitByID(unitID); ok {
		if bc := s.boardCellWidget(u); bc != nil {
			s.pulseWidgets = append(s.pulseWidgets, bc)
		}
	}

	s.rebuildQueuePanel()
}

func (s *RoomScreen) showAbilityPanel(unit ds.Unit) {
	s.hideAbilityPanel()

	if len(unit.Abilities) == 0 {
		return
	}

	s.abilityPanelRef = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Slategray)),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(len(unit.Abilities)),
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

	for _, abilityID := range unit.Abilities {
		ab := ability.ByID(string(abilityID))
		card := s.buildAbilityCard(ab)
		s.abilityPanelRef.AddChild(card)
	}

	s.footerRef.AddChild(s.abilityPanelRef)
	s.abilityPanelIn = true
}

func (s *RoomScreen) buildAbilityCard(ab ability.Ability) *widget.Container {
	card := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(colornames.Darkslateblue)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(cellSize, cellSize),
		),
	)

	card.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(abilityImage(ab.ID)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(highlightColor))
			}),
			widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(colornames.Darkslateblue))
			}),
		),
	))

	if ab.Cooldown > 0 {
		cdContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)

		cdContainer.AddChild(widget.NewGraphic(
			widget.GraphicOpts.Image(
				ImageAsset("turn.png", ImageSize{W: 64, H: 64}),
			),
			widget.GraphicOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		))

		tf := ui.TextFace(40)

		cdContainer.AddChild(widget.NewText(
			widget.TextOpts.Text(
				fmt.Sprintf("%d", ab.Cooldown),
				&tf, color.NRGBA{0, 0, 0, 180},
			),
			widget.TextOpts.Position(
				widget.TextPositionCenter,
				widget.TextPositionCenter,
			),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		))

		card.AddChild(cdContainer)
	}

	return card
}

func (s *RoomScreen) hideAbilityPanel() {
	if s.abilityPanelIn && s.abilityPanelRef != nil {
		s.footerRef.RemoveChild(s.abilityPanelRef)
		s.abilityPanelRef = nil
		s.abilityPanelIn = false
	}
}

func (s *RoomScreen) buildNextMoveButton() *widget.Button {
	size := 80

	tf := ui.TextFace(18)

	// TODO colornames
	idle := image.NewBorderedNineSliceColor(
		color.NRGBA{0x22, 0x8B, 0x22, 0xff},
		color.NRGBA{0x11, 0x55, 0x11, 0xff},
		3,
	)
	hover := image.NewBorderedNineSliceColor(
		color.NRGBA{0x32, 0xAB, 0x32, 0xff},
		color.NRGBA{0x11, 0x55, 0x11, 0xff},
		3,
	)
	pressed := image.NewBorderedNineSliceColor(
		color.NRGBA{0x12, 0x6B, 0x12, 0xff},
		color.NRGBA{0x11, 0x55, 0x11, 0xff},
		3,
	)
	disabled := image.NewBorderedNineSliceColor(
		color.NRGBA{0x88, 0x88, 0x88, 0xff},
		color.NRGBA{0x55, 0x55, 0x55, 0xff},
		3,
	)

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

			s.endTurnBtnPulseActive = false
			s.nextActionBtn.Image().Idle = image.NewBorderedNineSliceColor(
				color.NRGBA{0x22, 0x8B, 0x22, 0xff},
				color.NRGBA{0x11, 0x55, 0x11, 0xff},
				3,
			)

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

func lerpColor(a, b color.RGBA, t float64) color.NRGBA {
	lerp := func(x, y uint8, t float64) uint8 {
		return uint8(float64(x) + (float64(y)-float64(x))*t)
	}
	return color.NRGBA{
		R: lerp(a.R, b.R, t),
		G: lerp(a.G, b.G, t),
		B: lerp(a.B, b.B, t),
		A: lerp(a.A, b.A, t),
	}
}

func abilityImage(abilityID string) *ebiten.Image {
	up := path.Join("abilities", abilityID+".png")
	return ImageAsset(up, ImageSize{W: 64, H: 64})
}
