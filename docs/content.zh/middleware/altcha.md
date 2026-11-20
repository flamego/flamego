---
title: altcha
weight: 105
---
altcha 中间件为 [Flame 实例](/core-concepts#实例)提供 [ALTCHA](https://altcha.org) v3 格式的自托管工作量证明验证码。浏览器花一点时间寻找一个数字，服务端用两次 HMAC 校验答案。不依赖第三方服务，不使用 Cookie，也不做任何追踪。

你可以在 [GitHub](https://github.com/flamego/altcha) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/altcha?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/altcha
```

## 用法示例

{{< callout type="warning" >}}
本小结仅展示 altcha 中间件的相关用法，示例中的用户验证方案绝不可以直接被用于生产环境。
{{< /callout >}}

[`altcha.Challenger`](https://pkg.go.dev/github.com/flamego/altcha#Challenger) 可以配合 [`altcha.Options`](https://pkg.go.dev/github.com/flamego/altcha#Options) 对中间件进行配置，使用 `altcha.Challenge.HTML` 渲染组件，并使用 `altcha.Challenge.Verify` 校验组件写入名为 `altcha` 的表单字段中的答案：

{{< tabs >}}
{{< tab name="main.go" >}}
```go {hl_lines=["15", "21", "25"] linenostart=1}
package main

import (
	"net/http"
	"os"

	"github.com/flamego/altcha"
	"github.com/flamego/flamego"
	"github.com/flamego/template"
)

func main() {
	f := flamego.Classic()
	f.Use(template.Templater())
	f.Use(altcha.Challenger(
		altcha.Options{
			Secret: []byte(os.Getenv("ALTCHA_SECRET")),
		},
	))
	f.Get("/", func(t template.Template, data template.Data, c altcha.Challenge) {
		data["AltchaHTML"] = c.HTML()
		t.HTML(http.StatusOK, "home")
	})
	f.Post("/", func(w http.ResponseWriter, r *http.Request, c altcha.Challenge) {
		if err := c.Verify(r.PostFormValue(altcha.FieldName)); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
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
<form method="POST">
  {{.AltchaHTML}}
  <button>Submit</button>
</form>
```
{{< /tab >}}
{{< /tabs >}}

下图为程序运行时浏览器中所展示的内容：

![Form with altcha](/imgs/altcha-form.png)

`Secret` 必须是至少 16 字节的随机数据（例如由 `openssl rand -hex 32` 生成），并且所有校验答案的服务端都要使用相同的值。

中间件会在 `URLPrefix`（默认为 `/.altcha/`）下提供 `challenge.json`、`widget.js` 和 `worker.js`。用户点击复选框，或在此之前提交表单时，组件会获取一个挑战，在 Web Worker 中求解，并将答案写入表单。每个答案只能使用一次，因此下一次提交前组件会重新求解。

`altcha.Challenge.Verify` 在答案正确时返回 `nil`，否则返回以下错误之一：

- `altcha.ErrInvalid`：内容格式错误、并非由该服务端签发，或未能解出对应的挑战。
- `altcha.ErrExpired`：挑战的时长超过了 `Expiry`（默认为 5 分钟）。
- `altcha.ErrReplayed`：该答案已经被接受过一次。

### 难度

浏览器平均需要尝试 `MaxCounter / 2` 个计数值（默认为 `500`），每次尝试执行 `Cost` 次 PBKDF2 迭代（默认为 `5000`）。调高任意一项都会让每次求解的成本变高，对用户和机器人都是如此，因此调高之前请在访客实际使用的设备上进行测量：

```go
altcha.Options{
	Secret:     secret,
	Cost:       5000,
	MaxCounter: 200, // 调低更容易，调高更困难
}
```

无论难度如何，校验答案对服务端的开销都很小，但每签发一个挑战都需要执行一次 `Cost` 次迭代的 PBKDF2。任何人都可以在不求解的情况下请求 `challenge.json`，因此请按客户端对该路径进行限流。

{{< callout type="info" >}}
请只使用 `f.Use` 挂载一个 `altcha.Challenger`。挑战只能在签发它的 `Cost` 和 `URLPrefix` 下通过校验。
{{< /callout >}}

### 运行多个服务端

默认存储将已接受的答案保存在内存中，因此一个进程接受过的答案可能会被另一个进程再次接受。当有多个进程校验答案时，需要为它们配置相同的 `Secret`，并将共享的 [`altcha.ReplayStore`](https://pkg.go.dev/github.com/flamego/altcha#ReplayStore) 设置为 `Store`。它的 `Use` 方法必须以原子操作检查并标记键，例如使用 Redis 的 `SET key 1 NX EXAT expiresAt`。

### 翻译组件文字

组件在每个状态下显示一段简短的文字，默认为英文。可以向 `altcha.Challenge.HTML` 传入 [`altcha.Messages`](https://pkg.go.dev/github.com/flamego/altcha#Messages) 进行替换，例如使用 [i18n](/middleware/i18n) 中间件提供的翻译。留空的字段会继续使用英文默认值：

```go
f.Get("/", func(t template.Template, data template.Data, c altcha.Challenge, l i18n.Locale) {
	data["AltchaHTML"] = c.HTML(altcha.Messages{
		Idle:      l.Translate("altcha.idle"),
		Verifying: l.Translate("altcha.verifying"),
		Verified:  l.Translate("altcha.verified"),
		Error:     l.Translate("altcha.error"),
	})
	t.HTML(http.StatusOK, "home")
})
```

### 自定义组件样式

可以通过以下三种方式修改组件的外观：

1. 覆盖内置规则。所有内置规则都包裹在 `:where()` 中，因此你的任何规则都会生效，无需使用 `!important`。
1. 在 `.flamego-altcha` 上使用 CSS 自定义属性设置主题，例如 `--altcha-color`、`--altcha-background`、`--altcha-border`、`--altcha-radius`、`--altcha-accent`、`--altcha-success` 和 `--altcha-error`。
1. 编写自己的标记。添加 `data-unstyled` 以关闭内置规则，并在容器中放置任意带有 `altcha-check` 类的元素作为控件。容器的 `data-state` 的值为 `idle`、`verifying`、`verified` 或 `error`：

```html
<div class="flamego-altcha" data-unstyled data-challenge-url="/.altcha/challenge.json">
  <button type="button" class="altcha-check"><span class="altcha-label">证明你是人类</span></button>
</div>
<script src="/.altcha/widget.js" defer></script>
```

### 使用官方 ALTCHA 组件

官方组件（npm 上的 `altcha` 3.x）无需任何修改即可配合该中间件使用。将其指向挑战地址，而不是渲染 `altcha.Challenge.HTML`，它会将答案写入同一个 `altcha` 字段：

```html
<script type="module" src="https://cdn.jsdelivr.net/npm/altcha@3/dist/main/altcha.min.js"></script>
<form method="POST">
  <altcha-widget challenge="/.altcha/challenge.json"></altcha-widget>
  <button>Submit</button>
</form>
```

{{< callout type="warning" >}}
仅支持 v3 格式。v2 组件和其他 v1 客户端无法求解这些挑战。
{{< /callout >}}
