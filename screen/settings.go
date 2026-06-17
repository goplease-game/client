package screen

import (
	"fmt"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/config"
	"github.com/goplease-game/client/sfx"
	"github.com/goplease-game/client/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// SettingsScreen shows game settings. Currently only sound volume.
type SettingsScreen struct {
	previous   game.Screen
	ui         *ebitenui.UI
	nextScreen game.Screen
}

// NewSettingsScreen creates the settings screen. previous is the screen
// to return to when the player presses Back.
func NewSettingsScreen(previous game.Screen) *SettingsScreen {
	s := &SettingsScreen{previous: previous}

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	panel := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(15),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(300, 0),
		),
	)

	titleTF := ui.TextFace(30)
	title := widget.NewText(
		widget.TextOpts.Text("Settings", &titleTF, nameColor),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		})),
	)

	labelTF := ui.TextFace(16)
	volumeLabel := widget.NewText(
		widget.TextOpts.Text(volumeLabelText(sfx.Volume()), &labelTF, color.White),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		})),
	)

	volumeSlider := widget.NewSlider(
		widget.SliderOpts.Orientation(widget.DirectionHorizontal),
		widget.SliderOpts.MinMax(0, 100),
		widget.SliderOpts.Images(sliderTrackImage(), sliderHandleImage()),
		widget.SliderOpts.FixedHandleSize(20),
		widget.SliderOpts.TrackPadding(widget.NewInsetsSimple(2)),
		widget.SliderOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(260, 0),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			volume := float64(args.Current) / 100

			sfx.SetVolume(volume)
			volumeLabel.Label = volumeLabelText(volume)
			config.Get().Volume = volume
		}),
	)
	volumeSlider.Current = int(sfx.Volume() * 100)

	backButton := secondaryButton("Back", 12, func(_ *widget.ButtonClickedEventArgs) {
		err := config.Save()
		if err != nil {
			log.Printf("settings: failed to save config: %v", err)
		}
		s.nextScreen = s.previous
	})

	panel.AddChild(title)
	panel.AddChild(volumeLabel)
	panel.AddChild(volumeSlider)
	panel.AddChild(backButton)

	root.AddChild(panel)

	s.ui = &ebitenui.UI{Container: root}

	return s
}

// Update advances the settings UI and returns the next screen to
// transition to once the player presses Back.
func (s *SettingsScreen) Update(_ *game.Game) (game.Screen, error) {
	s.ui.Update()

	if s.nextScreen != nil {
		next := s.nextScreen
		s.nextScreen = nil
		return next, nil
	}

	return s, nil
}

// Draw renders the settings UI to screen.
func (s *SettingsScreen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)
}

// volumeLabelText formats volume (0.0-1.0) as a percentage label,
// e.g. "Volume: 50%".
func volumeLabelText(volume float64) string {
	return fmt.Sprintf("Volume: %d%%", int(volume*100))
}

// sliderTrackImage returns the background image for the volume slider's track.
func sliderTrackImage() *widget.SliderTrackImage {
	idle := image.NewNineSliceColor(ui.RGBFromHex("#2A3540"))
	return &widget.SliderTrackImage{
		Idle:     idle,
		Hover:    idle,
		Disabled: idle,
	}
}

// sliderHandleImage returns the button image for the volume slider's draggable handle.
func sliderHandleImage() *widget.ButtonImage {
	idle := image.NewNineSliceColor(menuButtonBgColor)
	hover := image.NewNineSliceColor(menuButtonHoverBgColor)
	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: hover,
	}
}
