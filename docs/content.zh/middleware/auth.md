---
title: auth
weight: 100
---
auth 中间件为 [Flame 实例](../core-concepts#实例)提供基于 HTTP Basic 和 Bearer 形式的请求认证服务。

你可以在 [GitHub](https://github.com/flamego/auth) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/auth?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/auth
```

## 用法示例

### Basic 认证

[`auth.Basic`](https://pkg.go.dev/github.com/flamego/auth#Basic) 支持基于一组静态的用户名和密码组合对请求进行认证，并在认证成功后将包含用户名的 [`auth.User`](https://pkg.go.dev/github.com/flamego/auth#User) 注入到请求上下文中：

```go
package main

import (
	"github.com/flamego/auth"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()
	f.Use(auth.Basic("username", "secretpassword"))
	f.Get("/", func(user auth.User) string {
		return "Welcome, " + string(user)
	})
	f.Run()
}
```

也可以使用 [`auth.BasicFunc`](https://pkg.go.dev/github.com/flamego/auth#BasicFunc) 支持基于动态的用户名和密码组合：

```go {hl_lines=["16"] linenostart=1}
package main

import (
	"github.com/flamego/auth"
	"github.com/flamego/flamego"
)

func main() {
	credentials := map[string]string{
		"alice": "pa$$word",
		"bob":   "secretpassword",
	}

	f := flamego.Classic()
	f.Use(auth.BasicFunc(func(username, password string) bool {
		return auth.SecureCompare(credentials[username], password)
	}))
	f.Get("/", func(user auth.User) string {
		return "Welcome, " + string(user)
	})
	f.Run()
}
```

使用 [`auth.SecureCompare`](https://pkg.go.dev/github.com/flamego/auth#SecureCompare) 对字符串进行比较可以预防基于时间的对比攻击。

### Bearer 认证

[`auth.Bearer`](https://pkg.go.dev/github.com/flamego/auth#Bearer) 支持基于静态令牌对请求进行认证，并在认证成功后将包含令牌的 [`auth.Token`](https://pkg.go.dev/github.com/flamego/auth#Token) 注入到请求上下文中：

```go
package main

import (
	"github.com/flamego/auth"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()
	f.Use(auth.Bearer("secrettoken"))
	f.Get("/", func(token auth.Token) string {
		return "Authenticated through " + string(token)
	})
	f.Run()
}
```

也可以使用 [`auth.BearerFunc`](https://pkg.go.dev/github.com/flamego/auth#BearerFunc) 支持基于动态令牌的认证：

```go
package main

import (
	"github.com/flamego/auth"
	"github.com/flamego/flamego"
)

func main() {
	tokens := map[string]struct{}{
		"token":       {},
		"secrettoken": {},
	}

	f := flamego.Classic()
	f.Use(auth.BearerFunc(func(token string) bool {
		_, ok := tokens[token]
		return ok
	}))
	f.Get("/", func(token auth.Token) string {
		return "Authenticated through " + string(token)
	})
	f.Run()
}
```