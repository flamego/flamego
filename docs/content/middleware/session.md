---
title: session
weight: 20
---
The session middleware provides user session management for [Flame instances](../core-concepts#instances), supporting various storage backends, including memory, file, PostgreSQL, MySQL, Redis, MongoDB and SQLite.

You can read source code of this middleware on [GitHub](https://github.com/flamego/session) and API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/flamego/session?tab=doc).

## Installation

```
go get github.com/flamego/session
```

## Storage backends

{{< callout type="warning" >}}
Examples included in this section is to demonstrate the usage of the session middleware, by no means illustrates the idiomatic or even correct way of doing user authentication.
{{< /callout >}}

### Memory

The [`session.Sessioner`](https://pkg.go.dev/github.com/flamego/session#Sessioner) works out-of-the-box with an optional [`session.Options`](https://pkg.go.dev/github.com/flamego/session#Options) and uses memory as the storage backend:

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

Because the memory is volatile, session data do not survive over restarts. Choose other storage backends if you need to persist session data.

### File

The [`session.FileIniter`](https://pkg.go.dev/github.com/flamego/session#FileIniter) is the function to initialize a file storage backend, used together with [`session.FileConfig`](https://pkg.go.dev/github.com/flamego/session#FileConfig) to customize the backend:

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

The [`postgres.Initer`](https://pkg.go.dev/github.com/flamego/session/postgres#Initer) is the function to initialize a PostgreSQL storage backend, used together with [`postgres.Config`](https://pkg.go.dev/github.com/flamego/session/postgres#Config) to customize the backend:

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

The [`mysql.Initer`](https://pkg.go.dev/github.com/flamego/session/mysql#Initer) is the function to initialize a MySQL storage backend, used together with [`mysql.Config`](https://pkg.go.dev/github.com/flamego/session/mysql#Config) to customize the backend:

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

The [`redis.Initer`](https://pkg.go.dev/github.com/flamego/session/redis#Initer) is the function to initialize a Redis storage backend, used together with [`redis.Config`](https://pkg.go.dev/github.com/flamego/session/redis#Config) to customize the backend:

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

The [`mongo.Initer`](https://pkg.go.dev/github.com/flamego/session/mongo#Initer) is the function to initialize a MongoDB storage backend, used together with [`mongo.Config`](https://pkg.go.dev/github.com/flamego/session/mongo#Config) to customize the backend:

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

The [`sqlite.Initer`](https://pkg.go.dev/github.com/flamego/session/sqlite#Initer) is the function to initialize a SQLite storage backend, used together with [`sqlite.Config`](https://pkg.go.dev/github.com/flamego/session/sqlite#Config) to customize the backend:

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

## Flash messages

The session middleware provides a mechanism for flash messages, which are always retrieved on the next access of the same session, once and only once (i.e. flash messages get deleted upon retrievals).

A flash message could just be a string in its simplest form:

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

The [`session.Flash`](https://pkg.go.dev/github.com/flamego/session#Flash) is just the value holder of the flash message, and it could be any type that fits your application's needs, and doesn't even have to be the same type for different routes in the same application!

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

In the above example, we use different types of flash messages (`string` and `Flash`) for different routes and both of them work!

## Supported value types

The default encoder and decoder of cache data use [gob](https://pkg.go.dev/encoding/gob), and only limited types are supported for values. When you encounter errors like `encode: gob: type not registered for interface: time.Duration`, you can use [`gob.Register`](https://pkg.go.dev/encoding/gob#Register) to register the type for encoding and decoding.

For example:

```go
gob.Register(time.Duration(0))
```

You only need to regsiter once for the entire lifecyle of your application.