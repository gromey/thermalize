package thermalize

import (
	"image"
	"io"
)

// NewStar returns the star set of printer commands.
func NewStar(cpl, ppl int, w io.Writer, _ bool) Cmd {
	return &star{Cmd: NewSkipper(cpl, ppl, w), hriPosition: 1, barcodeWidth: 1, barcodeHeight: 100}
}

type star struct {
	Cmd
	hriPosition, barcodeWidth, barcodeHeight byte
}

func (c *star) Init() {
	c.Write(ESC, '@')
}

func (c *star) LeftMargin(n int) {
	if n > c.CPL() {
		return
	}
	c.Write(ESC, 'l', byte(n))
}

func (c *star) WidthArea(n int) {
	if n > c.CPL() {
		return
	}
	c.Write(ESC, 'Q', byte(n))
}

func (c *star) AbsolutePosition(n int) {
	if n > c.PPL() {
		return
	}
	c.Write(ESC, GS, 'A', byte(n), byte(n>>8))
}

func (c *star) Align(b byte) {
	if b > 2 {
		b = 2
	}
	c.Write(ESC, GS, 'a', b)
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
	if h > 5 {
		h = 5
	}
	if w > 5 {
		w = 5
	}
	c.Write(ESC, 'i', h, w)
}

func (c *star) Bold(b bool) {
	if b {
		c.Write(ESC, 'E')
		return
	}
	c.Write(ESC, 'F')
}

func (c *star) Underling(b byte) {
	if b > 1 {
		b = 1
	}
	c.Write(ESC, '-', b)
}

// BarcodeWidth 1 <= b <= 9.
func (c *star) BarcodeWidth(b byte) {
	if b == 0 {
		b = 1
	} else if b > 9 {
		b = 9
	}
	c.barcodeWidth = b
}

func (c *star) BarcodeHeight(b byte) {
	if b == 0 {
		b = 1
	}
	c.barcodeHeight = b
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
	if b < 1 {
		b = 1
	} else if b > 8 {
		b = 8
	}
	c.Write(ESC, GS, 'y', 'S', '2', b)
}

func (c *star) QRCodeCorrectionLevel(b byte) {
	if b > 3 {
		b = 3
	}
	c.Write(ESC, GS, 'y', 'S', '1', b)
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
		if end > l {
			end = l
		}

		c.Write(ESC, 'X', xl, xh)
		c.Write(bs[start:end]...)
		c.Write(ESC, 'J', 12)

		start = end
	}
}

func (c *star) Feed(b byte) {
	if b == 0 {
		return
	}
	c.Write(ESC, 'J', b)
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
	if m > 3 {
		m = 3
	}
	c.Write(ESC, 'd', m)
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
	if m > 1 {
		m = 1
	}
	m += 1
	c.Write(ESC, GS, BEL, m, t1, t2)
}

func (c *star) barcodeType(m byte) byte {
	if m < 13 {
		m = 4
	}
	return [14]byte{49, 48, 50, 51, 52, 55, 54, 53, 56, 57, 65, 66, 67, 68}[m]
}
