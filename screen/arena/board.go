package arena

import (
	"math"
	"sort"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// boardShadowOffsetX and boardShadowOffsetY compensate for the padding baked
// around the board silhouette in the shadow PNG asset (assets/board_shadow.png).
// The asset's silhouette doesn't start at (0,0) — it's inset by this amount
// on each side to leave room for the shadow to spread past the board's edges.
// Subtracted from boardWidgetRef's Rect.Min when positioning the shadow, so
// the asset's silhouette aligns exactly with the real hex grid regardless of
// where the board is currently anchored on screen.
const (
	boardShadowOffsetX = 18
	boardShadowOffsetY = 14
)

// createBoardContainer builds the widget.Container that holds all hex cell
// widgets. It also populates boardCellWidgets and sortedCells.
func (s *Screen) createBoardContainer() *widget.Container {
	// Outer container stretches to fill the space between header and footer.
	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
				Padding: &widget.Insets{
					Top:    headerH,
					Bottom: footerH + statusH,
				},
			}),
		),
	)

	// Inner container uses HexLayout to position cells by axial coordinate.
	boardWidget := widget.NewContainer(
		widget.ContainerOpts.Layout(&ui.HexLayout{
			HexSize: ui.HexRadius,
		}),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	s.boardCellWidgets = make(map[ds.HexCoord]*ui.HexCellWidget)

	for coord, cellData := range s.board.Cells {
		cell := s.createCell(coord, cellData)
		s.boardCellWidgets[coord] = cell
		boardWidget.AddChild(cell)
	}

	// sortedCells provides a deterministic render order (row-major, left to right)
	// so that unit overlays and HUD badges always draw consistently at hex borders.
	s.sortedCells = make([]*ui.HexCellWidget, 0, len(s.boardCellWidgets))
	for _, cell := range s.boardCellWidgets {
		s.sortedCells = append(s.sortedCells, cell)
	}
	sort.Slice(s.sortedCells, func(i, j int) bool {
		if s.sortedCells[i].Coord.R != s.sortedCells[j].Coord.R {
			return s.sortedCells[i].Coord.R < s.sortedCells[j].Coord.R
		}
		return s.sortedCells[i].Coord.Q < s.sortedCells[j].Coord.Q
	})

	container.AddChild(boardWidget)

	s.boardContainerRef = container
	s.boardWidgetRef = boardWidget
	return container
}

// createCell creates a HexCellWidget for the given coordinate and cell data.
// Safe-zone cells receive additional drop-target widget options.
func (s *Screen) createCell(coord ds.HexCoord, data *ds.BoardCell) *ui.HexCellWidget {
	isDroppable := data != nil && data.IsSafeZone
	sc := &DropZoneCell{coord: coord}

	widgetOpts := []widget.WidgetOpt{
		widget.WidgetOpts.MinSize(
			int(math.Sqrt(3)*float64(ui.HexRadius)),
			int(2*float64(ui.HexRadius)),
		),
	}

	if isDroppable {
		widgetOpts = append(widgetOpts, s.dropZoneWidgetOpts(sc, coord)...)
		s.dropZoneCells = append(s.dropZoneCells, sc)
	}

	cell := ui.NewHexCellWidget(coord, s.board.Cells, widgetOpts...)
	cell.SetColor(boardCellBgColor)

	sc.cell = cell
	return cell
}

// dropZoneWidgetOpts returns widget options that make a cell accept unit drops
// via EbitenUI drag-and-drop. A drop is accepted only when the game is ready,
// the cell is unoccupied, and the player has not yet placed a unit this turn.
func (s *Screen) dropZoneWidgetOpts(sc *DropZoneCell, coord ds.HexCoord) []widget.WidgetOpt {
	return []widget.WidgetOpt{
		widget.WidgetOpts.CanDrop(func(args *widget.DragAndDropDroppedEventArgs) bool {
			_, ok := args.Data.(*ds.Unit)
			cell := s.board.Cells[coord]
			return ok && s.ready && !sc.occupied && !s.unitPlacedThisTurn &&
				cell != nil && cell.Unit == nil
		}),
		widget.WidgetOpts.Dropped(func(args *widget.DragAndDropDroppedEventArgs) {
			unit, ok := args.Data.(*ds.Unit)
			if !ok {
				return
			}
			s.onUnitDropped(sc, unit, coord)
		}),
	}
}

// onUnitDropped is called when a unit card is successfully dropped onto a safe-zone cell.
// It marks the cell as occupied, renders the unit card, and notifies the server.
func (s *Screen) onUnitDropped(sc *DropZoneCell, unit *ds.Unit, coord ds.HexCoord) {
	sc.occupied = true
	sc.baseColor = unitFriendlyBgColor
	s.unitPlacedThisTurn = true
	sc.activeGraphic = nil

	sc.cell.SetColor(unitFriendlyBgColor)
	s.buildBoardCard(sc.cell, unit)

	s.onUnitPlaced(unit, coord)
}

// boardCellWidget returns the HexCellWidget for the cell occupied by unit u,
// or nil if the cell does not exist.
func (s *Screen) boardCellWidget(u *ds.Unit) *ui.HexCellWidget {
	return s.boardCellWidgets[ds.HexCoord{Q: u.Pos.Q, R: u.Pos.R}]
}

// onCellClicked is the single entry point for board cell clicks.
// It dispatches based on the current selection state:
//   - clicking the active unit selects it for movement
//   - clicking a reachable cell moves the selected unit there
//   - any other click clears the current selection
func (s *Screen) onCellClicked(coord ds.HexCoord) {
	// If an ability is selected — try to apply it to this cell.
	if s.onCellClickedWithAbility(coord) {
		return
	}

	cell := s.board.Cells[coord]

	if cell != nil && cell.Unit != nil {
		s.infoPanelUnit = cell.Unit
		s.showInfoPanel(s.buildUnitInfoPanel(cell.Unit))
	}

	if cell != nil && cell.Unit != nil &&
		cell.Unit.ID == s.activeUnitID && !cell.Unit.IsOpponent {
		s.selectUnit(cell.Unit)
		s.showAbilityPanel(cell.Unit)
		return
	}

	if s.selectedUnitID != "" && isReachableHex(s.reachableCells, coord) {
		s.onReachableCellClicked(coord)
		return
	}

	if s.selectedUnitID != "" {
		u := s.unitByID(s.selectedUnitID)
		s.deselectUnit()
		s.showAbilityPanel(u)
	}
}

// drawBoard bakes all hex fills into a single offscreen image on first call
// (or after InvalidateBoardImage), then draws the shadow and the board with
// one DrawImage call each instead of per-cell rendering every frame.
func (s *Screen) drawBoard(screen *ebiten.Image) {
	for _, cell := range s.boardCellWidgets {
		cell.RenderFill(screen)
	}

	s.renderGrid(screen)

	boardRect := s.boardWidgetRef.GetWidget().Rect
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		float64(boardRect.Min.X-boardShadowOffsetX),
		float64(boardRect.Min.Y-boardShadowOffsetY),
	)

	screen.DrawImage(asset.Image("board-shadow.png"), op)
}
