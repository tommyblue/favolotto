package favolotto

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

type IoMgr struct {
	rstNfc   <-chan bool
	NfcReset *gpio.PinIO
	NfcIsp   *gpio.PinIO
	Btns     []*gpio.PinIO
}

func NewIoMap(conf Config, rstNfc <-chan bool) (*IoMgr, error) {
	if _, err := host.Init(); err != nil {
		return nil, errors.Wrap(err, "Unable to init gpio subsystem")
	}

	m := IoMgr{}

	if conf.NfcReset != "" {
		p := gpioreg.ByName(conf.NfcReset)
		if p == nil {
			return nil, fmt.Errorf("Unable to init gpio NfcReset: %v", p)
		}

		if err := p.Out(gpio.High); err != nil {
			return nil, errors.Wrap(err, "Unable to set direction gpio NfcReset")
		}

		m.NfcReset = &p
	}

	if conf.NfcIsp != "" {
		p := gpioreg.ByName(conf.NfcIsp)
		if p == nil {
			return nil, fmt.Errorf("Unable to init gpio NfcIsp: %v", p)
		}

		if err := p.Out(gpio.High); err != nil {
			return nil, errors.Wrap(err, "Unable to set direction gpio NfcIsp")
		}
		m.NfcIsp = &p
	}

	for _, pinName := range conf.BtnPins {
		p := gpioreg.ByName(pinName)
		if p == nil {
			return nil, fmt.Errorf("Unable to init gpio btn %s: %v", pinName, p)
		}

		if err := p.In(gpio.PullUp, gpio.BothEdges); err != nil {
			return nil, errors.Wrap(err, "Unable to set direction gpio btn")
		}

		m.Btns = append(m.Btns, &p)
	}

	return &m, nil
}

func (b *IoMgr) Run(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return
		case <-b.rstNfc:
			rst := *b.NfcReset
			if rst == nil {
				fmt.Println("No reset pin defined")
				return
			}
			if err := rst.Out(gpio.High); err != nil {
				fmt.Println("Unable to do reset")
				return
			}
			time.Sleep(time.Millisecond * 200)
			if err := rst.Out(gpio.Low); err != nil {
				fmt.Println("Unable to do reset")
				return
			}
			time.Sleep(time.Millisecond * 200)
			if err := rst.Out(gpio.High); err != nil {
				fmt.Println("Unable to do reset")
				return
			}
		}
	}
}
