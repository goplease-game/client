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

	ShieldWall = "shield_wall"
	Provoke    = "provoke"
	ShieldBash = "shield_bash"
	LastStand  = "last_stand"

	BattleCry = "battle_cry"
	Cleave    = "cleave"
	Slam      = "slam"
	Frenzy    = "frenzy"

	PiercingShot = "piercing_shot"
	Prey         = "prey"
	Disengage    = "disengage"
	CoverFire    = "cover_file"

	ShadowStep  = "shadow_step"
	Backstab    = "backstab"
	Eliminate   = "eliminate"
	Opportunity = "opportunity"

	Translocation = "translocation"
	TimeWarp      = "time_warp"
	Enfeeble      = "enfeeble"
	Meditation    = "meditation"

	Heal        = "heal"
	BalanceLife = "balance_life"
	Cleanse     = "cleanse"
	Renewal     = "renewal"
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

	ShieldWall: {
		Type:        Skill,
		Name:        "Shield Wall",
		Description: "You and adjacent allies gain +3 Shield.",
		Range:       1, Cooldown: 3,
	},
	Provoke: {
		Type:        Skill,
		Name:        "Provoke",
		Description: "Forces target enemies to attack you on their next turn.",
		Range:       3, Cooldown: 3,
	},
	ShieldBash: {
		Type:        Skill,
		Name:        "Shield Bash",
		Description: "Stuns the target for 1 turn.",
		Range:       2, Cooldown: 3,
	},
	LastStand: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Last Stand",
		Description: "If HP falls below 1, gain +3 Shield instead of dying.",
		Range:       0, Cooldown: 5,
	},

	BattleCry: {
		Type:        Skill,
		Name:        "Battle Cry",
		Description: "Grant +2 Attack to adjacent allies.",
		Range:       1, Cooldown: 3,
	},
	Cleave: {
		Type:        Skill,
		Name:        "Cleave",
		Description: "Attack all enemies in front of you.",
		Range:       1, Cooldown: 3,
	},
	Slam: {
		Type:        Skill,
		Name:        "Slam",
		Description: "Removes all shields from the target.",
		Range:       1, Cooldown: 3,
	},
	Frenzy: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Frenzy",
		Description: "+2 Attack if current HP is below 4.",
		Range:       0, Cooldown: 0,
	},

	PiercingShot: {
		Type:        Skill,
		Name:        "Piercing Shot",
		Description: "Fires a shot that passes through all enemies in a line.",
		Range:       3, Cooldown: 3,
	},
	Prey: {
		Type:        Skill,
		Name:        "Prey",
		Description: "Deals 2 damage and marks target for 3 turns. Allies deal +2 damage to marked target.",
		Range:       3, Cooldown: 4,
	},
	Disengage: {
		Type:        Skill,
		Name:        "Disengage",
		Description: "Retreat 2 cells back, breaking engagement.",
		Range:       0, Cooldown: 3,
	},
	CoverFire: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Cover Fire",
		Description: "Counter-attacks enemies that strike allies within your range.",
		Range:       3, Cooldown: 0,
	},

	ShadowStep: {
		Type:        Spell,
		Name:        "Shadow Step",
		Description: "Teleport to target cell and gain +2 Attack for 1 turn.",
		Range:       3, Cooldown: 3,
	},
	Backstab: {
		Type:        Skill,
		Name:        "Backstab",
		Description: "Deals 2x damage if an ally is on the opposite side of the target.",
		Range:       1, Cooldown: 3,
	},
	Eliminate: {
		Type:        Skill,
		Name:        "Eliminate",
		Description: "Deals 3 damage. If target dies, gain 1 AP.",
		Range:       1, Cooldown: 5,
	},
	Opportunity: {
		Type:        Skill,
		IsPassive:   true,
		Name:        "Opportunity",
		Description: "Strikes an enemy if an ally attacks them from melee range.",
		Range:       1, Cooldown: 0,
	},

	Translocation: {
		Type:        Spell,
		Name:        "Translocation",
		Description: "Swap places with any unit on the board.",
		Range:       3, Cooldown: 4,
	},
	TimeWarp: {
		Type:        Spell,
		Name:        "Time Warp",
		Description: "Target ally or self gains +1 AP on their next turn.",
		Range:       3, Cooldown: 5,
	},
	Enfeeble: {
		Type:        Spell,
		Name:        "Enfeeble",
		Description: "Reduces target's Attack by 50% for 1 turn.",
		Range:       3, Cooldown: 3,
	},
	Meditation: {
		Type:        Spell,
		IsPassive:   true,
		Name:        "Meditation",
		Description: "If no AP spent, no movement made, and no damage taken this turn: heal 1 HP, gain +1 AP and +1 Movement on next turn.",
		Range:       0, Cooldown: 0,
	},

	Heal: {
		Type:        Spell,
		Name:        "Heal",
		Description: "Restores 2 HP to target ally.",
		Range:       3, Cooldown: 2,
	},
	BalanceLife: {
		Type:        Spell,
		Name:        "Balance Life",
		Description: "Averages HP of all allies within 2 cells.",
		Range:       2, Cooldown: 5,
	},
	Cleanse: {
		Type:        Spell,
		Name:        "Cleanse",
		Description: "Removes all negative status effects from target ally.",
		Range:       3, Cooldown: 2,
	},
	Renewal: {
		Type:        Spell,
		IsPassive:   true,
		Name:        "Aura of Renewal",
		Description: "If no damage taken this turn, heal self and all adjacent allies for 1 HP at the start of your next turn.",
		Range:       1, Cooldown: 0,
	},
}

func ByID(id string) Ability {
	s, ok := Abilities[id]
	if ok {
		s.ID = id
	}

	return s
}
