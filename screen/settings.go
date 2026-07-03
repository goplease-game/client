package screen

import (
	"fmt"
	stdimage "image"
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
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// SettingsScreen shows game settings, split into a "General" tab (display,
// sound, gameplay toggles) and a "Keybinding" tab (placeholder for now).
type SettingsScreen struct {
	previous   game.Screen
	ui         *ebitenui.UI
	nextScreen game.Screen

	volumeSlider *widget.Slider
	volumeLabel  *widget.Text

	fullscreenCheckbox        *widget.Checkbox
	showGameLogCheckbox       *widget.Checkbox
	autoShowInfoPanelCheckbox *widget.Checkbox
}

// NewSettingsScreen creates the settings screen. previous is the screen
// to return to when the player presses Save.
func NewSettingsScreen(previous game.Screen) *SettingsScreen {
	s := &SettingsScreen{previous: previous}

	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0x13, 0x1a, 0x22, 0xff})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	panel := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(25),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionStart,
			}),
			widget.WidgetOpts.MinSize(460, 0),
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
	panel.AddChild(title)
	panel.AddChild(s.buildTabs())

	root.AddChild(panel)

	s.ui = &ebitenui.UI{Container: root}
	return s
}

// buildTabs constructs the General | Keybinding tab book.
func (s *SettingsScreen) buildTabs() *widget.TabBook {
	generalTab := widget.NewTabBookTab(
		widget.TabBookTabOpts.Label("General"),
		widget.TabBookTabOpts.ContainerOpts(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout(
				widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(10)),
			)),
		),
	)
	generalTab.AddChild(s.buildGeneralTab())

	keybindingFace := ui.TextFace(14)
	keybindingTab := widget.NewTabBookTab(
		widget.TabBookTabOpts.Label("Keybinding"),
		widget.TabBookTabOpts.ContainerOpts(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		),
	)
	keybindingTab.AddChild(widget.NewText(
		widget.TextOpts.Text("Coming soon.", &keybindingFace, ui.RGBFromHex("#8B98A5")),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})),
	))

	tabButtonImg := tabButtonImage()
	tabFace := ui.TextFace(16)

	return widget.NewTabBook(
		widget.TabBookOpts.TabButtonImage(tabButtonImg),
		widget.TabBookOpts.TabButtonText(&tabFace, &widget.ButtonTextColor{Idle: color.White, Disabled: color.White}),
		widget.TabBookOpts.TabButtonSpacing(5),
		widget.TabBookOpts.ContentPadding(widget.NewInsetsSimple(10)),
		widget.TabBookOpts.ContentSpacing(10),
		widget.TabBookOpts.TabButtonMinSize(&stdimage.Point{X: 130, Y: 36}),
		widget.TabBookOpts.ContainerOpts(
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			})),
		),
		widget.TabBookOpts.Tabs(generalTab, keybindingTab),
		widget.TabBookOpts.InitialTab(generalTab),
	)
}

// buildGeneralTab builds the General tab.
func (s *SettingsScreen) buildGeneralTab() *widget.Container {
	cfg := config.Get()
	labelFace := ui.TextFace(16)
	descFace := ui.TextFace(14)

	wrapper := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
		)),
	)

	table := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, nil),
			widget.GridLayoutOpts.Spacing(50, 30),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(400, 0),
		),
	)

	addRow := func(name, desc string, control widget.PreferredSizeLocateableWidget) {
		label := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Spacing(2),
			)),
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			})),
		)
		label.AddChild(widget.NewText(widget.TextOpts.Text(name, &labelFace, color.White)))
		if desc != "" {
			label.AddChild(widget.NewText(widget.TextOpts.Text(desc, &descFace, ui.RGBFromHex("#8B98A5"))))
		}
		table.AddChild(label)

		controlCell := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionEnd,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			})),
		)
		controlCell.AddChild(control)
		table.AddChild(controlCell)
	}

	// --- Fullscreen ---
	s.fullscreenCheckbox = settingsCheckbox(cfg.Fullscreen, func(checked bool) {
		ebiten.SetFullscreen(checked)
		config.Get().Fullscreen = checked
	})
	addRow("Fullscreen", "", s.fullscreenCheckbox)

	// --- Sound volume ---
	s.volumeLabel = widget.NewText(
		widget.TextOpts.Text(volumeLabelText(sfx.Volume()), &labelFace, color.White),
	)
	s.volumeSlider = widget.NewSlider(
		widget.SliderOpts.Orientation(widget.DirectionHorizontal),
		widget.SliderOpts.MinMax(0, 100),
		widget.SliderOpts.Images(sliderTrackImage(), sliderHandleImage()),
		widget.SliderOpts.FixedHandleSize(20),
		widget.SliderOpts.TrackPadding(widget.NewInsetsSimple(2)),
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.MinSize(180, 0)),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			volume := float64(args.Current) / 100
			sfx.SetVolume(volume)
			s.volumeLabel.Label = volumeLabelText(volume)
			config.Get().Volume = volume
		}),
	)
	s.volumeSlider.Current = int(sfx.Volume() * 100)

	volumeControl := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
		)),
	)
	volumeControl.AddChild(s.volumeLabel)
	volumeControl.AddChild(s.volumeSlider)
	addRow("Sound volume", "", volumeControl)

	// --- Show game log ---
	s.showGameLogCheckbox = settingsCheckbox(cfg.ShowGameLog, func(checked bool) {
		config.Get().ShowGameLog = checked
	})
	addRow("Show game log", "Display the game log panel on the left side of the screen.", s.showGameLogCheckbox)

	// --- Auto show info panel ---
	s.autoShowInfoPanelCheckbox = settingsCheckbox(cfg.AutoShowInfoPanel, func(checked bool) {
		config.Get().AutoShowInfoPanel = checked
	})
	addRow("Auto show info panel", "Open the unit details panel automatically for active unit.", s.autoShowInfoPanelCheckbox)

	// --- Tutorial ---
	tutorialResetButton := secondaryButton("Reset", 12, func(_ *widget.ButtonClickedEventArgs) {
		config.Get().SkipTutorial = false
	})
	addRow("Tutorial", "Show the tutorial again next time you start a practice match.", tutorialResetButton)

	// --- Save / Reset settings ---
	actions := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(12),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		})),
	)
	actions.AddChild(secondaryButton("Save", 14, func(_ *widget.ButtonClickedEventArgs) {
		if err := config.Save(); err != nil {
			log.Printf("settings: failed to save config: %v", err)
		}
		s.nextScreen = s.previous
	}))
	actions.AddChild(secondaryButton("Return", 14, func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = s.previous
	}))
	actions.AddChild(secondaryButton("Reset settings", 14, func(_ *widget.ButtonClickedEventArgs) {
		s.resetToDefaults()
	}))

	wrapper.AddChild(table)
	wrapper.AddChild(actions)

	return wrapper
}

// resetToDefaults restores every General setting to its default value.
func (s *SettingsScreen) resetToDefaults() {
	cfg := config.Get()

	cfg.Fullscreen = false
	cfg.Volume = 0.5
	cfg.ShowGameLog = true
	cfg.AutoShowInfoPanel = true

	ebiten.SetFullscreen(cfg.Fullscreen)
	sfx.SetVolume(cfg.Volume)

	setCheckboxState(s.fullscreenCheckbox, cfg.Fullscreen)
	setCheckboxState(s.showGameLogCheckbox, cfg.ShowGameLog)
	setCheckboxState(s.autoShowInfoPanelCheckbox, cfg.AutoShowInfoPanel)

	s.volumeSlider.Current = int(cfg.Volume * 100)
	s.volumeLabel.Label = volumeLabelText(cfg.Volume)
}

// Update advances the settings UI.
func (s *SettingsScreen) Update(_ *game.Game) (game.Screen, error) {
	s.ui.Update()

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.nextScreen = s.previous
	}

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

func volumeLabelText(volume float64) string {
	return fmt.Sprintf("Volume: %d%%", int(volume*100))
}

func sliderTrackImage() *widget.SliderTrackImage {
	idle := image.NewNineSliceColor(ui.RGBFromHex("#2A3540"))
	return &widget.SliderTrackImage{
		Idle:     idle,
		Hover:    idle,
		Disabled: idle,
	}
}

func sliderHandleImage() *widget.ButtonImage {
	idle := image.NewNineSliceColor(menuButtonBgColor)
	hover := image.NewNineSliceColor(menuButtonHoverBgColor)
	return &widget.ButtonImage{
		Idle:    idle,
		Hover:   hover,
		Pressed: hover,
	}
}

func tabButtonImage() *widget.ButtonImage {
	idle := image.NewNineSliceColor(ui.RGBFromHex("#1D262F"))
	hover := image.NewNineSliceColor(ui.RGBFromHex("#2A3540"))
	pressed := image.NewNineSliceColor(menuButtonBgColor)
	disabled := image.NewNineSliceColor(ui.RGBFromHex("#12181E"))
	return &widget.ButtonImage{
		Idle:     idle,
		Hover:    hover,
		Pressed:  pressed,
		Disabled: disabled,
	}
}

// settingsCheckbox builds a checkbox pre-set to initial under the new API.
func settingsCheckbox(initial bool, onChange func(checked bool)) *widget.Checkbox {
	cb := widget.NewCheckbox(
		widget.CheckboxOpts.Image(buildCheckboxImages()),
		widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
			onChange(args.State == widget.WidgetChecked)
		}),
	)
	setCheckboxState(cb, initial)
	return cb
}

// setCheckboxState sets cb using the new widget.WidgetState constants.
func setCheckboxState(cb *widget.Checkbox, checked bool) {
	if checked {
		cb.SetState(widget.WidgetChecked)
	} else {
		cb.SetState(widget.WidgetUnchecked)
	}
}

func buildCheckboxImages() *widget.CheckboxImage {
	bgIdleColor := ui.RGBFromHex("#2A3540")
	bgHoverColor := ui.RGBFromHex("#354250")
	markColor := menuButtonBgColor

	uncheckedIdle := checkboxGraphic(bgIdleColor, color.NRGBA{0, 0, 0, 0})
	uncheckedHover := checkboxGraphic(bgHoverColor, color.NRGBA{0, 0, 0, 0})
	checkedIdle := checkboxGraphic(bgIdleColor, markColor)
	checkedHover := checkboxGraphic(bgHoverColor, markColor)

	return &widget.CheckboxImage{
		Unchecked:         image.NewFixedNineSlice(uncheckedIdle),
		UncheckedHovered:  image.NewFixedNineSlice(uncheckedHover),
		UncheckedDisabled: image.NewFixedNineSlice(uncheckedIdle),

		Checked:         image.NewFixedNineSlice(checkedIdle),
		CheckedHovered:  image.NewFixedNineSlice(checkedHover),
		CheckedDisabled: image.NewFixedNineSlice(checkedIdle),

		Greyed:         image.NewFixedNineSlice(uncheckedIdle),
		GreyedHovered:  image.NewFixedNineSlice(uncheckedHover),
		GreyedDisabled: image.NewFixedNineSlice(uncheckedIdle),
	}
}

func checkboxGraphic(bg, fg color.Color) *ebiten.Image {
	const size = 20
	img := ebiten.NewImage(size, size)
	img.Fill(bg)

	if _, _, _, a := fg.RGBA(); a > 0 {
		const padding = 4
		mark := ebiten.NewImage(size-padding*2, size-padding*2)
		mark.Fill(fg)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(padding, padding)
		img.DrawImage(mark, op)
	}

	return img
}
