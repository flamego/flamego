---
title: gzip
weight: 70
---
The gzip middleware provides Gzip compression to responses for [Flame instances](/core-concepts#instances).

You can read source code of this middleware on [GitHub](https://github.com/flamego/gzip) and API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/flamego/gzip?tab=doc).

## Installation

```
go get github.com/flamego/gzip
```

## Usage examples

You should register the [`gzip.Gzip`](https://pkg.go.dev/github.com/flamego/gzip#Gzip) _before all other middleware or handlers_ that would write response to clients, and it works out-of-the-box with the default settings:

```go
package main

import (
	"github.com/flamego/flamego"
	"github.com/flamego/gzip"
)

func main() {
	f := flamego.Classic()
	f.Use(gzip.Gzip())
	f.Get("/", func() string {
		return "Hello, Gzip!"
	})
	f.Run()
}
```

The [`gzip.Options`](https://pkg.go.dev/github.com/flamego/gzip#Options) can be used to further customize the behavior of the middleware:

```go {hl_lines=["11-13"] linenostart=1}
package main

import (
	"github.com/flamego/flamego"
	"github.com/flamego/gzip"
)

func main() {
	f := flamego.Classic()
	f.Use(gzip.Gzip(
		gzip.Options{
			CompressionLevel: 9, // Best compression
		},
	))
	f.Get("/", func() string {
		return "Hello, Gzip!"
	})
	f.Run()
}
```