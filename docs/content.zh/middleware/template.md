---
title: template
weight: 10
---
template 中间件为 [Flame 实例](/core-concepts#实例)提供基于 [Go 模板引擎](https://pkg.go.dev/html/template)的 HTML 渲染服务。

你可以在 [GitHub](https://github.com/flamego/template) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/template?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/template
```

## 用法示例

{{< callout type="info" >}}
本小结仅展示 template 中间件的相关用法，如需了解模板引擎的用法请移步 [`html/template`](https://pkg.go.dev/html/template) 的文档。
{{< /callout >}}

[`template.Templater`](https://pkg.go.dev/github.com/flamego/template#Templater) 可以配合 [`template.Options`](https://pkg.go.dev/github.com/flamego/template#Options) 对中间件进行配置。

默认情况下，模板文件都需要被存放在 "templates" 目录内，并以 `.html` 或 `.tmpl` 作为文件名后缀。[`template.Data`](https://pkg.go.dev/github.com/flamego/template#Data) 是用于渲染模板的数据容器，即根对象。

### 使用本地文件

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

### 使用 `embed.FS`

[`template.EmbedFS`](https://pkg.go.dev/github.com/flamego/template#EmbedFS) 是用于将 `embed.FS` 转换为 [`template.FileSystem`](https://pkg.go.dev/github.com/flamego/template#FileSystem) 的辅助函数。

{{< tabs >}}
{{< tab name="目录" >}}
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

// 如果需要包含子目录中的模板文件则可以追加规则 "**/*"
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

## 模板缓存

当你的应用运行环境为 `flamego.EnvTypeDev`（默认运行环境）或 `flamego.EnvTypeTest` 时，每次响应客户端的请求都会对模板文件进行重新构建，便于开发调试。

通过 [Env](/core-concepts#运行环境) 函数将运行环境设置为 `flamego.EnvTypeProd` 可以启用模板缓存功能。