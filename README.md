# thermalize

![https://img.shields.io/github/v/tag/gromey/thermalize](https://img.shields.io/github/v/tag/gromey/thermalize)
![https://img.shields.io/github/license/gromey/thermalize](https://img.shields.io/github/license/gromey/thermalize)

`thermalize` is a library that generates commands for thermal printers.

![logo.png](logo.png)

## Installation

`thermalize` can be installed like any other Go library through `go get`:

```console
go get github.com/gromey/thermalize@latest
```

## Getting Started

### Example 1 (ESC/POS)

```go
package main

import (
	"bufio"
	"os"

	"github.com/gromey/thermalize"
)

func main() {
	f, err := os.OpenFile("/dev/ttyUSB0", os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)

	p := thermalize.NewEscape(48, 576, w)

	p.Init()
	p.CodePage(16)
	p.LineFeed()
	p.Align(thermalize.Center)
	p.Bold(true)
	p.Text("Hello world!", nil)
	p.LineFeed()
	p.FullCut()
	p.Print()

	if err = w.Flush(); err != nil {
		panic(err)
	}
}

```

### Example 2 (Star)

```go
package main

import (
	"bufio"
	"os"

	"github.com/gromey/thermalize"
)

func main() {
	f, err := os.OpenFile("/dev/ttyUSB0", os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)

	p := thermalize.NewStar(48, 576, w)

	p.Init()
	p.CodePage(32)
	p.LineFeed()
	p.Align(thermalize.Center)
	p.Bold(true)
	p.Text("Hello world!", nil)
	p.LineFeed()
	p.FullCut()
	p.Print()

	if err = w.Flush(); err != nil {
		panic(err)
	}
}

```

### Example 3 (Image)

```go
package main

import (
	"bufio"
	"os"

	"github.com/gromey/thermalize"
)

func main() {
	f, err := os.OpenFile("/dev/ttyUSB0", os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)

	p := thermalize.NewEscape(48, 576, w, thermalize.WithImageFuncVersion(1))

	// To change the level of gray that should be visible when printing, change GrayLevel setting.
	// Default is 127.
	thermalize.SetGrayLevel(150)

	p.Init()
	p.LineFeed()
	p.Align(thermalize.Center)

	// Before you send an image to print, you must ensure
	// that its width is less than or equal to the width of the print area.
	p.Image(thermalize.Logo(), false)
	p.LineFeed()
	p.FullCut()
	p.Print()

	if err = w.Flush(); err != nil {
		panic(err)
	}
}

```

### Example 4 (Postscript)

```go
package main

import (
	"bytes"
	"image"

	"github.com/gromey/thermalize"
	"github.com/phin1x/go-ipp"
)

func main() {
	w := new(bytes.Buffer)

	barcodeFunc := func(d string, o thermalize.BarcodeOptions) image.Image {
		// get Bar code here.
		return nil
	}

	qrcodeFunc := func(d string, o thermalize.QRCodeOptions) image.Image {
		// get QR code here.
		return nil
	}

	opts := []thermalize.Options{
		thermalize.WithPageHeight(5670),
		thermalize.WithBarcodeFunc(barcodeFunc),
		thermalize.WithQRCodeFunc(qrcodeFunc),
	}

	p := thermalize.NewPostscript(48, 576, w, opts...)

	p.Init()
	p.LineFeed()
	p.Align(thermalize.Center)
	p.Bold(true)
	p.Text("Hello world!", nil)
	p.LineFeed()
	p.FullCut()
	p.Print()

	defer w.Reset()

	doc := ipp.Document{
		Document: w,
		Size:     w.Len(),
		Name:     "Test Page",
		MimeType: ipp.MimeTypeOctetStream,
	}

	client := ipp.NewIPPClient("localhost", 631, "", "", true)

	if _, err := client.PrintJob(doc, "your_printer_name", nil); err != nil {
		panic(err)
	}
}

```
