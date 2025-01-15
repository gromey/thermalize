package thermalize

import (
	"image"
)

type Cmd interface {
	// Sizing sets the number of characters per line (CPL) and pixels per line (PPL) after the command set has been initialized.
	Sizing(cpl, ppl int)

	// CPL returns the set number of characters per line.
	CPL() int

	// PPL returns the set number of pixel per line.
	PPL() int

	// Write writes raw bytes.
	// If a writer is not provided or an error occurs during writing, it will panic.
	Write(bs ...byte)

	// Text adds printable string along with encoding, if an encoder is provided.
	//
	// Since Golang uses UTF-8 character encoding by default, you must provide an encoder
	// to convert the string according to the specified code page.
	//
	// If the encoder is not provided, the text will be printed using the default UTF-8 encoding,
	// which may result in incorrect printing.
	Text(s string, enc func(string) []byte)

	// Init initializes printer.
	// Clears the data in the print buffer and resets the printer modes.
	Init()

	// LeftMargin sets left margin.
	LeftMargin(n int)

	// WidthArea sets print area width.
	WidthArea(n int)

	// AbsolutePosition sets absolute print position.
	AbsolutePosition(n int)

	// Align specifies position alignment.
	//
	//	b = 0, left justification enabled;
	//	b = 1, center justification enabled;
	//	b = 2, right justification enabled.
	Align(b byte)

	// UpsideDown selects upside-down print mode on/off.
	UpsideDown(b bool)

	// TabPositions sets horizontal tab position.
	// Default 8, 16, 24, 32, 40, ..., 232, 240, 248
	TabPositions(bs ...byte)

	// Tab moves the print position to the next horizontal tab position.
	// Default 8, 16, 24, 32, 40, ..., 232, 240, 248
	Tab()

	// CodePage selects character code table.
	CodePage(b byte)

	// CharSize selects character width and height.
	CharSize(w byte, h byte)

	// Bold selects emphasized printing.
	Bold(b bool)

	// ClockwiseRotation turns 90' clockwise rotation mode on/off.
	ClockwiseRotation(b bool)

	// Underling selects/cancels underling mode.
	//
	//	b = 0, underline mode disabled;
	//	b = 1, underline mode (1-dot thick) enabled;
	//	b = 2, underline mode (2-dot thick) enabled.
	Underling(b byte)

	// BarcodeWidth sets the 1D barcode width multiplier.
	//
	//	1 <= b <= 6.
	BarcodeWidth(b byte)

	// BarcodeHeight sets the 1D barcode height, measured in dots.
	//
	//	1 <= b <= 255.
	BarcodeHeight(b byte)

	// HRIFont selects HRI character font.
	//
	//	b = 0, font A (12 x 24);
	//	b = 1, font B (9 x 17).
	HRIFont(b byte)

	// HRIPosition selects HRI character print position.
	//
	//	b = 0, not printed;
	//	b = 1, above the barcode;
	//	b = 2, below the barcode;
	//	b = 3, above and below the barcode.
	HRIPosition(b byte)

	// Barcode adds a barcode to print.
	//
	//	m = 0, UpcA;
	//	m = 1, UpcE;
	//	m = 2, JanEAN8;
	//	m = 3, JanEAN13;
	//	m = 4, Code39;
	//	m = 5, Code93;
	//	m = 6, Code128;
	//	m = 7, ITF;
	//	m = 8, NW7;
	//	m = 9, GS1128;
	//	m = 10, GS1Omnidirectional;
	//	m = 11, GS1Truncated;
	//	m = 12, GS1Limited;
	//	m = 13, GS1Expanded.
	// If m is out of range, Code39 will be used by default.
	Barcode(m byte, s string)

	// QRCodeSize sets the size of module.
	//
	//	1 <= b <= 8.
	QRCodeSize(b byte)

	// QRCodeCorrectionLevel sets the correction level.
	//
	//	b = 0, correction level 7 %;
	//	b = 1, correction level 15 %;
	//	b = 2, correction level 25 %;
	//	b = 3, correction level 30 %.
	QRCodeCorrectionLevel(b byte)

	// QRCode adds a QR code to print.
	QRCode(s string)

	// Image adds an image to print.
	Image(img image.Image, invert bool)

	// Feed prints current buffer and executes n/4mm paper feed.
	//
	//	0 <= b <= 255.
	Feed(b byte)

	// LineFeed prints the data in the print buffer and feeds one line, based on the current line spacing.
	LineFeed()

	// Cut executes the auto-cutter.
	//
	//	m = 0, full cut at the current position;
	//	m = 1, partial cut at the current position;
	//	m = 2, paper is fed to cutting position, then a full cut;
	//	m = 3, paper is fed to cutting position, then a partial cut;
	Cut(m byte, p byte)

	// FullCut executes the auto-cutter across the full width of the paper.
	FullCut()

	// OpenCashDrawer generates pulse to open a cache drawer.
	OpenCashDrawer(m byte, t1 byte, t2 byte)

	// Print performs final preparation of the document before printing.
	Print()
}
