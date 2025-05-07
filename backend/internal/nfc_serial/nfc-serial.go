package nfc_serial

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/warthog618/go-gpiocdev"
	"go.bug.st/serial"
)

type NFCSerial struct {
	tagChannel chan string
	port       serial.Port
	reset      *gpiocdev.Line
	isp        *gpiocdev.Line
}

func resetPulse(d *gpiocdev.Line) error {
	err := d.SetValue(1)
	if err != nil {
		log.Println(err)
		return err
	}
	time.Sleep(time.Millisecond * 400)
	err = d.SetValue(0)
	if err != nil {
		log.Println(err)
		return err
	}
	time.Sleep(time.Millisecond * 400)
	err = d.SetValue(1)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func New() (*NFCSerial, error) {

	log.Println("Resetting the reader..")
	c, err := gpiocdev.NewChip("gpiochip0")
	if err != nil {
		log.Printf("Unable to enable port %v", err)
		return nil, errors.Wrap(err, "Fail to init gpio port")
	}

	resetPin, err := c.RequestLine(4, gpiocdev.AsOutput(1))
	if err != nil {
		return nil, errors.Wrap(err, "Fail to init reset")
	}

	// The isp is connected to pin 27
	ispPin, err := c.RequestLine(27, gpiocdev.AsOutput(1))
	if err != nil {
		return nil, errors.Wrap(err, "Fail to init isp")
	}

	mode := &serial.Mode{
		BaudRate: 115200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open("/dev/ttyS0", mode)
	if err != nil {
		log.Printf("Wrong port to open %v", err)
		return nil, errors.Wrap(err, "Cannot initialize the serial Port")
	}
	defer port.Close()

	port.SetReadTimeout(serial.NoTimeout)
	port.ResetInputBuffer()
	port.ResetOutputBuffer()

	// The reset is connected to pin 4
	resetPulse(resetPin)
	return &NFCSerial{tagChannel: make(chan string, 10), port: port, reset: resetPin, isp: ispPin}, nil
}

func (d *NFCSerial) Run(ctx context.Context) error {
	log.Println("NFC over Serial driver is running")

	buff := make([]byte, 120)
	tag := string("")
	for {
		startFound := false
		endFound := false
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

			s := strings.TrimSpace(string(buff[:n]))
			idx_start := strings.Index(s, "$")
			if idx_start != -1 {
				startFound = true
			}

			log.Println(s)
			if !startFound {
				continue
			}

			tag = tag + s[idx_start+1:]

			idx_end := strings.Index(tag, "#")
			if idx_end != -1 {
				endFound = true
			}

			log.Println(tag)
			if endFound {
				tag = tag[:idx_end-1]
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
