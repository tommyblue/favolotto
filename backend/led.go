package favolotto

import (
	"context"
	"log"
	"time"

	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/sysfs"
)

type LED struct {
	conn spi.Conn
}

func NewLED() *LED {
	return &LED{
		conn: nil,
	}
}

func (l *LED) Run(ctx context.Context) {
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

	colors := [][]byte{
		l.makeAPA102Data(255, 0, 0), // LED 1 - Red
		l.makeAPA102Data(0, 255, 0), // LED 2 - Green
		l.makeAPA102Data(0, 0, 255), // LED 3 - Blue
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("LED context done")
			l.updateLEDs([][]byte{
				l.makeAPA102Data(0, 0, 0),
				l.makeAPA102Data(0, 0, 0),
				l.makeAPA102Data(0, 0, 0),
			})
			return
		default:
			l.updateLEDs(colors)
			time.Sleep(1 * time.Second)
			// update LED colors to loop through RGB
			colors = append(colors[1:], colors[0])
		}
	}
}

// Convert RGB to APA102
func (l *LED) makeAPA102Data(r, g, b byte) []byte {
	const brightness byte = 0xE0 | 31  // set brightness (0-31)
	return []byte{brightness, b, g, r} // Format: [Brightness, Blue, Green, Red]
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
