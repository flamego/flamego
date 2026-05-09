// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flamego/flamego/inject"
)

var _ inject.FastInvoker = (*testHandlerFastInvoker)(nil)

type testHandlerFastInvoker func() string

func (invoke testHandlerFastInvoker) Invoke([]interface{}) ([]reflect.Value, error) {
	return []reflect.Value{reflect.ValueOf(invoke())}, nil
}

type testHTTPHandler struct{}

func (testHTTPHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func TestValidateAndWrapHandler(t *testing.T) {
	t.Run("not a callable function", func(t *testing.T) {
		defer func() {
			assert.Contains(t, recover(), "handler must be a callable function or http.Handler")
		}()
		validateAndWrapHandler("string", nil)
	})

	t.Run("nil handler", func(t *testing.T) {
		defer func() {
			assert.Contains(t, recover(), "handler must be a callable function or http.Handler")
		}()
		validateAndWrapHandler(nil, nil)
	})

	t.Run("nil http.Handler interface", func(t *testing.T) {
		defer func() {
			assert.Contains(t, recover(), "handler must be a callable function or http.Handler")
		}()
		var hh http.Handler
		validateAndWrapHandler(hh, nil)
	})

	handlers := []Handler{
		func(Context) {},
		func(http.ResponseWriter, *http.Request) {},
		http.HandlerFunc(nil),
		func() string { return "" },
		testHTTPHandler{},
	}
	for _, h := range handlers {
		t.Run("handlers", func(t *testing.T) {
			assert.False(t, inject.IsFastInvoker(h))

			h = validateAndWrapHandler(h,
				func(handler Handler) Handler {
					v, ok := h.(func() string)
					if ok {
						return testHandlerFastInvoker(v)
					}
					return h
				},
			)
			assert.True(t, inject.IsFastInvoker(h))
		})
	}
}
