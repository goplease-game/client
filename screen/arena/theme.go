package arena

import (
	"image/color"

	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"golang.org/x/image/colornames"
)

var (
	// Layout colours.
	bodyBgColor      = color.NRGBA{0x13, 0x1a, 0x22, 0xff}
	footerBgColor    = ui.RGBFromHex("5682B1")
	headerBgColor    = ui.RGBFromHex("5682B1")
	statusBarBgColor = ui.DarkenRGB(footerBgColor, 5)

	// Unit panel and card colours.
	unitPanelBgColor       = ui.DarkenRGB(footerBgColor, 5)
	unitCardBgColor        = ui.DarkenRGB(unitPanelBgColor, 20)
	unitCardHoverBgColor   = ui.DarkenRGB(unitCardBgColor, 15)
	unitCardHoverFgColor   = ui.RGBFromHex("f5df4d")
	unitCardHighlightColor = colornames.Gold
	unitDragBgColor        = ui.RGBFromHex("78B3CE")
	unitFriendlyBgColor    = ui.RGBFromHex("B0DB9C")
	unitEnemyBgColor       = ui.RGBFromHex("EA7B7B")
	unitMoveToCellColor    = ui.RGBFromHex("6e8596")
	unitPulseColor1        = ui.RGBFromHex("FFC700")
	unitPulseColor2        = ui.DarkenRGB(unitPulseColor1, 80)

	// Tooltip colours.
	ttBgColor     = ui.RGBFromHex("42668d")
	ttBorderColor = ui.LightenRGB(ttBgColor, 50)
	ttTitleColor  = ui.RGBFromHex("f5df4d")
	ttTextColor   = colornames.Ghostwhite

	// HUD colours.
	hpColor                 = colornames.Tomato
	atkColor                = colornames.Orange
	mpColor                 = colornames.Skyblue
	shieldColor             = colornames.Gold
	statusBarTextColor      = colornames.Gold
	increasedStatValueColor = ui.RGBFromHex("08CB00")
	decreasedStatValueColor = ui.RGBFromHex("D70654")

	// Board colours.
	boardBgColor           = ui.RGBFromHex("607B8F")
	boardCellBgColor       = ui.DarkenRGB(boardBgColor, 10)
	boardGridColor         = color.RGBA{0x45, 0x63, 0x7a, 255}
	unitDropZoneColor      = ui.RGBFromHex("A7E399")
	unitDropZoneHoverColor = ui.DarkenRGB(unitDropZoneColor, 50)
	unitKilledBgColor      = ui.RGBFromHex("BFA28C")

	// Ability colours.
	abilitiesPanelBgColor       = ui.DarkenRGB(footerBgColor, 5)
	basicAttackBgColor          = ui.RGBFromHex("E97F4A")
	abilityBgColor              = ui.RGBFromHex("8CA9FF")
	passiveAbilityBgColor       = ui.RGBFromHex("9B8EC7")
	abilityEnemyTargetCellColor = ui.RGBFromHex("D70654")
	abilityAllyTargetCellColor  = ui.RGBFromHex("08CB00")
	abilityRangeCellColor       = ui.RGBFromHex("7c9493")
	// pulse colors for selected ability card
	activeAbilityBgColor        = ui.RGBFromHex("48A111")
	activeAbilityFgColor        = ui.RGBFromHex("f5df4d")
	abilitySelectedPulseColor1  = activeAbilityFgColor
	abilitySelectedPulseColor2  = activeAbilityBgColor
	abilityCooldownCounterColor = colornames.White

	// Unit status tooltip
	neutralStatusNameColor  = "f8f8ff"
	positiveStatusNameColor = "98fb98"
	negativeStatusNameColor = "ff6347"
	statusDurationColor     = "87ceeb"
)

// Font faces used across the arena package.
var (
	toolTipTitleTF           = ui.TextFace(28)
	toolTipTextTF            = ui.TextFace(18)
	abilityCooldownCounterTF = ui.TextFace(40)
)
