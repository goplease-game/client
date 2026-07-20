package arena

import (
	"github.com/goplease-game/game-server/ability"
)

// abilityFxRegistry maps ability IDs to their declarative Start/End fx
// and playback mode. Abilities needing more complex sequencing are
// defined in abilityComposerRegistry instead.
var abilityFxRegistry = map[ability.ID]AbilityFx{
	ability.BasicMeleeAttack: {Start: fxSwordAttack, End: fxSwordHit, PlayMode: FxDelayed, Delay: 0.1},
	ability.BasicRangeAttack: {Start: fxArrowShoot, End: fxArrowHit},
	ability.BasicMagicAttack: {Start: fxSpellCast, End: fxSpellHit},

	ability.Fortify:     {Start: fxShieldUp, End: fxNone},
	ability.Provoke:     {Start: fxProvoke, End: fxNone},
	ability.ShieldBash:  {Start: fxShieldAttack, End: fxHit, PlayMode: FxDelayed, Delay: 0.1},
	ability.UndyingWill: {Start: fxHeal, End: fxNone},

	ability.BattleCry:   {Start: fxBattleCry, End: fxNone},
	ability.IdolihuSpin: {Start: fxSwordSpin, End: fxSwordHit},
	ability.PowerPush:   {Start: fxHandPush, End: fxHit, PlayMode: FxDelayed, Delay: 0.1},

	ability.PiercingShot:  {Start: fxArrowShoot, End: fxArrowHit},
	ability.HuntersMark:   {Start: fxNone, End: fxMarkTarget},
	ability.HamstringShot: {Start: fxArrowShoot, End: fxHit},
	ability.CoverFire:     {Start: fxArrowShoot, End: fxHit},

	// ability.ShadowStep: defined in abilityComposerRegistry
	ability.GangUp:      {Start: fxSwordAttack, End: fxSwordHit, PlayMode: FxDelayed, Delay: 0.1},
	ability.Eliminate:   {Start: fxSwordAttack, End: fxSwordHit, PlayMode: FxDelayed, Delay: 0.1},
	ability.Opportunity: {Start: fxSwordAttack, End: fxSwordHit, PlayMode: FxDelayed, Delay: 0.1},

	// ability.Translocation: defined in abilityComposerRegistry
	ability.TimeWarp: {Start: fxNone, End: fxTimeWarp},
	ability.Purge:    {Start: fxNone, End: fxPurge},

	ability.Heal:           {Start: fxNone, End: fxHeal},
	ability.Equalize:       {Start: fxHeal, End: fxNone},
	ability.Purify:         {Start: fxNone, End: fxPurify},
	ability.BottomlessVial: {Start: fxPurify, End: fxNone},
}

// FxPlayMode controls the timing relationship between an AbilityFx's
// Start and End effects.
type FxPlayMode int

const (
	// FxSequential plays End after Start has fully finished.
	FxSequential FxPlayMode = iota
	// FxParallel plays End simultaneously with Start.
	FxParallel
	// FxDelayed plays End after a fixed number of frames, while Start may still be playing.
	FxDelayed
)

// AbilityFx declares the Start and End visual effects played for an
// ability, along with how they're timed relative to each other.
type AbilityFx struct {
	Start    FxName
	End      FxName
	PlayMode FxPlayMode
	// Delay is the time in seconds before End fx starts, used when PlayMode is FxDelayed.
	Delay float64
}
