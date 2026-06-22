// Package tutorial defines tutorial step sequences and trigger events
// used to guide the player through game mechanics.
package tutorial

// TriggerEvent is a game event that unlocks the next tutorial step.
type TriggerEvent int

const (
	// TriggerNone means the step advances immediately on button press.
	TriggerNone TriggerEvent = iota

	// TriggerPlayUnit unlocks the step when the first play_unit message is received,
	// indicating that the placement phase is complete and the first unit is active.
	TriggerPlayUnit
)

// HighlightTarget identifies a UI element to highlight after a step is acknowledged.
type HighlightTarget int

// Highlight targets
const (
	HighlightNone      HighlightTarget = iota
	HighlightUnitPanel                 // the unit hand panel
	HighlightQueue                     // the turn order queue
	HighlightEndTurn                   // the end turn button
)

// Chapter is a named sequence of tutorial steps.
type Chapter struct {
	Name  string
	Steps []Step
}

// Step is a single tutorial message shown to the player during a game session.
type Step struct {
	Message    string
	ButtonText string
	Highlight  HighlightTarget
	WaitFor    TriggerEvent
	Anchor     AnchorTarget
}

// AnchorTarget defines where the tutorial overlay is positioned on screen.
type AnchorTarget int

// Anchor тargets
const (
	AnchorBottomRight AnchorTarget = iota // default
	AnchorTopLeft
	AnchorTopCenter
	AnchorTopRight
	AnchorCenterLeft
	AnchorCenter
	AnchorCenterRight
	AnchorBottomLeft
	AnchorBottomCenter
)
