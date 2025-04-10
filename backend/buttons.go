package favolotto

import (
	"context"
	"fmt"
	"log"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/rpi"
)

var buttonsPin = map[string]gpio.PinIO{
	"btn1": rpi.P1_11,
	"btn2": rpi.P1_32,
	"btn3": rpi.P1_33,
}

var messages = map[string]string{
	"btn1": Volume,
	"btn2": Pause,
	"btn3": Volume,
}

type Buttons struct {
	in  chan<- string
	btn map[string]*gpio.PinIO
}

func NewButton(in chan<- string) (*Buttons, error) {

	return &Buttons{
		in:  in,
		btn: make(map[string]*gpio.PinIO),
	}, nil
}

func (b *Buttons) init() error {
	if _, err := host.Init(); err != nil {
		return fmt.Errorf("error during initialization: %v", err)
	}

	for label, pin := range buttonsPin {
		// configure as input with internal pull-up
		if err := pin.In(gpio.PullUp, gpio.BothEdges); err != nil {
			return fmt.Errorf("error configuring GPIO: %v", err)
		}

		b.btn[label] = &pin
	}

	return nil
}

func (b *Buttons) Run(ctx context.Context) {
	isDevelopment := ctx.Value(CtxDevelopment).(bool)
	if isDevelopment {
		log.Println("Button is disabled in development mode")
		return
	}

	if err := b.init(); err != nil {
		log.Fatalf("error initializing button: %v", err)
	}

	pressed := make(map[string]bool)
	for label := range b.btn {
		pressed[label] = false
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("button context done")
			return
		default:
			for label, pin := range b.btn {
				if (*pin).Read() == gpio.Low {
					pressed[label] = true
				} else if pressed[label] {
					pressed[label] = false
					b.in <- messages[label]
				}
			}
			time.Sleep(200 * time.Millisecond) // debounce time
		}
	}

}
