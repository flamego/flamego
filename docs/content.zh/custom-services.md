---
title: 自定义服务
weight: 40
---
Flamego 提供的[核心服务](/core-services)都很实用，但对于开发复杂的 Web 应用来说显然是远远不够的。届时，你必然会需要开发自己的[中间件](/core-concepts#中间件)来满足应用的实际需求。

## 注入服务

Flame 实例是基于 [`inject.TypeMapper`](https://pkg.go.dev/github.com/flamego/flamego/inject#TypeMapper) 来为处理器提供服务注入功能的，并且内置在了 [`flamego.Flame`](https://pkg.go.dev/github.com/flamego/flamego#Flame) 和 [`flamego.Context`](https://pkg.go.dev/github.com/flamego/flamego#Context) 这两个类型中，以便中间件和处理器进行服务注入的管理和使用。

`Map` 是用于注入服务类型本身的方法，可以是具体的类型（如 [`*log.Logger`](https://pkg.go.dev/log#Logger) ）或接口（如 [`io.Writer`](https://pkg.go.dev/io#Writer)）：

```go
l := log.New(os.Stdout, "[Flamego] ", 0)
f := flamego.New()
f.Map(l)

// 或

var w io.Writer = &bytes.Buffer{}
f := flamego.New()
f.Map(w)
```

`MapTo` 是用于将某个服务类型注入为其实现的某一接口的方法：

```go
buf := &bytes.Buffer{}
f := flamego.New()
f.MapTo(buf, (*io.Writer)(nil))
```

你也可以使用  `MapTo` 方法将一个接口变换为另一个接口：

```go
var w io.ReadCloser = io.NopCloser(&bytes.Buffer{})
f := flamego.New()
f.MapTo(w, (*io.Reader)(nil))
f.MapTo(w, (*io.Closer)(nil))
```

{{< callout type="warning" >}}
`MapTo` 仅实现无脑的类型映射，如果服务的底层类型并没有实现对应的接口则会在运行时发生错误。
{{< /callout >}}

### 全局服务

全局服务直接与整个 Flame 实例而非具体某个路由绑定，且可以被所有路由的处理器使用。

全局服务可以通过调用 Flame 实例的 `Map` 或 `MapTo` 方法完成注入，或通过调用 `Use` 方法在[全局中间件](/core-concepts#中间件)中注入：

```go
db := database.New()
f := flamego.New()
f.Map(db)

// 或

f := flamego.New()
f.Use(func(c flamego.Context) {
    db := database.New()
    c.Map(db)
})
```

### 组级服务

组级服务可以被组内的所有路由的处理器使用，并且只可以通过[组级中间件](/core-concepts#中间件)注入：

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
        // 由于 *database.User 对该路由并不可用，该处理器会发生运行时错误
    })
})
```

上例中，`*database.User` 仅可被用于位于第 3 至 7 行的路由组内，尝试在该路由组外使用它会导致运行时错误（第 14 行）。

### 路由级服务

路由级服务仅可被与该路由绑定的处理器使用，并且只可以通过[路由级中间件](/core-concepts#中间件)注入：

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
    // 由于 *database.User 对该路由并不可用，该处理器会发生运行时错误
})
```

上例中，`*database.User` 仅可被用于位于第 7 至 9 行的路由内，尝试在该路由外使用它会导致运行时错误（第 11 行）。

## 重载服务

你可以通过重载已注入的服务更变服务的状态或行为。

下面展示了如何在路由级重载一个全局服务：

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
				// 处理错误
			}
			return string(p)
		},
	)
	f.Run()
}
```

运行上面的程序并执行 `curl http://localhost:2830/` 后，可以在终端看到如下输出：

```
$ curl http://localhost:2830
this is from a route-level service
```