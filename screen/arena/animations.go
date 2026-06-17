package arena

import (
	"github.com/goplease-game/client/asset"
	"github.com/goplease-game/client/ds"
	"github.com/setanarut/anim"
)

type pendingVisuals struct {
	applyStates []ds.ApplyState
	fxDone      bool
	serverDone  bool
}

// tryFlushPendingVisuals fires visual feedback only when both fx and server response are ready.
func (s *Screen) tryFlushPendingVisuals(p *pendingVisuals) {
	if !p.fxDone || !p.serverDone {
		return
	}
	s.pendingVisuals = nil
	for _, st := range p.applyStates {
		if target := s.unitByID(st.ToUnitID); target != nil {
			s.applyStateVisuals(target, st)
		}
	}
}

// animDropArrow is the shared animation player for the drop-zone arrow.
// Initialised once via initDropPointAnim and advanced each frame via Update.
var animDropArrow *anim.AnimationPlayer

// initDropPointAnim loads the drop-arrow sprite sheet and registers the idle
// animation. Must be called once before the arena screen renders.
func initDropPointAnim() {
	img := asset.Image("drop_point_a.png")
	sheet := anim.Atlas{
		Name:  "Default",
		Image: img,
	}

	animDropArrow = anim.NewAnimationPlayer(sheet)
	animDropArrow.NewAnim("idle", 0, 0, 54, 54, 6, true, false, 30)
	animDropArrow.SetAnim("idle")
}
