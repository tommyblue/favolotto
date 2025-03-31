//go:build pn532

package pn532

import (
	"context"
	"encoding/hex"
	"log"
	"time"

	"github.com/clausecker/nfc/v2"
	"github.com/pkg/errors"
	"github.com/warthog618/gpiod"
)

var modulations = []nfc.Modulation{
	{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106},
	// {Type: nfc.ISO14443b, BaudRate: nfc.Nbr106},
	// {Type: nfc.Felica, BaudRate: nfc.Nbr212},
	// {Type: nfc.Felica, BaudRate: nfc.Nbr424},
	// {Type: nfc.Jewel, BaudRate: nfc.Nbr106},
	// {Type: nfc.ISO14443biClass, BaudRate: nfc.Nbr106},
}

type TagReader struct {
	tagChannel chan string
	reader     *nfc.Device
	resetPin   int
}

// Reset performs a hardware reset by pulling the ResetPin low and then releasing.
func reset(resetPin int) {
	log.Println("Resetting the reader..")

	c, err := gpiod.NewChip("gpiochip0")
	pin, err := c.RequestLine(resetPin, gpiod.AsOutput(0))
	if err != nil {
		log.Println(err)
		return
	}

	err = pin.SetValue(1)
	if err != nil {
		log.Println(err)
		return
	}

	time.Sleep(time.Millisecond * 400)
	err = pin.SetValue(0)
	if err != nil {
		log.Println(err)
		return
	}

	time.Sleep(time.Millisecond * 400)

	err = pin.SetValue(1)
	time.Sleep(time.Millisecond * 100)
	if err != nil {
		log.Println(err)
		return
	}
}

func New() (*TagReader, error) {
	var resetPin = 19

	dev, err := nfc.Open("")
	if err != nil {
		reset(resetPin)
		return nil, errors.Wrap(err, "Cannot communicate with the device")
	}

	if err := dev.InitiatorInit(); err != nil {
		return nil, errors.Wrap(err, "Cannot initialize the reader")
	}

	return &TagReader{reader: &dev, tagChannel: make(chan string, 10), resetPin: resetPin}, nil
}

func (reader *TagReader) Stop() error {
	return reader.reader.Close()
}

func (reader *TagReader) Read() <-chan string {
	return reader.tagChannel
}

func getIdFromTarget(target nfc.Target) (*string, error) {
	var UID string
	// Transform the target to a specific tag Type and send the UID to the channel
	switch target.Modulation() {
	case nfc.Modulation{Type: nfc.ISO14443a, BaudRate: nfc.Nbr106}:
		var card = target.(*nfc.ISO14443aTarget)
		var UIDLen = card.UIDLen
		var ID = card.UID
		UID = hex.EncodeToString(ID[:UIDLen])
	default:
		return nil, errors.New("Unknown modulation")
	}

	return &UID, nil
}

func (reader *TagReader) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Poll for 300ms
			tagCount, target, err := reader.reader.InitiatorPollTarget(modulations, 3, 200*time.Millisecond)
			if err != nil {
				log.Println("Error polling the reader", err)
				continue
			}

			// Check if a tag was detected
			if tagCount > 0 {
				id, err := getIdFromTarget(target)
				if err != nil {
					log.Println("Error getting ID from target", err)
					continue
				}
				reader.tagChannel <- *id
			}

			time.Sleep(time.Millisecond * 100)
		}
	}
}
