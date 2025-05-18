package favolotto

import (
	"context"
	"log"
	"time"

	"periph.io/x/conn/v3/gpio"
)

type Buttons struct {
	in     chan<- string
	btn    []string
	btnPin []*gpio.PinIO
}

func NewButton(in chan<- string, btnFn []string, btns []*gpio.PinIO) (*Buttons, error) {
	return &Buttons{
		in:     in,
		btn:    btnFn,
		btnPin: btns,
	}, nil
}

func (b *Buttons) Run(ctx context.Context) {
	isDevelopment := ctx.Value(CtxDevelopment).(bool)
	if isDevelopment {
		log.Println("Button is disabled in development mode")
		return
	}

	pressed := make(map[string]bool)
	for _, btnFn := range b.btn {
		pressed[btnFn] = false
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			for n, pin := range b.btnPin {
				btnFn := b.btn[n]
				if (*pin).Read() == gpio.Low {
					pressed[btnFn] = true
				} else if pressed[btnFn] {
					pressed[btnFn] = false
					b.in <- btnFn
				}
			}
			time.Sleep(100 * time.Millisecond) // debounce time
		}
	}

}
