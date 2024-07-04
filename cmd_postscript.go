package thermalize

import (
	"fmt"
	"image"
	"io"
	"strings"
)

const (
	lineFeed = 10.8

	styleRegular = "Regular"
	styleBold    = "Bold"
)

var (
	width      float64 = 204 // 3 inch = 205 / 2 inch = 145
	height     float64 = 400 // 1999.54mm = 5670
	qrCodeFunc func(string) (image.Image, error)
)

func SetPostscriptPageSize(w, h float64) {
	if w > 0 {
		width = w
	}
	if h > 0 {
		height = h
	}
}

func SetPostscriptQRCodeFunc(f func(string) (image.Image, error)) {
	qrCodeFunc = f
}

// NewPostscript returns the postscript set of printer commands. // 5660
// ImageFuncVersion is not used.
func NewPostscript(cpl, ppl int, w io.Writer, _ ImageFuncVersion) Cmd {
	return &postscript{
		Cmd:    NewSkipper(cpl, ppl, w),
		width:  width,
		height: height,
		y:      height,
		row:    row{pieces: make([]piece, 0)},
		font:   defaultFont,
		sizeX:  1,
		sizeY:  1,
	}
}

type postscript struct {
	Cmd

	width  float64
	height float64
	x, y   float64

	row   row
	font  font
	bold  bool
	sizeX byte
	sizeY byte

	align     byte
	underling byte

	openDrawer bool
}

func (c *postscript) Write(bs ...byte) {
	c.Text(string(bs), nil)
}

func (c *postscript) Text(s string, enc func(string) []byte) {
	if len(s) == 0 {
		return
	}

	if enc == nil {
		enc = encoder
	}

	c.row.align = c.align

	charSizeX := 4.25 * float64(c.sizeX)

	parts := c.splitString(s, c.row.width, charSizeX)
	for i, p := range parts {
		if i > 0 {
			c.LineFeed()
		}

		c.row.setHeight(c.sizeY)

		rowPiece := piece{
			data:      enc(p),
			w:         float64(len([]rune(p))) * charSizeX,
			sizeX:     c.sizeX,
			sizeY:     c.sizeY,
			underling: c.underling,
			bold:      c.bold,
		}

		c.row.width += rowPiece.w
		c.row.pieces = append(c.row.pieces, rowPiece)
	}
}

func (c *postscript) Init() {
	c.align = Left
	c.underling = NoUnderling
	c.font = defaultFont
	c.setPage()
}

func (c *postscript) Align(b byte) {
	c.align = minByte(b, 2)
}

func (c *postscript) CharSize(w, h byte) {
	c.sizeX = minByte(w, 5) + 1
	c.sizeY = minByte(h, 5) + 1
}

func (c *postscript) Bold(b bool) {
	c.bold = b
}

func (c *postscript) Underling(b byte) {
	c.underling = minByte(b, 2)
}

func (c *postscript) QRCode(s string) {
	if qrCodeFunc == nil || len(s) == 0 {
		return
	}
	qr, err := qrCodeFunc(s)
	if err != nil {
		c.Text(err.Error(), nil)
		return
	}
	c.Image(qr, false)
	c.LineFeed()
}

func (c *postscript) Image(img image.Image, invert bool) {
	w, bs := ImageToBytes(img, invert)
	h := img.Bounds().Size().Y

	c.image(w, h, bs)
}

func (c *postscript) LineFeed() {
	c.y -= c.row.height

	if c.y < lineFeed {
		c.newPage()
		c.y -= c.row.height
	}

	offset := c.getOffset(c.row.width)

	for _, p := range c.row.pieces {
		c.font.setStyle(p.bold, p.sizeX, p.sizeY)
		c.setFont()

		c.moveTo(offset, c.y)

		c.Cmd.Write([]byte(fmt.Sprintf("(%s) show\n", p.data))...)

		c.setLine(p.underling, offset, p.w)

		offset += p.w
	}

	c.row.reset()
	c.x = 0
}

func (c *postscript) Print() {
	c.LineFeed()
	c.showPage()
}

func (c *postscript) setPage() {
	s := fmt.Sprintf("%%!PS\n<< /PageSize [%.2f %.2f] >> setpagedevice\n", c.width, c.height)
	c.Cmd.Write([]byte(s)...)
}

func (c *postscript) setFont() {
	if !c.font.changed {
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("/NotoSansMono-%s findfont 9 scalefont\n", c.font.style))
	sb.WriteString(fmt.Sprintf("dup [%.2f 0 0 %d 0 0] makefont setfont\n", float64(c.font.sizeX)*0.79, c.font.sizeY))
	c.Cmd.Write([]byte(sb.String())...)

	c.font.changed = false
}

func (c *postscript) moveTo(x, y float64) {
	c.Cmd.Write([]byte(fmt.Sprintf("%.2f %.2f moveto\n", x, y))...)
}

func (c *postscript) setLine(underling byte, offset, width float64) {
	if underling == 0 {
		return
	}
	weight := 0.25
	if underling == 2 {
		weight = 0.75
	}
	y := c.y - 2

	c.Cmd.Write([]byte(fmt.Sprintf("%.1f setlinewidth\n%.2f %.2f moveto\n%.2f %.2f lineto\nstroke\n", weight, offset, y, offset+width, y))...)
}

func (c *postscript) image(width, height int, bs []byte) {
	w := float64(width) / (float64(c.PPL()) / c.width)
	h := w / (float64(width) / float64(height))

	if h > c.height {
		c.Text("the height of the image is greater than the height of the page", nil)
		return
	}

	c.y -= h
	if c.y < lineFeed {
		c.newPage()
		c.y -= h
	}

	c.y -= 4

	var sb strings.Builder
	sb.WriteString("gsave\n")
	sb.WriteString(fmt.Sprintf("/picstr %d string def\n%.2f %.2f translate\n", width, c.getOffset(w), c.y))
	sb.WriteString(fmt.Sprintf("%.2f %.2f scale\n%d %d 8\n", w, h, width, height))
	sb.WriteString(fmt.Sprintf("[%d 0 0 %d neg 0 %d]\n{ currentfile picstr readhexstring pop }\nimage\n", width, height, height))
	sb.WriteString(fmt.Sprintf("%X\n", bs))
	sb.WriteString("grestore\n")
	c.Cmd.Write([]byte(sb.String())...)
}

func (c *postscript) showPage() {
	c.y = c.height
	c.font.changed = true
	c.Cmd.Write([]byte("showpage\n")...)
}

func (c *postscript) newPage() {
	c.showPage()
	c.setPage()
}

func (c *postscript) getOffset(w float64) float64 {
	switch c.align {
	case Center:
		return (c.width - c.x - w) / 2
	case Right:
		return c.width - c.x - w
	default:
		return 0
	}
}

func (c *postscript) splitString(s string, offset, width float64) []string {
	n := int(c.width / width)

	start, end := 0, n

	if offset > 0 {
		end = int((c.width - offset) / width)
	}

	var chunks []string

	if end > len(s) {
		return append(chunks, s)
	}

	for end < len(s) {
		chunks = append(chunks, s[start:end])

		start = end
		end += n

		if end > len(s) {
			return append(chunks, s[start:])
		}
	}

	return chunks
}

type piece struct {
	data      []byte
	x, w      float64
	sizeX     byte
	sizeY     byte
	underling byte
	bold      bool
}

type row struct {
	pieces []piece
	height float64
	width  float64
	align  byte
}

func (r *row) setHeight(h byte) {
	var q float64 = 1
	if h > 1 {
		q = 0.79 * float64(h)
	}
	if y := lineFeed * q; y > r.height {
		r.height = y
	}
}

func (r *row) reset() {
	r.pieces = r.pieces[:0]
	r.height, r.width = lineFeed, 0
}

var defaultFont = font{
	style:   styleRegular,
	sizeX:   1,
	sizeY:   1,
	changed: true,
}

type font struct {
	style   string
	sizeX   byte
	sizeY   byte
	changed bool
}

func (f *font) setStyle(b bool, w, h byte) {
	style := styleRegular
	if b {
		style = styleBold
	}
	f.setChanged(swap(&f.style, style))
	f.setChanged(swap(&f.sizeX, w))
	f.setChanged(swap(&f.sizeY, h))
}

func (f *font) setChanged(b bool) {
	f.changed = f.changed || b
}

func swap[T comparable](dst *T, src T) (b bool) {
	if dst == nil {
		return
	}
	b = *dst != src
	*dst = src
	return
}

func encoder(s string) []byte {
	return []byte(s)
}
