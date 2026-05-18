package arena

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
)

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
		opts = append(opts, s.dropZoneOpts(sc, r, c)...)
	}

	cell := widget.NewContainer(opts...)
	sc.container = cell
	if isDroppable {
		s.safeZoneCells = append(s.safeZoneCells, sc)
	}

	return cell
}

// dropZoneOpts returns widget options that make a cell accept unit drops.
func (s *Screen) dropZoneOpts(sc *DropZoneCell, r, c int) []widget.ContainerOpt {
	return []widget.ContainerOpt{
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.CanDrop(func(args *widget.DragAndDropDroppedEventArgs) bool {
				_, ok := args.Data.(ds.Unit)
				return ok && s.ready && !sc.occupied && !s.unitPlacedThisTurn
			}),
			widget.WidgetOpts.Dropped(func(args *widget.DragAndDropDroppedEventArgs) {
				unit, ok := args.Data.(ds.Unit)
				if !ok {
					return
				}
				s.onUnitDropped(sc, unit, r, c)
			}),
		),
	}
}

// onUnitDropped is called when a unit card is successfully dropped onto a safe-zone cell.
func (s *Screen) onUnitDropped(sc *DropZoneCell, unit ds.Unit, r, c int) {
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
}

func (s *Screen) boardCellWidget(u ds.Unit) *widget.Container {
	if u.Row < 0 || u.Row >= len(s.boardCellWidgets) ||
		u.Col < 0 || u.Col >= len(s.boardCellWidgets[u.Row]) {
		return nil
	}
	return s.boardCellWidgets[u.Row][u.Col]
}
