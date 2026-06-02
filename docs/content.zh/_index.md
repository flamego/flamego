---
title: 
weight: 1
toc: false
linkTitle: Flamego
type: docs
---
![banner](/imgs/banner.png)

Flamego 是一款简洁的、极易扩展且模块化的 Go Web 框架。

作为 [Macaron](https://github.com/go-macaron/macaron) 框架的继任者，Flamego 具备了当下 Go 语言生态中最强大的路由配置语法，没有任何一款 Go Web 框架可以与之比肩。

## 下载安装
```
go get github.com/flamego/flamego
```

## 开始使用

```go
package main

import "github.com/flamego/flamego"

func main() {
	f := flamego.Classic()
	f.Get("/", func() string {
		return "Hello, Flamego!"
	})
	f.Run()
}
```

## 功能说明

- 在所有 Go Web 框架中都[无可匹敌的路由配置语法](/routing)
- [无限路由组合与嵌套](/routing#组路由)
- [在路由的任意位置注入中间件](/core-concepts#中间件)
- [无侵入式地集成到任何已有的 Web 应用程序中](/faqs#如何集成到现有的-web-应用)
- [通过函数签名实现依赖注入](/core-concepts#服务注入)以编写更易测试和维护的代码

## 相关信息

- 如果对 Flamego 还不了解，则可以先阅读[初学指南](/starter-guide)
- 查找已有的[中间件集成](/middleware/)
- 许多[常见问题](/faqs)可能都已经得到解答
- 请通过[提交工单](https://github.com/flamego/flamego/issues)或[发起讨论](https://github.com/flamego/flamego/discussions)向我们提供反馈
- 关注 [Twitter](https://twitter.com/flamego_dev) 官方账号以获取最新动态
- 还可以在 GitHub 上找到我们的[品牌套件](https://github.com/flamego/brand-kit)
- 加入 QQ 群 910083715 参与社区交流