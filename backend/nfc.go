package favolotto

import (
	"context"
	"log"
	"time"

	"github.com/tommyblue/favolotto/tagreader"
)

type Nfc struct {
	in chan<- string
}

func NewNFC(in chan<- string) *Nfc {
	return &Nfc{
		in: in,
	}
}

func (n *Nfc) Run(ctx context.Context) {
	isDevelopment := ctx.Value("development").(bool)
	if isDevelopment {
		log.Println("NFC is disabled in development mode")
		return
	}

	// Create an abstraction of the Reader, DeviceConnection string is empty if you want the library to autodetect your reader
	rfidReader := tagreader.NewTagReader("", 19)
	tagChannel := rfidReader.GetTagChannel()

	// Listen for an RFID/NFC tag in a separate goroutine
	go rfidReader.ListenForTags(ctx)

	for {
		select {
		case tagId := <-tagChannel:
			log.Printf("Read tag: %s \n", tagId)
			n.in <- tagId
		case <-ctx.Done():
			err := rfidReader.Cleanup()
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
