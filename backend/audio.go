package favolotto

import (
	"context"
	"fmt"
	"log"
	"os"
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
			fmt.Println("Audio file requested: ", fname)

			f, err := os.Open(fname)
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
			buffer := beep.NewBuffer(format)
			buffer.Append(streamer)
			streamer.Close()
			song := buffer.Streamer(0, buffer.Len())

			a.ctrl = &beep.Ctrl{Streamer: song, Paused: false}
			a.volume = &effects.Volume{
				Streamer: a.ctrl,
				Base:     2,
				Volume:   0,
				Silent:   false,
			}

			done := make(chan bool)
			speaker.Play(beep.Seq(song, beep.Callback(func() {
				done <- true
			})))
			<-done
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
