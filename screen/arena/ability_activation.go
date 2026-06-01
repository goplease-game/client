package arena

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/hex"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

// onAbilityCardClicked is called when the player clicks an ability card.
// Instant abilities fire immediately; others enter targeting mode.
func (s *Screen) onAbilityCardClicked(ab ability.Ability, card *widget.Container, bgColor color.Color) {
	if ab.IsPassive {
		return
	}

	// Another ability is already in targeting mode — ignore.
	if s.selectedAbility != nil && s.selectedAbility.ID != ab.ID {
		return
	}

	// Second click on the same ability — cancel targeting mode.
	if s.selectedAbility != nil && s.selectedAbility.ID == ab.ID {
		s.cancelAbilitySelection()
		return
	}

	if ab.Activation == ability.Instant {
		s.sendUseAbility(ab.ID, ds.HexCoord{})
		return
	}

	s.selectedAbility = &ab
	s.selectedAbilityCard = card
	s.selectedAbilityCardColor = bgColor
	s.selectedAbilityActiveColor = activeAbilityBgColor
	s.setStatus(fmt.Sprintf("Select target to use %s (Press ESC to cancel)", ab.Name))
}

// onCellClickedWithAbility checks if a selected ability can be applied to coord.
// Returns true if the click was consumed by ability activation.
func (s *Screen) onCellClickedWithAbility(coord ds.HexCoord) bool {
	if s.selectedAbility == nil {
		return false
	}

	ab := *s.selectedAbility
	cell := s.board.Cells[coord]

	if !s.isValidAbilityTarget(ab, coord, cell) {
		// Clicked invalid target — cancel ability selection.
		s.cancelAbilitySelection()
		return true
	}

	s.sendUseAbility(ab.ID, coord)
	s.cancelAbilitySelection()
	return true
}

// isValidAbilityTarget checks whether coord is a valid target for ab.
func (s *Screen) isValidAbilityTarget(ab ability.Ability, coord ds.HexCoord, cell *ds.BoardCell) bool {
	caster := s.unitByID(s.activeUnitID)

	// Must be within range.
	if ab.Range > 0 && hex.Distance(caster.Pos, coord) > ab.Range {
		return false
	}

	unit := (*ds.Unit)(nil)
	if cell != nil {
		unit = cell.Unit
	}

	switch ab.Activation {
	case ability.SelectEnemy:
		return unit != nil && unit.IsOpponent != caster.IsOpponent
	case ability.SelectAlly:
		return unit != nil && unit.IsOpponent == caster.IsOpponent && unit.ID != caster.ID
	case ability.SelectAllyOrSelf:
		return unit != nil && unit.IsOpponent == caster.IsOpponent
	case ability.SelectAnyUnit:
		return unit != nil
	case ability.SelectFreeCell:
		return unit == nil
	case ability.SelectAny:
		return true
	}

	return false
}

// cancelAbilitySelection clears the selected ability and restores board highlights.
func (s *Screen) cancelAbilitySelection() {
	if s.selectedAbilityCard != nil {
		s.selectedAbilityCard.SetBackgroundImage(
			image.NewNineSliceColor(s.selectedAbilityCardColor),
		)
		s.selectedAbilityCard = nil
		s.selectedAbilityCardColor = nil
		s.selectedAbilityActiveColor = nil
	}

	if s.selectedAbilityIcon != nil && s.selectedAbility != nil {
		s.selectedAbilityIcon.Image = abilityImage(string(s.selectedAbility.ID))
		s.selectedAbilityIcon = nil
	}

	s.selectedAbility = nil
	s.clearAbilityHighlight()
	s.updateActiveUnitStatusLabel()
}

// sendUseAbility sends the UseAbility message to the server.
func (s *Screen) sendUseAbility(abilityID ability.ID, target ds.HexCoord) {
	u := s.unitByID(s.activeUnitID)
	if u == nil {
		return
	}

	s.server.Send(ws.OutMessage{
		Action: ws.UseAbility,
		Data: ds.UseAbilityPayload{
			UnitID:    s.activeUnitID,
			AbilityID: abilityID,
			Target:    target,
		},
	})

	ab := ability.ByID(abilityID)
	if ab.Cooldown > 0 {
		if u.Cooldowns == nil {
			u.Cooldowns = make(map[ability.ID]int)
		}
		u.Cooldowns[abilityID] = ab.Cooldown
	}
	u.CurrentAP--

	s.clearAbilityHighlight()
	s.showAbilityPanel(u)
	s.updateNextActionLabel()
	s.updateActiveUnitStatusLabel()

	pending := &pendingVisuals{}
	s.pendingVisuals = pending
	s.playAbilityFx(abilityID, u, target, func() {
		pending.fxDone = true
		s.tryFlushPendingVisuals(pending)
	})
}
