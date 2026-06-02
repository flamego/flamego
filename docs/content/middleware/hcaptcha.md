---
title: hcaptcha
weight: 120
---
The hcaptcha middleware provides [hCaptcha](https://www.hcaptcha.com/) integration for [Flame instances](/core-concepts#instances).

You can read source code of this middleware on [GitHub](https://github.com/flamego/hcaptcha) and API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/flamego/hcaptcha?tab=doc).

## Installation

```
go get github.com/flamego/hcaptcha
```

## Usage examples

{{< callout type="warning" >}}
Examples included in this section is to demonstrate the usage of the hcaptcha middleware, by no means illustrates the idiomatic or even correct way of doing user authentication.
{{< /callout >}}

The [`hcaptcha.Captcha`](https://pkg.go.dev/github.com/flamego/hcaptcha#Captcha) is used in combination with [`hcaptcha.Options`](https://pkg.go.dev/github.com/flamego/hcaptcha#Options), and the `hcaptcha.HCaptcha.Verify` should be used to verify response tokens:

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

Below is how it would look like in your browser for the above example:

![Form with hCaptcha](https://user-images.githubusercontent.com/2946214/158646590-6e58234f-70ae-4afa-a9f4-b69ffaa5c04f.png)