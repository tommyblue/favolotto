package favolotto

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tommyblue/favolotto/internal/nfc_serial"
	"github.com/tommyblue/favolotto/internal/pn532"
)

type NfcDriver interface {
	Run(ctx context.Context) error
	Stop() error
	Read() <-chan string
}

type Nfc struct {
	driver NfcDriver
	in     chan<- string
}

func NewNFC(driverName string, in chan<- string, rstNfc chan<- bool) (*Nfc, error) {
	var driver NfcDriver
	var err error
	switch driverName {
	case "pn532":
		driver, err = pn532.New(rstNfc)
	case "serial":
		driver, err = nfc_serial.New(rstNfc)
	default:
		log.Printf("Unknown NFC driver: %s\n", driverName)
		return nil, fmt.Errorf("unknown NFC driver: %s", driverName)
	}

	if err != nil {
		log.Println("Error while init nfc device\n", err)
		return nil, err
	}

	return &Nfc{
		in:     in,
		driver: driver,
	}, nil
}

func (n *Nfc) Run(ctx context.Context) {
	isDevelopment := ctx.Value(CtxDevelopment).(bool)
	if isDevelopment {
		log.Println("NFC is disabled in development mode")
		return
	}
	// Listen for an RFID/NFC tag in a separate goroutine
	go n.driver.Run(ctx)

	for {
		select {
		case tagId := <-n.driver.Read():
			if tagId != "" {
				log.Printf("Tag: %v", tagId)
				n.in <- tagId
			}
		case <-ctx.Done():
			err := n.driver.Stop()
			if err != nil {
				log.Fatal("Error cleaning up the reader: ", err.Error())
			}
			log.Printf("NFC context done")
			return
		default:
			// log.Printf("%s: Waiting for a tag \n", time.Now().String())
			time.Sleep(time.Millisecond * 300)
		}
	}
}
