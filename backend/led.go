package favolotto

import (
	"context"
	"log"

	"github.com/tommyblue/favolotto/internal/colors"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/sysfs"
)

type LED struct {
	conn spi.Conn
	ch   <-chan colors.Color
}

func NewLED(inCh <-chan colors.Color) *LED {
	return &LED{
		conn: nil,
		ch:   inCh,
	}
}

func (l *LED) Run(ctx context.Context) {
	isDevelopment := ctx.Value(CtxDevelopment).(bool)
	if isDevelopment {
		log.Println("LED is disabled in development mode")
		return
	}

	if _, err := host.Init(); err != nil {
		log.Fatalf("led init error: %v", err)
	}

	// Open SPI0.0 (used by ReSpeaker HAT)
	spiPort, err := sysfs.NewSPI(0, 0)
	if err != nil {
		log.Fatalf("error opening SPI0: %v", err)
	}
	defer spiPort.Close()

	conn, err := spiPort.Connect(1*physic.MegaHertz, spi.Mode3, 8) // 1 MHz, Mode 3, 8-bit
	if err != nil {
		log.Fatalf("SPI connection error: %v", err)
	}

	l.conn = conn

	for {
		select {
		case <-ctx.Done():
			log.Println("LED context done")
			l.setColor(colors.Black)
			return
		case color := <-l.ch:
			l.setColor(color)
			// default:
			// 	l.updateLEDs(colors)
			// 	time.Sleep(1 * time.Second)
			// 	// update LED colors to loop through RGB
			// 	colors = append(colors[1:], colors[0])
		}
	}
}

func (l *LED) setColor(color colors.Color) {
	l.updateLEDs(color.ToLeds())
}

func (l *LED) updateLEDs(colors [][]byte) {
	startFrame := []byte{0x00, 0x00, 0x00, 0x00}
	endFrame := []byte{0xFF, 0xFF, 0xFF, 0xFF}

	var data []byte
	data = append(data, startFrame...)
	for _, color := range colors {
		data = append(data, color...)
	}
	data = append(data, endFrame...)

	if err := l.conn.Tx(data, nil); err != nil {
		log.Fatalf("error on SPI tx: %v", err)
	}
}
