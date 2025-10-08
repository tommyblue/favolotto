package favolotto

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/tommyblue/favolotto/internal/colors"
)

//go:embed store/assets/boot.mp3
var bootSound []byte

type Config struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Store       string `json:"store"`
	Development bool   `json:"development"`
	NfcDriver   string `json:"nfc_driver"`
}

type Favolotto struct {
	config Config
}

func New(config Config) *Favolotto {
	return &Favolotto{
		config: config,
	}

}

type CtxKey string

var (
	CtxDevelopment CtxKey = "development"
)

func (f *Favolotto) Run(ctx context.Context) error {
	ctx = context.WithValue(ctx, CtxDevelopment, f.config.Development)

	inNfc := make(chan string)          // channel for NFC tag IDs
	inFname := make(chan string)        // channel for audio files to play
	ctrl := make(chan string)           // channel for control commands
	ledColor := make(chan colors.Color) // channel for LED color commands

	if f.config.Development {
		go func() {
			fmt.Println("Control mode: type 'p' to pause, 'r' to resume, 's' to stop")

			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				s := scanner.Text()
				fmt.Printf("You typed: %s\n", s)

				switch s {
				case "1":
					ledColor <- colors.Red
				case "2":
					ledColor <- colors.Green
				case "3":
					ledColor <- colors.Blue
				case "4":
					ledColor <- colors.Brown
				case "5":
					ledColor <- colors.White
				case "6":
					ledColor <- colors.Violet
				case "7":
					ledColor <- colors.Black
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
	}

	wg := &sync.WaitGroup{}

	led := NewLED(ledColor)

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

	audio, err := NewAudio("store", inFname, ctrl, ledColor)
	if err != nil {
		log.Fatal("Error creating audio: ", err.Error())
	}
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

	nfc, err := NewNFC(f.config.NfcDriver, inNfc)
	if err != nil {
		log.Fatal("Error creating NFC: ", err.Error())
	}

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

	bootSoundPath := filepath.Join(f.config.Store, "assets", "boot.mp3")
	if err := ensureBootFile(bootSoundPath); err == nil {
		inFname <- bootSoundPath
	}

	wg.Wait()

	<-ctx.Done()
	return nil
}

func ensureBootFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("File '%s' not found. Creating it now.", path)

		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directories: %w", err)
		}

		if err := os.WriteFile(path, bootSound, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}

	return nil
}
