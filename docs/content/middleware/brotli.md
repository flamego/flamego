---
title: brotli
weight: 80
---
The brotli middleware provides Brotli compression to responses for [Flame instances](../core-concepts#instances).

You can read source code of this middleware on [GitHub](https://github.com/flamego/brotli) and API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/flamego/brotli?tab=doc).

## Installation

```
go get github.com/flamego/brotli
```

## Usage examples

You should register the [`brotli.Brotli`](https://pkg.go.dev/github.com/flamego/brotli#Brotli) _before all other middleware or handlers_ that would write response to clients, and it works out-of-the-box with the default settings:

```go
package main

import (
	"github.com/flamego/brotli"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()
	f.Use(brotli.Brotli())
	f.Get("/", func() string {
		return "Hello, Brotli!"
	})
	f.Run()
}
```

The [`brotli.Options`](https://pkg.go.dev/github.com/flamego/brotli#Options) can be used to further customize the behavior of the middleware:

```go {hl_lines=["11-13"] linenostart=1}
package main

import (
	"github.com/flamego/brotli"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()
	f.Use(brotli.Brotli(
		brotli.Options{
			CompressionLevel: 11, // Best compression
		},
	))
	f.Get("/", func() string {
		return "Hello, Brotli!"
	})
	f.Run()
}
```