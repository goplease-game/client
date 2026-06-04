package arena

import (
	"fmt"
	stdImg "image"
	"image/color"
	"log"
	"path"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ability/status"
	"github.com/ognev-dev/goplease-ebitengine-client/asset"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/sfx"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
	"golang.org/x/image/colornames"
)

// setupUnitPanel builds the hand panel in the footer showing the player's
// undeployed units. No-ops if the panel is already shown or there are no units.
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
		card := s.buildUnitCard(&u)
		s.unitPanelRef.AddChild(card)
		s.unitCards[u.ID] = card
	}

	s.footerRef.AddChild(s.unitPanelRef)
	s.unitPanelIn = true
}

// buildUnitCard creates a draggable unit card for the hand panel.
// The card shows hover and drag-and-drop behaviour and includes a tooltip.
func (s *Screen) buildUnitCard(u *ds.Unit) *widget.Container {
	dnd := &dndHandler{
		dndUnit:   &dndUnit{unit: u},
		safeCells: s.safeZoneCells,
		canDrag:   func() bool { return s.ready && !s.unitPlacedThisTurn },
	}

	var refs UnitCardRefs
	var card *widget.Container

	card = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(unitCardBgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(unitCardSize, unitCardSize),
			widget.WidgetOpts.CursorEnterHandler(func(_ *widget.WidgetCursorEnterEventArgs) {
				sfx.Play(unitHoverSound)
				card.SetBackgroundImage(image.NewNineSliceColor(unitCardHoverBgColor))
				refs.Icon.Image = refs.HoverIcon
			}),
			widget.WidgetOpts.CursorExitHandler(func(_ *widget.WidgetCursorExitEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(unitCardBgColor))
				refs.Icon.Image = refs.NormIcon
			}),
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

	refs = buildHandCard(card, u)
	return card
}

// onUnitPlaced is called after a successful drop onto a safe-zone cell.
// It removes the unit from the player's hand, updates the board state,
// adds the unit to the turn queue, and notifies the server.
func (s *Screen) onUnitPlaced(u *ds.Unit, coord ds.HexCoord) {
	sfx.Play(unitPlacedSound)
	s.removeUnitCard(u.ID)

	for i, pu := range s.player.Units {
		if pu.ID == u.ID {
			s.player.Units = append(s.player.Units[:i], s.player.Units[i+1:]...)
			break
		}
	}

	u.Pos.Q = coord.Q
	u.Pos.R = coord.R
	u.IsOpponent = false

	if cell := s.board.Cells[coord]; cell != nil {
		cell.Unit = u
	}

	s.addUnitToQueue(u)

	s.server.Send(ws.OutMessage{
		Action: ws.UnitPlacedAction,
		Data: ds.UnitPlacedPayload{
			TemplateID: u.TemplateID,
			Coord:      coord,
		},
	})
}

// removeUnitCard removes the card for unitID from the hand panel.
// If the panel becomes empty it is also removed from the footer.
func (s *Screen) removeUnitCard(unitID string) {
	card, ok := s.unitCards[unitID]
	if !ok {
		return
	}
	s.unitPanelRef.RemoveChild(card)
	delete(s.unitCards, unitID)

	if len(s.unitCards) == 0 && s.unitPanelIn {
		s.footerRef.RemoveChild(s.unitPanelRef)
		s.unitPanelIn = false
	}
}

// addUnitToQueue appends unit to the turn queue if not already present,
// then rebuilds the queue panel in the header.
func (s *Screen) addUnitToQueue(unit *ds.Unit) {
	for _, u := range s.unitsQueue {
		if u.ID == unit.ID {
			return
		}
	}
	s.unitsQueue = append(s.unitsQueue, unit)
	s.rebuildQueuePanel()
}

// rebuildQueuePanel clears and repopulates the queue panel in the header.
// The queue is displayed newest-first. The panel is hidden when the queue is empty.
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

	for _, u := range s.unitsQueue {
		s.queuePanelRef.AddChild(s.buildQueueCard(u, u.ID == s.activeUnitID))
	}
}

// buildQueueCard creates a unit card for the turn queue panel in the header.
// The active unit's card is added to pulseWidgets so it pulses.
// Hovering a card also highlights the corresponding hex cell on the board.
func (s *Screen) buildQueueCard(u *ds.Unit, isActive bool) *widget.Container {
	bgColor := unitFriendlyBgColor
	if u.IsOpponent {
		bgColor = unitEnemyBgColor
	}

	restoreColor := bgColor
	var card *widget.Container

	widgetOpts := []widget.WidgetOpt{
		widget.WidgetOpts.MinSize(unitCardSize, 54),
		widget.WidgetOpts.CursorEnterHandler(func(_ *widget.WidgetCursorEnterEventArgs) {
			sfx.Play(unitHoverSound)
			card.SetBackgroundImage(image.NewNineSliceColor(unitCardHighlightColor))
			current := s.unitByID(u.ID)
			if bc := s.boardCellWidget(current); bc != nil {
				bc.SetColor(unitCardHighlightColor)
			}
		}),
		widget.WidgetOpts.CursorExitHandler(func(_ *widget.WidgetCursorExitEventArgs) {
			card.SetBackgroundImage(image.NewNineSliceColor(restoreColor))
			current := s.unitByID(u.ID)
			if bc := s.boardCellWidget(current); bc != nil {
				bc.SetColor(restoreColor)
			}
		}),
		widget.WidgetOpts.ToolTip(
			widget.NewToolTip(
				widget.ToolTipOpts.Content(buildStatusTooltip(u)),
				widget.ToolTipOpts.Position(widget.TOOLTIP_POS_WIDGET),
				widget.ToolTipOpts.Offset(stdImg.Point{X: 0, Y: 8}),
				widget.ToolTipOpts.AnchorOriginHorizontal(widget.TOOLTIP_ANCHOR_MIDDLE),
				widget.ToolTipOpts.AnchorOriginVertical(widget.TOOLTIP_ANCHOR_END),
				widget.ToolTipOpts.ContentOriginHorizontal(widget.TOOLTIP_ANCHOR_MIDDLE),
				widget.ToolTipOpts.ContentOriginVertical(widget.TOOLTIP_ANCHOR_START),
			),
		),
	}

	card = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(bgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(widgetOpts...),
	)

	if isActive {
		s.pulseWidgets = append(s.pulseWidgets, card)
	}

	buildQueueUnitCard(NewContainerChildAdder(card), u)
	return card
}

// highlightActiveUnit updates board visuals when the active unit changes.
// It restores the previous unit's cell, sets the new active unit,
// starts the pulse on its hex cell, and rebuilds the queue panel.
func (s *Screen) highlightActiveUnit() {
	if s.prevActiveUnitID != "" {
		if prev := s.unitByID(s.prevActiveUnitID); prev != nil {
			restoreColor := unitFriendlyBgColor
			if prev.IsOpponent {
				restoreColor = unitEnemyBgColor
			}
			if bc := s.boardCellWidget(prev); bc != nil {
				bc.SetColor(restoreColor)
				bc.RemoveChildren()
				buildBoardCard(bc, prev, false)
			}
		}
	}

	s.setPulseTargets(nil)
	s.prevActiveUnitID = s.activeUnitID

	if s.activeUnitID != "" {
		if u := s.unitByID(s.activeUnitID); u != nil {
			if bc := s.boardCellWidget(u); bc != nil {
				s.setPulseHexTargets([]*ui.HexCellWidget{bc})
				bc.RemoveChildren()
				buildBoardCard(bc, u, !s.activeUnitMoved)
			}
		}
	}

	s.rebuildQueuePanel()
}

// setPulseTargets replaces the queue card pulse list and resets the tick.
func (s *Screen) setPulseTargets(widgets []*widget.Container) {
	s.pulseWidgets = widgets
	s.pulseTick = 0
}

// setPulseHexTargets replaces the hex cell pulse list and resets the tick.
func (s *Screen) setPulseHexTargets(widgets []*ui.HexCellWidget) {
	s.pulseHexWidgets = widgets
	s.pulseTick = 0
}

func (s *Screen) unitByID(id string) *ds.Unit {
	for _, u := range s.unitsQueue {
		if u.ID == id {
			return u
		}
	}

	return nil
}

// unitImage loads the portrait for the given template ID at the specified size.
// Size defaults to 64px if not provided.
func unitImage(templateID int, sizeOpt ...int) *ebiten.Image {
	size := 64
	if len(sizeOpt) > 0 {
		size = sizeOpt[0]
	}

	return asset.Image(path.Join("units", fmt.Sprintf("unit_%d_pic.png", templateID)), size)
}

// buildUnitToolTip constructs the tooltip content for a unit card,
// including icon, name, description, and a stat row (HP, ATK, Move).
func (s *Screen) buildUnitToolTip(u *ds.Unit) *widget.Container {
	c := buildToolTipBase(unitImage(u.TemplateID, 28), u.Name)

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
	stats.AddChild(tooltipStatRow("walk.png", fmt.Sprintf("Move: %d", u.CurrentMP), &toolTipTextTF, mpColor))
	c.AddChild(stats)

	return c
}

// tooltipStatRow returns a horizontal row with a small icon and a coloured label.
// Used inside unit tooltips to display individual stat values.
func tooltipStatRow(iconPath, label string, tf *text.Face, c color.Color) *widget.Container {
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

func (s *Screen) activeUnitAP() int {
	if u := s.unitByID(s.activeUnitID); u != nil {
		return u.CurrentAP
	}
	return 0
}

// hideUnitOnBoard removes the unit portrait from its board cell visually.
// The unit remains in unitsQueue and board state — only the visual is hidden.
func (s *Screen) hideUnitOnBoard(unit *ds.Unit) {
	w := s.boardCellWidgets[unit.Pos]

	if w == nil {
		return
	}

	w.SetColor(boardCellBgColor)
	w.RemoveChildren()
}

// showUnitOnBoard redraws the unit portrait on its current board cell.
func (s *Screen) showUnitOnBoard(unit *ds.Unit) {
	w := s.boardCellWidgets[unit.Pos]
	if w == nil {
		return
	}
	w.RemoveChildren()
	buildBoardCard(w, unit, unit.ID == s.activeUnitID && !s.activeUnitMoved)
}

// unitAtCoord returns the unit at the given hex coord, or nil if the cell is empty.
func (s *Screen) unitAtCoord(coord ds.HexCoord) *ds.Unit {
	cell := s.board.Cells[coord]
	if cell == nil || cell.Unit == nil {
		return nil
	}
	return s.unitByID(cell.Unit.ID)
}

// killUnit marks the unit as dead, removes it from the queue,
// and updates the board cell to show the dead overlay.
func (s *Screen) killUnit(u *ds.Unit) {
	u.IsDead = true

	// Remove from queue.
	for i, qu := range s.unitsQueue {
		if qu.ID == u.ID {
			s.unitsQueue = append(s.unitsQueue[:i], s.unitsQueue[i+1:]...)
			break
		}
	}

	// Update board cell — remove unit card, show dead overlay.
	w := s.boardCellWidgets[u.Pos]
	if w == nil {
		return
	}
	w.RemoveChildren()

	w.SetColor(unitKilledBgColor)

	deadImg := asset.Image(unitKilledPic, unitIconSize)
	w.AddToUnitLayer(widget.NewGraphic(
		widget.GraphicOpts.Image(deadImg),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	))

	s.rebuildQueuePanel()
}

// addUnitStatus adds a status effect to the unit and refreshes its board card.
func (s *Screen) addUnitStatus(u *ds.Unit, statusType status.Type, meta map[string]any) {
	st := status.ByType(statusType)
	if st == nil {
		log.Printf("addUnitStatus: unknown status type %s", statusType)
		return
	}
	if u.Statuses == nil {
		u.Statuses = make(map[status.Type]status.Value)
	}

	u.Statuses[statusType] = status.Value{
		UnitID:   u.ID,
		Duration: st.Duration,
		Value:    st.InitialValue,
		Status:   st,
		Meta:     meta,
	}

	col := colornames.Gold
	if st.Alignment == status.Negative {
		col = colornames.Red
	}
	s.showFloatingText(u.Pos, "+ "+st.Name, col)

	// s.showUnitOnBoard(u)
}

func (s *Screen) updateUnitStatusDuration(u *ds.Unit, statusDur map[status.Type]int) {
	for st, dur := range statusDur {
		sv, ok := u.Statuses[st]
		if !ok {
			continue
		}

		sv.Duration = dur
		u.Statuses[st] = sv
	}
}

// removeUnitStatus removes a status effect from the unit and refreshes its board card.
func (s *Screen) removeUnitStatus(u *ds.Unit, statusType status.Type) {
	st := status.ByType(statusType)
	if st != nil {
		delete(u.Statuses, statusType)
		s.showFloatingText(u.Pos, "- "+st.Name, colornames.White)
	}

	//s.showUnitOnBoard(u)
}

// getProvokingUnitID returns the ID of the unit that provoked this unit, or empty string.
func getProvokingUnitID(u *ds.Unit) string {
	us, ok := u.Statuses[status.Provoked]
	if !ok {
		return ""
	}
	provoker, ok := us.Meta["provoker"].(string)
	if !ok {
		return ""
	}
	return provoker
}
