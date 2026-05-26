package arena

type FxName int

const (
	fxNone FxName = iota
	fxSwordAttack
	fxSwordHit
	fxArrowShoot
	fxArrowHit
	fxSpellCast
	fxSpellHit
)

var fxRegistry = map[FxName]FxDefiner{
	fxSwordAttack: FxStep{
		Sprite: "sword_attack", Sound: "swoosh.ogg",
		FrameCount: 6, FrameSize: 512,
		DisplaySize: 100},
	fxSwordHit: FxStep{
		Sprite: "sword_hit", Sound: "sword_hit.ogg",
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
}
