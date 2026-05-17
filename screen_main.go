package game

import (
	"image/color"
	"log"
	"strconv"
	"strings"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
	"golang.org/x/image/colornames"
)

var (
	nameColor                = rgbFromHex("#00a8e8")
	menuButtonBgColor        = rgbFromHex("#73A5CA")
	menuButtonHoverBgColor   = lightenRGB(menuButtonBgColor, 35)
	menuButtonTextColor      = rgbFromHex("FFF8DE")
	menuButtonHoverTextColor = darkenRGB(menuButtonBgColor, 45)
)

func rgbFromHex(hex string) color.Color {
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		log.Fatalf("rgbFromHex: invalid hex length %d", len(hex))
	}

	value, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		log.Fatalf("rgbFromHex: parse hex: %s: %s", len(hex), err)
	}

	return color.NRGBA{
		R: uint8(value >> 16),
		G: uint8(value >> 8),
		B: uint8(value),
		A: 0xff,
	}
}

func lightenRGB(c color.Color, amount int) color.Color {
	rgba := color.NRGBAModel.Convert(c).(color.NRGBA)

	change := func(val uint8) uint8 {
		res := int(val) + amount
		if res > 255 {
			return 255
		}
		if res < 0 {
			return 0
		}
		return uint8(res)
	}

	rgba.R = change(rgba.R)
	rgba.G = change(rgba.G)
	rgba.B = change(rgba.B)

	return rgba
}

func darkenRGB(c color.Color, amount int) color.Color {
	return lightenRGB(c, -amount)
}

// MainScreen is the entry screen with the "Play" button.
type MainScreen struct {
	server     ws.Client
	ui         *ebitenui.UI
	nextScreen Screen
	exit       bool
}

func NewMainScreen(server ws.Client) *MainScreen {
	s := &MainScreen{
		server: server,
	}

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
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

	versionTF := ui.TextFace(12)
	versionText := widget.NewText(
		widget.TextOpts.Text("client v0.0.1\nserver v0.0.1", &versionTF, color.White),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)

	footer.AddChild(versionText)

	root.AddChild(s.mainMenu())
	root.AddChild(footer)

	s.ui = &ebitenui.UI{Container: root}

	return s
}

func (s *MainScreen) Update(g *Game) (Screen, error) {
	if s.exit {
		return nil, ebiten.Termination
	}

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

	titleTF := ui.TextFace(40)
	titleText := widget.NewText(
		widget.TextOpts.Text("go, please", &titleTF, nameColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	playButton, err := mainMenuButton("PLAY", 30, func(args *widget.ButtonClickedEventArgs) {
		s.nextScreen = NewSearchScreen(s.server)
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

	exitButton, err := mainMenuButton("Exit", 14, func(args *widget.ButtonClickedEventArgs) {
		s.exit = true
	})
	if err != nil {
		log.Fatal(err)
	}

	menuC.AddChild(titleText)
	menuC.AddChild(playButton)
	menuC.AddChild(tutButton)
	menuC.AddChild(settButton)
	menuC.AddChild(aboutButton)
	menuC.AddChild(exitButton)

	c.AddChild(menuC)

	return c
}

func mainMenuButton(text string, size float64, clickHandler widget.ButtonClickedHandlerFunc) (*widget.Button, error) {
	tf := ui.TextFace(size)
	tfHover := ui.TextFace(size + 5)
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
		widget.ButtonOpts.Text(text, &tf, &widget.ButtonTextColor{
			Idle:    menuButtonTextColor,
			Hover:   menuButtonHoverTextColor,
			Pressed: menuButtonTextColor,
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
			button.Text().SetPadding(&widget.Insets{Top: 1, Bottom: -1})
			button.Text().SetFace(&tfHover)
			button.GetWidget().Render(nil)
		}),
		widget.ButtonOpts.CursorExitedHandler(func(args *widget.ButtonHoverEventArgs) {
			button.Text().SetPadding(&widget.Insets{})
			button.Text().SetFace(&tf)
		}),
	)

	return button, nil
}

func mainMenuButtonImage() *widget.ButtonImage {
	idle := image.NewNineSliceColor(menuButtonBgColor)
	hover := image.NewNineSliceColor(menuButtonHoverBgColor)
	pressed := image.NewNineSliceColor(colornames.Gold)

	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: pressed,
	}
}
