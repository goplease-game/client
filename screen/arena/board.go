package arena

import (
	"math"
	"sort"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
)

func (s *Screen) createBoardContainer() *widget.Container {
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

	boardWidget := widget.NewContainer(
		//widget.ContainerOpts.BackgroundImage(
		//	image.NewNineSliceColor(boardBgColor),
		//),
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

	return container
}

func (s *Screen) createCell(coord ds.HexCoord, data *ds.BoardCell) *ui.HexCellWidget {
	isDroppable := data != nil && data.IsSafeZone && data.Unit == nil
	sc := &DropZoneCell{coord: coord}

	widgetOpts := []widget.WidgetOpt{
		widget.WidgetOpts.MinSize(
			int(math.Sqrt(3)*float64(ui.HexRadius)),
			int(2*float64(ui.HexRadius)),
		),
	}

	if isDroppable {
		widgetOpts = append(widgetOpts, s.dropZoneWidgetOpts(sc, coord)...)
		s.safeZoneCells = append(s.safeZoneCells, sc)
	}

	cell := ui.NewHexCellWidget(coord, widgetOpts...)
	cell.SetColor(boardCellBgColor)

	sc.cell = cell
	return cell
}

// dropZoneWidgetOpts returns widget options that make a cell accept unit drops.
func (s *Screen) dropZoneWidgetOpts(sc *DropZoneCell, coord ds.HexCoord) []widget.WidgetOpt {
	return []widget.WidgetOpt{
		widget.WidgetOpts.CanDrop(func(args *widget.DragAndDropDroppedEventArgs) bool {
			_, ok := args.Data.(ds.Unit)
			return ok && s.ready && !sc.occupied && !s.unitPlacedThisTurn
		}),
		widget.WidgetOpts.Dropped(func(args *widget.DragAndDropDroppedEventArgs) {
			unit, ok := args.Data.(ds.Unit)
			if !ok {
				return
			}
			s.onUnitDropped(sc, unit, coord)
		}),
	}
}

// onUnitDropped is called when a unit card is successfully dropped onto a safe-zone cell.
func (s *Screen) onUnitDropped(sc *DropZoneCell, unit ds.Unit, coord ds.HexCoord) {
	sc.occupied = true
	sc.baseColor = unitFriendlyBgColor
	s.unitPlacedThisTurn = true

	if sc.activeGraphic != nil {
		sc.activeGraphic = nil
	}

	sc.cell.SetColor(unitFriendlyBgColor)
	buildBoardCard(sc.cell, unit, false)

	s.onUnitPlaced(unit, coord)
}

func (s *Screen) boardCellWidget(u ds.Unit) *ui.HexCellWidget {
	cell := s.boardCellWidgets[ds.HexCoord{Q: u.Pos.Q, R: u.Pos.R}]
	if cell == nil {
		return nil
	}

	return cell
}

// onCellClicked is the single click handler attached to every board cell.
// It dispatches to the right action depending on current selection state.
func (s *Screen) onCellClicked(coord ds.HexCoord) {
	cell := s.board.Cells[coord]

	// If this cell holds the active friendly unit — select it for movement.
	if cell != nil && cell.Unit != nil &&
		cell.Unit.ID == s.activeUnitID && !cell.Unit.IsOpponent {

		s.selectUnit(*cell.Unit)
		return
	}

	// If a unit is selected and this is a reachable cell — move there.
	if s.selectedUnitID != "" && isReachableHex(s.reachableCells, coord) {
		s.onReachableCellClicked(coord)
		return
	}

	// Any other click clears the selection.
	if s.selectedUnitID != "" {
		s.deselectUnit()
	}
}

// cellsInRange returns all hex positions within rangeN of `from`.
// It uses hex distance (cube/axial equivalent), not square-grid diagonals.
// Unlike movement range, it does NOT consider occupancy.
func cellsInRange(from ds.HexCoord, rangeN int, board ds.Board) []ds.HexCoord {
	var result []ds.HexCoord

	for coord := range board.Cells {
		if HexDistance(from, coord) <= rangeN {
			result = append(result, coord)
		}
	}

	return result
}

func HexDistance(a, b ds.HexCoord) int {
	dq := a.Q - b.Q
	dr := a.R - b.R

	return max3(
		abs(dq),
		abs(dr),
		abs(dq+dr),
	)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max3(a, b, c int) int {
	return max(max(a, b), c)
}
