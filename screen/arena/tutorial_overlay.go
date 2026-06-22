package arena

import (
	"fmt"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/tutorial"
	"github.com/goplease-game/client/ui"
)

// TutorialOverlay renders tutorial panel over the arena.
type TutorialOverlay struct {
	UI                  *ebitenui.UI
	OnSkip              func()
	OnDone              func()
	OnStep              func(tutorial.Step)
	OnVisibilityChanged func(bool)

	tutorial    tutorial.Chapter
	steps       []tutorial.Step
	currentStep int
	triggers    map[tutorial.TriggerEvent]bool
	visible     bool

	messageW   *widget.Text
	buttonW    *widget.Button
	messageRef *widget.Container
	buttonRef  *widget.Container
	titleRef   *widget.Container
	panelRef   *widget.Container
}

// NewTutorialOverlay creates a tutorial overlay for the given steps.
// onSkip is called when the player skips the tutorial.
// onDone is called when the last step is acknowledged.
func NewTutorialOverlay(t tutorial.Chapter, onSkip, onDone func(), onStep func(tutorial.Step), onVisibilityChanged func(bool)) *TutorialOverlay {
	o := &TutorialOverlay{
		tutorial:            t,
		steps:               t.Steps,
		triggers:            make(map[tutorial.TriggerEvent]bool),
		visible:             true,
		OnSkip:              onSkip,
		OnDone:              onDone,
		OnStep:              onStep,
		OnVisibilityChanged: onVisibilityChanged,
	}
	o.UI = o.build()
	o.applyStep()
	return o
}

// Trigger notifies the overlay that a game event occurred,
// unlocking steps that are waiting for that event.
func (o *TutorialOverlay) Trigger(event tutorial.TriggerEvent) {
	if o.currentStep >= len(o.steps) {
		return
	}
	o.triggers[event] = true
	if o.steps[o.currentStep].WaitFor == event {
		o.applyStep()
	}
}

func (o *TutorialOverlay) IsVisible() bool {
	return o != nil && o.visible
}

func (o *TutorialOverlay) CurrentStep() (tutorial.Step, bool) {
	if o == nil || o.currentStep >= len(o.steps) {
		return tutorial.Step{}, false
	}
	return o.steps[o.currentStep], true
}

// advance moves to the next step or calls onDone if all steps are complete.
func (o *TutorialOverlay) advance() {
	o.currentStep++
	if o.currentStep >= len(o.steps) {
		if o.OnDone != nil {
			o.OnDone()
		}
		return
	}
	o.applyStep()
}

// applyStep updates the overlay text and button for the current step.
func (o *TutorialOverlay) applyStep() {
	step := o.steps[o.currentStep]
	if step.WaitFor != tutorial.TriggerNone && !o.triggers[step.WaitFor] {
		o.setVisible(false)
		return
	}

	o.setVisible(true)
	o.applyAnchor(step.Anchor)
	if o.OnStep != nil {
		o.OnStep(step)
	}

	title := fmt.Sprintf("[%d/%d] %s",
		o.currentStep+1,
		len(o.steps),
		o.tutorial.Name,
	)
	o.titleRef.RemoveChildren()
	titleW := widget.NewText(
		widget.TextOpts.Text(title, &tutorialTitleTF, tutorialTitleTextColor),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)
	o.titleRef.AddChild(titleW)

	// recreate message text
	o.messageRef.RemoveChildren()
	for _, seg := range parseTutorialMessage(step.Message) {
		if seg.image != "" {
			img := asset.Image(seg.image, seg.imgW, seg.imgH)
			o.messageRef.AddChild(widget.NewGraphic(
				widget.GraphicOpts.Image(img),
				widget.GraphicOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.RowLayoutData{
						Position: widget.RowLayoutPositionCenter,
					}),
				),
			))
		} else if seg.text != "" {
			tf := ui.TextFace(16)
			o.messageRef.AddChild(widget.NewText(
				widget.TextOpts.Text(seg.text, &tf, tutorialTextColor),
				widget.TextOpts.MaxWidth(480),
			))
		}
	}

	// recreate button
	o.buttonRef.RemoveChildren()
	btnTF := ui.TextFace(15)
	locked := step.WaitFor != tutorial.TriggerNone && !o.triggers[step.WaitFor]
	o.buttonW = widget.NewButton(
		widget.ButtonOpts.Text(step.ButtonText, &btnTF, &widget.ButtonTextColor{
			Idle:     tutorialTextColor,
			Disabled: tutorialBtnDisabledTextColor,
		}),
		widget.ButtonOpts.Image(tutorialButtonImage()),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 20, Right: 20, Top: 10, Bottom: 10}),
		widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
			o.advance()
		}),
	)
	o.buttonW.GetWidget().Disabled = locked
	o.buttonRef.AddChild(o.buttonW)

	skipTF := ui.TextFace(13)
	skipBtn := widget.NewButton(
		widget.ButtonOpts.Text("Skip tutorial", &skipTF, &widget.ButtonTextColor{
			Idle: tutorialSkipBtnTextColor,
		}),
		widget.ButtonOpts.Image(skipButtonImage()),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 12, Right: 12, Top: 8, Bottom: 8}),
		widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
			if o.OnSkip != nil {
				o.OnSkip()
			}
		}),
	)
	o.buttonRef.AddChild(skipBtn)
}

func (o *TutorialOverlay) setVisible(visible bool) {
	if o.visible == visible {
		return
	}

	o.visible = visible
	if o.OnVisibilityChanged != nil {
		o.OnVisibilityChanged(visible)
	}
}

// build constructs the ebitenui widget tree for the overlay.
func (o *TutorialOverlay) build() *ebitenui.UI {
	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	panel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewBorderedNineSliceColor(tutorialTextBgColor, tutorialTitleBgColor, 2)),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(0),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionEnd,
				Padding: &widget.Insets{
					Right:  12,
					Bottom: footerH + statusH + 8,
				},
			}),
			widget.WidgetOpts.MinSize(480, 0),
		),
	)
	o.panelRef = panel

	// title bar
	titleBar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(tutorialTitleBgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	o.titleRef = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				Padding:            widget.NewInsetsSimple(10),
			}),
		),
	)
	titleBar.AddChild(o.titleRef)

	// body
	body := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(12),
			widget.RowLayoutOpts.Padding(&widget.Insets{Left: 28, Right: 28, Top: 16, Bottom: 16}),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	o.messageRef = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(8),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
	)

	o.buttonRef = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(12),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionEnd,
				Stretch:  true,
			}),
		),
	)

	body.AddChild(o.messageRef)
	body.AddChild(o.buttonRef)

	panel.AddChild(titleBar)
	panel.AddChild(body)
	root.AddChild(panel)

	return &ebitenui.UI{Container: root}
}

func (o *TutorialOverlay) applyAnchor(anchor tutorial.AnchorTarget) {
	const (
		topPad    = headerH + 8
		bottomPad = footerH + statusH + 8
		sidePad   = 12
	)

	var ld widget.AnchorLayoutData

	switch anchor {
	case tutorial.AnchorTopLeft:
		ld = widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionStart,
			Padding:            &widget.Insets{Left: sidePad, Top: topPad},
		}
	case tutorial.AnchorTopCenter:
		ld = widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionStart,
			Padding:            &widget.Insets{Top: topPad},
		}
	case tutorial.AnchorTopRight:
		ld = widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionEnd,
			VerticalPosition:   widget.AnchorLayoutPositionStart,
			Padding:            &widget.Insets{Right: sidePad, Top: topPad},
		}
	case tutorial.AnchorCenterLeft:
		ld = widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
			Padding:            &widget.Insets{Left: sidePad},
		}
	case tutorial.AnchorCenter:
		ld = widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		}
	case tutorial.AnchorCenterRight:
		ld = widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionEnd,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
			Padding:            &widget.Insets{Right: sidePad},
		}
	case tutorial.AnchorBottomLeft:
		ld = widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionEnd,
			Padding:            &widget.Insets{Left: sidePad, Bottom: bottomPad},
		}
	case tutorial.AnchorBottomCenter:
		ld = widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionEnd,
			Padding:            &widget.Insets{Bottom: bottomPad},
		}
	default: // AnchorBottomRight
		ld = widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionEnd,
			VerticalPosition:   widget.AnchorLayoutPositionEnd,
			Padding:            &widget.Insets{Right: sidePad, Bottom: bottomPad},
		}
	}

	if o.panelRef != nil {
		o.panelRef.GetWidget().LayoutData = ld
	}
}

func tutorialButtonImage() *widget.ButtonImage {
	return &widget.ButtonImage{
		Idle:     image.NewNineSliceColor(tutorialBtnBgColor),
		Hover:    image.NewNineSliceColor(tutorialBtnHoverBgColor),
		Pressed:  image.NewNineSliceColor(tutorialBtnPressedBgColor),
		Disabled: image.NewNineSliceColor(tutorialBtnDisabledBgColor),
	}
}

func skipButtonImage() *widget.ButtonImage {
	return &widget.ButtonImage{
		Idle:    image.NewNineSliceColor(tutorialSkipBtnBgColor),
		Hover:   image.NewNineSliceColor(tutorialSkipBtnHoverBgColor),
		Pressed: image.NewNineSliceColor(tutorialSkipBtnPressedBgColor),
	}
}
