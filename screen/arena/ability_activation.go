package arena

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/ability"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/grid"
	"github.com/goplease-game/client/sfx"
	"github.com/goplease-game/client/ws"
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
		s.sendUseAbility(ab.ID, nil)
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
	if !s.isValidAbilityTarget(ab, coord) {
		sfx.Play(selectError)
		// Clicked invalid target — cancel ability selection.
		s.cancelAbilitySelection()
		return true
	}

	s.sendUseAbility(ab.ID, &coord)
	s.cancelAbilitySelection()
	return true
}

// isValidAbilityTarget checks whether coord is a valid target for ab.
func (s *Screen) isValidAbilityTarget(ab ability.Ability, coord ds.HexCoord) bool {
	caster := s.unitByID(s.activeUnitID)
	if caster == nil {
		return false
	}

	// Must be within range.
	if ab.Range > 0 && grid.Distance(caster.Pos, coord) > ab.Range {
		return false
	}

	cell, ok := s.board.Cells[coord]
	if !ok || cell == nil {
		return false
	}

	if ab.Activation == ability.SelectAny {
		return true
	}

	u := cell.Unit
	if u == nil {
		return ab.Activation == ability.SelectFreeCell
	}

	if u.IsDead {
		return false
	}

	// Only alive units left in here
	switch ab.TargetMode {
	case ability.TargetSelf:
		if u.ID == caster.ID {
			return true
		}
	case ability.TargetAllies:
		if u.IsAlly(caster) {
			return true
		}
	case ability.TargetAlliesAndSelf:
		if u.ID == caster.ID || u.IsAlly(caster) {
			return true
		}
	case ability.TargetEnemies:
		if u.IsEnemy(caster) {
			if provokerID := getProvokingUnitID(caster); provokerID != "" {
				return u.ID == provokerID
			}
			return true
		}
	case ability.TargetEnemiesAndSelf:
		if u.ID == caster.ID || u.IsEnemy(caster) {
			if provokerID := getProvokingUnitID(caster); provokerID != "" {
				return u.ID == provokerID
			}
			return true
		}
	}

	switch ab.Activation {
	case ability.SelectEnemy:
		if provokerID := getProvokingUnitID(caster); provokerID != "" {
			return u.ID == provokerID
		}
		return u.IsEnemy(caster)
	case ability.SelectAlly:
		return u.IsAlly(caster) && u.ID != caster.ID
	case ability.SelectAllyOrSelf:
		return u.IsAlly(caster)
	case ability.SelectAnyUnit:
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
func (s *Screen) sendUseAbility(abilityID ability.ID, target *ds.HexCoord) {
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
	u.SetCooldown(ab.ID, ab.Cooldown)

	if u.CurrentAP > 0 {
		u.CurrentAP--
	} else {
		u.PhantomAPUsedThisTurn++
		s.player.PhantomAP--
	}

	s.clearAbilityHighlight()
	s.showAbilityPanel(u)
	s.updateNextActionLabel()
	s.updateActiveUnitStatusLabel()
	s.showUnitOnBoard(u)

	pending := &pendingVisuals{}
	s.pendingVisuals = pending

	var fxTarget ds.HexCoord
	if target != nil {
		fxTarget = *target
	}
	s.playAbilityFx(abilityID, u, fxTarget, func() {
		pending.fxDone = true
		s.tryFlushPendingVisuals(pending)
	})
}
