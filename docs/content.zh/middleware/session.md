---
title: session
weight: 20
---
session 中间件为 [Flame 实例](/core-concepts#实例)提供用户会话管理服务，支持的存储后端包括内存、文件系统、PostgreSQL、MySQL、Redis、MongoDB 和 SQLite。

你可以在 [GitHub](https://github.com/flamego/session) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/session?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/session
```

## 存储后端

{{< callout type="warning" >}}
本小结仅展示 session 中间件的相关用法，示例中的用户验证方案绝不可以直接被用于生产环境。
{{< /callout >}}

### 内存

[`session.Sessioner`](https://pkg.go.dev/github.com/flamego/session#Sessioner) 可以配合 [`session.Options`](https://pkg.go.dev/github.com/flamego/session#Options) 对中间件进行配置，并默认使用内存作为存储后端：

```go
package main

import (
	"strconv"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
)

func main() {
	f := flamego.Classic()
	f.Use(session.Sessioner())
	f.Get("/set", func(s session.Session) string {
		s.Set("user_id", 123)
		return "Succeed"
	})
	f.Get("/get", func(s session.Session) string {
		userID, ok := s.Get("user_id").(int)
		if !ok || userID <= 0 {
			return "Not authenticated"
		}
		return "Authenticated as " + strconv.Itoa(userID)
	})
	f.Get("/clear", func(s session.Session) string {
		s.Delete("user_id")
		return "Cleared"
	})
	f.Run()
}
```

由于数据存储在内存中，因此会在应用退出后被清除。如需持久化会话数据，请选择其它存储后端。

### 文件系统

[`session.FileIniter`](https://pkg.go.dev/github.com/flamego/session#FileIniter) 是文件系统存储后端的初始化函数，并可以配合 [`session.FileConfig`](https://pkg.go.dev/github.com/flamego/session#FileConfig) 对其进行配置：

```go {hl_lines=["15-20"] linenostart=1}
package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
)

func main() {
	f := flamego.Classic()
	f.Use(session.Sessioner(
		session.Options{
			Initer: session.FileIniter(),
			Config: session.FileConfig{
				RootDir: filepath.Join(os.TempDir(), "sessions"),
			},
		},
	))
	f.Get("/set", func(s session.Session) string {
		s.Set("user_id", 123)
		return "Succeed"
	})
	f.Get("/get", func(s session.Session) string {
		userID, ok := s.Get("user_id").(int)
		if !ok || userID <= 0 {
			return "Not authenticated"
		}
		return "Authenticated as " + strconv.Itoa(userID)
	})
	f.Get("/clear", func(s session.Session) string {
		s.Delete("user_id")
		return "Cleared"
	})
	f.Run()
}
```

### PostgreSQL

[`postgres.Initer`](https://pkg.go.dev/github.com/flamego/session/postgres#Initer) 是 PostgreSQL 存储后端的初始化函数，并可以配合 [`postgres.Config`](https://pkg.go.dev/github.com/flamego/session/postgres#Config) 对其进行配置：

```go {hl_lines=["17-24"] linenostart=1}
package main

import (
	"os"
	"strconv"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/flamego/session/postgres"
)

func main() {
	f := flamego.Classic()

	dsn := os.ExpandEnv("postgres://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE?sslmode=$PGSSLMODE")
	f.Use(session.Sessioner(
		session.Options{
			Initer: postgres.Initer(),
			Config: postgres.Config{
				DSN:       dsn,
				Table:     "sessions",
				InitTable: true,
			},
		},
	))
	f.Get("/set", func(s session.Session) string {
		s.Set("user_id", 123)
		return "Succeed"
	})
	f.Get("/get", func(s session.Session) string {
		userID, ok := s.Get("user_id").(int)
		if !ok || userID <= 0 {
			return "Not authenticated"
		}
		return "Authenticated as " + strconv.Itoa(userID)
	})
	f.Get("/clear", func(s session.Session) string {
		s.Delete("user_id")
		return "Cleared"
	})
	f.Run()
}
```

### MySQL

[`mysql.Initer`](https://pkg.go.dev/github.com/flamego/session/mysql#Initer) 是 MySQL 存储后端的初始化函数，并可以配合 [`mysql.Config`](https://pkg.go.dev/github.com/flamego/session/mysql#Config) 对其进行配置：

```go {hl_lines=["17-24"] linenostart=1}
package main

import (
	"os"
	"strconv"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/flamego/session/mysql"
)

func main() {
	f := flamego.Classic()

	dsn := os.ExpandEnv("$MYSQL_USER:$MYSQL_PASSWORD@tcp($MYSQL_HOST:$MYSQL_PORT)/$MYSQL_DATABASE?charset=utf8&parseTime=true")
	f.Use(session.Sessioner(
		session.Options{
			Initer: mysql.Initer(),
			Config: mysql.Config{
				DSN:       dsn,
				Table:     "cache",
				InitTable: true,
			},
		},
	))
	f.Get("/set", func(s session.Session) string {
		s.Set("user_id", 123)
		return "Succeed"
	})
	f.Get("/get", func(s session.Session) string {
		userID, ok := s.Get("user_id").(int)
		if !ok || userID <= 0 {
			return "Not authenticated"
		}
		return "Authenticated as " + strconv.Itoa(userID)
	})
	f.Get("/clear", func(s session.Session) string {
		s.Delete("user_id")
		return "Cleared"
	})
	f.Run()
}
```

### Redis

[`redis.Initer`](https://pkg.go.dev/github.com/flamego/session/redis#Initer) 是 Redis 存储后端的初始化函数，并可以配合 [`redis.Config`](https://pkg.go.dev/github.com/flamego/session/redis#Config) 对其进行配置：

```go {hl_lines=["16-24"] linenostart=1}
package main

import (
	"os"
	"strconv"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/flamego/session/redis"
)

func main() {
	f := flamego.Classic()

	f.Use(session.Sessioner(
		session.Options{
			Initer: redis.Initer(),
			Config: redis.Config{
				Options: &redis.Options{
					Addr: os.ExpandEnv("$REDIS_HOST:$REDIS_PORT"),
					DB:   15,
				},
			},
		},
	))
	f.Get("/set", func(s session.Session) string {
		s.Set("user_id", 123)
		return "Succeed"
	})
	f.Get("/get", func(s session.Session) string {
		userID, ok := s.Get("user_id").(int)
		if !ok || userID <= 0 {
			return "Not authenticated"
		}
		return "Authenticated as " + strconv.Itoa(userID)
	})
	f.Get("/clear", func(s session.Session) string {
		s.Delete("user_id")
		return "Cleared"
	})
	f.Run()
}
```

### MongoDB

[`mongo.Initer`](https://pkg.go.dev/github.com/flamego/session/mongo#Initer) 是 MongoDB 存储后端的初始化函数，并可以配合 [`mongo.Config`](https://pkg.go.dev/github.com/flamego/session/mongo#Config) 对其进行配置：

```go {hl_lines=["17-24"] linenostart=1}
package main

import (
	"os"
	"strconv"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/flamego/session/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	f := flamego.Classic()

	f.Use(session.Sessioner(
		session.Options{
			Initer: mongo.Initer(),
			Config: mongo.Config{
				Options:    options.Client().ApplyURI(os.Getenv("MONGODB_URI")),
				Database:   os.Getenv("MONGODB_DATABASE"),
				Collection: "cache",
			},
		},
	))
	f.Get("/set", func(s session.Session) string {
		s.Set("user_id", 123)
		return "Succeed"
	})
	f.Get("/get", func(s session.Session) string {
		userID, ok := s.Get("user_id").(int)
		if !ok || userID <= 0 {
			return "Not authenticated"
		}
		return "Authenticated as " + strconv.Itoa(userID)
	})
	f.Get("/clear", func(s session.Session) string {
		s.Delete("user_id")
		return "Cleared"
	})
	f.Run()
}
```

### SQLite

[`sqlite.Initer`](https://pkg.go.dev/github.com/flamego/session/sqlite#Initer) 是 SQLite 存储后端的初始化函数，并可以配合 [`sqlite.Config`](https://pkg.go.dev/github.com/flamego/session/sqlite#Config) 对其进行配置：

```go {hl_lines=["16-23"] linenostart=1}
package main

import (
	"os"
	"strconv"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/flamego/session/sqlite"
)

func main() {
	f := flamego.Classic()

	f.Use(session.Sessioner(
		session.Options{
			Initer: sqlite.Initer(),
			Config: sqlite.Config{
				DSN:       "app.db",
				Table:     "sessions",
				InitTable: true,
			},
		},
	))
	f.Get("/set", func(s session.Session) string {
		s.Set("user_id", 123)
		return "Succeed"
	})
	f.Get("/get", func(s session.Session) string {
		userID, ok := s.Get("user_id").(int)
		if !ok || userID <= 0 {
			return "Not authenticated"
		}
		return "Authenticated as " + strconv.Itoa(userID)
	})
	f.Get("/clear", func(s session.Session) string {
		s.Delete("user_id")
		return "Cleared"
	})
	f.Run()
}
```

## 闪现消息

session 中间件提供了闪现消息的机制，闪现消息指的是在下次会话展现给用户的消息，并且只会展现一次。

闪现消息最简单的形式就是字符串：

```go
package main

import (
	"github.com/flamego/flamego"
	"github.com/flamego/session"
)

func main() {
	f := flamego.Classic()
	f.Use(session.Sessioner())
	f.Get("/set", func(s session.Session) string {
		s.SetFlash("This is a flash message")
		return "Succeed"
	})
	f.Get("/get", func(f session.Flash) string {
		s, ok := f.(string)
		if !ok || s == "" {
			return "No flash message"
		}
		return s
	})
	f.Run()
}
```

[`session.Flash`](https://pkg.go.dev/github.com/flamego/session#Flash) 是单纯作为闪现消息的载体而存在，你可以使用任意类型作为闪现消息的具体表现形式，且不同路由之间的闪现消息类型也可完全不同：

```go {hl_lines=["15", "31-33"] linenostart=1}
package main

import (
	"fmt"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
)

func main() {
	f := flamego.Classic()
	f.Use(session.Sessioner())

	f.Get("/set-simple", func(s session.Session) string {
		s.SetFlash("This is a flash message")
		return "Succeed"
	})
	f.Get("/get-simple", func(f session.Flash) string {
		s, ok := f.(string)
		if !ok || s == "" {
			return "No flash message"
		}
		return s
	})

	type Flash struct {
		Success string
		Error   string
	}
	f.Get("/set-complex", func(s session.Session) string {
		s.SetFlash(Flash{
			Success: "It worked!",
		})
		return "Succeed"
	})
	f.Get("/get-complex", func(f session.Flash) string {
		s, ok := f.(Flash)
		if !ok {
			return "No flash message"
		}
		return fmt.Sprintf("%#v", s)
	})
	f.Run()
}
```

上例展示了如何在不同的路由中使用不同类型的闪现消息（`string` and `Flash`）。

## 绝对截止时间

存储后端的生命周期是基于空闲时间的，即每次请求都会延长会话。如果会话必须在固定时间结束而不论是否有活动，则可以使用 [`session.SetDeadline`](https://pkg.go.dev/github.com/flamego/session#SetDeadline) 设置绝对截止时间，例如在用户登录时。截止时间之前的请求不会推迟它。如果请求在截止时间或之后开始，会话中间件会销毁该会话并使用新的 ID 创建一个新会话，即便客户端仍然发送旧的会话 ID。

例如，要求用户在每周一 00:00 UTC 重新登录，但截止时间距离登录必须超过五分钟：

```go
package main

import (
	"time"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
)

// nextDeadline 返回下一个周一 00:00 UTC，如果距离它不足或恰好五分钟，则返回再下一个周一。
func nextDeadline(now time.Time) time.Time {
	now = now.UTC()
	daysUntilMonday := (7 - int(now.Weekday()) + int(time.Monday)) % 7
	deadline := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 0, 0, 0, 0, time.UTC)
	if deadline.Sub(now) <= 5*time.Minute {
		deadline = deadline.AddDate(0, 0, 7)
	}
	return deadline
}

func main() {
	f := flamego.Classic()
	f.Use(session.Sessioner(
		session.Options{
			Config: session.MemoryConfig{
				// 由截止时间而非空闲时间结束会话
				Lifetime: 8 * 24 * time.Hour,
			},
		},
	))
	f.Get("/sign-in", func(s session.Session) string {
		s.Set("user_id", 123)
		session.SetDeadline(s, nextDeadline(time.Now()))
		return "Signed in"
	})
	f.Get("/deadline", func(s session.Session) string {
		deadline, ok := session.Deadline(s)
		if !ok {
			return "No deadline"
		}
		return deadline.UTC().String()
	})
	f.Run()
}
```

截止时间与会话数据存储在一起，因此在使用持久化存储后端时，截止时间在重启和多实例之间依然生效。它不会延长会话的寿命：如果存储后端的空闲生命周期或 Cookie 先过期，会话仍会提前结束。如果只希望截止时间生效，请将存储后端的 `Lifetime` 和 `Cookie.MaxAge` 设置得足够长。

关于截止时间还需要注意：

- [`Session.Flush`](https://pkg.go.dev/github.com/flamego/session#Session) 会保留截止时间，因此清空会话后重新登录不会解除截止时间。如需移除截止时间，请调用 `session.SetDeadline(s, time.Time{})`。
- 截止时间会在请求开始和结束时各检查一次。到达截止时间时仍在处理中的请求会继续使用旧会话直到返回，但随后该会话会被销毁而不是保存，之后使用旧 ID 的请求会得到该 ID 下的一个空会话。WebSocket 或服务器发送事件（SSE）等长连接需要自行将 `session.Deadline` 与当前时间进行比较。
- 过期会话的数据会保留在存储后端中，直到客户端再次使用该会话发送请求，或由存储后端的空闲生命周期回收，但这些数据无法再用于验证请求。
- 每个实例都使用自己的时钟比较截止时间，因此请保持多个实例的时钟同步。

## 存储类型支持

缓存数据的默认编解码格式为 [gob](https://pkg.go.dev/encoding/gob)，因此仅支持有限的值类型。如果遇到类似 `encode: gob: type not registered for interface: time.Duration` 这样的错误，则可以通过 [`gob.Register`](https://pkg.go.dev/encoding/gob#Register) 在应用中将该类型注册到编解码器中解决：

```go
gob.Register(time.Duration(0))
```

单个应用中对同一类型仅需注册一次。