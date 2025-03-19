package favolotto

import (
	"context"
	"log"

	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/host/v3"
	"periph.io/x/host/v3/sysfs"
)

var (
	// red
	red = []byte{0xE0 | 20, 255, 0, 0}
	// green
	green = []byte{0xE0 | 20, 0, 255, 0}
	// blue
	blue = []byte{0xE0 | 20, 0, 0, 255}
	//orange
	orange = []byte{0xE0 | 20, 255, 172, 28}
	// black
	black = []byte{0xE0 | 31, 0, 0, 0}
)

type LED struct {
	conn spi.Conn
	ch   <-chan string
}

func NewLED(inCh <-chan string) *LED {
	return &LED{
		conn: nil,
		ch:   inCh,
	}
}

func (l *LED) Run(ctx context.Context) {
	isDevelopment := ctx.Value("development").(bool)
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

	// colors := [][]byte{
	// 	l.makeAPA102Data(255, 0, 0), // LED 1 - Red
	// 	l.makeAPA102Data(0, 255, 0), // LED 2 - Green
	// 	l.makeAPA102Data(0, 0, 255), // LED 3 - Blue
	// }

	for {
		select {
		case <-ctx.Done():
			log.Println("LED context done")
			l.setColor("black")
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

func (l *LED) setColor(color string) {
	var colors [][]byte
	switch color {
	case "red":
		colors = [][]byte{red, red, red}
	case "green":
		colors = [][]byte{green, green, green}
	case "blue":
		colors = [][]byte{blue, blue, blue}
	case "orange":
		colors = [][]byte{orange, orange, orange}
	default:
		colors = [][]byte{black, black, black}
	}
	l.updateLEDs(colors)
}

// Convert RGB to APA102
// func (l *LED) makeAPA102Data(r, g, b byte) []byte {
// 	const brightness byte = 0xE0 | 31  // set brightness (0-31)
// 	return []byte{brightness, b, g, r} // Format: [Brightness, Blue, Green, Red]
// }

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
