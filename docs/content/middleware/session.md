---
title: session
weight: 20
---
The session middleware provides user session management for [Flame instances](/core-concepts#instances), supporting various storage backends, including memory, file, PostgreSQL, MySQL, Redis, MongoDB and SQLite.

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

## Absolute deadlines

The lifetime of storage backends is based on idle time, i.e., every request extends the session. When a session must end at a fixed time regardless of activity, use [`session.SetDeadline`](https://pkg.go.dev/github.com/flamego/session#SetDeadline) to set an absolute deadline, e.g., when the user signs in. Requests before the deadline do not move it. When a request starts at or after the deadline, the session middleware destroys the session and starts a new one with a new ID, even if the client still sends the old one.

For example, to require signing in again every Monday at 00:00 UTC, but never five minutes or less after signing in:

```go
package main

import (
	"time"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
)

// nextDeadline returns the next Monday 00:00 UTC, or the one after when the next
// one is five minutes or less away.
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
				// Let the deadline end the session rather than being idle.
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

The deadline is stored with the session data, so it is enforced across restarts and multiple instances when using a persistent storage backend. It does not keep the session alive any longer: the session still ends earlier if the idle lifetime of the storage backend or the cookie expires first, so set the `Lifetime` of the storage backend and `Cookie.MaxAge` long enough if you only want the deadline to apply.

A few more things to know about deadlines:

- [`Session.Flush`](https://pkg.go.dev/github.com/flamego/session#Session) keeps the deadline, so flushing and then signing the user in again does not lift it. Use `session.SetDeadline(s, time.Time{})` to remove the deadline.
- The deadline is checked when a request starts and again when it ends. A request that is still running at the deadline keeps the old session until it returns, but the session is then destroyed instead of saved, and the next request with the old ID gets an empty session under that ID. Long-lived connections such as WebSocket or server-sent events should compare `session.Deadline` with the current time themselves.
- Data of an expired session stays in the storage backend until the client sends a request with that session again, or until the idle lifetime of the storage backend recycles it. It can no longer be used to authenticate requests.
- Every instance compares the deadline with its own clock, so keep clocks of multiple instances in sync.

## Supported value types

The default encoder and decoder of cache data use [gob](https://pkg.go.dev/encoding/gob), and only limited types are supported for values. When you encounter errors like `encode: gob: type not registered for interface: time.Duration`, you can use [`gob.Register`](https://pkg.go.dev/encoding/gob#Register) to register the type for encoding and decoding.

For example:

```go
gob.Register(time.Duration(0))
```

You only need to regsiter once for the entire lifecyle of your application.