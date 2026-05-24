package ds

import "github.com/ognev-dev/goplease-ebitengine-client/ability"

type UseAbilityPayload struct {
	UnitID    string     `json:"unit_id"`
	AbilityID ability.ID `json:"ability_id"`
	Target    HexCoord   `json:"target,omitempty"`
}
