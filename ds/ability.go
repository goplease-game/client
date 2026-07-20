// Package ds ...
package ds

import "github.com/goplease-game/game-server/ability"

// UseAbilityPayload is the payload for requesting that a unit use an ability,
// optionally targeting a specific hex coordinate.
type UseAbilityPayload struct {
	UnitID    string     `json:"unit_id"`
	AbilityID ability.ID `json:"ability_id"`
	Target    *HexCoord  `json:"target,omitempty"`
}
