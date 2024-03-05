package thermalize

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
)

const defaultGrayLevel uint8 = 127

// GrayLevel the level of gray that should be visible when printing.
var GrayLevel = defaultGrayLevel

func Gray(c color.Color, invert bool) bool {
	if color.AlphaModel.Convert(c).(color.Alpha).A < GrayLevel {
		return invert
	}
	if invert {
		return color.GrayModel.Convert(c).(color.Gray).Y > GrayLevel
	}
	return color.GrayModel.Convert(c).(color.Gray).Y < GrayLevel
}

func ImageToBit(img image.Image, invert bool) (int, []byte) {
	sz := img.Bounds().Size()

	width := sz.X / 8
	if sz.X%8 != 0 {
		width += 1
	}

	data := make([]byte, width*sz.Y)

	for y := 0; y < sz.Y; y++ {
		for x := 0; x < sz.X; x++ {
			if Gray(img.At(x, y), invert) {
				data[y*width+x/8] |= 0x80 >> uint(x%8)
			}
		}
	}

	return width, data
}

func ImageToBin(img image.Image, invert bool) (int, []byte) {
	sz := img.Bounds().Size()

	rows := sz.Y / 24
	if sz.Y%24 != 0 {
		rows += 1
	}
	rows *= 3

	data := make([]byte, rows*sz.X)
	shift := 3 * (sz.X - 1)

	for y := 0; y < sz.Y; y++ {
		n := y/8 + y/24*shift
		for x := 0; x < sz.X; x++ {
			if Gray(img.At(x, y), invert) {
				data[n+x*3] |= 0x80 >> uint(y%8)
			}
		}
	}

	return sz.X, data
}

// Logo returns the library logo.
func Logo() (img image.Image) {
	file, err := os.ReadFile("logo.png")
	if err != nil {
		panic(err)
	}
	if img, err = png.Decode(bytes.NewReader(file)); err != nil {
		panic(err)
	}
	return
}
