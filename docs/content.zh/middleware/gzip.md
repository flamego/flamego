---
title: gzip
weight: 70
---
gzip 中间件为 [Flame 实例](/core-concepts#实例)提供基于 Gzip 的响应流压缩服务。

你可以在 [GitHub](https://github.com/flamego/gzip) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/gzip?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/gzip
```

## 用法示例

[`gzip.Gzip`](https://pkg.go.dev/github.com/flamego/gzip#Gzip) 需要在 _其它任何可能写入内容到响应流的中间件之前_ 被注册：

```go
package main

import (
	"github.com/flamego/flamego"
	"github.com/flamego/gzip"
)

func main() {
	f := flamego.Classic()
	f.Use(gzip.Gzip())
	f.Get("/", func() string {
		return "Hello, Gzip!"
	})
	f.Run()
}
```

[`gzip.Options`](https://pkg.go.dev/github.com/flamego/gzip#Options) 可以被用于配置该中间件的行为：

```go {hl_lines=["11-13"] linenostart=1}
package main

import (
	"github.com/flamego/flamego"
	"github.com/flamego/gzip"
)

func main() {
	f := flamego.Classic()
	f.Use(gzip.Gzip(
		gzip.Options{
			CompressionLevel: 9, // 最优压缩
		},
	))
	f.Get("/", func() string {
		return "Hello, Gzip!"
	})
	f.Run()
}
```