package arena

import (
	"image/color"

	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"golang.org/x/image/colornames"
)

// ---- COLORS ----
var (
	// LAYOUT
	bodyBgColor      = color.NRGBA{0x13, 0x1a, 0x22, 0xff}
	footerBgColor    = ui.RGBFromHex("5682B1")
	headerBgColor    = ui.RGBFromHex("5682B1")
	statusBarBgColor = ui.DarkenRGB(footerBgColor, 5)

	// UNITS
	unitPanelBgColor        = ui.DarkenRGB(footerBgColor, 5)
	unitCardBgColor         = ui.DarkenRGB(unitPanelBgColor, 20)
	unitCardHoverBgColor    = ui.DarkenRGB(unitCardBgColor, 15)
	unitCardHoverFgColor    = ui.RGBFromHex("f5df4d")
	unitCardHighlightColor  = colornames.Gold
	unitDragBgColor         = ui.RGBFromHex("78B3CE")
	unitFriendlyBgColor     = ui.RGBFromHex("B0DB9C")
	unitEnemyBgColor        = ui.RGBFromHex("EA7B7B")
	selectedUnitBorderColor = colornames.Red
	unitMoveToCellColor     = ui.RGBFromHex("49687e")
	unitPulseColor1         = ui.RGBFromHex("FFC700")
	unitPulseColor2         = ui.DarkenRGB(unitPulseColor1, 80)

	// TOOLTIPS
	ttBgColor     = ui.RGBFromHex("42668d")
	ttBorderColor = ui.LightenRGB(ttBgColor, 50)
	ttTitleColor  = ui.RGBFromHex("f5df4d")
	ttTextColor   = colornames.Ghostwhite

	// HUD
	hpColor            = colornames.Tomato
	atkColor           = colornames.Orange
	mpColor            = colornames.Skyblue
	shieldColor        = colornames.Gold
	statusBarTextColor = colornames.Gold

	// BOARD
	boardBgColor           = ui.RGBFromHex("607B8F")
	boardCellBgColor       = ui.DarkenRGB(boardBgColor, 10)
	unitDropZoneColor      = ui.RGBFromHex("A7E399")
	unitDropZoneHoverColor = ui.DarkenRGB(unitDropZoneColor, 50)

	// ABILITIES
	abilitiesPanelBgColor = ui.DarkenRGB(footerBgColor, 5)
	basicAttackBgColor    = ui.RGBFromHex("E97F4A")
	abilityBgColor        = ui.RGBFromHex("8CA9FF")
	passiveAbilityBgColor = ui.RGBFromHex("9B8EC7")
)

// FONTS
var (
	toolTipTitleTF = ui.TextFace(28)
	toolTipTextTF  = ui.TextFace(18)
)
