---
title: captcha
weight: 130
---
captcha 中间件为 [Flame 实例](../core-concepts#实例)提供验证码生成和验证服务，该中间件依赖于 [session](session) 中间件。

你可以在 [GitHub](https://github.com/flamego/captcha) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/captcha?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/captcha
```

## 用法示例

{{< callout type="warning" >}}
本小结仅展示 captcha 中间件的相关用法，示例中的用户验证方案绝不可以直接被用于生产环境。
{{< /callout >}}

[`captcha.Captchaer`](https://pkg.go.dev/github.com/flamego/captcha#Captchaer) 可以配合 [`captcha.Options`](https://pkg.go.dev/github.com/flamego/captcha#Options) 对中间件进行配置，并使用 `captcha.ValidText` 对验证码结果进行验证：

{{< tabs >}}
{{< tab name="main.go" >}}
```go {hl_lines=["19", "23"] linenostart=1}
package main

import (
	"net/http"

	"github.com/flamego/captcha"
	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/flamego/template"
)

func main() {
	f := flamego.Classic()
	f.Use(template.Templater())
	f.Use(session.Sessioner())
	f.Use(captcha.Captchaer())

	f.Get("/", func(t template.Template, data template.Data, captcha captcha.Captcha) {
		data["CaptchaHTML"] = captcha.HTML()
		t.HTML(http.StatusOK, "home")
	})
	f.Post("/", func(c flamego.Context, captcha captcha.Captcha) {
		if !captcha.ValidText(c.Request().FormValue("captcha")) {
			c.ResponseWriter().WriteHeader(http.StatusBadRequest)
			_, _ = c.ResponseWriter().Write([]byte(http.StatusText(http.StatusBadRequest)))
		} else {
			c.ResponseWriter().WriteHeader(http.StatusOK)
			_, _ = c.ResponseWriter().Write([]byte(http.StatusText(http.StatusOK)))
		}
	})

	f.Run()
}
```
{{< /tab >}}
{{< tab name="templates/home.tmpl" >}}
```html
<form method="POST">
  {{.CaptchaHTML}} <br>
  <input name="captcha">
  <button>Submit</button>
</form>
```
{{< /tab >}}
{{< /tabs >}}

下图为程序运行时浏览器中所展示的内容：

![Form with captcha](https://user-images.githubusercontent.com/2946214/158567419-9a9556ad-c9d0-48db-b96a-32b9d4b67045.png)

正如图中提示，当验证码图片无法识别时可以通过鼠标左键点击更换新的图片。