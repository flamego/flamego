---
title: Core concepts
weight: 20
---
This page describes foundational concepts that are required to be proficient of using Flamego to build web applications that are most optimal.

## Classic Flame

The classic Flame instance is the one that comes with a reasonable list of default middleware and could be your starting point for build web applications using Flamego.

A fresh classic Flame instance is returned every time you call [`flamego.Classic`](https://pkg.go.dev/github.com/flamego/flamego#Classic), and following middleware are registered automatically:

- [`flamego.Logger`](core-services#routing-logger) for logging requested routes.
- [`flamego.Recovery`](core-services#panic-recovery) for recovering from panic.
- [`flamego.Static`](core-services#serving-static-files) for serving static files.

{{< callout type="info" >}}
If you look up [the source code of the `flamego.Classic`](https://github.com/flamego/flamego/blob/8505d18c5243f797d5bb7160797d26454b9e5011/flame.go#L65-L77), it is fairly simple:

```go
func Classic() *Flame {
	f := New()
	f.Use(
		Logger(),
		Recovery(),
		Static(
			StaticOptions{
				Directory: "public",
			},
		),
	)
	return f
}
```

Do keep in mind that `flamego.Classic` may not always be what you want if you do not use these default middleware (e.g. for using custom implementations), or to use different config options, or even just want to change the order of middleware as sometimes the order matters (i.e. middleware are being invoked in the same order as they are registered).
{{< /callout >}}

## Instances

The function [`flamego.New`](https://pkg.go.dev/github.com/flamego/flamego#New) is used to create bare Flame instances that do not have default middleware registered, and any type that contains the [`flamego.Flame`](https://pkg.go.dev/github.com/flamego/flamego#Flame) can be seen as a Flame instance.

Each Flame instace is independent of other Flame instances in the sense that instance state is not shared and is maintained separately by each of them. For example, you can have two Flame instances simultaneously and both of them can have different middleware, routes and handlers registered or defined:

```go
func main() {
	f1 := flamego.Classic()

	f2 := flamego.New()
	f2.Use(flamego.Recovery())

    ...
}
```

In the above example, `f1` has some default middleware registered as a classic Flame instance, while `f2` only has a single middleware `flamego.Recovery`.

{{< callout type="info" >}}
**💬 Do you agree?**

Storing states in the way that is polluting global namespace is such a bad practice that not only makes the code hard to maintain in the future, but also creates more tech debt with every single new line of the code.

It feels so elegent to have isolated state managed by each Flame instance, and make it possible to migrate existing web applications to use Flamego progressively.
{{< /callout >}}

## Handlers

Flamego handlers are defined as [`flamego.Handler`](https://pkg.go.dev/github.com/flamego/flamego#Handler), and if you look closer, it is just an empty interface (`interface{}`):

```go
// Handler is any callable function or a value that implements http.Handler.
// Flamego attempts to inject services into the Handler's argument list and
// panics if any argument could not be fulfilled via dependency injection.
// Values implementing http.Handler are invoked via their ServeHTTP method.
type Handler interface{}
```

As being noted in the docstring, any callable function is a valid `flamego.Handler`, doesn't matter if it's an anonymous, a declared function or even a method of a type:

{{< tabs >}}
{{< tab name="Code" >}}
```go
package main

import (
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.New()
	f.Get("/anonymous", func() string {
		return "Respond from an anonymous function"
	})
	f.Get("/declared", declared)

	t := &customType{}
	f.Get("/method", t.handler)

	f.Run()
}

func declared() string {
	return "Respond from a declared function"
}

type customType struct{}

func (t *customType) handler() string {
	return "Respond from a method of a type"
}
```
{{< /tab >}}
{{< tab name="Test" >}}
```
$ curl http://localhost:2830/anonymous
Respond from an anonymous function

$ curl http://localhost:2830/declared
Respond from a declared function

$ curl http://localhost:2830/method
Respond from a method of a type
```
{{< /tab >}}
{{< /tabs >}}

{{< callout type="info" >}}
**🆕 Available in v1.9.11**

Any value that implements [`http.Handler`](https://pkg.go.dev/net/http#Handler) is also a valid `flamego.Handler`, so you can plug in standard library handlers or third-party `http.Handler` implementations directly:

```go
f.Get("/files/*", http.StripPrefix("/files/", http.FileServer(http.Dir("./public"))))
```
{{< /callout >}}

## Return values

Generally, your web application needs to write content directly to the [`http.ResponseWriter`](https://pkg.go.dev/net/http#ResponseWriter) (which you can retrieve using `ResponseWriter` method of [`flamego.Context`](https://pkg.go.dev/github.com/flamego/flamego#Context)). In some web frameworks, they offer returning an extra `error` as the indication of the server error as follows:

```go
func handler(w http.ResponseWriter, r *http.Request) error
```

However, you are still being limited to a designated list of return values from your handlers. In contrast, Flamego provides the flexibility of having different lists of return values from handlers based on your needs case by case, whether it's an error, a string, or just a status code.

Let's see some examples that you can use for your handlers:

{{< tabs >}}
{{< tab name="Code" >}}
```go
package main

import (
	"errors"

	"github.com/flamego/flamego"
)

func main() {
	f := flamego.New()
	f.Get("/string", func() string {
		return "Return a string"
	})
	f.Get("/bytes", func() []byte {
		return []byte("Return some bytes")
	})
	f.Get("/error", func() error {
		return errors.New("Return an error")
	})
	f.Run()
}
```
{{< /tab >}}
{{< tab name="Test" >}}
```
$ curl -i http://localhost:2830/string
HTTP/1.1 200 OK
...

Return a string

$ curl -i http://localhost:2830/bytes
HTTP/1.1 200 OK
...

Return some bytes

$ curl -i http://localhost:2830/error
HTTP/1.1 500 Internal Server Error
...

Return an error
...
```
{{< /tab >}}
{{< /tabs >}}

As you can see, if an error is returned, the Flame instance automatically sets the HTTP status code to be 500.

{{< callout type="info" >}}
Try returning `nil` for the error on line 18, then redo the test request and see what changes.
{{< /callout >}}

### Return with a status code

In the cases that you want to have complete control over the status code of your handlers, that is also possible!

{{< tabs >}}
{{< tab name="Code" >}}
```go
package main

import (
	"errors"
	"net/http"

	"github.com/flamego/flamego"
)

func main() {
	f := flamego.New()
	f.Get("/string", func() (int, string) {
		return http.StatusOK, "Return a string"
	})
	f.Get("/bytes", func() (int, []byte) {
		return http.StatusOK, []byte("Return some bytes")
	})
	f.Get("/error", func() (int, error) {
		return http.StatusForbidden, errors.New("Return an error")
	})
	f.Run()
}
```
{{< /tab >}}
{{< tab name="Test" >}}
```
$ curl -i http://localhost:2830/string
HTTP/1.1 200 OK
...

Return a string

$ curl -i http://localhost:2830/bytes
HTTP/1.1 200 OK
...

Return some bytes

$ curl -i http://localhost:2830/error
HTTP/1.1 403 Forbidden
...

Return an error
...
```
{{< /tab >}}
{{< /tabs >}}

### Return body with potential error

Body or error? Not a problem!

{{< tabs >}}
{{< tab name="Code" >}}
```go
package main

import (
	"errors"
	"net/http"

	"github.com/flamego/flamego"
)

func main() {
	f := flamego.New()
	f.Get("/string", func() (string, error) {
		return "Return a string", nil
	})
	f.Get("/bytes", func() ([]byte, error) {
		return []byte("Return some bytes"), nil
	})
	f.Run()
}
```
{{< /tab >}}
{{< tab name="Test" >}}
```
$ curl -i http://localhost:2830/string
HTTP/1.1 200 OK
...

Return a string

$ curl -i http://localhost:2830/bytes
HTTP/1.1 200 OK
...

Return some bytes
```
{{< /tab >}}
{{< /tabs >}}

If the handler returns a non-`nil` error, the error message will be responded to the client instead.

![How cool is that?](https://media0.giphy.com/media/hS4Dz87diTpnDXf98E/giphy.gif?cid=ecf05e47go1oiqgxj1ro7e3t1usexogh109gigssvhxlp93a&rid=giphy.gif&ct=g)

## Service injection

Flamego is claimed to be boiled with [dependency injection](https://en.wikipedia.org/wiki/Dependency_injection) because of the service injection, it is the soul of the framework. The Flame instance uses the [`inject.Injector`](https://pkg.go.dev/github.com/flamego/flamego/inject#Injector) to manage injected services and resolves dependencies of a handler's argument list at the time of the handler invocation.

Both dependency injection and service injection are very abstract concepts, so it is much easier to explain with examples:

```go
// Both `http.ResponseWriter` and `*http.Request` are injected,
// so they can be used as handler arguments.
f.Get("/", func(w http.ResponseWriter, r *http.Request) { ... })

// The `flamego.Context` is probably the most frequently used
// service in your web applications.
f.Get("/", func(c flamego.Context) { ... })
```

What happens if you try to use a service that hasn't been injected?

{{< tabs >}}
{{< tab name="Code" >}}
```go
package main

import (
	"github.com/flamego/flamego"
)

type myService struct{}

func main() {
	f := flamego.New()
	f.Get("/", func(s myService) {})
	f.Run()
}
```
{{< /tab >}}
{{< tab name="Test" >}}
```
http: panic serving 127.0.0.1:50061: unable to invoke the 0th handler [func(main.myService)]: value not found for type main.myService
...
```
{{< /tab >}}
{{< /tabs >}}

{{< callout type="info" >}}
If you're interested in learning how exactly the service injection works in Flamego, the [custom services](custom-services) has the best resources you would want.
{{< /callout >}}

### Builtin services

There are services that are always injected thus available to every handler, including [`*log.Logger`](https://pkg.go.dev/log#Logger), [`flamego.Context`](https://pkg.go.dev/github.com/flamego/flamego#Context), [`http.ResponseWriter`](https://pkg.go.dev/net/http#ResponseWriter) and [`*http.Request`](https://pkg.go.dev/net/http#Request).

## Middleware

Middleware are the special kind of handlers that are designed as reusable components, and often accepting configurable options. There is no difference between middleware and handlers from compiler's point of view.

Technically speaking, you may use the term middleware and handlers interchangably but the common sense would be that middleware are providing some services, either by [injecting to the context](https://github.com/flamego/session/blob/f8f1e1893ea6c15f071dd53aefd9494d41ce9e48/session.go#L183-L184) or [intercepting the request](https://github.com/flamego/auth/blob/dbec68df251ff382e908eb5659453d4918a042aa/basic.go#L38-L42), or both. On the other hand, handlers are mainly focusing on the business logic that is unique to your web application and the route that handlers are registered with.

Middleware can be used at anywhere that a `flamego.Handler` is accepted, including at global, group and route level.

```go {hl_lines=["6-9"] linenostart=1}
// Global middleware that are invoked before all other handlers.
f.Use(middleware1, middleware2, middleware3)

// Group middleware that are scoped down to a group of routes.
f.Group("/",
	func() {
		f.Get("/hello", func() { ... })
	},
	middleware4, middleware5, middleware6,
)

// Route-level middleware that are scoped down to a single route.
f.Get("/hi", middleware7, middleware8, middleware9, func() { ... })
```

Please be noted that middleware are always invoked first when a route is matched, i.e. even though that middleware on line 9 appear to be after the route handlers in the group (from line 6 to 8), they are being invoked first regardless.

{{< callout type="info" >}}
**💡 Did you know?**

Global middleware are always invoked regardless whether a route is matched.
{{< /callout >}}

{{< callout type="info" >}}
If you're interested in learning how to inject services for your middleware, the [custom services](custom-services) has the best resources you would want.
{{< /callout >}}

## Env

Flamego environment provides the ability to control behaviors of middleware and handlers based on the running environment of your web application. It is defined as the type [`EnvType`](https://pkg.go.dev/github.com/flamego/flamego#EnvType) and has some pre-defined values, including `flamego.EnvTypeDev`, `flamego.EnvTypeProd` and `flamego.EnvTypeTest`, which is for indicating development, production and testing environment respectively.

For example, the [template](./middleware/template#template-caching) middleware [rebuilds template files for every request when in `flamego.EnvTypeDev`](https://github.com/flamego/template/blob/ced6948bfc8cb49e32412380e407cbbe01485937/template.go#L229-L241), but caches the template files otherwise.

The Flamego environment is typically configured via the environment variable `FLAMEGO_ENV`:

```sh
export FLAMEGO_ENV=development
export FLAMEGO_ENV=production
export FLAMEGO_ENV=test
```

In case you want to retrieve or alter the environment in your web application, [`Env`](https://pkg.go.dev/github.com/flamego/flamego#Env) and [`SetEnv`](https://pkg.go.dev/github.com/flamego/flamego#SetEnv) methods are also available, and both of them are safe to be used concurrently.