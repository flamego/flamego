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

func TestValidateAndWrapHandler(t *testing.T) {
	t.Run("not a callable function", func(t *testing.T) {
		defer func() {
			assert.Contains(t, recover(), "handler must be a callable function")
		}()
		validateAndWrapHandler("string", nil)
	})

	handlers := []Handler{
		func(Context) {},
		func(http.ResponseWriter, *http.Request) {},
		http.HandlerFunc(nil),
		func() string { return "" },
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
