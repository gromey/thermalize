# thermalize

![https://img.shields.io/github/v/tag/gromey/thermalize](https://img.shields.io/github/v/tag/gromey/thermalize)
![https://img.shields.io/github/license/gromey/thermalize](https://img.shields.io/github/license/gromey/thermalize)

`thermalize` is a library that generates commands for thermal printers.

![logo.png](logo.png)

## Installation

`thermalize` can be installed like any other Go library through `go get`:

```console
go get github.com/gromey/thermalize
```

Or, if you are already using
[Go Modules](https://github.com/golang/go/wiki/Modules), you may specify a version number as well:

```console
go get github.com/gromey/thermalize@latest
```

## Getting Started

### Example 1 (ESC/POS)

```go
package main

import (
	"bytes"

	"github.com/gromey/thermalize"
	"github.com/phin1x/go-ipp"
)

func main() {
	w := new(bytes.Buffer)

	p := thermalize.NewEscape(48, 576, w, false)

	p.Init()
	p.CodePage(16)
	p.LineFeed()
	p.Align(thermalize.Center)
	p.Bold(true)
	p.Text("Hello world!", nil)
	p.LineFeed()
	p.FullCut()

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

	p := thermalize.NewStar(48, 576, w, false)

	p.Init()
	p.CodePage(32)
	p.LineFeed()
	p.Align(thermalize.Center)
	p.Bold(true)
	p.Text("Hello world!", nil)
	p.LineFeed()
	p.FullCut()

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

	p := thermalize.NewEscape(48, 576, w, false)

	// To change the level of gray that should be visible when printing, change GrayLevel setting.
	// Default is 127.
	thermalize.GrayLevel = 150

	p.Init()
	p.LineFeed()
	p.Align(thermalize.Center)

	// Before you send an image to print, you must ensure
	// that its width is less than or equal to the width of the print area.
	p.Image(thermalize.Logo(), false)
	p.LineFeed()
	p.FullCut()

	if err = w.Flush(); err != nil {
		panic(err)
	}
}

```
