package thermalize

import (
	"fmt"
	"image"
	"io"
	"strings"
)

const (
	charWidth = 4.25
	lineFeed  = 10.8

	styleRegular = "Regular"
	styleBold    = "Bold"

	header = `%%!PS
<< /PageSize [%.2f %.2f] >> setpagedevice
/showEuro {/Euro glyphshow} def
`
)

// NewPostscript returns the postscript set of printer commands configured with the specified parameters.
//
// This function creates a new postscript command set for printing with customizable options.
//
// Parameters:
//   - cpl: characters per line.
//   - ppl: pixels per line.
//   - w: the writer to which the commands will be sent.
//   - opts: a variadic list of options to customize the behavior of the command set.
//
// Options:
// You can customize various aspects of the postscript command set using the following options:
//   - WithBarcodeFunc(func(string, BarcodeOptions) image.Image): sets a function for generating barcodes.
//   - WithQRCodeFunc(func(string, QRCodeOptions) image.Image): sets a function for generating QR codes.
//   - WithPageHeight(h): sets the page height to the specified value.
//
// Example Usage:
//
// cmd := NewPostscript(48, 576, writer, WithPageHeight(5670), WithBarcodeFunc(barcodeFunc), WithQRCodeFunc(qrcodeFunc))
//
// In this example, a new postscript command set is created with 48 characters per line,
// 576 pixels per line. The page height is set to 5670 units,
// and functions for generating barcodes and QR codes are provided.
//
// Default Initialization:
// If no options are specified, the postscript command set initializes with height: 400 units.
//
// Note:
// If functions for generating barcodes and QR codes not provided, the call to print them will be skipped.
func NewPostscript(cpl, ppl int, w io.Writer, opts ...Options) Cmd {
	cmd := &postscript{
		skipper:      newSkipper(cpl, ppl, w),
		tabPositions: []float64{34, 68, 102, 136, 170, 204, 238, 272, 306, 340, 374, 408, 442, 476, 510, 544, 578, 612, 646, 680, 714, 748, 782, 816, 850, 884, 918, 952, 986, 1020, 1054},
		width:        float64(cpl)*charWidth + 1,
		height:       400,
		y:            400,
		row:          row{pieces: make([]piece, 0)},
		font:         defaultFont,
		sizeX:        1,
		sizeY:        1,
	}
	for _, opt := range opts {
		opt.apply(cmd)
		opt.apply(cmd.skipper)
	}
	return cmd
}

type postscript struct {
	*skipper

	tabPositions []float64

	width  float64
	height float64
	x, y   float64
	tab    float64

	row   row
	font  font
	bold  bool
	sizeX byte
	sizeY byte

	align     byte
	underling byte

	openDrawer bool
}

func (c *postscript) Sizing(cpl, ppl int) {
	c.skipper.Sizing(cpl, ppl)
	if cpl != 0 {
		c.width = float64(cpl) * charWidth
	}
}

func (c *postscript) Text(s string, enc func(string) []byte) {
	if len(s) == 0 {
		return
	}

	if enc == nil {
		enc = encoder
	}

	c.row.align = c.align

	charSizeX := float64(c.sizeX) * charWidth

	parts := c.splitString(string(enc(s)), c.tab+c.row.width, charSizeX)
	for i, p := range parts {
		if i > 0 {
			c.LineFeed()
		}

		c.row.setHeight(c.sizeY)

		d := strings.ReplaceAll(p, `\`, `\\`)
		d = strings.ReplaceAll(d, "\x80", `) show showEuro (`)

		rowPiece := piece{
			data:      []byte(d),
			w:         float64(len(p)) * charSizeX,
			tab:       c.tab,
			sizeX:     c.sizeX,
			sizeY:     c.sizeY,
			underling: c.underling,
			bold:      c.bold,
		}

		c.row.width += c.tab + rowPiece.w
		c.row.pieces = append(c.row.pieces, rowPiece)
		c.tab = 0
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

func (c *postscript) TabPositions(bs ...byte) {
	l := len(bs)
	if l == 0 {
		return
	} else if l > 16 {
		bs = bs[:16]
	}

	var previous byte
	buf := make([]float64, 0)
	for _, n := range bs {
		if n <= previous {
			continue
		}
		if tab := float64(n) * charWidth; tab < c.width {
			buf = append(buf, tab)
		} else {
			tab = c.width
			buf = append(buf, tab)
			break
		}
		previous = n
	}

	c.tabPositions = buf
}

func (c *postscript) Tab() {
	for _, x := range c.tabPositions {
		if c.row.width < x {
			c.tab = x - c.row.width
			if c.tab > c.width {
				c.LineFeed()
				c.tab = 0
			}
			break
		}
	}
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

func (c *postscript) Barcode(m byte, s string) {
	if c.barcodeFunc == nil || len(s) == 0 {
		return
	}
	code := c.barcodeFunc(s, BarcodeOptions{Mode: m, Width: c.barcodeWidth, Height: c.barcodeHeight})
	c.Image(code, false)
}

func (c *postscript) QRCode(s string) {
	if c.qrcodeFunc == nil || len(s) == 0 {
		return
	}
	code := c.qrcodeFunc(s, QRCodeOptions{CorrectionLevel: c.qrcodeCorrectionLevel, Size: c.qrcodeSize})
	c.Image(code, false)
}

func (c *postscript) Image(img image.Image, invert bool) {
	if img == nil {
		return
	}

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

		offset += p.tab

		c.moveTo(offset, c.y)
		c.Write([]byte(fmt.Sprintf("(%s) show\n", p.data))...)
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
	s := fmt.Sprintf(header, c.width, c.height)
	c.Write([]byte(s)...)
}

func (c *postscript) setFont() {
	if !c.font.changed {
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("/NotoSansMono-%s findfont 9 scalefont\n", c.font.style))
	sb.WriteString(fmt.Sprintf("dup [%.2f 0 0 %d 0 0] makefont setfont\n", float64(c.font.sizeX)*0.79, c.font.sizeY))
	c.Write([]byte(sb.String())...)

	c.font.changed = false
}

func (c *postscript) moveTo(x, y float64) {
	c.Write([]byte(fmt.Sprintf("%.2f %.2f moveto\n", x, y))...)
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

	c.Write([]byte(fmt.Sprintf("%.1f setlinewidth\n%.2f %.2f moveto\n%.2f %.2f lineto\nstroke\n", weight, offset, y, offset+width, y))...)
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
	c.Write([]byte(sb.String())...)
}

func (c *postscript) showPage() {
	c.y = c.height
	c.font.changed = true
	c.Write([]byte("showpage\n")...)
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

	if end >= len(s) {
		return append(chunks, s)
	}

	for end < len(s) {
		chunks = append(chunks, s[start:end])

		start = end
		end += n

		if end >= len(s) {
			return append(chunks, s[start:])
		}
	}

	return chunks
}

type piece struct {
	data      []byte
	x, w      float64
	tab       float64
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
