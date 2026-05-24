package arena

import (
	"github.com/ognev-dev/goplease-ebitengine-client/asset"
	"github.com/setanarut/anim"
)

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
