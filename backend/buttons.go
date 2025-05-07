package favolotto

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

var buttonsMap = map[int]string{
	22: Pause,  // "btnLeft":
	23: Resume, // "btnMid":
	24: Stop,   // "btnRight":
}

type Buttons struct {
	in      chan<- string
	pins    *gpiocdev.Lines
	offsets []int
}

func NewButton(in chan<- string) (*Buttons, error) {
	offsets := []int{}
	for x := range buttonsMap {
		offsets = append(offsets, x)
	}
	pins, err := gpiocdev.RequestLines("/dev/gpiochip0", offsets, gpiocdev.AsInput)
	if err != nil {
		fmt.Printf("Requesting lines returned error: %s\n", err)
	}
	//defer pins.Close()

	return &Buttons{
		in:      in,
		pins:    pins,
		offsets: offsets,
	}, nil
}

func (b *Buttons) Run(ctx context.Context) {
	isDevelopment := ctx.Value(CtxDevelopment).(bool)
	if isDevelopment {
		log.Println("Button is disabled in development mode")
		return
	}

	pressed := []bool{}
	for _ = range b.offsets {
		pressed = append(pressed, false)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("button context done")
			return
		default:
			time.Sleep(100 * time.Millisecond) // debounce time

			readValues := []int{0, 0, 0}
			err := b.pins.Values(readValues)
			if err != nil {
				continue
			}

			for idx := range b.offsets {
				if readValues[idx] == 0 {
					pressed[idx] = true
				} else if pressed[idx] {
					pressed[idx] = false
					b.in <- buttonsMap[b.offsets[idx]]
				}
			}
		}
	}

}
