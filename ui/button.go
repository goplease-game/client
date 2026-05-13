package ui

import (
	"bytes"
	"fmt"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

func Button(text string) (*widget.Button, error) {
	face := TextFace(30)

	var button *widget.Button
	// construct a button.
	button = widget.NewButton(
		// set general widget options
		widget.ButtonOpts.WidgetOpts(
			// instruct the container's anchor layout to center the button both horizontally and vertically.
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.MouseButtonLongPressedHandler(func(args *widget.WidgetMouseButtonLongPressedEventArgs) {
				fmt.Println("Long Press button ", args)
			}),
		),
		// specify the images to use.
		widget.ButtonOpts.Image(buttonImage()),

		// specify the button's text, the font face, and the color.
		widget.ButtonOpts.Text(text, &face, &widget.ButtonTextColor{
			Idle: color.NRGBA{0xdf, 0xf4, 0xff, 0xff},
		}),
		widget.ButtonOpts.TextProcessBBCode(false),
		// specify that the button's text needs some padding for correct display.
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   45,
			Right:  45,
			Top:    15,
			Bottom: 15,
		}),
		// Move the text down and right on press
		widget.ButtonOpts.PressedHandler(func(args *widget.ButtonPressedEventArgs) {
			button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			button.GetWidget().CustomData = true
		}),
		// Move the text back to start on press released
		widget.ButtonOpts.ReleasedHandler(func(args *widget.ButtonReleasedEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
			button.GetWidget().CustomData = false
		}),

		// add a handler that reacts to clicking the button.
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			println("button clicked")
		}),

		// add a handler that reacts to entering the button with the cursor
		widget.ButtonOpts.CursorEnteredHandler(func(args *widget.ButtonHoverEventArgs) {
			println("cursor entered button: entered =", args.Entered, "offsetX =", args.OffsetX, "offsetY =", args.OffsetY)
			// If we moved the Text because we clicked on this button previously, move the text down and right
			if button.GetWidget().CustomData == true {
				button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			}
		}),

		// add a handler that reacts to entering the button with the cursor.
		widget.ButtonOpts.CursorEnteredHandler(func(args *widget.ButtonHoverEventArgs) {
			println("cursor entered button: entered =", args.Entered, "offsetX =", args.OffsetX, "offsetY =", args.OffsetY)
		}),

		// add a handler that reacts to moving the cursor on the button.
		widget.ButtonOpts.CursorMovedHandler(func(args *widget.ButtonHoverEventArgs) {
			println("cursor moved on button: entered =", args.Entered, "offsetX =", args.OffsetX, "offsetY =", args.OffsetY, "diffX =", args.DiffX, "diffY =", args.DiffY)
		}),

		// add a handler that reacts to exiting the button with the cursor.
		widget.ButtonOpts.CursorExitedHandler(func(args *widget.ButtonHoverEventArgs) {
			println("cursor exited button: entered =", args.Entered, "offsetX =", args.OffsetX, "offsetY =", args.OffsetY)
			// Reset the Text inset if the cursor is no longer over the button
			button.Text().SetPadding(&widget.Insets{})
		}),

		// Indicate that this button should not be submitted when enter or space are pressed
		// widget.ButtonOpts.DisableDefaultKeys(),
	)

	return button, nil
}

func buttonImage() *widget.ButtonImage {
	idle := image.NewNineSliceColor(color.NRGBA{R: 170, G: 170, B: 180, A: 255})

	hover := image.NewBorderedNineSliceColor(color.NRGBA{R: 130, G: 130, B: 150, A: 255}, color.NRGBA{70, 70, 70, 255}, 1)

	pressed := image.NewAdvancedNineSliceColor(color.NRGBA{R: 130, G: 130, B: 150, A: 255}, image.NewBorder(3, 2, 2, 2, color.NRGBA{70, 70, 70, 255}))

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}
}

func TextFace(size float64) text.Face {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(fmt.Errorf("loading font: %w", err))
	}

	return &text.GoTextFace{
		Source: s,
		Size:   size,
	}
}
