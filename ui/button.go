// Package ui ...
package ui

import (
	"bytes"
	"log"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/sfx"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// Shared color palette for the main menu UI.
var (
	MenuButtonBgColor        = RGBFromHex("#73A5CA")
	MenuButtonHoverBgColor   = LightenRGB(MenuButtonBgColor, 35)
	MenuButtonTextColor      = RGBFromHex("FFF8DE")
	MenuButtonHoverTextColor = DarkenRGB(MenuButtonBgColor, 45)
)

// regularSource and boldSource are the parsed font faces used to render
// UI text in regular and bold weights.
var (
	regularSource *text.GoTextFaceSource
	boldSource    *text.GoTextFaceSource
)

// init loads the embedded regular and bold font faces used by TextFace
// and TextFaceBold.
func init() {
	var err error
	regularSource, err = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
	}
	boldSource, err = text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		log.Fatal(err)
	}
}

// SecondaryButton creates a smaller menu button styled like
// mainMenuButton, used for secondary actions like Back.
func SecondaryButton(text string, size float64, clickHandler widget.ButtonClickedHandlerFunc) *widget.Button {
	tf := TextFace(size)
	tfHover := TextFace(size + 3)

	var button *widget.Button
	button = widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ButtonOpts.Image(ButtonImage()),
		widget.ButtonOpts.Text(text, &tf, &widget.ButtonTextColor{
			Idle:    MenuButtonTextColor,
			Hover:   MenuButtonHoverTextColor,
			Pressed: MenuButtonTextColor,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   25,
			Right:  25,
			Top:    10,
			Bottom: 10,
		}),
		widget.ButtonOpts.PressedHandler(func(_ *widget.ButtonPressedEventArgs) {
			button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			button.GetWidget().CustomData = true
		}),
		widget.ButtonOpts.ReleasedHandler(func(_ *widget.ButtonReleasedEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
			button.GetWidget().CustomData = false
		}),
		widget.ButtonOpts.ClickedHandler(clickHandler),
		widget.ButtonOpts.CursorEnteredHandler(func(_ *widget.ButtonHoverEventArgs) {
			sfx.Play("button_hover.ogg")
			button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			button.Text().SetFace(&tfHover)
			button.GetWidget().Render(nil)
		}),
		widget.ButtonOpts.CursorExitedHandler(func(_ *widget.ButtonHoverEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
			button.Text().SetFace(&tf)
		}),
	)

	return button
}

// ImageButton creates a smaller menu button styled like
// mainMenuButton, showing an icon instead of text. Used for secondary
// actions like Back where an icon fits better than a label.
func ImageButton(icon *ebiten.Image, clickHandler widget.ButtonClickedHandlerFunc) *widget.Button {
	iconHover := asset.TintImage(icon, MenuButtonHoverTextColor)

	var button *widget.Button
	button = widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ButtonOpts.Image(ButtonImage()),
		widget.ButtonOpts.GraphicPadding(*widget.NewInsetsSimple(10)),

		widget.ButtonOpts.Graphic(&widget.GraphicImage{
			Idle:  icon,
			Hover: iconHover,
		}),
		widget.ButtonOpts.PressedHandler(func(_ *widget.ButtonPressedEventArgs) {
			button.GetWidget().CustomData = true
		}),
		widget.ButtonOpts.ReleasedHandler(func(_ *widget.ButtonReleasedEventArgs) {
			button.GetWidget().CustomData = false
		}),
		widget.ButtonOpts.ClickedHandler(clickHandler),
		widget.ButtonOpts.CursorEnteredHandler(func(_ *widget.ButtonHoverEventArgs) {
			sfx.Play("button_hover.ogg")
		}),
		widget.ButtonOpts.CursorExitedHandler(func(_ *widget.ButtonHoverEventArgs) {
		}),
	)

	return button
}

// ButtonImage returns the nine-slice background images for menu buttons.
func ButtonImage() *widget.ButtonImage {
	idle := image.NewBorderedNineSliceColor(MenuButtonBgColor, DarkenRGB(MenuButtonBgColor, 20), 2)
	hover := image.NewNineSliceColor(MenuButtonHoverBgColor)
	pressed := image.NewNineSliceColor(colornames.Gold)

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}
}

// TextFace returns a regular-weight text face at the given size.
func TextFace(size float64) text.Face {
	return &text.GoTextFace{Source: regularSource, Size: size}
}

// TextFaceBold returns a bold-weight text face at the given size.
func TextFaceBold(size float64) text.Face {
	return &text.GoTextFace{Source: boldSource, Size: size}
}
