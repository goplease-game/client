package tutorial

// Basics is the first tutorial sequence covering core game mechanics.
var Basics = Chapter{
	Name: "Get to know the Basics",
	Steps: []Step{
		{
			Message:    "Welcome! This is a turn-based tactical game with deterministic gameplay.",
			ButtonText: "Next",
			Anchor:     AnchorCenter,
		},
		{
			Message:    "Each player commands 6 units, deployed in waves: 3 units in round 1, then 2, then last one.",
			ButtonText: "Next",
			Highlight:  HighlightUnitPanel,
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "To deploy a unit, grab it from the panel below and drop it onto a highlighted zone. Go ahead and deploy any 3.",
			ButtonText: "Let's go!",
			Highlight:  HighlightUnitPanel,
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "Units act in the order they were deployed — that order is shown in the queue at the top. It stays the same every round.",
			ButtonText: "Got it",
			Anchor:     AnchorTopCenter,
			WaitFor:    TriggerPlayUnit,
			Highlight:  HighlightQueue,
		},
		{
			Message:    "Every unit has 5 abilities: 1 basic attack, 3 unique active abilities, and 1 passive. [@pic:tutorial/abilities.png;394x171]",
			ButtonText: "Next",
			Highlight:  HighlightAbilityPanel,
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "Active abilities cost 1 AP (Action Point) to use. Each unit gets 1 AP per turn — some abilities can grant more.",
			ButtonText: "Next",
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "Active unit's AP is shown as blue dots at the top center of the unit card. [@pic:tutorial/ap.png;387x198]",
			ButtonText: "Next",
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "Each unit also has MP (Movement Points) — how far it can move in one turn. Moving spends all MP at once, no matter how far you go. You can only move once per turn.",
			ButtonText: "Next",
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "If active unit can move this turn, you'll see a Move icon on its card. [@pic:walk_o.png;32x32]",
			ButtonText: "Next",
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "To move, click a unit on the board — reachable cells will light up. Click any highlighted cell to move there.",
			ButtonText: "Next",
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "You can move and use abilities in any order.",
			ButtonText: "Next",
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "That's the basics! Eliminate all enemy units to win.",
			ButtonText: "Let's go!",
			Anchor:     AnchorBottomCenter,
		},
	},
}
