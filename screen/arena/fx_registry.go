package arena

// FxName identifies a visual/audio effect sequence that can be played on the board.
type FxName int

// Effect identifiers used to look up definitions in fxRegistry.
const (
	fxNone FxName = iota
	fxSwordAttack
	fxSwordHit
	fxArrowShoot
	fxArrowHit
	fxSpellCast
	fxSpellHit
	fxShieldUp
	fxProvoke
	fxShieldAttack
	fxHit
	fxBattleCry
	fxSwordSpin
	fxHandPush
	fxMarkTarget
	fxTimeWarp
	fxPurge
	fxHeal
	fxPurify
)

// fxRegistry maps each FxName to its sprite, sound, and timing definition.
var fxRegistry = map[FxName]FxDefiner{
	fxSwordAttack: FxStep{
		Sprite: "sword_attack", Sound: "swoosh.ogg",
		FrameCount: 6, FrameSize: 512,
		DisplaySize: 100},
	fxSwordHit: FxStep{
		Sprite: "hit", Sound: "sword_hit.ogg",
		FrameCount: 6, FrameSize: 512,
		DisplaySize: 100},
	fxArrowShoot: FxStep{
		Sprite: "arrow_attack", Sound: "arrow_shoot.ogg",
		FrameCount: 6, FrameSize: 512,
		DisplaySize: 300},
	fxArrowHit: FxStep{
		Sprite: "arrow_hit", Sound: "arrow_hit.ogg",
		FrameCount: 6, FrameSize: 512,
		DisplaySize: 64},
	fxSpellCast: FxStep{
		Sprite: "spell_cast", Sound: "spell_cast.ogg",
		FrameCount: 6, FrameSize: 512,
		DisplaySize: 100},
	fxSpellHit: FxStep{
		Sprite: "spell_hit", Sound: "spell_hit.ogg",
		FrameCount: 6, FrameSize: 512,
		DisplaySize: 100},
	fxShieldUp: FxStep{
		Sprite: "shield_up", Sound: "shield_up.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 15,
		DisplaySize: 120},
	fxProvoke: FxStep{
		Sprite: "expanding_waves", Sound: "im_talking_to_you_speech.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 15,
		DisplaySize: 120},
	fxShieldAttack: FxStep{
		Sprite: "shield_attack", Sound: "shield_attack.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 30,
		DisplaySize: 100},
	fxHit: FxStep{
		Sprite: "hit", Sound: "hit.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 30,
		DisplaySize: 100},
	fxBattleCry: FxStep{
		Sprite: "expanding_waves", Sound: "male_brave_scream.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 15,
		DisplaySize: 100},
	fxSwordSpin: FxStep{
		Sprite: "sword_spin", Sound: "sword_spin.ogg",
		FrameCount: 12, FrameSize: 512, FPS: 20,
		DisplaySize: 100},
	fxHandPush: FxStep{
		Sprite: "hand_push", Sound: "hand_attack_whoosh.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 30,
		DisplaySize: 100},
	fxMarkTarget: FxStep{
		Sprite: "apply_mark", Sound: "appy_debuff.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 15,
		DisplaySize: 100},
	fxTimeWarp: FxStep{
		Sprite: "power_up", Sound: "time_warp.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 10,
		DisplaySize: 100},
	fxPurge: FxStep{
		Sprite: "power_down", Sound: "appy_debuff.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 10,
		DisplaySize: 100},
	fxHeal: FxStep{
		Sprite: "heal", Sound: "heal.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 10,
		DisplaySize: 100},
	fxPurify: FxStep{
		Sprite: "power_up", Sound: "heal.ogg",
		FrameCount: 6, FrameSize: 512, FPS: 10,
		DisplaySize: 100},
}
