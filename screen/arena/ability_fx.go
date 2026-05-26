package arena

import (
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
)

var abilityFxRegistry = map[ability.ID]AbilityFx{
	ability.BasicMeleeAttack: {Start: fxSwordAttack, End: fxSwordHit, PlayMode: FxDelayed, Delay: 0.1},
	ability.BasicRangeAttack: {Start: fxArrowShoot, End: fxArrowHit},
	ability.BasicMagicAttack: {Start: fxSpellCast, End: fxSpellHit},
}

type FxPlayMode int

const (
	// FxSequential plays End after Start has fully finished.
	FxSequential FxPlayMode = iota
	// FxParallel plays End simultaneously with Start.
	FxParallel
	// FxDelayed plays End after a fixed number of frames, while Start may still be playing.
	FxDelayed
)

type AbilityFx struct {
	Start    FxName
	End      FxName
	PlayMode FxPlayMode
	// Delay is the time in seconds before End fx starts, used when PlayMode is FxDelayed.
	Delay float64
}
