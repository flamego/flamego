---
title: sse
weight: 140
---
The sse middleware provides [Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) for sending updates from [Flame instances](/core-concepts#instances) to web clients over an HTTP connection.

You can read source code of this middleware on [GitHub](https://github.com/flamego/sse) and API documentation on [pkg.go.dev](https://pkg.go.dev/github.com/flamego/sse?tab=doc).

## Installation

```
go get github.com/flamego/sse
```

## Usage examples

The [`sse.Bind`](https://pkg.go.dev/github.com/flamego/sse#Bind) middleware configures the response as an event stream and injects a typed, send-only channel into the next handler. Pass a non-pointer value to `Bind`; a value of type `T` makes a `chan<- *T` available to the handler. Values sent through the channel are encoded as JSON event data.

The following example sends the current time to each connected client once per second:

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

Clients can consume the stream with the browser's [`EventSource`](https://developer.mozilla.org/en-US/docs/Web/API/EventSource) API:

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

The route handler remains active for the lifetime of the connection. Always observe `c.Request().Context().Done()` and stop timers or other resources when the client disconnects.

The middleware periodically sends a ping to keep the connection alive. The default interval is 10 seconds and can be changed with [`sse.Options`](https://pkg.go.dev/github.com/flamego/sse#Options):

```go
f.Get("/events",
	sse.Bind(update{}, sse.Options{
		PingInterval: 30 * time.Second,
	}),
	func(c flamego.Context, events chan<- *update) {
		// Produce events until the request context is canceled.
	},
)
```
