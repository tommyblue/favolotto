package favolotto

import (
	"context"
	"log"
	"time"

	"github.com/tommyblue/favolotto/internal/pn532"
	"github.com/tommyblue/favolotto/internal/pn7150"
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

func NewNFC(driverName string, in chan<- string) *Nfc {
	var driver NfcDriver
	switch driverName {
	case "pn7150":
		driver = pn7150.New()
	case "pn532":
		driver = pn532.New()
	}

	return &Nfc{
		in:     in,
		driver: driver,
	}
}

func (n *Nfc) Run(ctx context.Context) {
	isDevelopment := ctx.Value("development").(bool)
	if isDevelopment {
		log.Println("NFC is disabled in development mode")
		return
	}
	// Listen for an RFID/NFC tag in a separate goroutine
	go n.driver.Run(ctx)

	for {
		select {
		case tagId := <-n.driver.Read():
			log.Printf("Read tag: %s \n", tagId)
			n.in <- tagId
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
