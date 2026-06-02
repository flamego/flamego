---
title: hcaptcha
weight: 120
---
hcaptcha 中间件为 [Flame 实例](/core-concepts#实例)提供 [hCaptcha](https://www.hcaptcha.com/) 验证服务的集成。

你可以在 [GitHub](https://github.com/flamego/hcaptcha) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/hcaptcha?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/hcaptcha
```

## 用法示例

{{< callout type="warning" >}}
本小结仅展示 hcaptcha 中间件的相关用法，示例中的用户验证方案绝不可以直接被用于生产环境。
{{< /callout >}}

[`hcaptcha.Captcha`](https://pkg.go.dev/github.com/flamego/hcaptcha#Captcha) 可以配合 [`hcaptcha.Options`](https://pkg.go.dev/github.com/flamego/hcaptcha#Options) 对中间件进行配置，并使用 `hcaptcha.HCaptcha.Verify` 来进行验证码的校验：

{{< tabs >}}
{{< tab name="main.go" >}}
```go {hl_lines=["21", "27"] linenostart=1}
package main

import (
	"fmt"
	"net/http"

	"github.com/flamego/flamego"
	"github.com/flamego/hcaptcha"
	"github.com/flamego/template"
)

func main() {
	f := flamego.Classic()
	f.Use(template.Templater())
	f.Use(hcaptcha.Captcha(
		hcaptcha.Options{
			Secret: "<SECRET>",
		},
	))
	f.Get("/", func(t template.Template, data template.Data) {
		data["SiteKey"] = "<SITE KEY>"
		t.HTML(http.StatusOK, "home")
	})

	f.Post("/", func(w http.ResponseWriter, r *http.Request, h hcaptcha.HCaptcha) {
		token := r.PostFormValue("h-captcha-response")
		resp, err := h.Verify(token)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		} else if !resp.Success {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("Verification failed, error codes %v", resp.ErrorCodes)))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Verified!"))
	})

	f.Run()
}
```
{{< /tab >}}
{{< tab name="templates/home.tmpl" >}}
```html
<html>
<head>
  <script src="https://hcaptcha.com/1/api.js"></script>
</head>
<body>
  <form method="POST">
    <div class="h-captcha" data-sitekey="{{.SiteKey}}"></div>
    <input type="submit" name="button" value="Submit">
  </form>
</body>
</html>
```
{{< /tab >}}
{{< /tabs >}}

下图为程序运行时浏览器中所展示的内容：

![Form with hCaptcha](https://user-images.githubusercontent.com/2946214/158646590-6e58234f-70ae-4afa-a9f4-b69ffaa5c04f.png)