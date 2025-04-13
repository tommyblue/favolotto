package nfc_serial

import (
	"context"
	"log"
	"strings"

	"github.com/pkg/errors"
	"go.bug.st/serial"
)

type NFCSerial struct {
	tagChannel chan string
	port       serial.Port
}

func New() (*NFCSerial, error) {

	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open("/tmp/ttyTMP0", mode)
	if err != nil {
		log.Fatal(err)
		return nil, errors.Wrap(err, "Cannot initialize the serial Port")
	}
	port.SetReadTimeout(serial.NoTimeout)
	port.ResetInputBuffer()
	port.ResetOutputBuffer()
	return &NFCSerial{tagChannel: make(chan string, 10), port: port}, nil
}

func (d *NFCSerial) Run(ctx context.Context) error {
	log.Println("NFC over Serial driver is running")

	buff := make([]byte, 120)
	for {
		n, err := d.port.Read(buff)
		if err != nil {
			errors.Wrap(err, "unable to read from Port")
			break
		}
		if n == 0 {
			errors.Wrap(err, "EOF")
			continue
		}
		s := string(buff[:n])
		log.Printf("read: %v", s)
		l := strings.Split(s, ";")
		if len(l) == 0 {
			log.Printf("Invalid scring %v", l)
			continue
		}
		d.tagChannel <- strings.TrimSpace(l[0])
	}
	return nil
}

func (d *NFCSerial) Stop() error {
	d.port.Close()
	return nil
}

func (d *NFCSerial) Read() <-chan string {
	return d.tagChannel
}
