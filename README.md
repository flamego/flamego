![Flamego](https://github.com/flamego/brand-kit/raw/main/banner/banner-01.jpg#gh-light-mode-only)
![Flamego](https://github.com/flamego/brand-kit/raw/main/banner/banner-02.jpg#gh-dark-mode-only)

[![GitHub Workflow Status](https://img.shields.io/github/checks-status/flamego/flamego/main?logo=github&style=for-the-badge)](https://github.com/flamego/flamego/actions?query=branch%3Amain)
[![Codecov](https://img.shields.io/codecov/c/gh/flamego/flamego?logo=codecov&style=for-the-badge)](https://app.codecov.io/gh/flamego/flamego)
[![GoDoc](https://img.shields.io/badge/GoDoc-Reference-blue?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/flamego/flamego?tab=doc)
[![Sourcegraph](https://img.shields.io/badge/view%20on-Sourcegraph-brightgreen.svg?style=for-the-badge&logo=sourcegraph)](https://sourcegraph.com/github.com/flamego/flamego)

Flamego is a fantastic modular Go web framework with a slim core but limitless extensibility.

It is the successor of the [Macaron](https://github.com/go-macaron/macaron), and equips the most powerful routing syntax among all web frameworks within the Go ecosystem.

## Installation

The minimum requirement of Go is **1.19**.

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

- [The most powerful routing syntax](https://flamego.dev/routing.html) among all web frameworks within the Go ecosystem.
- [Limitless routes nesting and grouping](https://flamego.dev/routing.html#group-routes).
- [Inject middleware at wherever you want](https://flamego.dev/core-concepts.html#middleware).
- [Integrate with any existing Go web application non-intrusively](https://flamego.dev/faqs.html#how-do-i-integrate-into-existing-applications).
- [Dependency injection via function signature](https://flamego.dev/core-concepts.html#service-injection) to write testable and maintainable code.

## Middleware

- [Logger](https://flamego.dev/core-services.html#routing-logger) - Log requests and response status code
- [Recovery](https://flamego.dev/core-services.html#panic-recovery) - Automatic recovery from panics
- [Static](https://flamego.dev/core-services.html#serving-static-files) - Serve static files
- [Renderer](https://flamego.dev/core-services.html#rendering-content) - Render content
- [template](https://flamego.dev/middleware/template.html) - Go template rendering
- [session](https://flamego.dev/middleware/session.html) - User session management
- [recaptcha](https://flamego.dev/middleware/recaptcha.html) - Google reCAPTCHA verification
- [csrf](https://flamego.dev/middleware/csrf.html) - Generate and validate CSRF tokens
- [cors](https://flamego.dev/middleware/cors.html) - Cross-Origin Resource Sharing
- [binding](https://flamego.dev/middleware/binding.html) - Request data binding and validation
- [gzip](https://flamego.dev/middleware/gzip.html) - Gzip compression to responses
- [cache](https://flamego.dev/middleware/cache.html) - Cache management
- [brotli](https://flamego.dev/middleware/brotli.html) - Brotli compression to responses
- [auth](https://flamego.dev/middleware/auth.html) - Basic and bearer authentication
- [i18n](https://flamego.dev/middleware/i18n.html) - Internationalization and localization
- [captcha](https://flamego.dev/middleware/captcha.html) - Captcha service
- [hcaptcha](https://flamego.dev/middleware/hcaptcha.html) - hCaptcha verification

## Getting help

- New to Flamego? Check out the [starter guide](https://flamego.dev/starter-guide.html)!
- Have any questions? Answers may be found in our [FAQs](https://flamego.dev/faqs.html).
- Please [file an issue](https://github.com/flamego/flamego/issues) or [start a discussion](https://github.com/flamego/flamego/discussions) if you want to reach out.
- Follow our [Twitter](https://twitter.com/flamego_dev) to stay up to the latest news.
- Our [brand kit](https://github.com/flamego/brand-kit) is also available on GitHub!

## Users and projects

- [Cardinal](https://github.com/vidar-team/Cardinal): Attack-defence CTF platform.
- [mebeats](https://github.com/wuhan005/mebeats): Realtime heartbeat monitor service based on Mi band.
- [ASoulDocs](https://github.com/asoul-sig/asouldocs): Ellien's documentation server.
- [NekoBox](https://github.com/NekoWheel/NekoBox): Anonymous question box.
- [Codenotify.run](https://github.com/codenotify/codenotify.run): Codenotify as a Service.
- [Relay](https://github.com/bytebase/relay): A web server for forwarding events from service A to service B.
- [bilibili-lottery](https://github.com/flamego-examples/bilibili-lottery): 一款支持对哔哩哔哩视频或动态评论进行抽奖的小程序
- [pgrok](https://github.com/pgrok/pgrok): Poor man's ngrok.
- [Caramelverse](https://caramelverse.com): A fashion brand.
- [Sourcegraph Accounts](https://accounts.sourcegraph.com/): Centralized accounts system for all of the Sourcegraph-operated services
- _Just send a PR to add yours!_

## Development

Install "go-mockgen" and "goimports" to re-generate mocks:

```sh
go install github.com/derision-test/go-mockgen/cmd/go-mockgen@latest
go install golang.org/x/tools/cmd/goimports@latest

go generate ./...
```

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
