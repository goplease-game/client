package game

import (
	"fmt"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"golang.org/x/image/colornames"
)

// MainScreen is the entry screen with the "Play" button.
type MainScreen struct {
	ui         *ebitenui.UI
	nextScreen Screen
}

func NewMainScreen() *MainScreen {
	s := &MainScreen{}

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout(
		//widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(50)),
		)),
	)

	footer := widget.NewContainer(
		//widget.ContainerOpts.BackgroundImage(
		//	image.NewNineSliceColor(colornames.Steelblue),
		//),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				StretchHorizontal:  true,
			}),
		),
	)

	versionFace, err := ui.TextFace(12)
	if err != nil {
		log.Fatal(err)
	}

	versionText := widget.NewText(
		widget.TextOpts.Text("client v0.0.1\nserver v0.0.1", &versionFace, color.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	footer.AddChild(versionText)

	root.AddChild(s.mainMenu())
	root.AddChild(footer)

	s.ui = &ebitenui.UI{Container: root}

	return s
}

func (s *MainScreen) Update(g *Game) (Screen, error) {
	s.ui.Update()

	if s.nextScreen != nil {
		next := s.nextScreen
		s.nextScreen = nil
		g.Server.Connect(g.PlayerID)
		return next, nil
	}

	return s, nil
}

func (s *MainScreen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

func (s *MainScreen) mainMenu() *widget.Container {
	c := widget.NewContainer(
		//widget.ContainerOpts.BackgroundImage(
		//	image.NewNineSliceColor(colornames.Steelblue),
		//),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				StretchHorizontal:  true,
			}),
		),
	)

	menuC := widget.NewContainer(
		//widget.ContainerOpts.BackgroundImage(
		//	image.NewNineSliceColor(colornames.Goldenrod),
		//),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(5),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(300, 0),
		),
	)

	titleFace, err := ui.TextFace(40)
	if err != nil {
		log.Fatal(err)
	}

	titleText := widget.NewText(
		widget.TextOpts.Text("go, please", &titleFace, colornames.Deepskyblue),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	playButton, err := mainMenuButton("PLAY", 30, func(args *widget.ButtonClickedEventArgs) {
		s.nextScreen = NewSearchScreen()
	})
	if err != nil {
		log.Fatal(err)
	}

	tutButton, err := mainMenuButton("How to play", 16, func(args *widget.ButtonClickedEventArgs) {
		println("I DON'T KNOW HOW TO PLAY")
	})
	if err != nil {
		log.Fatal(err)
	}

	settButton, err := mainMenuButton("Settings", 16, func(args *widget.ButtonClickedEventArgs) {
		println("NO SETTINGS YET")
	})
	if err != nil {
		log.Fatal(err)
	}

	aboutButton, err := mainMenuButton("About", 16, func(args *widget.ButtonClickedEventArgs) {
		println("ABOUT WHAT?")
	})
	if err != nil {
		log.Fatal(err)
	}

	menuC.AddChild(titleText)
	menuC.AddChild(playButton)
	menuC.AddChild(tutButton)
	menuC.AddChild(settButton)
	menuC.AddChild(aboutButton)

	c.AddChild(menuC)

	return c
}

func mainMenuButton(text string, size float64, clickHandler widget.ButtonClickedHandlerFunc) (*widget.Button, error) {
	face, err := ui.TextFace(size)
	if err != nil {
		return nil, fmt.Errorf("create button: %w", err)
	}

	var button *widget.Button
	button = widget.NewButton(
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
		widget.ButtonOpts.Image(mainMenuButtonImage()),
		widget.ButtonOpts.Text(text, &face, &widget.ButtonTextColor{
			Idle:    colornames.White,
			Hover:   colornames.Black,
			Pressed: colornames.Black,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left:   45,
			Right:  45,
			Top:    15,
			Bottom: 15,
		}),
		widget.ButtonOpts.PressedHandler(func(args *widget.ButtonPressedEventArgs) {
			button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			button.GetWidget().CustomData = true
		}),
		widget.ButtonOpts.ReleasedHandler(func(args *widget.ButtonReleasedEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
			button.GetWidget().CustomData = false
		}),
		widget.ButtonOpts.ClickedHandler(clickHandler),
		widget.ButtonOpts.CursorEnteredHandler(func(args *widget.ButtonHoverEventArgs) {
			if button.GetWidget().CustomData == true {
				button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			}
		}),
	)

	return button, nil
}

func mainMenuButtonImage() *widget.ButtonImage {
	idle := image.NewNineSliceColor(colornames.Steelblue)

	hover := image.NewBorderedNineSliceColor(colornames.White, color.NRGBA{70, 70, 70, 255}, 1)

	pressed := image.NewAdvancedNineSliceColor(colornames.Gold, image.NewBorder(3, 2, 2, 2, color.NRGBA{70, 70, 70, 255}))

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}
}
