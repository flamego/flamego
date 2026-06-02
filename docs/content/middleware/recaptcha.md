---
title: recaptcha
weight: 110
---
The recaptcha middleware provides [Google reCAPTCHA](https://www.google.com/recaptcha/about/) integration for [Flame instances](/core-concepts#instances).

You can read source code of this middleware on [GitHub](https://github.com/flamego/recaptcha) and API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/flamego/recaptcha?tab=doc).

## Installation

```
go get github.com/flamego/recaptcha
```

## Usage examples

{{< callout type="warning" >}}
Examples included in this section is to demonstrate the usage of the recaptcha middleware, by no means illustrates the idiomatic or even correct way of doing user authentication.
{{< /callout >}}

### reCAPTCHA v3

The [`recaptcha.V3`](https://pkg.go.dev/github.com/flamego/recaptcha#V3) is used in combination with [`recaptcha.Options`](https://pkg.go.dev/github.com/flamego/recaptcha#Options) for [reCAPTCHA v3](https://developers.google.com/recaptcha/docs/v3) integration, and the `recaptcha.RecaptchaV3.Verify` should be used to verify response tokens:

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

Because reCAPTCHA v3 is non-interruptive, you will not see any captcha image in the browser.

### reCAPTCHA v2

The [`recaptcha.V2`](https://pkg.go.dev/github.com/flamego/recaptcha#V2) is used in combination with [`recaptcha.Options`](https://pkg.go.dev/github.com/flamego/recaptcha#Options) for [reCAPTCHA v2](https://developers.google.com/recaptcha/docs/display) integration, and the `recaptcha.RecaptchaV2.Verify` should be used to verify response tokens.

The following example is using the **"I'm not a robot" Checkbox** type of challenge:

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

Below is how it would look like in your browser for the above example:

![Form with reCAPTCHA](https://user-images.githubusercontent.com/2946214/158651864-1cd14d53-9a41-496f-a2e3-f3e03ac305d1.png)