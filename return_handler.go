// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/flamego/flamego/inject"
)

// ReturnHandler is a service that is called when a route handler returns
// something. The ReturnHandler is responsible for writing to the ResponseWriter
// based on the values that are passed into this function.
type ReturnHandler func(Context, []reflect.Value)

// TypedReturnHandler is a function that handles route handler return values.
//
// It must accept Context as its first argument and must not return values. Its
// remaining arguments are matched against route handler return values by exact
// type first, then by assignability in registration order.
type TypedReturnHandler any

var contextType = inject.InterfaceOf((*Context)(nil))

type returnHandlers struct {
	handlers []typedReturnHandler
}

type typedReturnHandler struct {
	handler     TypedReturnHandler
	argTypes    []reflect.Type
	returnTypes []reflect.Type
}

func newReturnHandlers() *returnHandlers {
	hs := &returnHandlers{}
	hs.Register(func(c Context, body string) {
		writeReturnValue(c, reflect.ValueOf(body))
	})
	hs.Register(func(c Context, body []byte) {
		writeReturnValue(c, reflect.ValueOf(body))
	})
	hs.Register(func(c Context, err error) {
		writeReturnValue(c, reflect.ValueOf(err))
	})
	hs.Register(func(c Context, status int, body string) {
		c.ResponseWriter().WriteHeader(status)
		writeReturnValue(c, reflect.ValueOf(body))
	})
	hs.Register(func(c Context, status int, body []byte) {
		c.ResponseWriter().WriteHeader(status)
		writeReturnValue(c, reflect.ValueOf(body))
	})
	hs.Register(func(c Context, status int, err error) {
		c.ResponseWriter().WriteHeader(status)
		writeReturnValue(c, reflect.ValueOf(err))
	})
	hs.Register(func(c Context, body string, err error) {
		if err != nil {
			writeReturnValue(c, reflect.ValueOf(err))
			return
		}
		writeReturnValue(c, reflect.ValueOf(body))
	})
	hs.Register(func(c Context, body []byte, err error) {
		if err != nil {
			writeReturnValue(c, reflect.ValueOf(err))
			return
		}
		writeReturnValue(c, reflect.ValueOf(body))
	})
	return hs
}

func (hs *returnHandlers) Register(handler TypedReturnHandler) {
	typedHandler := newTypedReturnHandler(handler)

	for _, h := range hs.handlers {
		if sameTypes(h.returnTypes, typedHandler.returnTypes) {
			panic(fmt.Sprintf("return handler already registered for return values (%s)", formatTypes(typedHandler.returnTypes)))
		}
	}
	hs.handlers = append(hs.handlers, typedHandler)
}

func (hs *returnHandlers) Handle(c Context, vals []reflect.Value) {
	for _, h := range hs.handlers {
		if h.matches(vals, false) {
			h.invoke(c, vals)
			return
		}
	}
	for _, h := range hs.handlers {
		if h.matches(vals, true) {
			h.invoke(c, vals)
			return
		}
	}
}

func newTypedReturnHandler(handler TypedReturnHandler) typedReturnHandler {
	if handler == nil {
		panic("return handler must be a callable function, but got nil")
	}

	t := reflect.TypeOf(handler)
	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("return handler must be a callable function, but got %T", handler))
	}
	if t.NumOut() > 0 {
		panic("return handler must not return values")
	}
	if t.NumIn() == 0 || t.In(0) != contextType {
		panic("return handler must accept flamego.Context as its first argument")
	}

	h := typedReturnHandler{
		handler:  handler,
		argTypes: make([]reflect.Type, 0, t.NumIn()),
	}
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		h.argTypes = append(h.argTypes, argType)
		if i == 0 {
			continue
		}
		h.returnTypes = append(h.returnTypes, argType)
	}
	if len(h.returnTypes) == 0 {
		panic("return handler must accept at least one returned value after flamego.Context")
	}
	return h
}

func (h typedReturnHandler) matches(vals []reflect.Value, assignable bool) bool {
	if len(vals) != len(h.returnTypes) {
		return false
	}

	for i, val := range vals {
		if !val.IsValid() {
			return false
		}

		valType := val.Type()
		returnType := h.returnTypes[i]
		if valType == returnType {
			continue
		}
		if assignable && valType.AssignableTo(returnType) {
			continue
		}
		return false
	}
	return true
}

func (h typedReturnHandler) invoke(c Context, vals []reflect.Value) {
	args := make([]reflect.Value, 0, len(h.argTypes))
	returnIndex := 0
	for i := range h.argTypes {
		if i == 0 {
			args = append(args, reflect.ValueOf(c))
			continue
		}

		args = append(args, vals[returnIndex])
		returnIndex++
	}
	reflect.ValueOf(h.handler).Call(args)
}

func sameTypes(a, b []reflect.Type) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func formatTypes(types []reflect.Type) string {
	parts := make([]string, len(types))
	for i, typ := range types {
		parts[i] = typ.String()
	}
	return strings.Join(parts, ", ")
}

// ReturnHandler registers a handler for route handler return values.
//
// The handler must accept Context as its first argument. Its remaining
// arguments are matched against route handler return values by exact type first,
// then by assignability in registration order. For example, registering
// `func(Context, int, string)` handles route handlers that return `(int,
// string)`.
func (f *Flame) ReturnHandler(handler TypedReturnHandler) {
	f.returnHandlers.Register(handler)
}

func writeReturnValue(c Context, respVal reflect.Value) {
	if !respVal.IsValid() {
		return
	}

	w := c.ResponseWriter()
	if err, ok := respVal.Interface().(error); ok && err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if respVal.IsZero() {
		return
	}

	if canDeref(respVal) {
		respVal = respVal.Elem()
	}

	if isByteSlice(respVal) {
		_, _ = w.Write(respVal.Bytes())
	} else {
		_, _ = w.Write([]byte(respVal.String()))
	}
}

func canDeref(val reflect.Value) bool {
	return val.Kind() == reflect.Interface || val.Kind() == reflect.Pointer
}

func isByteSlice(val reflect.Value) bool {
	return val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.Uint8
}
