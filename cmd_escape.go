package thermalize

import (
	"image"
	"io"
)

// NewEscape returns the most popular set of printer commands.
// ImageFuncVersion tells which command to use to print an image.
//
// The set of Escape commands supports the following versions:
//   - ImageFuncV0 [GS v ...] the obsolete print image command;
//   - ImageFuncV1 [GS 8 L ... GS ( L]
//   - ImageFuncV2 [ESC * ! ... ESC J]
func NewEscape(cpl, ppl int, w io.Writer, v ImageFuncVersion) Cmd {
	return &escape{Cmd: NewSkipper(cpl, ppl, w), version: v}
}

type escape struct {
	Cmd
	version ImageFuncVersion
}

func (c *escape) Init() {
	c.Write(ESC, '@')
}

func (c *escape) LeftMargin(n int) {
	if n < c.PPL() {
		c.Write(GS, 'L', byte(n), byte(n>>8))
	}
}

func (c *escape) WidthArea(n int) {
	if n < c.PPL() {
		c.Write(GS, 'W', byte(n), byte(n>>8))
	}
}

func (c *escape) AbsolutePosition(n int) {
	if n < c.PPL() {
		c.Write(ESC, '$', byte(n), byte(n>>8))
	}
}

func (c *escape) Align(b byte) {
	c.Write(ESC, 'a', minByte(b, 2))
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
	c.Write(GS, '!', minByte(w, 7)<<4+minByte(h, 7))
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
	c.Write(ESC, '-', minByte(b, 2))
}

// BarcodeWidth
//
//	1 <= b <= 6.
func (c *escape) BarcodeWidth(b byte) {
	b = maxByte(b, 1)
	c.Write(GS, 'w', minByte(b, 6))
}

func (c *escape) BarcodeHeight(b byte) {
	c.Write(GS, 'h', maxByte(b, 1))
}

func (c *escape) HRIFont(b byte) {
	c.Write(GS, 'f', minByte(b, 1))
}

func (c *escape) HRIPosition(b byte) {
	c.Write(GS, 'H', minByte(b, 3))
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
	b = maxByte(b, 1)
	c.Write(GS, '(', 'k', 3, 0, 49, 67, minByte(b, 16))
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
	switch c.version {
	case ImageFuncV1:
		c.imageV1(img, invert)
	case ImageFuncV2:
		c.imageV2(img, invert)
	//case 3:
	//	c.imageV3(img, invert)
	default:
		c.imageV0(img, invert)
	}
}

// Obsolete command to print an image
func (c *escape) imageV0(img image.Image, invert bool) {
	w, bs := ImageToBit(img, invert)

	l := len(bs)
	if l == 0 {
		return
	}

	h := l / w

	c.Write(GS, 'v', 0, 0, byte(w), byte(w>>8), byte(h), byte(h>>8))
	c.Write(bs...)
}

func (c *escape) imageV1(img image.Image, invert bool) {
	w, bs := ImageToBit(img, invert)

	l := len(bs)
	if l == 0 {
		return
	}

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

func (c *escape) imageV2(img image.Image, invert bool) {
	w, bs := ImageToBin(img, invert)

	xl, xh := byte(w), byte(w>>8)

	l := len(bs)
	block := w * 3
	start := 0

	for end := block; start < l; end += block {
		end = minByte(end, l)

		c.Write(ESC, '*', '!', xl, xh)
		c.Write(bs[start:end]...)
		c.Write(ESC, 'J', 24)

		start = end
	}
}

func (c *escape) Feed(b byte) {
	if b > 0 {
		c.Write(ESC, 'J', b)
	}
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
	c.Write(ESC, 'p', minByte(m, 1), t1, t2)
}

func (c *escape) barcodeType(m byte) byte {
	if m > 13 {
		m = 4
	}
	return [14]byte{65, 66, 68, 67, 69, 72, 73, 70, 71, 74, 75, 76, 77, 78}[m]
}
