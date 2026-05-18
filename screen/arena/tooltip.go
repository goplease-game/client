package arena

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// buildToolTipBase creates the common tooltip shell: a bordered container
// with a header row consisting of an icon and a title. Both abilities.go and
// units.go call this and then append their own content rows.
func buildToolTipBase(icon *ebiten.Image, title string) *widget.Container {
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
		widget.GraphicOpts.Image(icon),
		widget.GraphicOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(24, 24),
		),
	))
	header.AddChild(widget.NewText(
		widget.TextOpts.Text(title, &toolTipTitleTF, ttTitleColor),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	))

	c.AddChild(header)

	return c
}
