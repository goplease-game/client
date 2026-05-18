package arena

import (
	"github.com/ognev-dev/goplease-ebitengine-client/asset"
	"github.com/setanarut/anim"
)

var animDropArrow *anim.AnimationPlayer

func initDropPointAnim() {
	img := asset.Image("drop_point_a.png")
	sheet := anim.Atlas{
		Name:  "Default",
		Image: img,
	}

	animDropArrow = anim.NewAnimationPlayer(sheet)
	animDropArrow.NewAnim("idle", 0, 0, 64, 64, 3, true, false, 6)
	animDropArrow.SetAnim("idle")
}
