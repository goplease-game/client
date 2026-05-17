package ability

type Type int

const (
	Skill Type = iota + 1
	Spell
)

const (
	BasicMeleeAttack = "basic_melee_attack"
	BasicRangeAttack = "basic_range_attack"
	BasicMagicAttack = "basic_magic_attack"

	Phalanx     = "phalanx"
	Provoke     = "provoke"
	ShieldBash  = "shield_bash"
	UndyingWill = "undying_will"

	BattleCry = "battle_cry"
	Cleave    = "cleave"
	PowerPush = "power_push"
	Frenzy    = "frenzy"

	PiercingShot  = "piercing_shot"
	HuntersMark   = "hunters_mark"
	HamstringShot = "hamstring_shot"
	CoverFire     = "cover_fire"

	ShadowStep  = "shadow_step"
	GangUp      = "gang_up"
	Eliminate   = "eliminate"
	Opportunity = "opportunity"

	Translocation = "translocation"
	TimeWarp      = "time_warp"
	Purge         = "purge"
	ArcaneChaos   = "arcane_chaos"

	Heal           = "heal"
	Equalize       = "equalize"
	Purify         = "purify"
	BottomlessVial = "bottomless_vial"
)

type Ability struct {
	ID          string `json:"id"`
	Type        Type   `json:"type"`
	IsPassive   bool   `json:"is_passive"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Range       int    `json:"range"`
	Cooldown    int    `json:"cooldown"`
}

// TODO game client should download this from game server
var Abilities = map[string]Ability{
	BasicMeleeAttack: {
		Name:        "Strike",
		Description: "Delivers direct blow to a nearby enemy.",
		Range:       1,
		Cooldown:    0,
	},
	BasicRangeAttack: {
		Name:        "Shoot",
		Description: "Fires projectile at a distant target.",
		Range:       3,
		Cooldown:    0,
	},
	BasicMagicAttack: {
		Name:        "Blast",
		Description: "Hurls bolt of arcane energy.",
		Range:       3,
		Cooldown:    0,
	},

	Phalanx: {
		Type:        Skill,
		Name:        "Phalanx",
		Description: "You and adjacent allies gain +3 Shield. Shield decays by 1 at the start of each turn.",
		Range:       1, Cooldown: 3,
	},
	Provoke: {
		Type:        Skill,
		Name:        "Provoke",
		Description: "Forces target enemies to attack you on their turn.",
		Range:       3, Cooldown: 3,
	},
	ShieldBash: {
		Type:        Skill,
		Name:        "Shield Bash",
		Description: "Stuns the target for 1 turn.",
		Range:       1, Cooldown: 3,
	},
	UndyingWill: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Undying Will",
		Description: "When receiving fatal damage, prevent death: set HP to 1 and gain 3 Shield.",
		Range:       0, Cooldown: 5,
	},

	BattleCry: {
		Type:        Skill,
		Name:        "Battle Cry",
		Description: "Grants +3 Attack to adjacent allies. Bonus decays by 1 at the start of each turn.",
		Range:       1, Cooldown: 3,
	},
	Cleave: {
		Type:        Skill,
		Name:        "Cleave",
		Description: "Attacks all enemies in front of you for base damage.",
		Range:       1, Cooldown: 3,
	},
	PowerPush: {
		Type:        Skill,
		Name:        "Power Push",
		Description: "Deals 3 damage and pushes the target back 1 tile. If the target cannot be pushed, deals 5 damage instead.",
		Range:       1, Cooldown: 3,
	},
	Frenzy: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Frenzy",
		Description: "Gains +2 Attack if there are 2 or more enemies within 2 cells.",
		Range:       0, Cooldown: 0,
	},

	PiercingShot: {
		Type:        Skill,
		Name:        "Piercing Shot",
		Description: "Fires a shot that passes through all enemies in a line.",
		Range:       3, Cooldown: 3,
	},
	HuntersMark: {
		Type:        Skill,
		Name:        "Hunter's Mark",
		Description: "Marks target for 3 turns. Allies deal +1 damage to marked target.",
		Range:       3, Cooldown: 4,
	},
	HamstringShot: {
		Type:        Skill,
		Name:        "Hamstring Shot",
		Description: "Deals 2 damage and reduces target's Move Range to 1 for next turn.",
		Range:       3, Cooldown: 3,
	},
	CoverFire: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Cover Fire",
		Description: "Once per turn, counter-attacks the first enemy that strikes an ally within your range, dealing 3 flat damage.",
		Range:       3, Cooldown: 0,
	},

	ShadowStep: {
		Type:        Spell,
		Name:        "Shadow Step",
		Description: "Teleport to target cell and gain +2 Attack for next attack.",
		Range:       3, Cooldown: 3,
	},
	GangUp: {
		Type:        Skill,
		Name:        "Gang Up",
		Description: "Executes a melee attack. Deals +2 bonus damage if an ally is directly on the opposite side of the target.",
		Range:       1, Cooldown: 3,
	},
	Eliminate: {
		Type:        Skill,
		Name:        "Eliminate",
		Description: "Deals 3 damage. If the damage is fatal, gain 1 AP.",
		Range:       1, Cooldown: 5,
	},
	Opportunity: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Opportunity",
		Description: "Once per turn, automatically strikes an adjacent enemy when an ally attacks them in melee.",
		Range:       1, Cooldown: 0,
	},

	Translocation: {
		Type:        Spell,
		Name:        "Translocation",
		Description: "Swap places with any ally or enemy within range.",
		Range:       3, Cooldown: 4,
	},
	TimeWarp: {
		Type:        Spell,
		Name:        "Time Warp",
		Description: "Target ally or self gains +1 AP. At the end of their turn, their HP, Shield, and position are reverted to their state at the start of the turn.",
		Range:       3, Cooldown: 5,
	},
	Purge: {
		Type:        Spell,
		Name:        "Purge",
		Description: "Removes all positive effects from target enemy.",
		Range:       3, Cooldown: 3,
	},
	ArcaneChaos: {
		Type:        Spell,
		IsPassive:   true,
		Name:        "Arcane Chaos",
		Description: "End of turn triggers: 1. No movement: +1 Move Range next turn. 2. No enemies in radius 3: +1 Attack Range next turn. 3. No damage taken: heal 1 HP next turn. 4. Damage taken: gain 1 Shield. If 3 of 4 triggers are met, also gain +1 ATK for the next turn.",
		Range:       0, Cooldown: 0,
	},

	Heal: {
		Type:        Spell,
		Name:        "Heal",
		Description: "Restores 4 HP to target ally.",
		Range:       3, Cooldown: 1,
	},
	Equalize: {
		Type:        Spell,
		Name:        "Equalize",
		Description: "Averages HP of all allies within 2 cells.",
		Range:       2, Cooldown: 4,
	},
	Purify: {
		Type:        Spell,
		Name:        "Purify",
		Description: "Removes all negative status effects from an ally, heals them for 2 HP, and grants immunity to new debuffs for 1 turn.",
		Range:       3, Cooldown: 3,
	},
	BottomlessVial: {
		Type:        Spell,
		IsPassive:   true,
		Name:        "Bottomless Vial",
		Description: "Once per turn, when July takes damage, her Max HP permanently increases by 1.",
		Range:       0, Cooldown: 0,
	},
}

func ByID(id string) Ability {
	s, ok := Abilities[id]
	if ok {
		s.ID = id
	}

	return s
}

func (a Ability) IsBasicAttack() bool {
	switch a.ID {
	case BasicMeleeAttack, BasicRangeAttack, BasicMagicAttack:
		return true
	}

	return false
}
