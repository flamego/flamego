// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"net/http"
	"reflect"

	"github.com/flamego/flamego/inject"
)

// ReturnHandler is a service that is called when a route handler returns
// something. The ReturnHandler is responsible for writing to the ResponseWriter
// based on the values that are passed into this function.
type ReturnHandler func(Context, []reflect.Value)

func defaultReturnHandler() ReturnHandler {
	canDeref := func(val reflect.Value) bool {
		return val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr
	}

	checkError := func(val reflect.Value) (error, bool) {
		err, ok := val.Interface().(error)
		return err, ok
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

			if err, ok := checkError(respVal); ok && err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

		case 2: // (int, string), (int, []byte)
			if vals[0].Kind() != reflect.Int {
				break
			}

			w.WriteHeader(int(vals[0].Int()))
			respVal = vals[1]
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
