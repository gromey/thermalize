package thermalize

import (
	"image"
	"io"
)

// NewStar returns the star set of printer commands for the given configuration.
//
// This function creates a new star sequence command set for printing images and text.
//
// Parameters:
//   - cpl: characters per line.
//   - ppl: pixels per line.
//   - w: the writer to which the commands will be sent.
//   - opts: a variadic list of options to customize the behavior of the command set.
//
// Example Usage:
//
// cmd := NewStar(48, 576, writer)
//
// In this example, a new star sequence command set is created with 48 characters per line,
// 576 pixels per line.
func NewStar(cpl, ppl int, w io.Writer, opts ...Options) Cmd {
	cmd := &star{Cmd: NewSkipper(cpl, ppl, w), hriPosition: 1, barcodeWidth: 1, barcodeHeight: 100}
	for _, opt := range opts {
		opt.apply(cmd)
	}
	return cmd
}

type star struct {
	Cmd
	hriPosition, barcodeWidth, barcodeHeight byte
}

func (c *star) Init() {
	c.Write(ESC, '@')
}

func (c *star) LeftMargin(n int) {
	if n < c.CPL() {
		c.Write(ESC, 'l', byte(n))
	}
}

func (c *star) WidthArea(n int) {
	if n < c.CPL() {
		c.Write(ESC, 'Q', byte(n))
	}
}

func (c *star) AbsolutePosition(n int) {
	if n < c.PPL() {
		c.Write(ESC, GS, 'A', byte(n), byte(n>>8))
	}
}

func (c *star) Align(b byte) {
	c.Write(ESC, GS, 'a', minByte(b, 2))
}

func (c *star) UpsideDown(b bool) {
	if b {
		c.Write(SI)
		return
	}
	c.Write(DC2)
}

// TabPositions maximum of 16 horizontal tabs can be set.
func (c *star) TabPositions(bs ...byte) {
	l := len(bs)
	if l == 0 {
		return
	} else if l > 16 {
		bs = bs[:16]
	}

	buf := make([]byte, 0, 3+l)
	buf = append(buf, ESC, 'D')

	var previous byte
	for _, n := range bs {
		if n <= previous {
			continue
		}
		buf = append(buf, n)
		previous = n
	}

	buf = append(buf, NUL)
	c.Write(buf...)
}

func (c *star) Tab() {
	c.Write(HT)
}

func (c *star) CodePage(b byte) {
	c.Write(ESC, GS, 't', b)
}

// CharSize
//
//	h: character height (0 - x1 `normal`, 1 - x2, 2 - x3, 3 - x4, 5 - x5, 5 - x6)
//	w: character width (0 - x1 `normal`, 1 - x2, 2 - x3, 3 - x4, 4 - x5, 5 - x6)
func (c *star) CharSize(h, w byte) {
	c.Write(ESC, 'i', minByte(w, 5), minByte(h, 5))
}

func (c *star) Bold(b bool) {
	if b {
		c.Write(ESC, 'E')
		return
	}
	c.Write(ESC, 'F')
}

func (c *star) Underling(b byte) {
	c.Write(ESC, '-', minByte(b, 1))
}

// BarcodeWidth 1 <= b <= 9.
func (c *star) BarcodeWidth(b byte) {
	b = maxByte(b, 1)
	c.barcodeWidth = minByte(b, 9)
}

func (c *star) BarcodeHeight(b byte) {
	c.barcodeHeight = maxByte(b, 1)
}

func (c *star) HRIPosition(b byte) {
	c.hriPosition = 1
	if b > 1 {
		c.hriPosition = 2
	}
}

func (c *star) Barcode(m byte, s string) {
	if len(s) == 0 {
		return
	}

	c.Write(ESC, 'b', c.barcodeType(m), c.hriPosition, c.barcodeWidth, c.barcodeHeight)
	c.Text(s, nil)
	c.Write(RS)
}

// QRCodeSize
//
//	1 <= b <= 8.
func (c *star) QRCodeSize(b byte) {
	b = maxByte(b, 1)
	c.Write(ESC, GS, 'y', 'S', '2', minByte(b, 8))
}

func (c *star) QRCodeCorrectionLevel(b byte) {
	c.Write(ESC, GS, 'y', 'S', '1', minByte(b, 3))
}

func (c *star) QRCode(s string) {
	l := len(s)
	if l == 0 {
		return
	}

	h, w := byte(l), byte(l>>8)

	// Store the data in the symbol storage area.
	c.Write(ESC, GS, 'y', 'D', '1', 0, h, w)
	c.Text(s, nil)

	// Print the symbol data in the symbol storage area.
	c.Write(ESC, GS, 'y', 'P')
}

func (c *star) Image(img image.Image, invert bool) {
	w, bs := ImageToBin(img, invert)

	xl, xh := byte(w), byte(w>>8)

	l := len(bs)
	block := w * 3
	start := 0

	for end := block; start < l; end += block {
		end = minByte(end, l)

		c.Write(ESC, 'X', xl, xh)
		c.Write(bs[start:end]...)
		c.Write(ESC, 'J', 12)

		start = end
	}
}

func (c *star) Feed(b byte) {
	if b > 0 {
		c.Write(ESC, 'J', b)
	}
}

func (c *star) LineFeed() {
	c.Write(LF)
}

// Cut
//
//	m = 0, full cut at the current position;
//	m = 1, partial cut at the current position;
//	m = 2, paper is fed to cutting position, then a full cut;
//	m = 3, paper is fed to cutting position, then a partial cut;
func (c *star) Cut(m, _ byte) {
	c.Write(ESC, 'd', minByte(m, 3))
}

func (c *star) FullCut() {
	c.Cut(2, 0)
}

// OpenCashDrawer
//
//	1 <= t1 <= 255 - specifies the pulse on time (20 ms x t1).
//	1 <= t2 <= 255 - specifies the pulse off time (20 ms x t2).
func (c *star) OpenCashDrawer(m, t1, t2 byte) {
	if t1 == 0 || t2 == 0 {
		return
	}
	c.Write(ESC, GS, BEL, minByte(m, 1)+1, t1, t2)
}

func (c *star) barcodeType(m byte) byte {
	if m > 13 {
		m = 4
	}
	return [14]byte{49, 48, 50, 51, 52, 55, 54, 53, 56, 57, 65, 66, 67, 68}[m]
}
