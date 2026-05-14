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
type TypedReturnHandler interface{}

var contextType = inject.InterfaceOf((*Context)(nil))

type returnHandlers struct {
	fallback ReturnHandler
	handlers []typedReturnHandler
}

type typedReturnHandler struct {
	handler     TypedReturnHandler
	argTypes    []reflect.Type
	returnTypes []reflect.Type
}

func newReturnHandlers(fallback ReturnHandler) *returnHandlers {
	return &returnHandlers{fallback: fallback}
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
	handler, ok, fallback := hs.match(vals)
	if ok {
		handler.invoke(c, vals)
		return
	}

	if fallback != nil {
		fallback(c, vals)
	}
}

func (hs *returnHandlers) match(vals []reflect.Value) (typedReturnHandler, bool, ReturnHandler) {
	for _, h := range hs.handlers {
		if h.matches(vals, false) {
			return h, true, hs.fallback
		}
	}
	for _, h := range hs.handlers {
		if h.matches(vals, true) {
			return h, true, hs.fallback
		}
	}
	return typedReturnHandler{}, false, hs.fallback
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

func defaultReturnHandler() ReturnHandler {
	canDeref := func(val reflect.Value) bool {
		return val.Kind() == reflect.Interface || val.Kind() == reflect.Pointer
	}

	isByteSlice := func(val reflect.Value) bool {
		return val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.Uint8
	}

	return func(c Context, vals []reflect.Value) {
		v := c.Value(inject.InterfaceOf((*http.ResponseWriter)(nil)))
		w := v.Interface().(http.ResponseWriter)
		var respVal reflect.Value

		switch len(vals) {
		case 1: // string, []byte, error
			respVal = vals[0]

		case 2:
			// (int, string), (int, []byte), (int, error)
			if vals[0].Kind() == reflect.Int {
				w.WriteHeader(int(vals[0].Int()))
				respVal = vals[1]
				break
			}

			// (string, error), ([]byte, error)
			if vals[0].Kind() == reflect.String || isByteSlice(vals[0]) {
				respVal = vals[0]
				if _, ok := vals[1].Interface().(error); ok {
					respVal = vals[1]
				}
				break
			}
		}

		if !respVal.IsValid() {
			return
		}

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
}
