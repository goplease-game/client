package ability

const (
	BasicMeleeAttack ID = "basic_melee_attack"
	BasicRangeAttack ID = "basic_range_attack"
	BasicMagicAttack ID = "basic_magic_attack"

	// Tank
	Fortify     ID = "fortify"
	Provoke     ID = "provoke"
	ShieldBash  ID = "shield_bash"
	UndyingWill ID = "undying_will"

	// warrior
	BattleCry   ID = "battle_cry"
	IdolihuSpin ID = "idolihu_spin"
	PowerPush   ID = "power_push"
	Frenzy      ID = "frenzy"

	// ranger
	PiercingShot  ID = "piercing_shot"
	HuntersMark   ID = "hunters_mark"
	HamstringShot ID = "hamstring_shot"
	CoverFire     ID = "cover_fire"

	// rogue
	ShadowStep  ID = "shadow_step"
	GangUp      ID = "gang_up"
	Eliminate   ID = "eliminate"
	Opportunity ID = "opportunity"

	// mage
	Translocation ID = "translocation"
	TimeWarp      ID = "time_warp"
	Purge         ID = "purge"
	ArcaneChaos   ID = "arcane_chaos"

	// support
	Heal           ID = "heal"
	Equalize       ID = "equalize"
	Purify         ID = "purify"
	BottomlessVial ID = "bottomless_vial"
)

// Damage hint format (ATK will be replaced with unit.CurrentAtk):
// - ATK       — base attack damage
// - ATK+2     — base attack plus flat bonus
// - ATK/ATK+2 — base attack or base attack with bonus
// - 3         — flat damage value
// - 2/3       — two distinct possible values
// - 2–5       — damage range

const HintCurrentATK = "ATK"

var Abilities = map[ID]Ability{
	BasicMeleeAttack: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Strike",
		Description: "Delivers direct blow to a nearby enemy.",
		Cooldown:    0,
		Range:       1,
		Activation:  SelectEnemy,
		TargetMode:  TargetEnemies,
		DamageHint:  HintCurrentATK,
	},
	BasicRangeAttack: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Shoot",
		Description: "Fires projectile at a distant target.",
		Cooldown:    0,
		Range:       4,
		Activation:  SelectEnemy,
		TargetMode:  TargetEnemies,
		DamageHint:  HintCurrentATK,
	},
	BasicMagicAttack: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Arcane Bolt",
		Description: "Hurls bolt of arcane energy.",
		Cooldown:    0,
		Range:       4,
		Activation:  SelectEnemy,
		TargetMode:  TargetEnemies,
		DamageHint:  HintCurrentATK,
	},

	// --- TANK ---
	Fortify: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Fortify",
		Description: "You and adjacent allies gain +4 Shield. Shield decays by 1 at the start of each turn.",
		Cooldown:    3,
		Range:       0,
		Activation:  Instant,
		TargetMode:  TargetAlliesAndSelf,
		Area:        AreaCircle,
		AreaRadius:  2,
	},
	Provoke: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Provoke",
		Description: "Forces enemies to attack you on their turn.",
		Cooldown:    3,
		Range:       2,
		TargetMode:  TargetEnemies,
		Activation:  Instant,
		Area:        AreaCircle,
		AreaRadius:  2,
	},
	ShieldBash: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Shield Bash",
		Description: "Stuns an enemy, preventing their next action.",
		Cooldown:    3,
		Range:       1,
		TargetMode:  TargetEnemies,
		Activation:  SelectEnemy,
	},
	UndyingWill: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Undying Will",
		Description: "When receiving fatal damage, prevent death: set HP to 1 and gain 5 Shield.",
		Cooldown:    5,
		Range:       0,
	},

	// --- WARRIOR ---
	BattleCry: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Battle Cry",
		Description: "Grants +1 Attack to nearby allies for 2 turns",
		Cooldown:    3,
		TargetMode:  TargetAllies,
		Activation:  Instant,
		Area:        AreaCircle,
		AreaRadius:  2,
	},
	IdolihuSpin: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "IDOLIHU! Spin",
		Description: "Strikes all adjacent enemies in a single sweeping motion.",
		Cooldown:    3,
		TargetMode:  TargetEnemies,
		Activation:  Instant,
		Area:        AreaCircle,
		AreaRadius:  1,
		DamageHint:  HintCurrentATK,
	},
	PowerPush: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Power Push",
		Description: "Deals 2 damage and pushes the target back 1 tile. If the target cannot be pushed, deals 3 damage instead.",
		Cooldown:    3,
		Range:       1,
		TargetMode:  TargetEnemies,
		Activation:  SelectEnemy,
		DamageHint:  "2/3",
	},
	Frenzy: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Frenzy",
		Description: "Gains +1 Attack if there are 2 or more enemies within 2 cells.",
		AreaRadius:  2,
	},

	// --- RANGER ---
	PiercingShot: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Piercing Shot",
		Description: "Fires a piercing shot that deals 2 damage to each enemy in a straight line.",
		Cooldown:    3,
		TargetMode:  TargetEnemies,
		Activation:  SelectAny,
		Area:        AreaLine,
		AreaRadius:  4,
		DamageHint:  "2",
	},
	HuntersMark: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Hunter's Mark",
		Description: "Marks target for 3 turns. Allies deal +1 damage to marked target.",
		Cooldown:    4,
		Range:       3,
		TargetMode:  TargetEnemies,
		Activation:  SelectEnemy,
	},
	HamstringShot: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Hamstring Shot",
		Description: "Deals 2 damage and reduces target's Move Range to 1 for next turn.",
		Cooldown:    3,
		Range:       4,
		TargetMode:  TargetEnemies,
		Activation:  SelectEnemy,
		DamageHint:  "2",
	},
	CoverFire: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Cover Fire",
		Description: "Once per turn, counter-attacks the first enemy that strikes an ally within your range, dealing 3 flat damage.",
		Range:       3,
		DamageHint:  "3",
	},

	// --- ROGUE ---
	ShadowStep: {
		Type:        Spell,
		IsPassive:   false,
		Name:        "Shadow Step",
		Description: "Teleport to target cell and gain +1 Attack until the end of your next turn.",
		Cooldown:    3,
		Range:       4,
		Activation:  SelectFreeCell,
	},
	GangUp: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Gang Up",
		Description: "Executes a melee attack. Deals +2 bonus damage if an ally is on the opposite side of the target",
		Cooldown:    3,
		Range:       1,
		TargetMode:  TargetEnemies,
		Activation:  SelectEnemy,
		DamageHint:  "ATK/ATK+2",
	},
	Eliminate: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Eliminate",
		Description: "Deals 3 damage. If this attack kills the target, gain 1 AP.",
		Cooldown:    5,
		Range:       1,
		TargetMode:  TargetEnemies,
		Activation:  SelectEnemy,
		DamageHint:  "3",
	},
	Opportunity: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Opportunity",
		Description: "Once per turn, attacks an adjacent enemy when an ally hits them with a melee attack.",
		Cooldown:    0,
		Range:       1,
		DamageHint:  HintCurrentATK,
	},

	// --- MAGE ---
	Translocation: {
		Type:        Spell,
		IsPassive:   false,
		Name:        "Translocation",
		Description: "Swap places with any ally or enemy within range.",
		Cooldown:    3,
		Range:       4,
		TargetMode:  TargetAny,
		Activation:  SelectAnyUnit,
	},
	TimeWarp: {
		Type:        Spell,
		IsPassive:   false,
		Name:        "Time Warp",
		Description: "Target ally or self gains +1 AP at the start of their next turn. At the end of that turn, their HP, Shield, and position are restored to their state at the start of the turn.",
		Cooldown:    5,
		Range:       3,
		TargetMode:  TargetAlliesAndSelf,
		Activation:  SelectAllyOrSelf,
	},
	Purge: {
		Type:        Spell,
		IsPassive:   false,
		Name:        "Purge",
		Description: "Removes all positive effects from target enemy.",
		Cooldown:    3,
		Range:       3,
		TargetMode:  TargetEnemies,
		Activation:  SelectEnemy,
	},
	ArcaneChaos: {
		Type:        Spell,
		IsPassive:   true,
		Name:        "Arcane Chaos",
		Description: "At the end of your turn, gain bonuses based on actions taken during the turn:\n- If you did not move: gain +1 Movement Range next turn\n- If no enemies were within 3 tiles: gain +1 Attack Range next turn\n- If you took no damage: restore 1 HP next turn\n- If you took damage: gain 1 Shield\n\nIf 3 or more conditions are met, also gain +1 Attack next turn.",
	},

	// --- SUPPORT ---
	Heal: {
		Type:        Spell,
		IsPassive:   false,
		Name:        "Heal",
		Description: "Restores 4 HP to the target ally or self.",
		Cooldown:    1,
		Range:       3,
		TargetMode:  TargetAlliesAndSelf,
		Activation:  SelectAllyOrSelf,
	},
	Equalize: {
		Type:        Spell,
		IsPassive:   false,
		Name:        "Equalize",
		Description: "Equalizes the HP of all allied units within 3 tiles, setting each to the average HP of the affected units.",
		Cooldown:    3,
		TargetMode:  TargetAlliesAndSelf,
		Activation:  Instant,
		Area:        AreaCircle,
		AreaRadius:  3,
	},
	Purify: {
		Type:        Skill,
		IsPassive:   false,
		Name:        "Purify",
		Description: "Removes all negative status effects from the target ally or self, restores 2 HP, and grants immunity to new debuffs for 1 turn.",
		Cooldown:    2,
		Range:       3,
		TargetMode:  TargetAlliesAndSelf,
		Activation:  SelectAllyOrSelf,
	},
	BottomlessVial: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Bottomless Vial",
		Description: "The first time each turn July loses HP, her maximum HP permanently increases by 1.",
	},
}
