package arena

import (
	"encoding/json"
	"fmt"
	stdImg "image"
	"image/color"
	"log"
	"math"
	"path"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	game "github.com/ognev-dev/goplease-ebitengine-client"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/asset"
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

type Screen struct {
	server       ws.Client
	ui           *ebitenui.UI
	board        ds.Board
	roomID       string
	player       ds.Player
	opponentName string
	isMyTurn     bool

	safeZoneCells    []*DropZoneCell
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

func NewScreen(payload json.RawMessage, server ws.Client) game.Screen {
	var data ds.NewGamePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		log.Fatalf("failed to unmarshal: %v", err)
	}

	s := &Screen{
		server:       server,
		board:        data.Board,
		roomID:       data.RoomID,
		player:       *data.Player,
		unitCards:    make(map[string]*widget.Container),
		turnNumber:   1,
		opponentName: data.Opponent,
	}

	initDropPointAnim()

	s.setupUI(data)
	return s
}

func (s *Screen) Update(g *game.Game) (game.Screen, error) {
	for {
		select {
		case msg := <-g.Server.Inbox():
			s.handleMessage(msg)
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
		c := ui.LerpColor(unitPulseColor1, unitPulseColor2, t)
		for _, w := range s.pulseWidgets {
			w.SetBackgroundImage(image.NewNineSliceColor(c))
		}
	}

	if s.endTurnBtnPulseActive && s.nextActionBtn != nil {
		borderColor := ui.LerpColor(
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

	animDropArrow.Update()
	for _, sc := range s.safeZoneCells {
		if sc.activeGraphic != nil {
			sc.activeGraphic.Image = animDropArrow.CurrentFrame
		}
	}

	s.ui.Update()
	return s, nil
}

func (s *Screen) Draw(screen *ebiten.Image) {
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

func (s *Screen) setupUI(data ds.NewGamePayload) {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	s.headerRef = s.createHeader()
	s.footerRef = s.createFooter()
	board := s.createBoardContainer(data.Board)
	statusBar := s.createStatusBar()

	root.AddChild(board)
	root.AddChild(s.headerRef)
	root.AddChild(statusBar)
	root.AddChild(s.footerRef)

	s.ui = &ebitenui.UI{Container: root}
}

func (s *Screen) createHeader() *widget.Container {
	h := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(headerBgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{StretchHorizontal: true}),
			widget.WidgetOpts.MinSize(0, headerH),
		),
	)

	s.queuePanelRef = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(unitPanelBgColor)),
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

func (s *Screen) createFooter() *widget.Container {
	footer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(footerBgColor)),
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

func (s *Screen) createBoardContainer(boardData ds.Board) *widget.Container {
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
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(boardBgColor)),
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

func (s *Screen) createStatusBar() *widget.Container {
	bar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(statusBarBgColor)),
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

func (s *Screen) createCell(r, c int, data *ds.BoardCell) *widget.Container {
	isDroppable := data != nil && data.IsSafeZone && data.Unit == nil
	sc := &DropZoneCell{row: r, col: c}

	opts := []widget.ContainerOpt{
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(boardCellBgColor)),
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
				sc.baseColor = unitFriendlyBgColor
				s.unitPlacedThisTurn = true

				if sc.activeGraphic != nil {
					sc.container.RemoveChild(sc.activeGraphic)
					sc.activeGraphic = nil
				}

				sc.container.SetBackgroundImage(image.NewNineSliceColor(unitFriendlyBgColor))
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

func (s *Screen) setupUnitPanel() {
	if s.unitPanelIn || len(s.player.Units) == 0 {
		return
	}

	s.unitPanelRef = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(unitPanelBgColor)),
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

func (s *Screen) buildUnitCard(u ds.Unit) *widget.Container {
	dnd := &dndHandler{
		dndUnit:   &dndUnit{unit: u},
		safeCells: s.safeZoneCells,
		// Drag is only allowed once the server has confirmed placement phase
		// AND no unit has been placed in the current turn yet.
		canDrag: func() bool { return s.ready && !s.unitPlacedThisTurn },
	}

	card := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(unitCardBgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(cellSize, cellSize),
			widget.WidgetOpts.EnableDragAndDrop(widget.NewDragAndDrop(
				widget.DragAndDropOpts.ContentsCreater(dnd),
				widget.DragAndDropOpts.MinDragStartDistance(10),
				widget.DragAndDropOpts.ContentsOriginVertical(widget.DND_ANCHOR_START),
				widget.DragAndDropOpts.ContentsOriginHorizontal(widget.DND_ANCHOR_START),
				widget.DragAndDropOpts.Offset(stdImg.Point{-5, -5}),
			)),
			widget.WidgetOpts.ToolTip(
				widget.NewToolTip(
					widget.ToolTipOpts.Content(s.buildUnitToolTip(u)),
					widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
					widget.ToolTipOpts.Offset(stdImg.Point{X: 0, Y: -8}),
					widget.ToolTipOpts.AnchorOriginHorizontal(widget.TOOLTIP_ANCHOR_MIDDLE),
					widget.ToolTipOpts.AnchorOriginVertical(widget.TOOLTIP_ANCHOR_START),
					widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_MIDDLE),
					widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_END),
				),
			),
		),
	)

	normalIcon := unitImage(u.TemplateID)
	hoverIcon := ui.TintImage(normalIcon, unitCardHoverFgColor)

	var graphic *widget.Graphic
	graphic = widget.NewGraphic(
		widget.GraphicOpts.Image(normalIcon),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(unitCardHoverBgColor))
				graphic.Image = hoverIcon
			}),
			widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(unitCardBgColor))
				graphic.Image = normalIcon
			}),
		),
	)
	card.AddChild(graphic)

	return card
}

func (s *Screen) onUnitPlaced(u ds.Unit, r, c int) {
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

func (s *Screen) addUnitToQueue(unitID string) {
	for _, id := range s.unitsQueue {
		if id == unitID {
			return
		}
	}
	s.unitsQueue = append(s.unitsQueue, unitID)
	s.rebuildQueuePanel()
}

func (s *Screen) rebuildQueuePanel() {
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

		bgColor := unitFriendlyBgColor
		restoreBoardColor := unitFriendlyBgColor
		if u.IsOpponent {
			bgColor = unitEnemyBgColor
			restoreBoardColor = unitEnemyBgColor
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
					card.SetBackgroundImage(image.NewNineSliceColor(unitCardHighlightColor))
					if bc := s.boardCellWidget(u); bc != nil {
						bc.SetBackgroundImage(image.NewNineSliceColor(unitCardHighlightColor))
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

func (s *Screen) handleMessage(msg ws.InMessage) {
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
			cell.SetBackgroundImage(image.NewNineSliceColor(unitEnemyBgColor))
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

func (s *Screen) unitByID(id string) (ds.Unit, bool) {
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

func (s *Screen) boardCellWidget(u ds.Unit) *widget.Container {
	if u.Row < 0 || u.Row >= len(s.boardCellWidgets) || u.Col < 0 || u.Col >= len(s.boardCellWidgets[u.Row]) {
		return nil
	}
	return s.boardCellWidgets[u.Row][u.Col]
}

func (s *Screen) setStatus(text string) {
	if s.statusLabel != nil {
		s.statusLabel.Label = text
	}
}

func (s *Screen) highlightActiveUnit(unitID string) {
	if s.activeUnitID != "" {
		if prev, ok := s.unitByID(s.activeUnitID); ok {
			restoreColor := unitFriendlyBgColor
			if prev.IsOpponent {
				restoreColor = unitEnemyBgColor
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

func (s *Screen) showAbilityPanel(unit ds.Unit) {
	s.hideAbilityPanel()

	if len(unit.Abilities) == 0 {
		return
	}

	s.abilityPanelRef = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(abilitiesPanelBgColor)),
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

func (s *Screen) buildAbilityCard(ab ability.Ability) *widget.Container {
	var card *widget.Container

	bgColor := abilityBgColor
	if ab.IsBasicAttack() {
		bgColor = basicAttackBgColor
	}
	if ab.IsPassive {
		bgColor = passiveAbilityBgColor
	}

	card = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(bgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(cellSize, cellSize),
			widget.WidgetOpts.ToolTip(
				widget.NewToolTip(
					widget.ToolTipOpts.Content(s.buildAbilityToolTip(ab)),
					widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
					widget.ToolTipOpts.Offset(stdImg.Point{X: 0, Y: -8}),
					widget.ToolTipOpts.AnchorOriginHorizontal(widget.TOOLTIP_ANCHOR_MIDDLE),
					widget.ToolTipOpts.AnchorOriginVertical(widget.TOOLTIP_ANCHOR_START),
					widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_MIDDLE),
					widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_END),
				),
			),
			widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(ui.DarkenRGB(bgColor, 30)))
			}),
			widget.WidgetOpts.CursorExitHandler(func(args *widget.WidgetCursorExitEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(bgColor))
			}),
		),
	)

	card.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(abilityImage(ab.ID)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	if ab.Cooldown > 0 {
		cdContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
				widget.RowLayoutOpts.Spacing(2),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionStart,
					VerticalPosition:   widget.AnchorLayoutPositionStart,
					Padding:            &widget.Insets{Top: 3, Left: 3},
				}),
			),
		)
		cdContainer.AddChild(widget.NewGraphic(
			widget.GraphicOpts.Image(asset.Image("turn.png", 12)),
		))
		tf := ui.TextFace(11)
		cdContainer.AddChild(widget.NewText(
			widget.TextOpts.Text(fmt.Sprintf("%d", ab.Cooldown), &tf, colornames.White),
		))
		card.AddChild(cdContainer)
	}

	return card
}

func (s *Screen) hideAbilityPanel() {
	if s.abilityPanelIn && s.abilityPanelRef != nil {
		s.footerRef.RemoveChild(s.abilityPanelRef)
		s.abilityPanelRef = nil
		s.abilityPanelIn = false
	}
}

func (s *Screen) buildNextMoveButton() *widget.Button {
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

func unitImage(templateID int, sizeOpt ...int) *ebiten.Image {
	sizeDefault := 64
	if len(sizeOpt) > 0 {
		sizeDefault = sizeOpt[0]
	}

	up := path.Join("units", fmt.Sprintf("unit_%d_pic.png", templateID))

	return asset.Image(up, sizeDefault)
}

func abilityImage(abilityID string, sizeOpt ...int) *ebiten.Image {
	size := 64
	if len(sizeOpt) > 0 {
		size = sizeOpt[0]
	}

	return asset.Image(path.Join("abilities", abilityID+".png"), size)
}

func (s *Screen) buildAbilityToolTip(ab ability.Ability) *widget.Container {
	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewBorderedNineSliceColor(ttBgColor, ttBorderColor, 2)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(8)),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.AutoDisableChildren(),
	)

	header := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(8),
		)),
	)
	header.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(abilityImage(ab.ID, 28)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(24, 24),
		),
	))
	header.AddChild(widget.NewText(
		widget.TextOpts.Text(ab.Name, &toolTipTitleTF, ttTitleColor),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	))
	c.AddChild(header)

	c.AddChild(widget.NewText(
		widget.TextOpts.Text(ab.Description, &toolTipTextTF, ttTextColor),
		widget.TextOpts.MaxWidth(350),
	))

	if ab.Cooldown > 0 {
		cdRow := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
				widget.RowLayoutOpts.Spacing(4),
			)),
		)
		cdRow.AddChild(widget.NewText(
			widget.TextOpts.Text(fmt.Sprintf("Cooldown: %d", ab.Cooldown), &toolTipTextTF, colornames.Skyblue),
		))
		c.AddChild(cdRow)
	}
	if ab.Range > 0 {
		rangeRow := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
				widget.RowLayoutOpts.Spacing(4),
			)),
		)
		rangeRow.AddChild(widget.NewText(
			widget.TextOpts.Text(fmt.Sprintf("Range: %d", ab.Range), &toolTipTextTF, colornames.Palegreen),
		))
		c.AddChild(rangeRow)
	}

	if ab.IsPassive {
		c.AddChild(widget.NewText(
			widget.TextOpts.Text("Passive", &toolTipTextTF, colornames.Plum),
		))
	}

	return c
}

func (s *Screen) buildUnitToolTip(u ds.Unit) *widget.Container {
	c := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewBorderedNineSliceColor(ttBgColor, ttBorderColor, 2)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(15)),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.AutoDisableChildren(),
	)

	header := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(8),
		)),
	)
	header.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(unitImage(u.TemplateID, 28)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(28, 28),
		),
	))
	header.AddChild(widget.NewText(
		widget.TextOpts.Text(u.Name, &toolTipTitleTF, ttTitleColor),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	))
	c.AddChild(header)

	c.AddChild(widget.NewText(
		widget.TextOpts.Text(u.Description, &toolTipTextTF, ttTextColor),
		widget.TextOpts.MaxWidth(350),
	))

	stats := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(12),
		)),
	)
	stats.AddChild(tooltipStatRow("heart.png", fmt.Sprintf("HP: %d", u.CurrentHP), &toolTipTextTF, hpColor))
	stats.AddChild(tooltipStatRow("hit.png", fmt.Sprintf("ATK: %d", u.CurrentAtk), &toolTipTextTF, atkColor))
	stats.AddChild(tooltipStatRow("walk.png", fmt.Sprintf("Move: %d", u.MP), &toolTipTextTF, mpColor))
	c.AddChild(stats)

	return c
}

func tooltipStatRow(iconPath string, label string, tf *text.Face, c color.Color) *widget.Container {
	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(4),
		)),
	)
	row.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(asset.Image(iconPath, 18)),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	))
	row.AddChild(widget.NewText(
		widget.TextOpts.Text(label, tf, c),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	))
	return row
}
