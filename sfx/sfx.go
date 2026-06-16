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

var volume float64 = 1.0

// Volume returns the current global sound volume in range [0, 1].
func Volume() float64 {
	return volume
}

// SetVolume sets the global sound volume, clamping the value to [0, 1].
func SetVolume(v float64) {
	switch {
	case v < 0:
		v = 0
	case v > 1:
		v = 1
	}
	volume = v
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

	p.SetVolume(volume)
	p.Play()
}
