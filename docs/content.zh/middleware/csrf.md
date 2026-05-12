---
title: csrf
weight: 50
---
csrf 中间件为 [Flame 实例](../core-concepts#实例)提供 CSRF 令牌的生成和验证服务，该中间件依赖于 [session](session) 中间件。

你可以在 [GitHub](https://github.com/flamego/csrf) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/csrf?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/csrf
```

## 用法示例

{{< callout type="warning" >}}
本小结仅展示 csrf 中间件的相关用法，示例中的用户验证方案绝不可以直接被用于生产环境。
{{< /callout >}}

[`csrf.Csrfer`](https://pkg.go.dev/github.com/flamego/csrf#Csrfer) 可以配合 [`csrf.Options`](https://pkg.go.dev/github.com/flamego/csrf#Options) 对中间件进行配置，并使用 [`csrf.Validate`](https://pkg.go.dev/github.com/flamego/csrf#Validate) 来进行 CSRF 令牌的验证：

{{< tabs >}}
{{< tab name="main.go" >}}
```go {hl_lines=["41-42", "46-47"] linenostart=1}
package main

import (
	"net/http"

	"github.com/flamego/csrf"
	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/flamego/template"
)

func main() {
	f := flamego.Classic()
	f.Use(template.Templater())
	f.Use(session.Sessioner())
	f.Use(csrf.Csrfer())

	// 模拟会话认证，若 userID 存在则重定向到包含 CSRF 令牌的表单页面
	f.Get("/", func(c flamego.Context, s session.Session) {
		if s.Get("userID") == nil {
			c.Redirect("/login")
			return
		}
		c.Redirect("/protected")
	})

	// 设置会话的 uid
	f.Get("/login", func(c flamego.Context, s session.Session) {
		s.Set("userID", 123)
		c.Redirect("/")
	})

	// 使用 x.Token() 将 CSRF 令牌渲染到模板中
	f.Get("/protected", func(c flamego.Context, s session.Session, x csrf.CSRF, t template.Template, data template.Data) {
		if s.Get("userID") == nil {
			c.Redirect("/login", http.StatusUnauthorized)
			return
		}

		// 传递令牌到被渲染的模板
		data["CSRFToken"] = x.Token()
		t.HTML(http.StatusOK, "protected")
	})

	// 应用 CSRF 验证
	f.Post("/protected", csrf.Validate, func(c flamego.Context, s session.Session, t template.Template) {
		if s.Get("userID") != nil {
			c.ResponseWriter().Write([]byte("You submitted with a valid CSRF token"))
			return
		}
		c.Redirect("/login", http.StatusUnauthorized)
	})

	f.Run()
}
```
{{< /tab >}}
{{< tab name="templates/protected.tmpl" >}}
```html
<form action="/protected" method="POST">
  <input type="hidden" name="_csrf" value="{{.CSRFToken}}">
  <button>Submit</button>
</form>
```
{{< /tab >}}
{{< /tabs >}}