---
title: FAQs
weight: 70
---
## How do I change the listen address?

If you're using the `Run` method to start the web server, you may change the listen address using either the environment variable `FLAMEGO_ADDR`:

```sh
export FLAMEGO_ADDR=localhost:8888
```

Or the variable arguments of the `Run` method:

```go
f.Run("localhost")       // => localhost:2830
f.Run(8888)              // => 0.0.0.0:8888
f.Run("localhost", 8888) // => localhost:8888
```

Alternatively, [`http.ListenAndServe`](https://pkg.go.dev/net/http#ListenAndServe) or [`http.ListenAndServeTLS`](https://pkg.go.dev/net/http#ListenAndServeTLS) can also be used to change the listen address:

```go
http.ListenAndServe("localhost:8888", f)
http.ListenAndServeTLS("localhost:8888", "certFile", "keyFile", f)
```

## How do I do graceful shutdown?

The [github.com/ory/graceful](https://github.com/ory/graceful) package can be used to do graceful shutdown with the Flame instance:

```go
package main

import (
	"net/http"

	"github.com/flamego/flamego"
	"github.com/ory/graceful"
)

func main() {
	f := flamego.New()

	...

	server := graceful.WithDefaults(
		&http.Server{
			Addr:    "0.0.0.0:2830",
			Handler: f,
		},
	)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		// Handler error
	}
}
```

## How do I serve file downloads?

```go
import (
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/flamego/flamego"
	"golang.org/x/exp/utf8string"
)

func main() {
	f := flamego.Classic()
	f.Get("/download", func(w http.ResponseWriter, r *http.Request) {
		fpath := "your filepath"
		filename := filepath.Base(fpath)
		if utf8string.NewString(filename).IsASCII() {
			w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		} else {
			w.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
		}
		http.ServeFile(w, r, fpath)
	})
	f.Run()
}
```

## How do I integrate into existing applications?

Because Flame instances implement the [`http.Handler`](https://pkg.go.dev/net/http#Handler) interface, a Flame instance can be plugged into anywhere that accepts a `http.Handler`.

### Example: Integrating with `net/http`

Below is an example of integrating with the `net/http` router for a single route `"/user/info"`:

{{< tabs >}}
{{< tab name="Code" >}}
```go
package main

import (
	"log"
	"net/http"

	"github.com/flamego/flamego"
)

func main() {
	f := flamego.New()
	f.Get("/user/info", func() string {
		return "The user is Joe"
	})

	// Pass on all routes under "/user/" to the Flame isntance
	http.Handle("/user/", f)

	if err := http.ListenAndServe("0.0.0.0:2830", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```
{{< /tab >}}
{{< tab name="Test" >}}
```
$ curl -i http://localhost:2830/user/info
The user is Joe
```
{{< /tab >}}
{{< /tabs >}}

### Example: Integrating with Macaron

Below is an example of integrating with the Macaron router for a single route `"/user/info"`:

{{< tabs >}}
{{< tab name="Code" >}}
```go
package main

import (
	"log"
	"net/http"

	"github.com/flamego/flamego"
	"gopkg.in/macaron.v1"
)

func main() {
	f := flamego.New()
	f.Get("/user/info", func() string {
		return "The user is Joe"
	})

	// Pass on all routes under "/user/" to the Flame isntance
	m := macaron.New()
	m.Any("/user/*", f.ServeHTTP)

	if err := http.ListenAndServe("0.0.0.0:2830", m); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```
{{< /tab >}}
{{< tab name="Test" >}}
```
$ curl -i http://localhost:2830/user/info
The user is Joe
```
{{< /tab >}}
{{< /tabs >}}

## What is the difference between `inject.Invoker` and `inject.FastInvoker`?

The [`inject.Invoker`](https://pkg.go.dev/github.com/flamego/flamego/inject#Invoker) is the default way that the Flame instance uses to invoke a function through reflection.

In 2016, [@tupunco](https://github.com/tupunco) [contributed a patch](https://github.com/go-macaron/inject/commit/07e997cf1c187f573791bd7680cfdcba43161c22) with the concept and the implementation of the [`inject.FastInvoker`](https://pkg.go.dev/github.com/flamego/flamego/inject#FastInvoker), which invokes a function through interface. The `inject.FastInvoker` is about 30% faster to invoke a function and uses less memory.

## What is the idea behind this other than Macaron/Martini?

Martini brought the brilliant idea of build a web framework with dependency injection in a magical experience. However, it has terrible performance and high memory usage. Some people are blaming the use of reflection for its slowness and memory footprint, but that is not fair by the way, most of people are using reflections every single day with marshalling and unmarshalling JSON in Go.

Macaron achieved the reasonable performance and much lower memory usage. Unfortunately, it was not a properly designed product, or let's be honest, there was no design. The origin of Macaron was to support the rapid development of the [Gogs](https://gogs.io) project, thus almost all things were inherited from some other web frameworks at the time.

Absence of holistic architecture view and design principles have caused many bad decisions, including but not limited to:

- The [`*macaron.Context`](https://pkg.go.dev/github.com/go-macaron/macaron#Context) is very heavy, thus [separation of concerns](https://en.wikipedia.org/wiki/Separation_of_concerns) wasn't a thing.
- The choice of using the opening-only style (e.g. `:name`) for [named parameters](https://go-macaron.com/middlewares/routing#named-parameters) has limited capability and extensibility of the routing syntax.
- Being too opinionated in many aspects, a simple example is the existence of [`SetConfig`](https://pkg.go.dev/github.com/go-macaron/macaron#SetConfig)/[`Config`](https://pkg.go.dev/github.com/go-macaron/macaron#Config) that kinda kidnaps all users to import the package `"gopkg.in/ini.v1"` but not using it at 99% of the time.
- The way to [set a cookie](https://go-macaron.com/core_services#cookie) is a disaster.

All in all, Macaron is still an excellent web framework, and Flamego is just better as the successor. 🙂

## Why the default port is 2830?

![keyboard layout 2830](/imgs/keyboard-layout-2830.png)