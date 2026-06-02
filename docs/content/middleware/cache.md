---
title: cache
weight: 30
---
The cache middleware provides cache data management for [Flame instances](/core-concepts#instances), supporting various storage backends, including memory, file, PostgreSQL, MySQL, Redis, MongoDB and SQLite.

You can read source code of this middleware on [GitHub](https://github.com/flamego/cache) and API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/flamego/cache?tab=doc).

## Installation

```
go get github.com/flamego/cache
```

## Storage backends

### Memory

The [`cache.Cacher`](https://pkg.go.dev/github.com/flamego/cache#Cacher) works out-of-the-box with an optional [`cache.Options`](https://pkg.go.dev/github.com/flamego/cache#Options) and uses memory as the storage backend:

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

Because the memory is volatile, cache data do not survive over restarts. Choose other storage backends if you need to persist cache data.

### File

The [`cache.FileIniter`](https://pkg.go.dev/github.com/flamego/cache#FileIniter) is the function to initialize a file storage backend, used together with [`cache.FileConfig`](https://pkg.go.dev/github.com/flamego/cache#FileConfig) to customize the backend:

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

The [`postgres.Initer`](https://pkg.go.dev/github.com/flamego/cache/postgres#Initer) is the function to initialize a PostgreSQL storage backend, used together with [`postgres.Config`](https://pkg.go.dev/github.com/flamego/cache/postgres#Config) to customize the backend:

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

The [`mysql.Initer`](https://pkg.go.dev/github.com/flamego/cache/mysql#Initer) is the function to initialize a MySQL storage backend, used together with [`mysql.Config`](https://pkg.go.dev/github.com/flamego/cache/mysql#Config) to customize the backend:

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

The [`redis.Initer`](https://pkg.go.dev/github.com/flamego/cache/redis#Initer) is the function to initialize a Redis storage backend, used together with [`redis.Config`](https://pkg.go.dev/github.com/flamego/cache/redis#Config) to customize the backend:

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

The [`mongo.Initer`](https://pkg.go.dev/github.com/flamego/cache/mongo#Initer) is the function to initialize a MongoDB storage backend, used together with [`mongo.Config`](https://pkg.go.dev/github.com/flamego/cache/mongo#Config) to customize the backend:

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

The [`sqlite.Initer`](https://pkg.go.dev/github.com/flamego/cache/sqlite#Initer) is the function to initialize a SQLite storage backend, used together with [`sqlite.Config`](https://pkg.go.dev/github.com/flamego/cache/sqlite#Config) to customize the backend:

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

## Supported value types

The default encoder and decoder of cache data use [gob](https://pkg.go.dev/encoding/gob), and only limited types are supported for values. When you encounter errors like `encode: gob: type not registered for interface: time.Duration`, you can use [`gob.Register`](https://pkg.go.dev/encoding/gob#Register) to register the type for encoding and decoding.

For example:

```go
gob.Register(time.Duration(0))
```

You only need to regsiter once for the entire lifecyle of your application.