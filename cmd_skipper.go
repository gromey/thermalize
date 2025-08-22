package thermalize

import (
	"image"
	"io"
)

// newSkipper returns a set of methods that skip the execution of unimplemented commands.
// This writes raw bytes and text to a writer.
func newSkipper(cpl, ppl int, w io.Writer) *skipper {
	return &skipper{cpl: cpl, ppl: ppl, w: w, barcodeWidth: 1, barcodeHeight: 100, qrcodeCorrectionLevel: Q, qrcodeSize: 5}
}

type skipper struct {
	cpl int
	ppl int
	w   io.Writer

	barcodeFunc   barcodeFunc
	barcodeWidth  byte
	barcodeHeight byte
	hriPosition   byte

	qrcodeFunc            qrcodeFunc
	qrcodeCorrectionLevel byte
	qrcodeSize            byte

	initFunc func(Cmd)
}

func (c *skipper) Sizing(cpl, ppl int) {
	if cpl != 0 {
		c.cpl = cpl
	}
	if ppl != 0 {
		c.ppl = ppl
	}
}

func (c *skipper) CPL() int {
	return c.cpl
}

func (c *skipper) PPL() int {
	return c.ppl
}

func (c *skipper) Write(bs ...byte) {
	if c.w == nil {
		panic("writer not specified")
	}
	if _, err := c.w.Write(bs); err != nil {
		panic(err.Error())
	}
}

func (c *skipper) Text(str string, enc func(string) []byte) {
	if enc != nil {
		c.Write(enc(str)...)
		return
	}
	c.Write([]byte(str)...)
}

func (c *skipper) Init() {}

func (c *skipper) LeftMargin(int) {}

func (c *skipper) WidthArea(int) {}

func (c *skipper) AbsolutePosition(int) {}

func (c *skipper) Align(byte) {}

func (c *skipper) UpsideDown(bool) {}

func (c *skipper) TabPositions(...byte) {}

func (c *skipper) Tab() {}

func (c *skipper) CodePage(byte) {}

func (c *skipper) CharSize(byte, byte) {}

func (c *skipper) Bold(bool) {}

func (c *skipper) ClockwiseRotation(bool) {}

func (c *skipper) Underling(byte) {}

func (c *skipper) BarcodeWidth(b byte) {
	c.barcodeWidth = minByte(maxByte(b, 1), 6)
}

func (c *skipper) BarcodeHeight(b byte) {
	c.barcodeHeight = maxByte(b, 1)
}

func (c *skipper) HRIFont(byte) {}

func (c *skipper) HRIPosition(b byte) {
	c.hriPosition = minByte(b, 3)
}

func (c *skipper) Barcode(byte, string) {}

func (c *skipper) QRCodeSize(b byte) {
	c.qrcodeSize = minByte(maxByte(b, 1), 8)
}

func (c *skipper) QRCodeCorrectionLevel(b byte) {
	c.qrcodeCorrectionLevel = b
}

func (c *skipper) QRCode(string) {}

func (c *skipper) Image(image.Image, bool) {}

func (c *skipper) Feed(byte) {}

func (c *skipper) LineFeed() {}

func (c *skipper) Cut(byte, byte) {}

func (c *skipper) FullCut() {}

func (c *skipper) OpenCashDrawer(byte, byte, byte) {}

func (c *skipper) Print() {}

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func minByte[T number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func maxByte[T number](a, b T) T {
	if a > b {
		return a
	}
	return b
}
