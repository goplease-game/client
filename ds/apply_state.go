package ds

import "github.com/ognev-dev/goplease-ebitengine-client/ability/effect"

// ApplyState represents a single, atomic state mutation applied to a unit.
// Sequential execution of these states forms the visual timeline on the client side.
// TODO need review:
//
//	if I apply ChangeX, I always need to apply SetX, Change & Set can be combined into one struct?
//	ApplyState constructors also will be simpler
type ApplyState struct {
	ToUnitID string `json:"to_unit_id"`

	// Movement
	MoveTo *HexCoord `json:"move_to,omitempty"` // New position on the grid

	// Delta changes used to trigger floating text or combat UI animations
	ChangeHP     *int `json:"change_hp,omitempty"`
	ChangeAP     *int `json:"change_ap,omitempty"`
	ChangeMP     *int `json:"change_mp,omitempty"`
	ChangeShield *int `json:"change_shield,omitempty"`
	ChangeAtk    *int `json:"change_atk,omitempty"`

	// Absolute values used for hard state synchronization after the animation plays
	SetHP     *int `json:"set_hp,omitempty"`
	SetAP     *int `json:"set_ap,omitempty"`
	SetMP     *int `json:"set_mp,omitempty"`
	SetShield *int `json:"set_shield,omitempty"`
	SetAtk    *int `json:"set_atk,omitempty"`

	// Statuses and effects
	IsDead       bool               `json:"is_dead,omitempty"`
	AddStatus    *StatusWithMeta    `json:"add_status,omitempty"`
	RemoveStatus *effect.StatusType `json:"remove_status,omitempty"`
}

// ApplyStates represents a collection of atomic state mutations bound to a single unit.
type ApplyStates []ApplyState

// NewUnitStates initializes a new ApplyStates slice dedicated to a specific unit's timeline events.
func NewUnitStates(ss ...ApplyState) ApplyStates {
	return ss
}

// Add appends new states to this unit's timeline.
func (s *ApplyStates) Add(ss ...ApplyState) ApplyStates {
	*s = append(*s, ss...)
	return *s
}

// ToUnitID explicitly assigns or overwrites the target unit ID for all currently accumulated states in this slice.
// Be careful
func (s *ApplyStates) ToUnitID(id string) ApplyStates {
	valueS := *s
	for i := range valueS {
		valueS[i].ToUnitID = id
	}
	return valueS
}
