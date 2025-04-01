package colors

type Color int

const (
	Red Color = iota
	Green
	Blue
	Brown
	Violet
	White
	Black
)

var Colors = map[Color][]byte{
	Red:    red,
	Green:  green,
	Blue:   blue,
	Brown:  brown,
	Violet: violet,
	White:  white,
	Black:  black,
}

// BGR version
var (
	red    = []byte{0xE0 | 10, 0, 0, 255}
	green  = []byte{0xE0 | 10, 0, 255, 0}
	blue   = []byte{0xE0 | 10, 255, 0, 0}
	brown  = []byte{0xE0 | 10, 0, 51, 102}
	violet = []byte{0xE0 | 10, 255, 51, 204}
	white  = []byte{0xE0 | 10, 255, 255, 255}
	black  = []byte{0xE0 | 31, 0, 0, 0}
)

func (c Color) Bytes() []byte {
	if color, ok := Colors[c]; ok {
		return color
	}
	return black
}

func (c Color) ToLeds() [][]byte {
	return [][]byte{c.Bytes(), c.Bytes(), c.Bytes()}
}

// Uncomment after changing the bytes colors to RGB
// func (c Color) APA102Data() []byte {
// 	return []byte{0xE0 | 10, c.Bytes()[2], c.Bytes()[1], c.Bytes()[3]}
// }

// func (c Color) APA102DataWithBrightness(brightness byte) []byte {
// 	return []byte{0xE0 | brightness, c.Bytes()[2], c.Bytes()[1], c.Bytes()[3]}
// }

// func (c Color) APA102DataWithBrightnessAndColor(brightness byte, color []byte) []byte {
// 	return []byte{0xE0 | brightness, color[2], color[1], color[3]}
// }
