---
title: cache
weight: 30
---
cache 中间件为 [Flame 实例](../core-concepts#实例)提供缓存数据管理服务，支持的存储后端包括内存、文件系统、PostgreSQL、MySQL、Redis、MongoDB 和 SQLite。

你可以在 [GitHub](https://github.com/flamego/cache) 上阅读该中间件的源码或通过 [pkg.go.dev](https://pkg.go.dev/github.com/flamego/cache?tab=doc) 查看 API 文档。

## 下载安装

```
go get github.com/flamego/cache
```

## 存储后端

### 内存

[`cache.Cacher`](https://pkg.go.dev/github.com/flamego/cache#Cacher) 可以配合 [`cache.Options`](https://pkg.go.dev/github.com/flamego/cache#Options) 对中间件进行配置，并默认使用内存作为存储后端：

```go
package main

import (
	"net/http"
	"time"

	"github.com/flamego/cache"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()
	f.Use(cache.Cacher())
	f.Get("/set", func(r *http.Request, cache cache.Cache) error {
		return cache.Set(r.Context(), "cooldown", true, time.Minute)
	})
	f.Get("/get", func(r *http.Request, cache cache.Cache) string {
		v, err := cache.Get(r.Context(), "cooldown")
		if err != nil && err != os.ErrNotExist {
			return err.Error()
		}

		cooldown, ok := v.(bool)
		if !ok || !cooldown {
			return "It has been cooled"
		}
		return "Still hot"
	})
	f.Run()
}
```

由于数据存储在内存中，因此会在应用退出后被清除。如需持久化缓存数据，请选择其它存储后端。

### 文件系统

[`cache.FileIniter`](https://pkg.go.dev/github.com/flamego/cache#FileIniter) 是文件系统存储后端的初始化函数，并可以配合 [`cache.FileConfig`](https://pkg.go.dev/github.com/flamego/cache#FileConfig) 对其进行配置：

```go {hl_lines=["16-21"] linenostart=1}
package main

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/flamego/cache"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()
	f.Use(cache.Cacher(
		cache.Options{
			Initer: cache.FileIniter(),
			Config: cache.FileConfig{
				RootDir: filepath.Join(os.TempDir(), "cache"),
			},
		},
	))
	f.Get("/set", func(r *http.Request, cache cache.Cache) error {
		return cache.Set(r.Context(), "cooldown", true, time.Minute)
	})
	f.Get("/get", func(r *http.Request, cache cache.Cache) string {
		v, err := cache.Get(r.Context(), "cooldown")
		if err != nil && err != os.ErrNotExist {
			return err.Error()
		}

		cooldown, ok := v.(bool)
		if !ok || !cooldown {
			return "It has been cooled"
		}
		return "Still hot"
	})
	f.Run()
}
```

### PostgreSQL

[`postgres.Initer`](https://pkg.go.dev/github.com/flamego/cache/postgres#Initer) 是 PostgreSQL 存储后端的初始化函数，并可以配合 [`postgres.Config`](https://pkg.go.dev/github.com/flamego/cache/postgres#Config) 对其进行配置：

```go {hl_lines=["18-25"] linenostart=1}
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/flamego/cache"
	"github.com/flamego/cache/postgres"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()

	dsn := os.ExpandEnv("postgres://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE?sslmode=$PGSSLMODE")
	f.Use(cache.Cacher(
		cache.Options{
			Initer: postgres.Initer(),
			Config: postgres.Config{
				DSN:       dsn,
				Table:     "cache",
				InitTable: true,
			},
		},
	))
	f.Get("/set", func(r *http.Request, cache cache.Cache) error {
		return cache.Set(r.Context(), "cooldown", true, time.Minute)
	})
	f.Get("/get", func(r *http.Request, cache cache.Cache) string {
		v, err := cache.Get(r.Context(), "cooldown")
		if err != nil && err != os.ErrNotExist {
			return err.Error()
		}

		cooldown, ok := v.(bool)
		if !ok || !cooldown {
			return "It has been cooled"
		}
		return "Still hot"
	})
	f.Run()
}
```

### MySQL

[`mysql.Initer`](https://pkg.go.dev/github.com/flamego/cache/mysql#Initer) 是 MySQL 存储后端的初始化函数，并可以配合 [`mysql.Config`](https://pkg.go.dev/github.com/flamego/cache/mysql#Config) 对其进行配置：

```go {hl_lines=["18-25"] linenostart=1}
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/flamego/cache"
	"github.com/flamego/cache/mysql"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()

	dsn := os.ExpandEnv("$MYSQL_USER:$MYSQL_PASSWORD@tcp($MYSQL_HOST:$MYSQL_PORT)/$MYSQL_DATABASE?charset=utf8&parseTime=true")
	f.Use(cache.Cacher(
		cache.Options{
			Initer: mysql.Initer(),
			Config: mysql.Config{
				DSN:       dsn,
				Table:     "cache",
				InitTable: true,
			},
		},
	))
	f.Get("/set", func(r *http.Request, cache cache.Cache) error {
		return cache.Set(r.Context(), "cooldown", true, time.Minute)
	})
	f.Get("/get", func(r *http.Request, cache cache.Cache) string {
		v, err := cache.Get(r.Context(), "cooldown")
		if err != nil && err != os.ErrNotExist {
			return err.Error()
		}

		cooldown, ok := v.(bool)
		if !ok || !cooldown {
			return "It has been cooled"
		}
		return "Still hot"
	})
	f.Run()
}
```

### Redis

[`redis.Initer`](https://pkg.go.dev/github.com/flamego/cache/redis#Initer) 是 Redis 存储后端的初始化函数，并可以配合 [`redis.Config`](https://pkg.go.dev/github.com/flamego/cache/redis#Config) 对其进行配置：

```go {hl_lines=["17-25"] linenostart=1}
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/flamego/cache"
	"github.com/flamego/cache/redis"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()

	f.Use(cache.Cacher(
		cache.Options{
			Initer: redis.Initer(),
			Config: redis.Config{
				Options: &redis.Options{
					Addr: os.ExpandEnv("$REDIS_HOST:$REDIS_PORT"),
					DB:   15,
				},
			},
		},
	))
	f.Get("/set", func(r *http.Request, cache cache.Cache) error {
		return cache.Set(r.Context(), "cooldown", true, time.Minute)
	})
	f.Get("/get", func(r *http.Request, cache cache.Cache) string {
		v, err := cache.Get(r.Context(), "cooldown")
		if err != nil && err != os.ErrNotExist {
			return err.Error()
		}

		cooldown, ok := v.(bool)
		if !ok || !cooldown {
			return "It has been cooled"
		}
		return "Still hot"
	})
	f.Run()
}
```

### MongoDB

[`mongo.Initer`](https://pkg.go.dev/github.com/flamego/cache/mongo#Initer) 是 MongoDB 存储后端的初始化函数，并可以配合 [`mongo.Config`](https://pkg.go.dev/github.com/flamego/cache/mongo#Config) 对其进行配置：

```go {hl_lines=["18-25"] linenostart=1}
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/flamego/cache"
	"github.com/flamego/cache/mongo"
	"github.com/flamego/flamego"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	f := flamego.Classic()

	f.Use(cache.Cacher(
		cache.Options{
			Initer: mongo.Initer(),
			Config: mongo.Config{
				Options:    options.Client().ApplyURI(os.Getenv("MONGODB_URI")),
				Database:   os.Getenv("MONGODB_DATABASE"),
				Collection: "cache",
			},
		},
	))
	f.Get("/set", func(r *http.Request, cache cache.Cache) error {
		return cache.Set(r.Context(), "cooldown", true, time.Minute)
	})
	f.Get("/get", func(r *http.Request, cache cache.Cache) string {
		v, err := cache.Get(r.Context(), "cooldown")
		if err != nil && err != os.ErrNotExist {
			return err.Error()
		}

		cooldown, ok := v.(bool)
		if !ok || !cooldown {
			return "It has been cooled"
		}
		return "Still hot"
	})
	f.Run()
}
```

### SQLite

[`sqlite.Initer`](https://pkg.go.dev/github.com/flamego/cache/sqlite#Initer) 是 SQLite 存储后端的初始化函数，并可以配合 [`sqlite.Config`](https://pkg.go.dev/github.com/flamego/cache/sqlite#Config) 对其进行配置：

```go {hl_lines=["17-24"] linenostart=1}
package main

import (
	"net/http"
	"os"
	"time"

	"github.com/flamego/cache"
	"github.com/flamego/cache/sqlite"
	"github.com/flamego/flamego"
)

func main() {
	f := flamego.Classic()

	f.Use(cache.Cacher(
		cache.Options{
			Initer: sqlite.Initer(),
			Config: sqlite.Config{
				DSN:       "app.db",
				Table:     "cache",
				InitTable: true,
			},
		},
	))
	f.Get("/set", func(r *http.Request, cache cache.Cache) error {
		return cache.Set(r.Context(), "cooldown", true, time.Minute)
	})
	f.Get("/get", func(r *http.Request, cache cache.Cache) string {
		v, err := cache.Get(r.Context(), "cooldown")
		if err != nil && err != os.ErrNotExist {
			return err.Error()
		}

		cooldown, ok := v.(bool)
		if !ok || !cooldown {
			return "It has been cooled"
		}
		return "Still hot"
	})
	f.Run()
}
```

## 存储类型支持

缓存数据的默认编解码格式为 [gob](https://pkg.go.dev/encoding/gob)，因此仅支持有限的值类型。如果遇到类似 `encode: gob: type not registered for interface: time.Duration` 这样的错误，则可以通过 [`gob.Register`](https://pkg.go.dev/encoding/gob#Register) 在应用中将该类型注册到编解码器中解决：

```go
gob.Register(time.Duration(0))
```

单个应用中对同一类型仅需注册一次。