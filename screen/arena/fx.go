package arena

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
)

type FxDefiner interface {
	Define() FxDefinition
}

// FxDefinition describes a reusable atomic visual effect.
// It can be referenced by name from fxRegistry and composed into ability sequences.
// Multiple groups allow internal sequencing within the effect itself —
// for example: flash first, then shake. For ability-level sequencing use AbilityFxComposer.
type FxDefinition struct {
	Groups []FxGroup
}

// FxGroup is a set of steps that play simultaneously.
// All steps in a group start at the same time and the next group
// begins only after all steps in the current group have finished.
type FxGroup struct {
	Steps []FxStep
}

// FxSteps is a flat list of steps that all play simultaneously.
// Implements FxDefiner — use when no internal sequencing is needed.
type FxSteps []FxStep

func (s FxSteps) Define() FxDefinition {
	return FxDefinition{
		Groups: []FxGroup{
			{Steps: s},
		},
	}
}

// FxGroups is a sequenced list of groups.
// Implements FxDefiner — use when internal sequencing within the fx is needed.
type FxGroups []FxGroup

func (g FxGroups) Define() FxDefinition {
	return FxDefinition{
		Groups: g,
	}
}

// FxStep describes a single visual or audio action within an FxGroup.
// Either Sprite (spritesheet animation) or ProgramFx (code-driven animation)
// should be set, not both. Sound is optional and plays when DelayFrames elapses.
type FxStep struct {
	// Sprite is the spritesheet asset name without path or extension (e.g. "ab_basic_melee_attack").
	// Leave empty if using ProgramFx instead.
	Sprite string

	// Sound is the audio asset filename including extension (e.g. "swoosh.ogg").
	// Plays when DelayFrames elapses, simultaneously with the animation.
	Sound string

	// DelaySeconds is the time to wait before starting this step.
	// For example, 0.2 means the step starts 200ms after the group begins.
	DelaySeconds float64

	// FrameSize is the width and height of a single frame in the original spritesheet (pixels).
	FrameSize int

	// FrameCount is the total number of frames in the spritesheet animation.
	FrameCount int

	// DisplaySize is the width and height at which the animation is rendered on screen (pixels).
	// The spritesheet is scaled down to this size at load time.
	// Default to FrameSize
	DisplaySize int

	// ProgramFx is a code-driven animation called every frame with progress t in [0, 1].
	// Use instead of Sprite for effects like fade, shake, or scale.
	ProgramFx ProgramFx

	// ProgramDuration is the length of the ProgramFx animation in frames.
	ProgramDuration int

	// FPS is the playback speed of the spritesheet animation.
	// Defaults to 30 if not set.
	FPS float64
}

func (s FxStep) Define() FxDefinition {
	return FxDefinition{
		Groups: []FxGroup{
			{Steps: []FxStep{s}},
		},
	}
}

type ProgramFxContext struct {
	Screen *Screen
	Coord  ds.HexCoord
	Unit   *ds.Unit          // unit at coord, nil if empty
	Widget *ui.HexCellWidget // cell widget at coord
	T      float64           // progress 0.0 to 1.0
}

type ProgramFx func(ctx ProgramFxContext)
