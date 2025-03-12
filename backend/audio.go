package favolotto

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

type Audio struct {
	in        <-chan string
	storePath string

	ctrl        *beep.Ctrl
	volume      *effects.Volume
	currentFile string
}

func NewAudio(storePath string, in <-chan string) *Audio {
	return &Audio{
		in:        in,
		storePath: storePath,
	}
}

func (a *Audio) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			a.Cleanup()
			return
		case fname := <-a.in:
			// if the same file is requested, do nothing
			if a.currentFile == fname {
				break
			}
			a.Cleanup()

			f, err := os.Open(filepath.Join(a.storePath, fname))
			if err != nil {
				log.Printf("Error opening file %s: %v", fname, err)
				continue
			}

			streamer, format, err := mp3.Decode(f)
			if err != nil {
				log.Printf("Error decoding file %s: %v", fname, err)
				continue
			}

			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			a.ctrl = &beep.Ctrl{Streamer: streamer, Paused: false}
			a.volume = &effects.Volume{
				Streamer: a.ctrl,
				Base:     2,
				Volume:   0,
				Silent:   false,
			}

			speaker.Play(streamer)
			// done := make(chan bool)
			// speaker.Play(beep.Seq(streamer, beep.Callback(func() {
			// 	done <- true
			// })))
			// <-done
			streamer.Close()
		}
	}
}

func (a *Audio) Toggle() {
	speaker.Lock()
	a.ctrl.Paused = !a.ctrl.Paused
	speaker.Unlock()
}

func (a *Audio) VolumeUp() {
	if a.volume != nil {
		speaker.Lock()
		a.volume.Volume += 0.5
		speaker.Unlock()
	}
}

func (a *Audio) VolumeDown() {
	if a.volume != nil {
		speaker.Lock()
		a.volume.Volume -= 0.5
		speaker.Unlock()
	}
}

func (a *Audio) Cleanup() {
	if a.ctrl != nil {
		speaker.Lock()
		a.ctrl.Paused = true
		speaker.Unlock()
	}

	a.ctrl = nil
	a.volume = nil
}
