// Package ui ...
package ui

import (
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/asset"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	panelShadowCorner = 20
	panelShadowSpread = 20

	panelTitleBgColor    = "547292"
	panelTitleTextColor  = "244463"
	panelWidth           = 400
	panelContentBgColor  = "5d7b9b"
	panelControlsBgColor = "547292"

	panelBorderWidth = 5
	panelBorderColor = "FFFFFF"
)

// Panel is a builder for a centered, shadowed panel with a title bar,
// a content area, and a controls area stacked vertically. Use NewPanel
// to create one, AddContent/AddControl to populate it, and Build to
// obtain the finished container.
type Panel struct {
	titleText     *widget.Text
	title         *widget.Container // outer AnchorLayout: bg + inner
	content       *widget.Container // outer AnchorLayout: bg + inner
	contentInner  *widget.Container
	controls      *widget.Container // outer AnchorLayout: bg + inner
	controlsInner *widget.Container
}

// NewPanel creates a Panel with the given title. The content and controls
// areas start empty; populate them with AddContent and AddControl before
// calling Build.
func NewPanel(title string) *Panel {
	p := &Panel{}

	// Title
	titleOuter, titleInner := NoiseContainer(
		panelTitleBgColor,
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(15)),
		)),
	)

	titleTF := TextFace(30)
	titleText := widget.NewText(
		widget.TextOpts.Text(title, &titleTF, RGBFromHex(panelTitleTextColor)),
	)
	titleInner.AddChild(titleText)

	p.title = titleOuter
	p.titleText = titleText

	// Content
	contentOuter, contentInner := NoiseContainer(
		panelContentBgColor,
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(25)),
		)),
	)
	p.content = contentOuter
	p.contentInner = contentInner

	// Controls
	controlsOuter, controlsInner := NoiseContainer(
		panelControlsBgColor,
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(15)),
			widget.RowLayoutOpts.Spacing(10),
		)),
	)
	p.controls = controlsOuter
	p.controlsInner = controlsInner

	return p
}

// AddContent appends the given widgets to the panel's content area.
func (p *Panel) AddContent(data ...widget.PreferredSizeLocateableWidget) {
	for _, c := range data {
		p.contentInner.AddChild(c)
	}
}

// AddControl appends the given widgets to the panel's controls area.
func (p *Panel) AddControl(data ...widget.PreferredSizeLocateableWidget) {
	for _, c := range data {
		p.controlsInner.AddChild(c)
	}
}

// Title updates the panel's title.
func (p *Panel) Title(t string) {
	p.titleText.Label = t
}

// Build assembles the panel into a single container.
func (p *Panel) Build() *widget.Container {
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
	)

	shadowWrap := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(asset.NineSlice("main_window_shadow.png", panelShadowCorner)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(panelShadowSpread)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(panelWidth, 0),
		),
	)

	frame := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(
			image.NewBorderedNineSliceColor(color.NRGBA{0, 0, 0, 0}, RGBFromHex(panelBorderColor, 10), panelBorderWidth),
		),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
			widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(panelBorderWidth)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
	)

	stack := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(0),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
	)

	stretchChild(p.title)
	stretchChild(p.content)
	stretchChild(p.controls)

	stack.AddChild(p.title)
	stack.AddChild(p.content)
	stack.AddChild(p.controls)

	frame.AddChild(stack)
	shadowWrap.AddChild(frame)
	root.AddChild(shadowWrap)

	return root
}

// NoiseContainer builds an AnchorLayout container ("outer") holding a
// tiled-noise background layer and an "inner" container laid
// out with innerOpts. inner is stretched to fill outer, and the
// background is size-proxied to inner so outer reports inner's
// preferred size. outer is meant to be added to the panel's vertical
// stack; children belong in inner.
func NoiseContainer(bgColor string, innerOpts ...widget.ContainerOpt) (outer, inner *widget.Container) {
	opts := append([]widget.ContainerOpt(nil), innerOpts...)
	opts = append(opts,
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
	)
	inner = widget.NewContainer(opts...)

	bg := NewTiledBackground(
		NoiseBg(),
		RGBFromHex(bgColor),
		inner,
		widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		}),
	)

	outer = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	outer.AddChild(bg)
	outer.AddChild(inner)

	return outer, inner
}

// NoiseBg returns the cached noise texture used as a subtle background overlay.
func NoiseBg() *ebiten.Image {
	return asset.Image("noise-bg.png")
}

// stretchChild sets RowLayoutData.Stretch on c's widget so it fills the
// full width of a vertical RowLayout parent.
func stretchChild(c *widget.Container) {
	c.GetWidget().LayoutData = widget.RowLayoutData{
		Stretch: true,
	}
}
