package arena

import (
	"image"
	"log"
	"math"

	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/sfx"
	"github.com/goplease-game/server/ability"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/setanarut/anim"
)

// FxContext holds the positional context for a single fx playback.
type FxContext struct {
	Px    image.Point // screen pixel position
	Coord ds.HexCoord // hex coord for program fx
}

// ActiveFxAnim is a single fx animation currently playing on screen,
// either sprite-sheet-driven (player) or code-driven (programFx).
type ActiveFxAnim struct {
	player          *anim.AnimationPlayer
	pos             image.Point
	onDone          func()
	finished        bool
	delayFrames     int
	sound           string
	soundPlayed     bool
	programFx       ProgramFx
	programDuration int
	programTick     int
	coord           ds.HexCoord // hex coord for program fx
}

// playAbilityFx plays the visual effects for the given ability, preferring
// a registered composer over a plain fx-sequence registry entry, and
// calling onDone immediately if neither is registered.
func (s *Screen) playAbilityFx(abilityID ability.ID, unit *ds.Unit, target ds.HexCoord, onDone func()) {
	if composer, ok := abilityComposerRegistry[abilityID]; ok {
		composer(s, unit, target, onDone)
		return
	}

	if abFx, ok := abilityFxRegistry[abilityID]; ok {
		s.abilityFxComposer(abFx, abilityID, unit, target, onDone)
		return
	}

	onDone()
}

// playFxGroups plays the fx groups in sequence starting at idx, advancing
// to the next group once every step in the current group has finished,
// and calling onDone after the last group completes.
func (s *Screen) playFxGroups(groups []FxGroup, idx int, ctx FxContext, onDone func()) {
	if idx >= len(groups) {
		onDone()
		return
	}

	group := groups[idx]
	remaining := len(group.Steps)
	if remaining == 0 {
		s.playFxGroups(groups, idx+1, ctx, onDone)
		return
	}

	groupDone := func() {
		remaining--
		if remaining == 0 {
			s.playFxGroups(groups, idx+1, ctx, onDone)
		}
	}

	for _, step := range group.Steps {
		s.playFxStep(step, ctx, groupDone)
	}
}

// playFxStep starts a single fx step at the position specified by ctx.
// If the step has a ProgramFx, a code-driven animation is queued.
// If the step has a Sprite, a spritesheet animation is queued.
// Sound is stored and played once DelayFrames elapses in updateFxAnims.
// If neither Sprite nor ProgramFx is set, onDone is called immediately.
func (s *Screen) playFxStep(step FxStep, ctx FxContext, onDone func()) {
	if step.DisplaySize == 0 {
		step.DisplaySize = step.FrameSize
	}

	if step.ProgramFx != nil {
		s.activeFxAnims = append(s.activeFxAnims, &ActiveFxAnim{
			pos:             ctx.Px,
			onDone:          onDone,
			delayFrames:     int(step.DelaySeconds * 60),
			sound:           step.Sound,
			programFx:       step.ProgramFx,
			programDuration: step.ProgramDuration,
			coord:           ctx.Coord,
		})
		return
	}

	if step.Sprite == "" {
		onDone()
		return
	}

	originalFrameSize := step.FrameSize
	displaySize := step.DisplaySize
	scale := float64(displaySize) / float64(originalFrameSize)

	sheetW := int(math.Ceil(float64(originalFrameSize*step.FrameCount) * scale))
	sheetH := int(math.Ceil(float64(originalFrameSize) * scale))

	img := asset.Image("vfx/"+step.Sprite+".png", sheetW, sheetH)

	fps := step.FPS
	if fps == 0 {
		fps = 30
	}
	player := anim.NewAnimationPlayer(anim.Atlas{
		Name:  step.Sprite,
		Image: img,
	})
	player.NewAnim("play", 0, 0, displaySize, displaySize, step.FrameCount, false, false, fps)
	player.SetAnim("play")

	s.activeFxAnims = append(s.activeFxAnims, &ActiveFxAnim{
		player:      player,
		pos:         ctx.Px,
		onDone:      onDone,
		delayFrames: int(step.DelaySeconds * 60),
		sound:       step.Sound,
	})
}

// updateFxAnims advances all active fx animations by one frame, triggering
// sounds, detecting completion, and removing finished animations.
func (s *Screen) updateFxAnims() {
	current := s.activeFxAnims
	s.activeFxAnims = nil

	for _, fx := range current {
		if fx.delayFrames > 0 {
			fx.delayFrames--
			s.activeFxAnims = append(s.activeFxAnims, fx)
			continue
		}

		if !fx.soundPlayed {
			fx.soundPlayed = true
			if fx.sound != "" {
				sfx.Play(fx.sound)
			}
		}

		if fx.programFx != nil {
			fx.programTick++
			if fx.programTick > fx.programDuration {
				fx.finished = true
			}
		} else {
			prevIdx := fx.player.Data.CurrentIndex
			fx.player.Update()
			if fx.player.Data.CurrentIndex < prevIdx {
				fx.finished = true
			}
		}

		if fx.finished {
			fx.onDone()
		} else {
			s.activeFxAnims = append(s.activeFxAnims, fx)
		}
	}
}

// drawActiveFxAnims renders all active fx animations onto screen, skipping
// any still in their delay period.
func (s *Screen) drawActiveFxAnims(screen *ebiten.Image) {
	for _, fx := range s.activeFxAnims {
		if fx.delayFrames > 0 {
			continue
		}
		if fx.programFx != nil {
			t := float64(fx.programTick) / float64(fx.programDuration)
			if t > 1.0 {
				t = 1.0
			}
			fx.programFx(ProgramFxContext{
				Screen:     s,
				Coord:      fx.coord,
				Px:         fx.pos,
				Unit:       s.unitAtCoord(fx.coord),
				Widget:     s.boardCellWidgets[fx.coord],
				T:          t,
				DrawTarget: screen,
			})
			continue
		}
		if fx.player == nil {
			continue
		}
		frame := fx.player.CurrentFrame
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(fx.pos.X)-float64(frame.Bounds().Dx())/2,
			float64(fx.pos.Y)-float64(frame.Bounds().Dy())/2,
		)
		screen.DrawImage(frame, op)
	}
}

// playFxAt plays a named fx definition at the given position.
// Logs a warning if the fx name is not found in fxRegistry.
func (s *Screen) playFxAt(name FxName, ctx FxContext, onDone func()) {
	fx, ok := fxRegistry[name]
	if !ok {
		log.Printf("playFxAt: fx %d not found in fxRegistry", name)
		onDone()
		return
	}
	s.playFxGroups(fx.Define().Groups, 0, ctx, onDone)
}

// scheduleDelayed queues fn to be called after the given number of seconds.
func (s *Screen) scheduleDelayed(seconds float64, fn func()) {
	frames := int(seconds * 60)
	if frames <= 0 {
		fn()
		return
	}
	s.delayedActions = append(s.delayedActions, delayedAction{frames: frames, fn: fn})
}

// delayedAction is a callback scheduled to run after a fixed number of frames.
type delayedAction struct {
	frames int
	fn     func()
}

// updateDelayedActions advances all pending delayed actions.
func (s *Screen) updateDelayedActions() {
	alive := s.delayedActions[:0]
	for _, a := range s.delayedActions {
		a.frames--
		if a.frames <= 0 {
			a.fn()
		} else {
			alive = append(alive, a)
		}
	}
	s.delayedActions = alive
}
