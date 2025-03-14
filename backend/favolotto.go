package favolotto

import (
	"bufio"
	"context"
	"fmt"
	"os"
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

	// nfc := NewNFC(inNfc)
	// go nfc.Run(ctx)
	go func() {
		fmt.Println("Type whatever you want!")

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Printf("You typed: %s\n", scanner.Text())
			if scanner.Text() == "a" {
				inNfc <- "1234"
			}
			if scanner.Text() == "b" {
				inNfc <- "5678"
			}
		}
	}()

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
