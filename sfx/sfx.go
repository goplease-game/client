// Package sfx provides functionality for playing sound effects and multiple background music tracks simultaneously.
package sfx

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/goplease-game/client/asset"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

var audioCtx *audio.Context

func init() {
	audioCtx = audio.NewContext(44100)
}

var (
	volume      = 1.0
	musicVolume = 1.0

	// activeTracks keeps track of all currently playing music tracks to allow global volume updates.
	activeTracks   = make(map[*MusicTrack]struct{})
	activeTracksMu sync.Mutex
)

// MusicTrack represents an individual playing music stream with its own lifecycle and effects.
type MusicTrack struct {
	player   *audio.Player
	stopChan chan struct{}
	mu       sync.Mutex
	isClosed bool
}

// Volume returns the current global sound volume in range [0, 1].
func Volume() float64 {
	return volume
}

// SetVolume sets the global sound volume, clamping the value to [0, 1].
// It dynamically updates the volume of all currently active music tracks.
func SetVolume(v float64) {
	switch {
	case v < 0:
		v = 0
	case v > 1:
		v = 1
	}
	volume = v

	activeTracksMu.Lock()
	defer activeTracksMu.Unlock()
	for track := range activeTracks {
		track.mu.Lock()
		if !track.isClosed && track.player.IsPlaying() {
			// In a production scenario, you might want to store the track's individual base volume.
			// For simplicity, we apply the global modifiers.
			track.player.SetVolume(musicVolume * volume)
		}
		track.mu.Unlock()
	}
}

// Play plays a sound effect file from the sfx asset directory once.
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

// PlayMusic starts playing a music file from the music asset directory and returns a MusicTrack controller.
// If loop is true, the track repeats indefinitely. fadeInDuration specifies how long it takes for the track
// to reach full volume (use 0 for immediate playback).
func PlayMusic(name string, loop bool, fadeInDurationOpt ...time.Duration) *MusicTrack {
	data := asset.Load("music/" + name)
	stream, err := vorbis.DecodeWithSampleRate(44100, bytes.NewReader(data))
	if err != nil {
		log.Printf("sfx.PlayMusic: failed to decode %s: %v", name, err)
		return nil
	}

	var finalStream io.ReadSeeker = stream
	if loop {
		finalStream = audio.NewInfiniteLoop(stream, stream.Length())
	}

	p, err := audioCtx.NewPlayer(finalStream)
	if err != nil {
		log.Printf("sfx.PlayMusic: failed to create player %s: %v", name, err)
		return nil
	}

	track := &MusicTrack{
		player:   p,
		stopChan: make(chan struct{}),
	}

	// Register track globally
	activeTracksMu.Lock()
	activeTracks[track] = struct{}{}
	activeTracksMu.Unlock()

	targetVolume := musicVolume * volume

	var fadeInDuration time.Duration
	if len(fadeInDurationOpt) > 0 {
		fadeInDuration = fadeInDurationOpt[0]
	}
	if fadeInDuration > 0 {
		p.SetVolume(0)
		p.Play()

		// Run Fade-In effect in a separate goroutine
		go func(t *MusicTrack, duration time.Duration, maxVol float64) {
			steps := 20
			stepDuration := duration / time.Duration(steps)

			for i := 1; i <= steps; i++ {
				select {
				case <-t.stopChan:
					return
				default:
					t.mu.Lock()
					if t.isClosed {
						t.mu.Unlock()
						return
					}
					currentFadeVol := maxVol * (float64(i) / float64(steps))
					t.player.SetVolume(currentFadeVol)
					t.mu.Unlock()
					time.Sleep(stepDuration)
				}
			}
		}(track, fadeInDuration, targetVolume)
	} else {
		p.SetVolume(targetVolume)
		p.Play()
	}

	return track
}

// Stop immediately pauses and closes the music track, removing it from active tracking.
func (t *MusicTrack) Stop() {
	if t == nil {
		return
	}
	t.mu.Lock()
	if t.isClosed {
		t.mu.Unlock()
		return
	}
	t.isClosed = true
	close(t.stopChan)
	t.player.Pause()
	err := t.player.Close()
	if err != nil {
		fmt.Println("could not stop player:", err)
	}
	t.mu.Unlock()

	// Unregister from active tracks
	activeTracksMu.Lock()
	delete(activeTracks, t)
	activeTracksMu.Unlock()
}

// FadeOut smoothly decreases the volume of the track to 0 over the specified duration, then stops it.
func (t *MusicTrack) FadeOut(duration time.Duration) {
	if t == nil {
		return
	}

	go func(track *MusicTrack, dur time.Duration) {
		track.mu.Lock()
		if track.isClosed || !track.player.IsPlaying() {
			track.mu.Unlock()
			return
		}
		startVolume := track.player.Volume()
		track.mu.Unlock()

		steps := 20
		stepDuration := dur / time.Duration(steps)

		for i := steps; i >= 0; i-- {
			select {
			case <-track.stopChan:
				return
			default:
				track.mu.Lock()
				if track.isClosed {
					track.mu.Unlock()
					return
				}
				targetVol := startVolume * (float64(i) / float64(steps))
				track.player.SetVolume(targetVol)
				track.mu.Unlock()
				time.Sleep(stepDuration)
			}
		}

		track.Stop()
	}(t, duration)
}

// StopAll immediately stops and closes all currently active music tracks.
func StopAll() {
	activeTracksMu.Lock()
	// Create a temporary slice of tracks to avoid modifying the map while iterating over it.
	tracksToStop := make([]*MusicTrack, 0, len(activeTracks))
	for track := range activeTracks {
		tracksToStop = append(tracksToStop, track)
	}
	activeTracksMu.Unlock()

	// Stop each track. The individual Stop() method will handle internal locks
	// and unregister the track from the activeTracks map safely.
	for _, track := range tracksToStop {
		track.Stop()
	}
}
