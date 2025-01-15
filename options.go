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

func (h pageHeight) apply(cmd Cmd) {
	if c, ok := cmd.(*postscript); ok {
		c.height = float64(h)
		c.y = float64(h)
	}
}

func WithPageHeight(height float64) Options {
	return pageHeight(height)
}

type BarcodeOptions struct {
	// Mode defines the barcode type. Possible values are:
	//   - 0: UpcA
	//   - 1: UpcE
	//   - 2: JanEAN8
	//   - 3: JanEAN13
	//   - 4: Code39
	//   - 5: Code93
	//   - 6: Code128
	//   - 7: ITF
	//   - 8: NW7
	//   - 9: GS1128
	//   - 10: GS1Omnidirectional
	//   - 11: GS1Truncated
	//   - 12: GS1Limited
	//   - 13: GS1Expanded
	Mode byte

	// Width specifies the barcodes line width. Possible values are:
	//   1 to 6 (where 1 represents the thinnest lines and 6 the thickest).
	Width byte

	// Height specifies the barcodes height in arbitrary units. Possible values are:
	//   1 to 255.
	Height byte
}

type barcodeFunc func(string, BarcodeOptions) image.Image

func (f barcodeFunc) apply(cmd Cmd) {
	if c, ok := cmd.(*skipper); ok {
		c.barcodeFunc = f
	}
}

func WithBarcodeFunc(fn func(string, BarcodeOptions) image.Image) Options {
	return barcodeFunc(fn)
}

type QRCodeOptions struct {
	// CorrectionLevel determines the level of error correction in the QR code.
	// Higher levels can recover more data but reduce the amount of storable data.
	// Possible values are:
	//   - 0: L (Low) - Recovers up to 7% of data.
	//   - 1: M (Medium) - Recovers up to 15% of data.
	//   - 2: Q (Quartile) - Recovers up to 25% of data.
	//   - 3: H (High) - Recovers up to 30% of data.
	CorrectionLevel byte

	// Size determines the overall size of the QR code matrix. Larger sizes
	// allow more data to be encoded but result in a bigger QR code. Possible values are:
	//   1 to 8.
	Size byte
}

type qrcodeFunc func(data string, opts QRCodeOptions) image.Image

func (f qrcodeFunc) apply(cmd Cmd) {
	if c, ok := cmd.(*skipper); ok {
		c.qrcodeFunc = f
	}
}

func WithQRCodeFunc(fn func(data string, opts QRCodeOptions) image.Image) Options {
	return qrcodeFunc(fn)
}
