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

type Button struct {
	in  chan<- string
	btn gpio.PinIO
}

func NewButton(in chan<- string) (*Button, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("error during initialization: %v", err)
	}

	// Define GPIO17 pin
	btn := rpi.P1_11

	// configure as input with internal pull-up
	if err := btn.In(gpio.PullUp, gpio.BothEdges); err != nil {
		return nil, fmt.Errorf("error configuring GPIO: %v", err)
	}

	return &Button{
		in:  in,
		btn: btn,
	}, nil
}

func (b *Button) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("button context done")
			return
		default:
			if b.btn.Read() == gpio.Low {
				log.Println("button pressed")
				b.in <- Pause
				time.Sleep(200 * time.Millisecond) // Debounce
			}
		}
	}

}
