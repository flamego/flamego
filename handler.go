// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"
	"reflect"

	"github.com/flamego/flamego/internal/inject"
)

// Handler is any callable function. Flamego attempts to inject services into
// the Handler's argument list and panics if any argument could not be fulfilled
// via dependency injection.
type Handler interface{}

var _ inject.FastInvoker = (*contextInvoker)(nil)

// contextInvoker is an inject.FastInvoker implementation of `func(*Context)`.
type contextInvoker func(ctx *Context)

func (invoke contextInvoker) Invoke(args []interface{}) ([]reflect.Value, error) {
	invoke(args[0].(*Context))
	return nil, nil
}

var _ inject.FastInvoker = (*httpHandlerFuncInvoker)(nil)

// httpHandlerFuncInvoker is an inject.FastInvoker implementation of
// `func(http.ResponseWriter, *http.Request)`.
type httpHandlerFuncInvoker func(http.ResponseWriter, *http.Request)

func (invoke httpHandlerFuncInvoker) Invoke(args []interface{}) ([]reflect.Value, error) {
	invoke(args[0].(http.ResponseWriter), args[1].(*http.Request))
	return nil, nil
}

// validateAndWrapHandler makes sure the handler is a callable function, it
// panics if not. When the handler is also convertible to any built-in
// inject.FastInvoker implementations, it wraps the handler automatically to
// gain up to 3x performance improvement.
func validateAndWrapHandler(h Handler, wrapper func(Handler) Handler) Handler {
	if reflect.TypeOf(h).Kind() != reflect.Func {
		panic("handler must be a callable function")
	}

	if inject.IsFastInvoker(h) {
		return h
	}

	switch v := h.(type) {
	case func(*Context):
		return contextInvoker(v)
	case func(http.ResponseWriter, *http.Request):
		return httpHandlerFuncInvoker(v)
	case http.HandlerFunc:
		return httpHandlerFuncInvoker(v)
	}

	if wrapper != nil {
		h = wrapper(h)
	}
	return h
}

// validateAndWrapHandlers preforms validation and wrapping for given handlers.
func validateAndWrapHandlers(handlers []Handler, wrapper func(Handler) Handler) {
	for i, h := range handlers {
		handlers[i] = validateAndWrapHandler(h, wrapper)
	}
}
