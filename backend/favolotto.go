package favolotto

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
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
	ctrl := make(chan string)    // channel for control commands

	go func() {
		fmt.Println("Type whatever you want!")

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Printf("You typed: %s\n", scanner.Text())
			switch scanner.Text() {
			case "1":
				inNfc <- "1234"
			case "2":
				inNfc <- "5678"
			case "p":
				ctrl <- Pause
			case "r":
				ctrl <- Resume
			case "s":
				ctrl <- Stop
			default:
				fmt.Println("Unknown command")
			}
		}
	}()

	wg := &sync.WaitGroup{}

	// TODO: Use LEDs to indicate the current state of the audio player
	led := NewLED()

	wg.Add(1)
	go func() {
		defer wg.Done()
		led.Run(ctx)
	}()

	store, err := NewStore(f.config.Store, inNfc, inFname)
	if err != nil {
		log.Fatal("Error creating store: ", err.Error())
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		store.Run(ctx)
	}()

	audio := NewAudio("store", inFname, ctrl)
	wg.Add(1)
	go func() {
		defer wg.Done()
		audio.Run(ctx)
	}()

	// initialize web server
	httpServer := NewHTTPServer(f.config.Host, f.config.Port, store)

	wg.Add(1)
	go func() {
		defer wg.Done()
		httpServer.Run(ctx)
	}()

	nfc := NewNFC(inNfc)

	wg.Add(1)
	go func() {
		defer wg.Done()
		nfc.Run(ctx)
	}()

	button, err := NewButton(ctrl)
	if err != nil {
		log.Fatal("Error creating button: ", err.Error())
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		button.Run(ctx)
	}()

	wg.Wait()

	<-ctx.Done()
	return nil
}
