---
title: Routing
weight: 50
---
Every request goes through the routing in order to reach its handlers, that is how important the routing is to be ergonomic. Flamego spends great amount of effort on designing and implementing the routing capability and future extensibility to ensure your developer happiness.

In Flamego, a route is an HTTP method paired with a URL-matching pattern, and each route takes _a chain of handlers_.

Below are the helpers to register routes for each HTTP method:

```go
f.Get("/", ...)
f.Patch("/", ...)
f.Post("/", ...)
f.Put("/", ...)
f.Delete("/", ...)
f.Options("/", ...)
f.Head("/", ...)
f.Connect("/", ...)
f.Trace("/", ...)
```

If you want to match all HTTP methods for a single route, `Any` is available for you:

```go
f.Any("/", ...)
```

When you want to match a selected list of HTTP methods for a single route, `Routes` is your friend:

```go
f.Routes("/", "GET,POST", ...)

// or

f.Routes("/", http.MethodGet, http.MethodPost, ...)
```

## Terminology

- A **URL path segment** is the portion between two forward slashes, e.g. `/<segment>/`, the trailing forward slash may not present.
- A **bind parameter** uses curly brackets (`{}`) as its notation, e.g. `{<bind_parameter>}`, bind parameters are only available in [dynamic routes](#dynamic-routes).

## Static routes

The static routes are probably the most common routes you have been seeing and using, routes are defined in literals and only looking for _exact matches_:

```go
f.Get("/user", ...)
f.Get("/repo", ...)
```

In the above example, any request that is not a GET request to `/user` or `/repo` will result in 404.

{{< callout type="warning" >}}
Unlike the router in `net/http`, where you may use `/user/` to match all subpaths under the `/user` path, it is still a static route in Flamego, and only matches a route IFF the request path is exactly the `/user/`.

Let's write an example:

```go
package main

import (
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.New()
	f.Get("/user/", func() string {
		return "You got it!"
	})
	f.Run()
}
```

Then run some tests as follows:

```
$ curl http://localhost:2830/user
404 page not found

$ curl http://localhost:2830/user/
You got it!

$ curl http://localhost:2830/user/info
404 page not found
```
{{< /callout >}}

## Dynamic routes

The dynamic routes, by its name, they match request paths dynamically. Flamego provides most powerful dynamic routes in the Go ecosystem, at the time of writing, there is simply no feature parity you can find in all other existing Go web frameworks.

The `flamego.Context` provides a family of `Param` methods to access values that are captured by bind parameters, including:

- `Params` returns all bind parameters.
- `Param` returns value of the given bind parameter.
- `ParamInt` returns value parsed as int.
- `ParamInt64` returns value parsed as int64.

### Placeholders

A placeholder captures anything but a forward slash (`/`), and you may have one or more placeholders within a URL path segment.

Below are all valid usages of placeholders:

```go
f.Get("/users/{name}", ...)
f.Get("/posts/{year}-{month}-{day}.html", ...)
f.Get("/geo/{state}/{city}", ...)
```

On line 1, the placeholder named `{name}` to capture everything in a URL path segment.

On line 2, three placeholders `{year}`, `{month}` and `{day}` are used to capture different portions in a URL path segment.

On line 3, two placeholders are used independently in different URL path segments.

Let's see some examples:

{{< tabs >}}
{{< tab name="Code" >}}
```go
package main

import (
	"fmt"
	"strings"

	"github.com/flamego/flamego"
)

func main() {
	f := flamego.New()
	f.Get("/users/{name}", func(c flamego.Context) string {
		return fmt.Sprintf("The user is %s", c.Param("name"))
	})
	f.Get("/posts/{year}-{month}-{day}.html", func(c flamego.Context) string {
		return fmt.Sprintf(
			"The post date is %d-%d-%d",
			c.ParamInt("year"), c.ParamInt("month"), c.ParamInt("day"),
		)
	})
	f.Get("/geo/{state}/{city}", func(c flamego.Context) string {
		return fmt.Sprintf(
			"Welcome to %s, %s!",
			strings.Title(c.Param("city")),
			strings.ToUpper(c.Param("state")),
		)
	})
	f.Run()
}
```
{{< /tab >}}
{{< tab name="Test" >}}
```
$ curl http://localhost:2830/users/joe
The user is joe

$ curl http://localhost:2830/posts/2021-11-26.html
The post date is 2021-11-26

$ curl http://localhost:2830/geo/ma/boston
Welcome to Boston, MA!
```
{{< /tab >}}
{{< /tabs >}}

{{< callout type="info" >}}
Try a test request using `curl http://localhost:2830/posts/2021-11-abc.html` and see what changes.
{{< /callout >}}

{{< callout type="info" >}}
**🆕 Available in v1.9.10**

Captured bind parameters are also populated onto the request via [`(*http.Request).SetPathValue`](https://pkg.go.dev/net/http#Request.SetPathValue), so `r.PathValue("name")` returns the same value as `c.Param("name")`.
{{< /callout >}}

### Regular expressions

A bind parameter can be defined with a custom regular expression to capture characters in a URL path segment, and you may have one or more such bind parameters within a URL path segment. The regular expressions are needed to be surrounded by a pair of forward slashes (`/<regexp>/`).

Below are all valid usages of bind parameters with regular expressions:

```go
f.Get("/users/{name: /[a-zA-Z0-9]+/}", ...)
f.Get("/posts/{year: /[0-9]{4}/}-{month: /[0-9]{2}/}-{day: /[0-9]{2}/}.html", ...)
f.Get("/geo/{state: /[A-Z]{2}/}/{city}", ...)
```

On line 1, the placeholder named `{name}` to capture any alphanumeric in a URL path segment.

On line 2, three placeholders `{year}`, `{month}` and `{day}` are used to capture certain length of digits in different portions in a URL path segment.

On line 3, the placeholder `state` will only match two-digit and upper-case alphabets.

{{< callout type="info" >}}
Because forward slashes are used to indicate the use of regular expressions, they cannot be captured via regular expressions, and will cause a routing parser error when you are trying to do so:

```
panic: unable to parse route "/{name: /abc\\//}": 1:15: unexpected token "/" (expected "}")
```
{{< /callout >}}

Let's see some examples:

{{< tabs >}}
{{< tab name="Code" >}}
```go
package main

import (
	"fmt"
	"strings"

	"github.com/flamego/flamego"
)

func main() {
	f := flamego.New()
	f.Get("/users/{name: /[a-zA-Z0-9]+/}",
		func(c flamego.Context) string {
			return fmt.Sprintf("The user is %s", c.Param("name"))
		},
	)
	f.Get("/posts/{year: /[0-9]{4}/}-{month: /[0-9]{2}/}-{day: /[0-9]{2}/}.html",
		func(c flamego.Context) string {
			return fmt.Sprintf(
				"The post date is %d-%d-%d",
				c.ParamInt("year"), c.ParamInt("month"), c.ParamInt("day"),
			)
		},
	)
	f.Get("/geo/{state: /[A-Z]{2}/}/{city}",
		func(c flamego.Context) string {
			return fmt.Sprintf(
				"Welcome to %s, %s!",
				strings.Title(c.Param("city")),
				c.Param("state"),
			)
		},
	)
	f.Run()
}
```
{{< /tab >}}
{{< tab name="Test" >}}
```
$ curl http://localhost:2830/users/joe
The user is joe

$ curl http://localhost:2830/posts/2021-11-26.html
The post date is 2021-11-26

$ curl http://localhost:2830/geo/MA/boston
Welcome to Boston, MA!
```
{{< /tab >}}
{{< /tabs >}}

{{< callout type="info" >}}
Try doing following test requests and see what changes:

```
$ curl http://localhost:2830/users/logan-smith
$ curl http://localhost:2830/posts/2021-11-abc.html
$ curl http://localhost:2830/geo/ma/boston
```
{{< /callout >}}

### Globs

A bind parameter can be defined with globs to capture characters across URL path segments (including forward slashes). The only notation for the globs is `**` and allows an optional argument `capture` to define how many URL path segments to capture _at most_.

Below are all valid usages of bind parameters with globs:

```go
f.Get("/posts/{**}", ...) // A shorthand for "{**: **}"
f.Get("/webhooks/{repo: **}/events", ...)
f.Get("/geo/{**: **, capture: 2}", ...)
```

On line 1, the glob captures everything under the `/posts/` path.

On line 2, the glob captures everything in between a path starts with `/webhooks/` and ends with `/events`.

On line 3, the glob captures at most two URL path segments under the `/geo/` path.

Let's see some examples:

{{< tabs >}}
{{< tab name="Code" >}}
```go
package main

import (
	"fmt"
	"strings"

	"github.com/flamego/flamego"
)

func main() {
	f := flamego.New()
	f.Get("/posts/{**}",
		func(c flamego.Context) string {
			return fmt.Sprintf("The post is %s", c.Param("**"))
		},
	)
	f.Get("/webhooks/{repo: **}/events",
		func(c flamego.Context) string {
			return fmt.Sprintf("The event is for %s", c.Param("repo"))
		},
	)
	f.Get("/geo/{**: **, capture: 2}",
		func(c flamego.Context) string {
			fields := strings.Split(c.Param("**"), "/")
			return fmt.Sprintf(
				"Welcome to %s, %s!",
				strings.Title(fields[1]),
				strings.ToUpper(fields[0]),
			)
		},
	)
	f.Run()
}
```
{{< /tab >}}
{{< tab name="Test" >}}
```
$ curl http://localhost:2830/posts/2021/11/26.html
The post is 2021-11-26.html

$ curl http://localhost:2830/webhooks/flamego/flamego/events
The event is for flamego/flamego

$ curl http://localhost:2830/geo/ma/boston
Welcome to Boston, MA!
```
{{< /tab >}}
{{< /tabs >}}

{{< callout type="info" >}}
Try doing following test requests and see what changes:

```
$ curl http://localhost:2830/webhooks/flamego/flamego
$ curl http://localhost:2830/geo/ma/boston/02125
```
{{< /callout >}}

#### Multiple globs per route

{{< callout type="info" >}}
**🆕 Available in v1.10.0**

{{< /callout >}}

A single route may contain more than one glob. Two globs in the same route must be separated by either:

- a static or regex segment, or
- a capture limit on the earlier glob, which bounds how many segments it can consume.

Static segments pin the path text exactly. Regex segments are accepted as separators regardless of how broadly the regex matches — a tight pattern like `/[0-9]+/` gives a real disambiguation point, while a permissive pattern like `/.+/` does not, but the regex is taken as your explicit opt-in to that route shape and the resulting bindings follow the normal [matching priority](#matching-priority).

Placeholder segments (`{name}`) do **not** count as separators because they accept any one segment of any content with no opt-in, which would leave the split between the surrounding globs silently ambiguous.

Below are valid multi-glob routes:

```go
// Static separator between unbounded globs.
f.Get("/files/{prefix: **}/blob/{path: **}", ...)

// Regex separator between unbounded globs.
f.Get("/repos/{owner: **}/{id: /[0-9]+/}/{path: **}", ...)

// Capture limit on the earlier glob.
f.Get("/archive/{head: **, capture: 2}/{tail: **}", ...)

// Three globs with mixed separators.
f.Get("/api/{a: **, capture: 2}/sep/{b: **, capture: 2}/end/{c: **}", ...)
```

These routes are rejected at registration:

```go
// Two unbounded globs with no separator.
f.Get("/api/{a: **}/{b: **}", ...)

// Placeholder is not an anchor — the split is ambiguous.
f.Get("/api/{a: **}/{id}/{b: **}", ...)

// A capture limit on the *later* glob does not help, since the earlier
// (unbounded) glob is what consumes path segments first.
f.Get("/api/{a: **}/{b: **, capture: 2}", ...)
```

For a non-final bounded glob, matching grows the captured segment up to the capture limit and prefers the longest partition that lets the rest of the route match. If a partition matches through a more-specific sibling (static, regex, or placeholder), that match wins immediately even when a longer partition would also succeed through a glob sibling. In other words, the documented [matching priority](#matching-priority) outranks partition length.

## Combo routes

The `Combo` method can create combo routes when you have different handlers for different HTTP methods of the same route:

```go
f.Combo("/").Get(...).Post(...)
```

## Group routes

Organizing routes in groups not only help code readability, but also encourages code reuse in terms of shared middleware.

It is as easy as wrapping your routes with the `Group` method, and there is no limit on how many level of nestings you may have:

```go {hl_lines=["4"] linenostart=1}
f.Group("/user", func() {
    f.Get("/info", ...)
    f.Group("/settings", func() {
        f.Get("", ...)
        f.Get("/account_security", ...)
    }, middleware3)
}, middleware1, middleware2)
```

The line 4 in the above example may seem unusual to you, but that is not a mistake! The equivalent version of it is as follows:

```go
f.Get("/user/settings", ...)
```

![how does that work](https://media0.giphy.com/media/2gUHK3J2NCMsqsz1su/200w.webp?cid=ecf05e47d3syetfd9ja7nr3qwjfdrs4mnhjh46xq1numt01p&rid=200w.webp&ct=g)

That's because the Flamego router uses [string concatenation to combine group routes](https://github.com/flamego/flamego/blob/503ddd0f43a7281b5d334173aba5e32de2d2b31f/router.go#L201-L205).

Huh, so this also works?

```go {hl_lines=["3-5"] linenostart=1}
f.Group("/user", func() {
    f.Get("/info", ...)
    f.Group("/sett", func() {
        f.Get("ings", ...)
        f.Get("ings/account_security", ...)
    }, middleware3)
}, middleware1, middleware2)
```

Yes!

## Optional routes

Optional routes may be used for both static and dynamic routes, and use question mark (`?`) as the notation:

```go
f.Get("/user/?settings", ...)
f.Get("/users/?{name}", ...)
```

The above example is essentially a shorthand for the following:

```go
f.Get("/user", ...)
f.Get("/user/settings", ...)
f.Get("/users", ...)
f.Get("/users/{name}", ...)
```

{{< callout type="warning" >}}
The optional routes can only be used for the last URL path segment.
{{< /callout >}}

## Matching headers

{{< callout type="info" >}}
**🆕 Available in v1.5.0**

{{< /callout >}}

In addition to match a request path, you may also configure request headers to be matched for a given route:

```go
f.Get("/", ...).Headers(
	"User-Agent", "Chrome",   // Loose match
	"Host", "^flamego\.dev$", // Exact match
	"Cache-Control", "",      // As long as "Cache-Control" is not empty
)
```

The `Headers` method accepts key-value pairs as the list of matching criteria for request headers, where key is the header name and value is a regex.

When a route fails on matching request headers, the Flame instance continues to match other routes instead of halting the route matching process.

## Matching custom predicates

{{< callout type="info" >}}
**🆕 Available in v1.11.0**

{{< /callout >}}

When path and `Headers` matching are not enough, you may attach an arbitrary predicate to a route with `Match`:

```go
f.Get("/admin", ...).Match(func(r *http.Request) bool {
	return strings.HasPrefix(r.RemoteAddr, "10.")
})
```

The `Match` method accepts a `func(*http.Request) bool`. The predicate is evaluated only after the request path (and any `Headers` matchers) match. If it returns false, the request falls through to the next candidate route just like a failed `Headers` match — it does not halt the route matching process.

Multiple `Match` calls on the same route accumulate and are combined with AND: every predicate must return true for the route to match. When combined with `Headers`, both must pass:

```go
f.Get("/", ...).
	Headers("Accept", "application/json").
	Match(func(r *http.Request) bool { return r.TLS != nil })
```

The predicate does not affect the matching priority of the route, nor does it participate in `URLPath` construction. Passing a nil function to `Match` panics at registration.

## Matching priority

When your web application grows large enough, you'll start to want to make sense of which route gets matched at when. This is where the matching priority comes into play.

The matching priority is based on different URL-matching patterns, the matching scope (the narrower scope has the higher priority), and the order of registration.

Here is the breakdown:

1. Static routes are always being matched first, e.g. `/users/settings`.
1. Dynamic routes with placeholders not capturing everything, e.g. `/users/{name}.html`
1. Dynamic routes with single placeholder captures everything, e.g. `/users/{name}`.
1. Dynamic routes with globs in the middle, e.g. `/users/{**}/events`.
1. Dynamic routes with globs in the end, e.g. `/users/{**}`.

## Constructing URL paths

The URL path can be constructed using the `URLPath` method if you give the corresponding route a name, which helps prevent URL paths are getting out of sync spread across your codebase:

```go
f.Get("/user/?settings", ...).Name("UserSettings")
f.Get("/users/{name}", ...).Name("UsersName")

f.Get(..., func(c flamego.Context) {
   c.URLPath("UserSettings")                         // => /user
   c.URLPath("UserSettings", "withOptional", "true") // => /user/settings
   c.URLPath("UsersName", "name", "joe")             // => /users/joe
})
```

## Customizing the `NotFound` handler

By default, the [`http.NotFound`](https://pkg.go.dev/net/http#NotFound) is invoked for 404 pages, you can customize the behavior using the `NotFound` method:

```go
f.NotFound(func() string {
    return "This is a cool 404 page"
})
```

## Auto-registering `HEAD` method

By default, only GET requests is accepted when using the `Get` method to register a route, but it is not uncommon to allow HEAD requests to your web application.

The `AutoHead` method can automatically register HEAD method with same chain of handlers to the same route whenever a GET method is registered:

```go
f.Get("/without-head", ...)
f.AutoHead(true)
f.Get("/with-head", ...)
```

Please note that only routes that are registered after call of the `AutoHead(true)` method will be affected, existing routes remain unchanged.

In the above example, only GET requests are accepted for the `/without-head` path. Both GET and HEAD requests are accepted for the `/with-head` path.