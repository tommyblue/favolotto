package nfc_serial

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/warthog618/gpiod"
	"go.bug.st/serial"
)

type NFCSerial struct {
	tagChannel chan string
	port       serial.Port
}

func reset(resetPin int) {
	log.Println("Resetting the reader..")
	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		log.Println(err)
		return
	}
	pin, err := c.RequestLine(resetPin, gpiod.AsOutput(1))
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
	if err != nil {
		log.Println(err)
		return
	}
}

func New() (*NFCSerial, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open("/dev/ttyS0", mode)
	if err != nil {
		log.Fatal("Wrong port to open %v", err)
		return nil, errors.Wrap(err, "Cannot initialize the serial Port")
	}
	port.SetReadTimeout(serial.NoTimeout)
	port.ResetInputBuffer()
	port.ResetOutputBuffer()

	// The reset is connected to pin 4
	reset(4)
	return &NFCSerial{tagChannel: make(chan string, 10), port: port}, nil
}

func (d *NFCSerial) Run(ctx context.Context) error {
	log.Println("NFC over Serial driver is running")

	buff := make([]byte, 120)
	tag := string("")
	for {
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
			s = strings.TrimSpace(s)
			tag = tag + s
			log.Printf("read: %v", s)
			if strings.Contains(tag, ";") {
				tag = strings.ReplaceAll(tag, ";", "")
				break
			}
		}
		d.tagChannel <- tag
		tag = string("")
	}
}

func (d *NFCSerial) Stop() error {
	d.port.Close()
	return nil
}

func (d *NFCSerial) Read() <-chan string {
	return d.tagChannel
}
