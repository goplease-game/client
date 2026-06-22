package tutorial

// Basics is the first tutorial sequence covering core game mechanics.
var Basics = Chapter{
	Name: "Get to know the Basics",
	Steps: []Step{
		{
			Message:    "This is a turn-based game on a hex board with deterministic combat.",
			ButtonText: "Next",
			Anchor:     AnchorCenter,
		},
		{
			Message:    "Each player controls 6 units, deployed in waves of 3 -> 2 -> 1 across the first three rounds. Now deploy any of your 3 units to the board.",
			ButtonText: "Let's go",
			Highlight:  HighlightUnitPanel,
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "Deployed units build the queue in order of deployment. Every round units take turns in that exact order.",
			ButtonText: "Next",
			Anchor:     AnchorTopCenter,
			WaitFor:    TriggerPlayUnit,
			Highlight:  HighlightQueue,
		},
		{
			Message:    "Every unit has its own set of abilities: 1 basic attack, 3 unique active abilities, and 1 unique passive ability.",
			ButtonText: "Next",
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "Each active ability costs 1 AP (Action Point) to use. Each unit has 1 AP per turn, unless granted extra AP by another ability or game rule.",
			ButtonText: "Next",
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "Each unit has movement points (MP) which determine how far it can move. A unit can move once per turn. Once moved - even less than its full MP - the unit loses all remaining MP and cannot move again this turn.",
			ButtonText: "Next",
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "A unit can move and use abilities in any order, as long as it has not already moved or still has AP to spend.",
			ButtonText: "Next",
			Anchor:     AnchorBottomCenter,
		},
		{
			Message:    "That concludes the basics of the game. Eliminate all enemy units to win.",
			ButtonText: "Great, let's go!",
			Anchor:     AnchorBottomCenter,
		},
	},
}
