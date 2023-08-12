// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package flamego is a fantastic modular Go web framework with a slim core but limitless extensibility.
package flamego

import (
	gocontext "context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"

	"github.com/flamego/flamego/inject"
	"github.com/flamego/flamego/internal/route"
)

// Flame is the top-level web application instance, which manages global states
// of injected services and middleware.
type Flame struct {
	inject.Injector
	Router

	urlPrefix string // The URL prefix to be trimmed for every request.

	befores  []BeforeHandler // The list of handlers to be called before matching route.
	handlers []Handler       // The list of middleware handlers.
	action   Handler         // The last action handler to be executed.
	logger   *log.Logger     // The default request logger.

	stop chan struct{} // The signal to stop the HTTP server.
}

// NewWithLogger creates and returns a bare bones Flame instance. Use this
// function if you want to have full control over log destination and middleware
// that are used.
func NewWithLogger(w io.Writer) *Flame {
	f := &Flame{
		Injector: inject.New(),
		logger: log.NewWithOptions(
			w,
			log.Options{
				TimeFormat:      "2006-01-02 15:04:05", // TODO(go1.20): Use time.DateTime
				Level:           log.DebugLevel,
				ReportTimestamp: true,
			},
		),
		stop: make(chan struct{}),
	}
	f.Router = newRouter(f.createContext)
	f.NotFound(http.NotFound)

	f.Map(f.logger)
	f.Map(f.logger.StandardLog())
	f.Map(defaultReturnHandler())
	return f
}

// New creates and returns a bare bones Flame instance with default logger
// writing to os.Stdout. Use this function if you want to have full control over
// middleware that are used.
func New() *Flame {
	return NewWithLogger(os.Stdout)
}

// Classic creates and returns a classic Flame instance with default middleware:
// `flamego.Logger`, `flamego.Recovery` and `flamego.Static`.
func Classic() *Flame {
	f := New()
	f.Use(
		Logger(),
		Recovery(),
		Static(),
	)
	return f
}

func (f *Flame) createContext(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) internalContext {
	// Allocate a new slice to avoid mutating the original "handlers" and that could
	// potentially cause data race.
	hs := make([]Handler, 0, len(f.handlers)+len(handlers))
	hs = append(hs, f.handlers...)
	hs = append(hs, handlers...)

	c := newContext(w, r, params, hs, urlPath)
	c.SetParent(f)

	if f.action != nil {
		c.setAction(f.action)
	}
	return c
}

// Use adds handlers of middleware to the Flame instance, and panics if any of
// the handler is not a callable func. Middleware handlers are invoked in the
// same order as they are added.
func (f *Flame) Use(handlers ...Handler) {
	validateAndWrapHandlers(handlers, nil)
	f.handlers = append(f.handlers, handlers...)
}

// Handlers sets the entire middleware stack with the given Handlers. This will
// clear any current middleware handlers, and panics if any of the handlers is
// not a callable function.
func (f *Flame) Handlers(handlers ...Handler) {
	f.handlers = make([]Handler, 0, len(handlers))
	for _, handler := range handlers {
		f.Use(handler)
	}
}

// Action sets the final handler that will be called after all handlers have
// been invoked.
func (f *Flame) Action(h Handler) {
	f.action = validateAndWrapHandler(h, nil)
}

// BeforeHandler is a handler executes at beginning of every request. Flame
// instance stops further process when it returns true.
type BeforeHandler func(rw http.ResponseWriter, req *http.Request) bool

// Before allows for a handler to be called before matching any route. Multiple
// calls to this method will queue up handlers, and handlers will be called in
// the FIFO manner.
func (f *Flame) Before(h BeforeHandler) {
	f.befores = append(f.befores, h)
}

func (f *Flame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f.urlPrefix != "" {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, f.urlPrefix)
	}
	for _, h := range f.befores {
		if h(w, r) {
			return
		}
	}
	f.Router.ServeHTTP(w, r)
}

// Run starts the HTTP server on "0.0.0.0:2830". The listen address can be
// altered by the environment variable "FLAMEGO_ADDR". The instance can be
// stopped by calling `Flame.Stop`.
func (f *Flame) Run(args ...interface{}) {
	logger := f.logger.WithPrefix("🧙 Flamego")

	host := "0.0.0.0"
	port := "2830"

	if addr := os.Getenv("FLAMEGO_ADDR"); addr != "" {
		fields := strings.SplitN(addr, ":", 2)
		host = fields[0]
		port = fields[1]
	}

	if len(args) == 1 {
		switch arg := args[0].(type) {
		case string:
			host = arg
		case int:
			port = strconv.Itoa(arg)
		default:
			logger.Print("Ignoring invalid type of argument", "type", fmt.Sprintf("%T", arg), "value", arg)
		}
	} else if len(args) >= 2 {
		if arg, ok := args[0].(string); ok {
			host = arg
		}
		if arg, ok := args[1].(int); ok {
			port = strconv.Itoa(arg)
		}
	}

	addr := host + ":" + port
	logger.Print("Serving on http://localhost:"+port, "env", Env())

	server := &http.Server{
		Addr:              addr,
		Handler:           f,
		ReadHeaderTimeout: 3 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	<-f.stop

	if err := server.Shutdown(gocontext.Background()); err != nil {
		logger.Fatal("Failed to shutdown server", "error", err)
	}
	logger.Print("Server stopped")
}

// Stop stops the server started by the Flame instance.
func (f *Flame) Stop() {
	f.stop <- struct{}{}
}

// EnvType defines the runtime environment.
type EnvType string

const (
	EnvTypeDev  EnvType = "development"
	EnvTypeProd EnvType = "production"
	EnvTypeTest EnvType = "test"
)

var env = func() atomic.Value {
	var v atomic.Value
	v.Store(EnvTypeDev)
	return v
}()

// Env returns the current runtime environment. It can be altered by SetEnv or
// the environment variable "FLAMEGO_ENV".
func Env() EnvType {
	return env.Load().(EnvType)
}

// SetEnv sets the current runtime environment. Valid values are EnvTypeDev,
// EnvTypeProd and EnvTypeTest, all else ignored.
func SetEnv(e EnvType) {
	if e == EnvTypeDev ||
		e == EnvTypeProd ||
		e == EnvTypeTest {
		env.Store(e)
	}
}

func init() {
	SetEnv(EnvType(os.Getenv("FLAMEGO_ENV")))
}
