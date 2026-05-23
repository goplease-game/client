package arena

import (
	"fmt"
	stdImg "image"
	"image/color"
	"path"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/asset"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

// ---------------------------------------------------------------------------
// Unit panel (footer)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Placement
// ---------------------------------------------------------------------------

func (s *Screen) onUnitPlaced(u ds.Unit, coord ds.HexCoord) {
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
		cell.Unit = &u
	}

	s.addUnitToQueue(u.ID)

	s.server.Send(ws.OutMessage{
		Action: ws.UnitPlacedAction,
		Data: ds.UnitPlacedPayload{
			TemplateID: u.TemplateID,
			Coord:      coord,
		},
	})
}

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

// ---------------------------------------------------------------------------
// Queue panel (header)
// ---------------------------------------------------------------------------

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

	// Queue is displayed newest-first.
	for i := len(s.unitsQueue) - 1; i >= 0; i-- {
		uID := s.unitsQueue[i]
		u, ok := s.unitByID(uID)
		if !ok {
			continue
		}
		s.queuePanelRef.AddChild(s.buildQueueCard(u, uID == s.activeUnitID))
	}
}

func (s *Screen) buildQueueCard(u ds.Unit, isActive bool) *widget.Container {
	bgColor := unitFriendlyBgColor
	if u.IsOpponent {
		bgColor = unitEnemyBgColor
	}

	restoreColor := bgColor // capture for closures
	var card *widget.Container
	card = widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(bgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(unitCardSize, 54),
			widget.WidgetOpts.CursorEnterHandler(func(_ *widget.WidgetCursorEnterEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(unitCardHighlightColor))
				if current, ok := s.unitByID(u.ID); ok {
					if bc := s.boardCellWidget(current); bc != nil {
						bc.SetColor(unitCardHighlightColor)
					}
				}
			}),
			widget.WidgetOpts.CursorExitHandler(func(_ *widget.WidgetCursorExitEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(restoreColor))
				if current, ok := s.unitByID(u.ID); ok {
					if bc := s.boardCellWidget(current); bc != nil {
						bc.SetColor(restoreColor)
					}
				}
			}),
		),
	)

	if isActive {
		s.pulseWidgets = append(s.pulseWidgets, card)
	}

	buildBoardCard(NewContainerChildAdder(card), u, false)
	return card
}

// ---------------------------------------------------------------------------
// Highlight / pulse
// ---------------------------------------------------------------------------

func (s *Screen) highlightActiveUnit(unitID string) {
	if s.activeUnitID != "" {
		if prev, ok := s.unitByID(s.activeUnitID); ok {
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
	s.activeUnitID = unitID

	if unitID != "" {
		if u, ok := s.unitByID(unitID); ok {
			if bc := s.boardCellWidget(u); bc != nil {
				s.setPulseHexTargets([]*ui.HexCellWidget{bc})
				bc.RemoveChildren()
				buildBoardCard(bc, u, !s.activeUnitMoved)
			}
		}
	}

	s.rebuildQueuePanel()
}

// setPulseTargets replaces the current pulse widget list and resets the tick.
func (s *Screen) setPulseTargets(widgets []*widget.Container) {
	s.pulseWidgets = widgets
	s.pulseTick = 0
}

// setPulseHexTargets replaces the current pulse hex cell list and resets the tick.
func (s *Screen) setPulseHexTargets(widgets []*ui.HexCellWidget) {
	s.pulseHexWidgets = widgets
	s.pulseTick = 0
}

// ---------------------------------------------------------------------------
// Lookup helpers
// ---------------------------------------------------------------------------

func (s *Screen) unitByID(id string) (ds.Unit, bool) {
	for _, cell := range s.board.Cells {
		if cell == nil || cell.Unit == nil {
			continue
		}
		if cell.Unit.ID == id {
			return *cell.Unit, true
		}
	}

	return ds.Unit{}, false
}

// ---------------------------------------------------------------------------
// Assets
// ---------------------------------------------------------------------------

func unitImage(templateID int, sizeOpt ...int) *ebiten.Image {
	size := 64
	if len(sizeOpt) > 0 {
		size = sizeOpt[0]
	}
	return asset.Image(path.Join("units", fmt.Sprintf("unit_%d_pic.png", templateID)), size)
}

// ---------------------------------------------------------------------------
// Tooltip
// ---------------------------------------------------------------------------

func (s *Screen) buildUnitToolTip(u ds.Unit) *widget.Container {
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
	stats.AddChild(tooltipStatRow("walk.png", fmt.Sprintf("Move: %d", u.MP), &toolTipTextTF, mpColor))
	c.AddChild(stats)

	return c
}

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
