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
	"github.com/goplease-game/client/backdrop"
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
	bg         backdrop.Backdrop
	nextScreen game.Screen

	volumeSlider *widget.Slider
	volumeLabel  *widget.Text

	fullscreenCheckbox        *widget.Checkbox
	showGameLogCheckbox       *widget.Checkbox
	autoShowInfoPanelCheckbox *widget.Checkbox

	// Holds references to buttons mapped by their config key pointers
	bindingButtons map[**ebiten.Key]*widget.Button
}

// NewSettingsScreen creates the settings screen. previous is the screen
// to return to when the player presses Save.
func NewSettingsScreen(prevScreen *MainScreen) *SettingsScreen {
	s := &SettingsScreen{
		previous: prevScreen,
		bg:       prevScreen.bg,
	}

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	panel := ui.NewPanel("Settings: General")
	panel.AddContent(s.buildTabs(panel))

	// --- Save / Reset settings ---
	panel.AddControl(secondaryButton("Save", 14, func(_ *widget.ButtonClickedEventArgs) {
		err := config.Save()
		if err != nil {
			log.Printf("settings: failed to save config: %v", err)
		}
		s.nextScreen = s.previous
	}))
	panel.AddControl(secondaryButton("Return", 14, func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = s.previous
	}))
	panel.AddControl(secondaryButton("Reset settings", 14, func(_ *widget.ButtonClickedEventArgs) {
		s.resetToDefaults()
	}))

	root.AddChild(panel.Build())

	s.ui = &ebitenui.UI{Container: root}
	return s
}

// Update advances the settings UI.
func (s *SettingsScreen) Update(_ *game.Game) (game.Screen, error) {
	// 1. Handle Escape key logic first
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if activeBindingKey != nil {
			// Cancel active rebinding, revert button text, and clear listening state
			activeButton.Text().Label = game.KeyName(*activeBindingKey)
			activeBindingKey = nil
			activeButton = nil
		} else {
			// Exit the settings menu and return to the previous screen
			s.nextScreen = s.previous
		}
		return s, nil
	}

	// 2. Handle active key re-binding state
	if activeBindingKey != nil {
		for k := range ebiten.KeyMax {
			// Skip Escape here since it's already caught above as a cancel action
			if k == ebiten.KeyEscape {
				continue
			}

			if inpututil.IsKeyJustPressed(k) {
				for keyRef := range s.bindingButtons {
					if keyRef == nil || *keyRef == nil || keyRef == activeBindingKey {
						continue
					}
					if **keyRef == k {
						*keyRef = nil
					}
				}

				kk := k
				*activeBindingKey = &kk

				activeBindingKey = nil
				activeButton = nil

				s.refreshKeyLabels()
				break
			}
		}
	}

	s.bg.Update()
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
	s.bg.Draw(screen)
	s.ui.Draw(screen)
}

// Resize updates the backdrop dimensions when the screen or window is resized.
func (s *SettingsScreen) Resize(width, height int) {
	s.bg.Resize(width, height)
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

	wrapper.AddChild(table)
	return wrapper
}

var (
	activeBindingKey **ebiten.Key   // Pointer to the key currently being rebound
	activeButton     *widget.Button // Pointer to the button currently waiting for input
)

// buildKeybindingTab constructs the grid of actions and their corresponding key buttons.
func (s *SettingsScreen) buildKeybindingTab() *widget.Container {
	cfg := config.Get()
	labelFace := ui.TextFace(16)

	// Initialize or clear the buttons map
	s.bindingButtons = make(map[**ebiten.Key]*widget.Button)

	wrapper := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(20),
		)),
	)

	table := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, nil),
			widget.GridLayoutOpts.Spacing(30, 10),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(400, 0),
		),
	)

	// Helper function to insert a row into the grid
	addKeyRow := func(actionName string, keyRef **ebiten.Key) {
		// Left column: Action Label
		label := widget.NewText(
			widget.TextOpts.Text(actionName, &labelFace, color.White),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionStart,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			})),
		)
		table.AddChild(label)

		// Right column: Key Assignment Button
		var btn *widget.Button
		btn = secondaryButton(game.KeyName(*keyRef), 14, func(_ *widget.ButtonClickedEventArgs) {
			// If another button was waiting, revert its text before switching
			if activeButton != nil {
				activeButton.Text().Label = game.KeyName(*activeBindingKey)
			}

			// Enter waiting state for the clicked row
			activeBindingKey = keyRef
			activeButton = btn
			btn.Text().Label = "[...]"
		})

		s.bindingButtons[keyRef] = btn

		controlCell := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				HorizontalPosition: widget.GridLayoutPositionEnd,
				VerticalPosition:   widget.GridLayoutPositionCenter,
			})),
		)
		controlCell.AddChild(btn)
		table.AddChild(controlCell)
	}

	addKeyRow("Move", &cfg.Keybindings.Move)
	addKeyRow("Ability 1", &cfg.Keybindings.Ability1)
	addKeyRow("Ability 2", &cfg.Keybindings.Ability2)
	addKeyRow("Ability 3", &cfg.Keybindings.Ability3)
	addKeyRow("Ability 4", &cfg.Keybindings.Ability4)
	addKeyRow("Show game log", &cfg.Keybindings.ShowGameLog)
	addKeyRow("Show coordinates", &cfg.Keybindings.ShowCoordinates)
	addKeyRow("End Turn", &cfg.Keybindings.EndTurn)

	wrapper.AddChild(table)
	return wrapper
}

// refreshKeyLabels updates the rendered labels on every binding button to match the current configuration state.
func (s *SettingsScreen) refreshKeyLabels() {
	if s.bindingButtons == nil {
		return
	}
	for keyRef, button := range s.bindingButtons {
		if button != nil && button.Text() != nil {
			button.Text().Label = game.KeyName(*keyRef)
		}
	}
}

// buildTabs constructs the General | Keybinding tab book.
func (s *SettingsScreen) buildTabs(panel *ui.Panel) *widget.TabBook {
	generalTab := widget.NewTabBookTab(
		widget.TabBookTabOpts.Label("General"),
		widget.TabBookTabOpts.ContainerOpts(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout(
				widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(10)),
			)),
		),
	)
	generalTab.AddChild(s.buildGeneralTab())

	keybindingTab := widget.NewTabBookTab(
		widget.TabBookTabOpts.Label("Keybinding"),
		widget.TabBookTabOpts.ContainerOpts(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout(
				widget.AnchorLayoutOpts.Padding(widget.NewInsetsSimple(10)),
			)),
		),
	)
	keybindingTab.AddChild(s.buildKeybindingTab())

	tabTitles := map[*widget.TabBookTab]string{
		generalTab:    "Settings: General",
		keybindingTab: "Settings: Keybinding",
	}

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
		widget.TabBookOpts.TabSelectedHandler(func(args *widget.TabBookTabSelectedEventArgs) {
			if title, ok := tabTitles[args.Tab]; ok {
				panel.Title(title)
			}
		}),
	)
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
