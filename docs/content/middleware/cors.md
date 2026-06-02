---
title: cors
weight: 60
---
The cors middleware configures [Cross-Origin Resource Sharing](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) for [Flame instances](/core-concepts#instances).

You can read source code of this middleware on [GitHub](https://github.com/flamego/cors) and API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/flamego/cors?tab=doc).

## Installation

```
go get github.com/flamego/cors
```

## Usage examples

The [`cors.CORS`](https://pkg.go.dev/github.com/flamego/cors#CORS) works out-of-the-box with an optional [`cors.Options`](https://pkg.go.dev/github.com/flamego/cors#Options):

```go
package main

import (
	"github.com/flamego/cors"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()
	f.Get("/",
		cors.CORS(),
		func(c flamego.Context) string {
			return "This endpoint can be shared cross-origin"
		},
	)
	f.Run()
}
```

The [`cors.Options`](https://pkg.go.dev/github.com/flamego/cors#Options) can be used to further customize the behavior of the middleware:

```go {hl_lines=["12-14"] linenostart=1}
package main

import (
	"github.com/flamego/cors"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()
	f.Get("/",
		cors.CORS(
            cors.Options{
			    AllowDomain: []string{"cors.example.com"},
		    },
        ),
		func(c flamego.Context) string {
			return "This endpoint can be shared cross-origin"
		},
	)
	f.Run()
}
```