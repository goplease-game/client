package sfx

import (
	"bytes"
	"log"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/ognev-dev/goplease-ebitengine-client/asset"
)

var audioCtx *audio.Context

func init() {
	audioCtx = audio.NewContext(44100)
}

// Play plays a sound file from the sfx asset directory.
// audioCtx must not be nil.
func Play(name string) {
	data := asset.Load("sfx/" + name)
	stream, err := vorbis.DecodeWithoutResampling(bytes.NewReader(data))
	if err != nil {
		log.Printf("sfx.Play: failed to decode %s: %v", name, err)
		return
	}
	p, err := audioCtx.NewPlayer(stream)
	if err != nil {
		log.Printf("sfx.Play: failed to create player %s: %v", name, err)
		return
	}

	p.Play()
}
