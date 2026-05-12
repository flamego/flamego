---
title: recaptcha
weight: 110
---
recaptcha 中间件为 [Flame 实例](../core-concepts#实例)提供 [Google reCAPTCHA](https://www.google.com/recaptcha/about/) 验证服务的集成。

你可以在 [GitHub](https://github.com/flamego/recaptcha) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/recaptcha?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/recaptcha
```

## 用法示例

{{< callout type="warning" >}}
本小结仅展示 recaptcha 中间件的相关用法，示例中的用户验证方案绝不可以直接被用于生产环境。
{{< /callout >}}

### reCAPTCHA v3

[`recaptcha.V3`](https://pkg.go.dev/github.com/flamego/recaptcha#V3) 可以配合 [`recaptcha.Options`](https://pkg.go.dev/github.com/flamego/recaptcha#Options) 用于集成 [reCAPTCHA v3](https://developers.google.com/recaptcha/docs/v3)，并使用 `recaptcha.RecaptchaV3.Verify` 来进行验证码的校验：

{{< tabs >}}
{{< tab name="main.go" >}}
```go {hl_lines=["22", "27"] linenostart=1}
package main

import (
	"fmt"
	"net/http"

	"github.com/flamego/flamego"
	"github.com/flamego/recaptcha"
	"github.com/flamego/template"
)

func main() {
	f := flamego.Classic()
	f.Use(template.Templater())
	f.Use(recaptcha.V3(
		recaptcha.Options{
			Secret:    "<SECRET>",
			VerifyURL: recaptcha.VerifyURLGoogle,
		},
	))
	f.Get("/", func(t template.Template, data template.Data) {
		data["SiteKey"] = "<SITE KEY>"
		t.HTML(http.StatusOK, "home")
	})
	f.Post("/", func(w http.ResponseWriter, r *http.Request, re recaptcha.RecaptchaV3) {
		token := r.PostFormValue("g-recaptcha-response")
		resp, err := re.Verify(token)
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
  <script src="https://www.google.com/recaptcha/api.js"></script>
</head>
<body>
<script>
  function onSubmit(token) {
    document.getElementById("demo-form").submit();
  }
</script>
<form id="demo-form" method="POST">
  <button class="g-recaptcha"
    data-sitekey="{{.SiteKey}}"
    data-callback='onSubmit'
    data-action='submit'>Submit</button>
</form>
</body>
</html>
```
{{< /tab >}}
{{< /tabs >}}

因为 reCAPTCHA v3 采用非侵入式的校验方式，所以不会在网页中看到任何验证码。

### reCAPTCHA v2

[`recaptcha.V2`](https://pkg.go.dev/github.com/flamego/recaptcha#V2) 可以配合 [`recaptcha.Options`](https://pkg.go.dev/github.com/flamego/recaptcha#Options) 用于集成 [reCAPTCHA v2](https://developers.google.com/recaptcha/docs/display)，并使用 `recaptcha.RecaptchaV2.Verify` 来进行验证码的校验。

下面的例子才采用了 **"I'm not a robot" Checkbox** 类型的校验形式：

{{< tabs >}}
{{< tab name="main.go" >}}
```go {hl_lines=["22", "27"] linenostart=1}
package main

import (
	"fmt"
	"net/http"

	"github.com/flamego/flamego"
	"github.com/flamego/recaptcha"
	"github.com/flamego/template"
)

func main() {
	f := flamego.Classic()
	f.Use(template.Templater())
	f.Use(recaptcha.V2(
		recaptcha.Options{
			Secret:    "<SECRET>",
			VerifyURL: recaptcha.VerifyURLGoogle,
		},
	))
	f.Get("/", func(t template.Template, data template.Data) {
		data["SiteKey"] = "<SITE KEY>"
		t.HTML(http.StatusOK, "home")
	})
	f.Post("/", func(w http.ResponseWriter, r *http.Request, re recaptcha.RecaptchaV2) {
		token := r.PostFormValue("g-recaptcha-response")
		resp, err := re.Verify(token)
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
  <script src="https://www.google.com/recaptcha/api.js"></script>
</head>
<body>
  <form method="POST">
    <div class="g-recaptcha" data-sitekey="{{.SiteKey}}"></div>
    <input type="submit" name="button" value="Submit">
  </form>
</body>
</html>
```
{{< /tab >}}
{{< /tabs >}}

下图为程序运行时浏览器中所展示的内容：

![Form with reCAPTCHA](https://user-images.githubusercontent.com/2946214/158651864-1cd14d53-9a41-496f-a2e3-f3e03ac305d1.png)