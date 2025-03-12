package favolotto

import (
	"context"
)

type Config struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	Store string `json:"store"`
}

type Favolotto struct {
	config Config
}

func New(config Config) *Favolotto {
	return &Favolotto{
		config: config,
	}

}

func (f *Favolotto) Run(ctx context.Context) error {
	inNfc := make(chan string)   // channel for NFC tag IDs
	inFname := make(chan string) // channel for audio files to play

	// TODO: manage errors from the following goroutines

	nfc := NewNFC(inNfc)
	go nfc.Run(ctx)

	// Listen for GPIO button press and play/pause audio or volume up/down
	// Use GPIO LEDs to indicate the current state of the audio player

	store := NewStore(f.config.Store, inNfc, inFname)
	go store.Run(ctx)

	audio := NewAudio("store", inFname)
	go audio.Run(ctx)

	// initialize web server
	// httpServer := NewHTTPServer(f.config.Host, f.config.Port)
	// go httpServer.Run(ctx)

	<-ctx.Done()
	return nil
}
