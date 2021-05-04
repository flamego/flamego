// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package flamego is a fantastic modular Go web framework boiled with black magic.
package flamego

import (
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/flamego/flamego/internal/inject"
	"github.com/flamego/flamego/internal/route"
)

// Flame is the top-level web application instance, which manages global states
// of injected services and middleware.
type Flame struct {
	inject.Injector
	Router

	urlPrefix string // The URL prefix to be trimmed for every request.

	befores  []BeforeHandler
	handlers []Handler
	action   Handler
	logger   *log.Logger
}

// NewWithLogger creates and returns a bare bones Flame instance. Use this
// function if you want to have full control over log destination and middleware
// that are used.
func NewWithLogger(w io.Writer) *Flame {
	f := &Flame{
		Injector: inject.New(),
		logger:   log.New(w, "[Flamego] ", 0),
	}
	f.Router = newRouter(f.createContext)
	f.NotFound(http.NotFound)

	f.Map(f.logger)
	return f
}

// New creates and returns a bare bones Flame instance with default logger
// writing to os.Stdout. Use this function if you want to have full control over
// middleware that are used.
func New() *Flame {
	return NewWithLogger(os.Stdout)
}

func (f *Flame) createContext(w http.ResponseWriter, r *http.Request, params route.Params, handlers []Handler, urlPath urlPather) Context {
	// Allocate a new slice to avoid mutating the original "handlers" and that could
	// potentially cause data race.
	hs := make([]Handler, 0, len(f.handlers)+len(handlers))
	hs = append(hs, f.handlers...)
	hs = append(hs, handlers...)

	c := newContext(w, r, params, hs, urlPath)
	c.SetParent(f)
	return c
}

// Use adds a handler of a middleware to the Flame instance, and panics if the
// handler is not a callable func. Middleware handlers are invoked in the same
// order as they are added.
func (f *Flame) Use(h Handler) {
	f.handlers = append(f.handlers, validateAndWrapHandler(h, nil))
}

// Handlers sets the entire middleware stack with the given Handlers. This will
// clear any current middleware handlers, and panics if any of the handlers is
// not a callable function
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

func (f *Flame) Before(h BeforeHandler) {
	f.befores = append(f.befores, h)
}

// ServeHTTP is the HTTP Entry point for a Macaron instance.
// Useful if you want to control your own HTTP server.
// Be aware that none of middleware will run without registering any router.
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

// Run starts the HTTP server on "0.0.0.0:4000". The listen address can be
// altered by the environment variable "FLAMEGO_ADDR".
func (f *Flame) Run(args ...interface{}) {
	host := "0.0.0.0"
	port := "4000"

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
	logger := f.Value(reflect.TypeOf(f.logger)).Interface().(*log.Logger)
	logger.Printf("Listening on %s (%s)\n", addr, Env())
	logger.Fatalln(http.ListenAndServe(addr, f))
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

// Env returns the current runtime environment.
func Env() EnvType {
	return env.Load().(EnvType)
}
