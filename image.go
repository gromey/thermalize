package thermalize

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"sync/atomic"
)

func init() {
	grayLevel.Store(uint32(defaultGrayLevel))
}

const defaultGrayLevel uint8 = 127

var grayLevel atomic.Uint32

// SetGrayLevel sets the level of gray that should be visible when printing.
func SetGrayLevel(l uint8) {
	grayLevel.Store(uint32(l))
}

// ResetGrayLevel resets the level of gray that should be visible when printing to default value.
func ResetGrayLevel() {
	grayLevel.Store(uint32(defaultGrayLevel))
}

func gray(c color.Color, level uint8, invert bool) bool {
	if color.AlphaModel.Convert(c).(color.Alpha).A < level {
		return invert
	}
	if invert {
		return color.GrayModel.Convert(c).(color.Gray).Y > level
	}
	return color.GrayModel.Convert(c).(color.Gray).Y < level
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

	lvl := uint8(grayLevel.Load())

	for y := 0; y < sz.Y; y++ {
		n := y/8 + y/24*shift
		for x := 0; x < sz.X; x++ {
			if gray(img.At(x, y), lvl, invert) {
				data[n+x*3] |= 0x80 >> uint(y%8)
			}
		}
	}

	return sz.X, data
}

func ImageToBit(img image.Image, invert bool) (int, []byte) {
	sz := img.Bounds().Size()

	w := sz.X / 8
	if sz.X%8 != 0 {
		w += 1
	}

	data := make([]byte, w*sz.Y)

	lvl := uint8(grayLevel.Load())

	for y := 0; y < sz.Y; y++ {
		for x := 0; x < sz.X; x++ {
			if gray(img.At(x, y), lvl, invert) {
				data[y*w+x/8] |= 0x80 >> uint(x%8)
			}
		}
	}

	return w, data
}

func ImageToBytes(img image.Image, invert bool) (int, []byte) {
	sz := img.Bounds().Size()

	data := make([]byte, sz.X*sz.Y)

	lvl := uint8(grayLevel.Load())

	for y := 0; y < sz.Y; y++ {
		for x := 0; x < sz.X; x++ {
			if !gray(img.At(x, y), lvl, invert) {
				data[y*sz.X+x] = 255
			}
		}
	}

	return sz.X, data
}

// Logo returns the library logo.
func Logo() (img image.Image) {
	file, err := base64.StdEncoding.DecodeString(logo)
	if err != nil {
		panic(err)
	}
	if img, err = png.Decode(bytes.NewReader(file)); err != nil {
		panic(err)
	}
	return
}

const logo = `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
AASUVORK5CYII=`
