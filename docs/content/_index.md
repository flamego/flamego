---
title: 
weight: 1
toc: false
linkTitle: Flamego
type: docs
---
![banner](/imgs/banner.png)

Flamego is a modular Go framework for building composable systems, and equips the most powerful routing syntax among all web frameworks within the Go ecosystem.

## Installation
```
go get github.com/flamego/flamego
```

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

- [The most powerful routing syntax](routing) among all web frameworks within the Go ecosystem.
- [Limitless routes nesting and grouping](routing#group-routes).
- [Inject middleware at wherever you want](core-concepts#middleware).
- [Integrate with any existing Go web application non-intrusively](faqs#how-do-i-integrate-into-existing-applications).
- [Dependency injection via function signature](core-concepts#service-injection) to write testable and maintainable code.

## Exploring more

- New to Flamego? Check out the [Starter guide](starter-guide)!
- Look up [Middleware](middleware/) that are built for Flamego.
- Have any questions? Answers may be found in our [FAQs](faqs).
- Please [file an issue](https://github.com/flamego/flamego/issues) or [start a discussion](https://github.com/flamego/flamego/discussions) if you want to reach out.
- Follow our [Twitter](https://twitter.com/flamego_dev) to stay up to the latest news.
- Our [brand kit](https://github.com/flamego/brand-kit) is also available on GitHub!