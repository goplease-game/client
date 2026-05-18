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

// ---------------------------------------------------------------------------
// Placement
// ---------------------------------------------------------------------------

func (s *Screen) onUnitPlaced(u ds.Unit, r, c int) {
	s.removeUnitCard(u.ID)

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
			widget.WidgetOpts.CursorEnterHandler(func(_ *widget.WidgetCursorEnterEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(unitCardHighlightColor))
				if bc := s.boardCellWidget(u); bc != nil {
					bc.SetBackgroundImage(image.NewNineSliceColor(unitCardHighlightColor))
				}
			}),
			widget.WidgetOpts.CursorExitHandler(func(_ *widget.WidgetCursorExitEventArgs) {
				card.SetBackgroundImage(image.NewNineSliceColor(restoreColor))
				if bc := s.boardCellWidget(u); bc != nil {
					bc.SetBackgroundImage(image.NewNineSliceColor(restoreColor))
				}
			}),
		),
	))

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
				bc.SetBackgroundImage(image.NewNineSliceColor(restoreColor))
			}
		}
	}

	s.setPulseTargets(nil)
	s.activeUnitID = unitID

	if unitID != "" {
		if u, ok := s.unitByID(unitID); ok {
			if bc := s.boardCellWidget(u); bc != nil {
				s.setPulseTargets([]*widget.Container{bc})
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

// ---------------------------------------------------------------------------
// Lookup helpers
// ---------------------------------------------------------------------------

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
