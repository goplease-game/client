package arena

import (
	"image/color"

	"github.com/goplease-game/client/ui"
	"golang.org/x/image/colornames"
)

const (
	unitStunnedPic = "knockout.png"
	unitKilledPic  = "dead-head.png"
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
	negativeStatusIconColor = ui.RGBFromHex("D51C39")
	positiveStatusIconColor = ui.RGBFromHex("059212") // + neutral

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
	// pulse colors for selected ability card.
	activeAbilityBgColor        = ui.RGBFromHex("48A111")
	activeAbilityFgColor        = ui.RGBFromHex("f5df4d")
	abilitySelectedPulseColor1  = activeAbilityFgColor
	abilitySelectedPulseColor2  = activeAbilityBgColor
	abilityCooldownCounterColor = colornames.White

	// Unit status tooltip.
	neutralStatusNameColor  = "f8f8ff"
	positiveStatusNameColor = "98fb98"
	negativeStatusNameColor = "ff6347"
	statusDurationColor     = "87ceeb"

	// game menu.
	nameColor                = ui.RGBFromHex("#00a8e8")
	menuButtonBgColor        = ui.RGBFromHex("#73A5CA")
	menuButtonHoverBgColor   = ui.LightenRGB(menuButtonBgColor, 35)
	menuButtonTextColor      = ui.RGBFromHex("FFF8DE")
	menuButtonHoverTextColor = ui.DarkenRGB(menuButtonBgColor, 45)

	gameOverLoseColor = ui.RGBFromHex("F5004F")
	gameOverWinColor  = ui.RGBFromHex("F9E400")

	// tutorial
	tutorialTitleTextColor        = ui.RGBFromHex("315a87")
	tutorialTextColor             = colornames.Whitesmoke
	tutorialBtnDisabledTextColor  = color.NRGBA{0x88, 0x88, 0x88, 0xff}
	tutorialSkipBtnTextColor      = color.NRGBA{0x88, 0x88, 0x88, 0xff}
	tutorialTextBgColor           = ui.RGBFromHex("1a3a5c")
	tutorialTitleBgColor          = colornames.Whitesmoke
	tutorialBtnBgColor            = color.NRGBA{0x1a, 0x6b, 0xc4, 0xff}
	tutorialBtnHoverBgColor       = color.NRGBA{0x22, 0x88, 0xf0, 0xff}
	tutorialBtnPressedBgColor     = color.NRGBA{0x10, 0x50, 0x99, 0xff}
	tutorialBtnDisabledBgColor    = color.NRGBA{0x33, 0x33, 0x33, 0xff}
	tutorialSkipBtnBgColor        = color.NRGBA{0x22, 0x22, 0x22, 0x00}
	tutorialSkipBtnHoverBgColor   = color.NRGBA{0x33, 0x33, 0x33, 0x88}
	tutorialSkipBtnPressedBgColor = color.NRGBA{0x11, 0x11, 0x11, 0x88}
)

// Font faces used across the arena package.
var (
	toolTipTitleTF           = ui.TextFace(28)
	toolTipTextTF            = ui.TextFace(18)
	abilityCooldownCounterTF = ui.TextFace(40)

	// tutorial
	tutorialTitleTF = ui.TextFace(14)
	tutorialTextTF  = ui.TextFace(16)
)
