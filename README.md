![Flamego](https://github.com/flamego/brand-kit/raw/main/banner/banner-01.jpg)

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/flamego/flamego/Go?logo=github&style=for-the-badge)](https://github.com/flamego/flamego/actions?query=workflow%3AGo)
[![Codecov](https://img.shields.io/codecov/c/gh/flamego/flamego?logo=codecov&style=for-the-badge)](https://app.codecov.io/gh/flamego/flamego)
[![GoDoc](https://img.shields.io/badge/GoDoc-Reference-blue?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/flamego/flamego?tab=doc)
[![Sourcegraph](https://img.shields.io/badge/view%20on-Sourcegraph-brightgreen.svg?style=for-the-badge&logo=sourcegraph)](https://sourcegraph.com/github.com/flamego/flamego)

Flamego is a fantastic modular Go web framework boiled with dependency injection.

It is the successor of the [Macaron](https://github.com/go-macaron/macaron), and equips the most powerful routing syntax among all web frameworks within the Go ecosystem.

## Installation

The minimum requirement of Go is **1.16**.

	go get github.com/flamego/flamego

## Getting started

```go
package main

import "github.com/flamego/flamego"

func main() {
	f := flamego.Classic()
	f.Get("/", func() string {
		return "Hello, Flamego!"
	})
	f.Run()
}
```

## Features

_Stay tuned!_

## Middleware

- [`Logger`](logger.go) - Log requests and response status code
- [`Recovery`](recovery.go) - Automatic recovery from panics
- [`Static`](static.go) - Serve static files
- [template](https://github.com/flamego/template) - Go template rendering
- [session](https://github.com/flamego/session) - User session management
- [recaptcha](https://github.com/flamego/recaptcha) - Google reCAPTCHA verification
- [csrf](https://github.com/flamego/csrf) - Generate and validate CSRF tokens
- [cors](https://github.com/flamego/cors) - Cross-Origin Resource Sharing
- [binding](https://github.com/flamego/binding) - Request data binding and validation
- [gzip](https://github.com/flamego/gzip) - Gzip compression to responses
- [cache](https://github.com/flamego/cache) - Cache management
- [brotli](https://github.com/flamego/brotli) - Brotli compression to responses

## Getting help

_Stay tuned!_

## Users and projects

- [Cardinal](https://github.com/vidar-team/Cardinal): Attack-defence CTF platform.
- [mebeats](https://github.com/wuhan005/mebeats): Realtime heartbeat monitor service based on Mi band.
- _Just send a PR to add yours!_

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
