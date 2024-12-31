package thermalize

import "image"

const (
	Left = iota
	Center
	Right
)

const (
	NoUnderling = iota
	OneDotUnderling
	TwoDotsUnderling
)

const (
	HRIFontA = iota // (12 x 24)
	HRIFontB        // (9 x 17)
)

const (
	HRINotPrinted = iota
	HRIAbove
	HRIBelow
	HRIAboveAndBelow
)

const (
	UpcA = iota
	UpcE
	JanEAN8
	JanEAN13
	Code39
	Code93
	Code128
	ITF
	NW7
	GS1128
	GS1Omnidirectional
	GS1Truncated
	GS1Limited
	GS1Expanded
)

const (
	L = iota // L recovers 7% of data
	M        // M recovers 15% of data
	Q        // Q recovers 25% of data
	H        // H recovers 30% of data
)

const (
	DrawerPin2 = iota
	DrawerPin5
)

type Options interface {
	apply(Cmd)
}

type imageFuncVersionOption byte

func (ifv imageFuncVersionOption) apply(cmd Cmd) {
	if c, ok := cmd.(*escape); ok {
		switch ifv {
		case 1:
			c.imageFunc = c.imageV1
		case 2:
			c.imageFunc = c.imageV2
		default:
			c.imageFunc = c.imageObsolete
		}
	}
}

func WithImageFuncVersion(v byte) Options {
	return imageFuncVersionOption(v)
}

type pageHeight float64

func (ph pageHeight) apply(cmd Cmd) {
	if c, ok := cmd.(*postscript); ok {
		c.height = float64(ph)
		c.y = float64(ph)
	}
}

func WithPageHeight(height float64) Options {
	return pageHeight(height)
}

type barCodeFuncOption struct {
	fn func(byte, string) image.Image
}

func (cfo barCodeFuncOption) apply(cmd Cmd) {
	switch cmd.(type) {
	case *escape:
		cmd.(*escape).barCodeFunc = cfo.fn
	case *postscript:
		cmd.(*postscript).barCodeFunc = cfo.fn
	case *star:
		cmd.(*star).barCodeFunc = cfo.fn
	}
}

func WithBarCodeFunc(fn func(byte, string) image.Image) Options {
	return barCodeFuncOption{fn: fn}
}

type qrCodeFuncOption struct {
	fn func(string) image.Image
}

func (cfo qrCodeFuncOption) apply(cmd Cmd) {
	switch cmd.(type) {
	case *escape:
		cmd.(*escape).qrCodeFunc = cfo.fn
	case *postscript:
		cmd.(*postscript).qrCodeFunc = cfo.fn
	case *star:
		cmd.(*star).qrCodeFunc = cfo.fn
	}
}

func WithQRCodeFunc(fn func(string) image.Image) Options {
	return qrCodeFuncOption{fn: fn}
}
