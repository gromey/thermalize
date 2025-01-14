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

type imageFuncVersion byte

func (v imageFuncVersion) apply(cmd Cmd) {
	if c, ok := cmd.(*escape); ok {
		switch v {
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
	return imageFuncVersion(v)
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

type barCodeFunc func(byte, string) image.Image

func (fn barCodeFunc) apply(cmd Cmd) {
	switch cmd.(type) {
	case *escape:
		cmd.(*escape).barCodeFunc = fn
	case *postscript:
		cmd.(*postscript).barCodeFunc = fn
	case *star:
		cmd.(*star).barCodeFunc = fn
	}
}

func WithBarCodeFunc(fn func(byte, string) image.Image) Options {
	return barCodeFunc(fn)
}

type qrCodeFunc func(string) image.Image

func (fn qrCodeFunc) apply(cmd Cmd) {
	switch cmd.(type) {
	case *escape:
		cmd.(*escape).qrCodeFunc = fn
	case *postscript:
		cmd.(*postscript).qrCodeFunc = fn
	case *star:
		cmd.(*star).qrCodeFunc = fn
	}
}

func WithQRCodeFunc(fn func(string) image.Image) Options {
	return qrCodeFunc(fn)
}
