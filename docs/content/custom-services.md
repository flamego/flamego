---
title: Custom services
weight: 40
---
The [core services](core-services) from Flamego are great, but they are certainly not enough for your web applications, Inevitably, you will start want to make your own [middleware](core-concepts#middleware) and custom services that are specifically fitting your needs.

## Injecting services

The Flame instance is building on top of the [`inject.TypeMapper`](https://pkg.go.dev/github.com/flamego/flamego/inject#TypeMapper) to provide service injections for your handlers. Both [`flamego.Flame`](https://pkg.go.dev/github.com/flamego/flamego#Flame) and [`flamego.Context`](https://pkg.go.dev/github.com/flamego/flamego#Context) have embeded the `inject.TypeMapper` that allow you to inject services at anywhere you want.

The `Map` method is used to inject services (aka. map values to their own types), the injected service can be a concrete type ([`*log.Logger`](https://pkg.go.dev/log#Logger)) or an interface ([`io.Writer`](https://pkg.go.dev/io#Writer)):

```go
l := log.New(os.Stdout, "[Flamego] ", 0)
f := flamego.New()
f.Map(l)

// or

var w io.Writer = &bytes.Buffer{}
f := flamego.New()
f.Map(w)
```

The `MapTo` method works similar to `Map` but instead of ~~mapping values to their own types~~, it allows you to _map values to interfaces_:

```go
buf := &bytes.Buffer{}
f := flamego.New()
f.MapTo(buf, (*io.Writer)(nil))
```

You may also use the `MapTo` method to transform interfaces to other interfaces:

```go
var w io.ReadCloser = io.NopCloser(&bytes.Buffer{})
f := flamego.New()
f.MapTo(w, (*io.Reader)(nil))
f.MapTo(w, (*io.Closer)(nil))
```

{{< callout type="warning" >}}
The `MapTo` method does a naive mapping and runtime panic could occur if the interface you're mapping to is not implemented by the type of the underlying value you're giving.
{{< /callout >}}

### Global services

When you inject services to the Flame instance without attaching to any route, these injected services are considered as global services, which are available for all handlers of the Flame instance.

There are two ways you can inject a global service, directly call `Map` or `MapTo` on the Flame instance, or through a [global middleware](core-concepts#middleware) using the `Use` method:

```go
db := database.New()
f := flamego.New()
f.Map(db)

// or

f := flamego.New()
f.Use(func(c flamego.Context) {
    db := database.New()
    c.Map(db)
})
```

### Group services

When you inject services to a group of routes, these injected services are considered as group services, which are only available for all handlers within the group.

You can only inject a group service through a [group middleware](core-concepts#middleware):

```go {hl_lines=["3-7", "14"] linenostart=1}
f := flamego.New()
f.Group("/user",
    func() {
        f.Get("/settings", func(user *database.User) {
            ...
        })
    },
    func(c flamego.Context) {
        user := database.GetUser()
        c.Map(user)
    },
)
f.Group("/repo", func() {
    f.Get("/settings", func(user *database.User) {
        // This handler will cause panic because *database.User is not available
    })
})
```

In the above example, the `*database.User` is only available to the group of routes on line 3 to 7. Trying to use it outside the group will cause panic as illustrated on line 14.

### Route-level services

When you inject services to a single route, these injected services are considered as route-level services, which are only available the handlers of that particular route.

You can only inject a route-level service through a [route-level middleware](core-concepts#middleware):

```go {hl_lines=["7-9", "11"] linenostart=1}
f := flamego.New()
f.Get("/user",
    func(c flamego.Context) {
        user := database.GetUser()
        c.Map(user)
    },
    f.Get("/settings", func(user *database.User) {
        ...
    }),
)
f.Get("/repo", func(user *database.User) {
    // This handler will cause panic because *database.User is not available
})
```

In the above example, the `*database.User` is only available to the route on line 7 to 9. Trying to use it in all other routes will cause panic as illustrated on line 11.

## Overriding services

Injected services can be overridden when you're not happy with the service functionality or behaviors provided by the other middleware.

Here is an example of overriding a global service at the route level:

```go {hl_lines=["13-14", "18-19"] linenostart=1}
package main

import (
	"bytes"
	"io"

	"github.com/flamego/flamego"
)

func main() {
	f := flamego.New()
	f.Use(func(c flamego.Context) {
		buf := bytes.NewBufferString("this is from a global service")
		f.MapTo(buf, (*io.Reader)(nil))
	})
	f.Get("/",
		func(c flamego.Context) {
			buf := bytes.NewBufferString("this is from a route-level service")
			f.MapTo(buf, (*io.Reader)(nil))
		},
		func(r io.Reader) string {
			p, err := io.ReadAll(r)
			if err != nil {
				// Handler error
			}
			return string(p)
		},
	)
	f.Run()
}
```

When you run the above program and do `curl http://localhost:2830/`, the following content are printed to your terminal:

```
$ curl http://localhost:2830
this is from a route-level service
```