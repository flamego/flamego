---
title: template
weight: 10
---
The template middleware provides HTML rendering using [Go template](https://pkg.go.dev/html/template) for [Flame instances](/core-concepts#instances).

You can read source code of this middleware on [GitHub](https://github.com/flamego/template) and API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/flamego/template?tab=doc).

## Installation

```
go get github.com/flamego/template
```

## Usage examples

{{< callout type="info" >}}
Examples included in this section is to demonstrate the usage of the template middleware, please refer to the documentation of [`html/template`](https://pkg.go.dev/html/template) package for syntax and constraints.
{{< /callout >}}

The [`template.Templater`](https://pkg.go.dev/github.com/flamego/template#Templater) works out-of-the-box with an optional [`template.Options`](https://pkg.go.dev/github.com/flamego/template#Options).

By default, the templates files should resides in the "templates" directory and has extension of either `.html` or `.tmpl`. The special data type [`template.Data`](https://pkg.go.dev/github.com/flamego/template#Data) is provided as container to store any data you would want to be used in rendering the template.

### Using local files

{{< tabs >}}
{{< tab name="main.go" >}}
```go
package main

import (
	"net/http"

	"github.com/flamego/flamego"
	"github.com/flamego/template"
)

func main() {
	f := flamego.Classic()
	f.Use(template.Templater())

	type Book struct {
		Name   string
		Author string
	}
	f.Get("/", func(t template.Template, data template.Data) {
		data["Name"] = "Joe"
		data["Books"] = []*Book{
			{
				Name:   "Designing Data-Intensive Applications",
				Author: "Martin Kleppmann",
			},
			{
				Name:   "Shape Up",
				Author: "Basecamp",
			},
		}
		t.HTML(http.StatusOK, "home")
	})
	f.Run()
}
```
{{< /tab >}}
{{< tab name="templates/home.tmpl" >}}
```html
<p>
  Hello, <b>{{.Name}}</b>!
</p>
<p>
  Books to read:
  <ul>
    {{range .Books}}
      <li><i>{{.Name}}</i> by {{.Author}}</li>
    {{end}}
  </ul>
</p>
```
{{< /tab >}}
{{< /tabs >}}

### Using the `embed.FS`

The [`template.EmbedFS`](https://pkg.go.dev/github.com/flamego/template#EmbedFS) is a handy helper to convert your `embed.FS` into the [`template.FileSystem`](https://pkg.go.dev/github.com/flamego/template#FileSystem).

{{< tabs >}}
{{< tab name="Directory" >}}
```
$ tree .
.
├── templates
│   ├── embed.go
│   ├── home.tmpl
├── go.mod
├── go.sum
└── main.go
```
{{< /tab >}}
{{< tab name="main.go" >}}
```go
package main

import (
	"net/http"

	"github.com/flamego/flamego"
	"github.com/flamego/template"

	"main/templates"
)

func main() {
	f := flamego.Classic()

	fs, err := template.EmbedFS(templates.Templates, ".", []string{".tmpl"})
	if err != nil {
		panic(err)
	}
	f.Use(template.Templater(
		template.Options{
			FileSystem: fs,
		},
	))

	type Book struct {
		Name   string
		Author string
	}
	f.Get("/", func(t template.Template, data template.Data) {
		data["Name"] = "Joe"
		data["Books"] = []*Book{
			{
				Name:   "Designing Data-Intensive Applications",
				Author: "Martin Kleppmann",
			},
			{
				Name:   "Shape Up",
				Author: "Basecamp",
			},
		}
		t.HTML(http.StatusOK, "home")
	})
	f.Run()
}
```
{{< /tab >}}
{{< tab name="embed.go" >}}
```go
package templates

import "embed"

// Append "**/*" if you also have template files in subdirectories
//go:embed *.tmpl
var Templates embed.FS
```
{{< /tab >}}
{{< tab name="home.tmpl" >}}
```html
<p>
  Hello, <b>{{.Name}}</b>!
</p>
<p>
  Books to read:
  <ul>
    {{range .Books}}
      <li><i>{{.Name}}</i> by {{.Author}}</li>
    {{end}}
  </ul>
</p>
```
{{< /tab >}}
{{< /tabs >}}

## Template caching

When your application is running with `flamego.EnvTypeDev` (default) or `flamego.EnvTypeTest`, template files are reloaded and recomplied upon every client request.

Set the [Env](/core-concepts#env) to `flamego.EnvTypeProd` to enable template caching in production.