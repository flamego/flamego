---
title: cors
weight: 60
---
cors 中间件为 [Flame 实例](../core-concepts#实例)提供[跨站资源共享](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)服务。

你可以在 [GitHub](https://github.com/flamego/cors) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/cors?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/cors
```

## 用法示例

[`cors.CORS`](https://pkg.go.dev/github.com/flamego/cors#CORS) 开箱即用：

```go
package main

import (
	"github.com/flamego/cors"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()
	f.Get("/",
		cors.CORS(),
		func(c flamego.Context) string {
			return "This endpoint can be shared cross-origin"
		},
	)
	f.Run()
}
```

也可以配合 [`cors.Options`](https://pkg.go.dev/github.com/flamego/cors#Options) 对中间件进行配置：

```go {hl_lines=["12-14"] linenostart=1}
package main

import (
	"github.com/flamego/cors"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()
	f.Get("/",
		cors.CORS(
            cors.Options{
			    AllowDomain: []string{"cors.example.com"},
		    },
        ),
		func(c flamego.Context) string {
			return "This endpoint can be shared cross-origin"
		},
	)
	f.Run()
}
```