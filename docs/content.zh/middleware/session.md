---
title: session
weight: 20
---
session 中间件为 [Flame 实例](../core-concepts#实例)提供用户会话管理服务，支持的存储后端包括内存、文件系统、PostgreSQL、MySQL、Redis、MongoDB 和 SQLite。

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

## 存储类型支持

缓存数据的默认编解码格式为 [gob](https://pkg.go.dev/encoding/gob)，因此仅支持有限的值类型。如果遇到类似 `encode: gob: type not registered for interface: time.Duration` 这样的错误，则可以通过 [`gob.Register`](https://pkg.go.dev/encoding/gob#Register) 在应用中将该类型注册到编解码器中解决：

```go
gob.Register(time.Duration(0))
```

单个应用中对同一类型仅需注册一次。