package thermalize

import (
	"image"
	"io"
)

// NewSkipper returns a set of methods that skip the execution of unimplemented commands.
// This writes raw bytes and text to a writer.
func NewSkipper(cpl, ppl int, w io.Writer) Cmd {
	return &skipper{cpl: cpl, ppl: ppl, w: w}
}

type skipper struct {
	cpl int
	ppl int
	w   io.Writer
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

func (c *skipper) TabPositions(...byte) {}

func (c *skipper) Tab() {}

func (c *skipper) CodePage(byte) {}

func (c *skipper) CharSize(byte, byte) {}

func (c *skipper) Bold(bool) {}

func (c *skipper) Underling(byte) {}

func (c *skipper) BarcodeWidth(byte) {}

func (c *skipper) BarcodeHeight(byte) {}

func (c *skipper) HRIFont(byte) {}

func (c *skipper) HRIPosition(byte) {}

func (c *skipper) Barcode(byte, string) {}

func (c *skipper) QRCodeSize(byte) {}

func (c *skipper) QRCodeCorrectionLevel(byte) {}

func (c *skipper) QRCode(string) {}

func (c *skipper) Image(image.Image, bool) {}

func (c *skipper) Feed(byte) {}

func (c *skipper) LineFeed() {}

func (c *skipper) Cut(byte, byte) {}

func (c *skipper) FullCut() {}

func (c *skipper) OpenCashDrawer(byte, byte, byte) {}
