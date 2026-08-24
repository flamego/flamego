---
title: sse
weight: 140
---
sse 中间件为 [Flame 实例](/core-concepts#实例)提供[服务器发送事件（Server-Sent Events）](https://developer.mozilla.org/zh-CN/docs/Web/API/Server-sent_events)，用于通过 HTTP 连接向 Web 客户端发送更新。

你可以在 [GitHub](https://github.com/flamego/sse) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/sse?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/sse
```

## 用法示例

[`sse.Bind`](https://pkg.go.dev/github.com/flamego/sse#Bind) 中间件会将响应配置为事件流，并向后续处理器注入一个带类型的只发送通道。请向 `Bind` 传递非指针值；传递 `T` 类型的值会使处理器可以获取 `chan<- *T` 类型的通道。通过该通道发送的值会被编码为 JSON 事件数据。

以下示例每秒向每个已连接的客户端发送一次当前时间：

```go
package main

import (
	"time"

	"github.com/flamego/flamego"
	"github.com/flamego/sse"
)

type update struct {
	Time time.Time `json:"time"`
}

func main() {
	f := flamego.Classic()
	f.Get("/events", sse.Bind(update{}), func(c flamego.Context, events chan<- *update) {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.Request().Context().Done():
				return
			case now := <-ticker.C:
				select {
				case events <- &update{Time: now}:
				case <-c.Request().Context().Done():
					return
				}
			}
		}
	})
	f.Run()
}
```

客户端可以使用浏览器的 [`EventSource`](https://developer.mozilla.org/zh-CN/docs/Web/API/EventSource) API 接收事件流：

```html
<p id="time"></p>
<script>
  const events = new EventSource("/events");
  events.onmessage = (event) => {
    const update = JSON.parse(event.data);
    document.querySelector("#time").textContent = update.time;
  };
</script>
```

路由处理器会在连接的整个生命周期内保持运行。请始终监听 `c.Request().Context().Done()`，并在客户端断开连接时停止计时器或释放其它资源。

该中间件会定期发送 ping 以保持连接。默认间隔为 10 秒，可以通过 [`sse.Options`](https://pkg.go.dev/github.com/flamego/sse#Options) 修改：

```go
f.Get("/events",
	sse.Bind(update{}, sse.Options{
		PingInterval: 30 * time.Second,
	}),
	func(c flamego.Context, events chan<- *update) {
		// 在请求上下文被取消之前持续生成事件。
	},
)
```
