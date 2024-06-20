package thermalize

import (
	"image"
	"io"
)

// NewEscape returns the most popular set of printer commands.
// If the obsolete is true, the obsolete print image command will be used.
func NewEscape(cpl, ppl int, w io.Writer, obsolete bool) Cmd {
	return &escape{Cmd: NewSkipper(cpl, ppl, w), obsolete: obsolete}
}

type escape struct {
	Cmd
	obsolete bool
}

func (c *escape) Init() {
	c.Write(ESC, '@')
}

func (c *escape) LeftMargin(n int) {
	if n > c.PPL() {
		return
	}
	c.Write(GS, 'L', byte(n), byte(n>>8))
}

func (c *escape) WidthArea(n int) {
	if n > c.PPL() {
		return
	}
	c.Write(GS, 'W', byte(n), byte(n>>8))
}

func (c *escape) AbsolutePosition(n int) {
	if n > c.PPL() {
		return
	}
	c.Write(ESC, '$', byte(n), byte(n>>8))
}

func (c *escape) Align(b byte) {
	if b > 2 {
		b = 2
	}
	c.Write(ESC, 'a', b)
}

func (c *escape) UpsideDown(b bool) {
	if b {
		c.Write(ESC, '{', 1)
		return
	}
	c.Write(ESC, '{', 0)
}

// TabPositions maximum of 32 horizontal tabs can be set.
func (c *escape) TabPositions(bs ...byte) {
	l := len(bs)
	if l == 0 {
		return
	} else if l > 32 {
		bs = bs[:32]
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

func (c *escape) Tab() {
	c.Write(HT)
}

func (c *escape) CodePage(b byte) {
	c.Write(ESC, 't', b)
}

// CharSize
//
//	w: character width (0 - x1 `normal`, 1 - x2, 2 - x3, 3 - x4, 4 - x5, 5 - x6, 6 - x7, 7 - x8)
//	h: character height (0 - x1 `normal`, 1 - x2, 2 - x3, 3 - x4, 4 - x5, 5 - x6, 6 - x7, 7 - x8)
func (c *escape) CharSize(w, h byte) {
	if w > 7 {
		w = 7
	}
	if h > 7 {
		h = 7
	}
	c.Write(GS, '!', w<<4+h)
}

func (c *escape) Bold(b bool) {
	if b {
		c.Write(ESC, 'E', 1)
		return
	}
	c.Write(ESC, 'E', 0)
}

func (c *escape) ClockwiseRotation(b bool) {
	if b {
		c.Write(ESC, 'V', 1)
		return
	}
	c.Write(ESC, 'V', 0)
}

func (c *escape) Underling(b byte) {
	if b > 2 {
		b = 2
	}
	c.Write(ESC, '-', b)
}

// BarcodeWidth
//
//	1 <= b <= 6.
func (c *escape) BarcodeWidth(b byte) {
	if b == 0 {
		b = 1
	} else if b > 6 {
		b = 6
	}
	c.Write(GS, 'w', b)
}

func (c *escape) BarcodeHeight(b byte) {
	if b == 0 {
		b = 1
	}
	c.Write(GS, 'h', b)
}

func (c *escape) HRIFont(b byte) {
	if b > 1 {
		b = 1
	}
	c.Write(GS, 'f', b)
}

func (c *escape) HRIPosition(b byte) {
	if b > 3 {
		b = 3
	}
	c.Write(GS, 'H', b)
}

func (c *escape) Barcode(m byte, s string) {
	l := len(s)
	if l == 0 {
		return
	}

	c.Write(GS, 'k', c.barcodeType(m), byte(l))
	c.Text(s, nil)
}

// QRCodeSize (cn = 49, fn = 67).
//
//	1 <= b <= 16.
func (c *escape) QRCodeSize(b byte) {
	if b < 1 {
		b = 1
	} else if b > 16 {
		b = 16
	}
	c.Write(GS, '(', 'k', 3, 0, 49, 67, b)
}

// QRCodeCorrectionLevel (cn = 49, fn = 69).
func (c *escape) QRCodeCorrectionLevel(b byte) {
	b += 48
	c.Write(GS, '(', 'k', 3, 0, 49, 69, b)
}

func (c *escape) QRCode(s string) {
	l := len(s)
	if l == 0 {
		return
	}

	l += 3
	h, w := byte(l), byte(l>>8)

	// Store the data in the symbol storage area (cn = 49, fn = 80).
	c.Write(GS, '(', 'k', h, w, 49, 80, 48)
	c.Text(s, nil)

	// Print the symbol data in the symbol storage area (cn = 49, fn = 81).
	c.Write(GS, '(', 'k', 3, 0, 49, 81, 48)
}

func (c *escape) Image(img image.Image, invert bool) {
	w, bs := ImageToBit(img, invert)

	l := len(bs)
	if l == 0 {
		return
	}

	if c.obsolete {
		c.imageObsolete(w, l, bs)
		return
	}

	c.image(w, l, bs)
}

func (c *escape) imageObsolete(w, l int, bs []byte) {
	h := l / w

	c.Write(GS, 'v', 0, 0, byte(w), byte(w>>8), byte(h), byte(h>>8))
	c.Write(bs...)
}

func (c *escape) image(w, l int, bs []byte) {
	p := 10 + l
	p1, p2, p3, p4 := byte(p), byte(p>>8), byte(p>>16), byte(p>>24)

	var bx, by byte = 1, 1

	x := w * 8
	xl, xh := byte(x), byte(x>>8)

	y := l / w
	yl, yh := byte(y), byte(y>>8)

	// Store the graphics data in the print buffer (fn = 112).
	c.Write(GS, '8', 'L', p1, p2, p3, p4, 48, 112, 48, bx, by, 49, xl, xh, yl, yh)
	c.Write(bs...)

	// Print the graphics data in the print buffer (fn = 2, 50).
	c.Write(GS, '(', 'L', 2, 0, 48, 2)
}

func (c *escape) Feed(b byte) {
	if b == 0 {
		return
	}
	c.Write(ESC, 'J', b)
}

func (c *escape) LineFeed() {
	c.Write(LF)
}

// Cut
//
//	m = 0  | 1  - cuts paper;
//	m = 65 | 66 - feeds paper to  (cutting position + [p x (vertical motion unit)]) and cuts the paper;
func (c *escape) Cut(m, p byte) {
	switch m {
	case 0, 1:
		c.Write(GS, 'V', m)
	case 65, 66:
		c.Write(GS, 'V', m, p)
	}
}

func (c *escape) FullCut() {
	c.Cut(65, 10)
}

// OpenCashDrawer
//
//	1 <= t1 <= 255 - specifies the pulse on time (2 ms x t1).
//	1 <= t2 <= 255 - specifies the pulse off time (2 ms x t2).
//	t1 must be less than t2.
func (c *escape) OpenCashDrawer(m byte, t1, t2 byte) {
	switch {
	case t1 == 0 || t2 == 0:
		return
	case t1 > t2:
		t1, t2 = t2, t1
	}
	if m > 1 {
		m = 1
	}
	c.Write(ESC, 'p', m, t1, t2)
}

func (c *escape) barcodeType(m byte) byte {
	if m < 13 {
		m = 4
	}
	return [14]byte{65, 66, 68, 67, 69, 72, 73, 70, 71, 74, 75, 76, 77, 78}[m]
}
