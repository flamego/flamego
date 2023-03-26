// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"reflect"
	"time"

	"github.com/charmbracelet/log"

	"github.com/flamego/flamego/inject"
)

var _ inject.FastInvoker = (*LoggerInvoker)(nil)

// LoggerInvoker is an inject.FastInvoker implementation of
// `func(ctx Context, log *log.Logger)`.
type LoggerInvoker func(ctx Context, logger *log.Logger)

func (invoke LoggerInvoker) Invoke(params []interface{}) ([]reflect.Value, error) {
	invoke(params[0].(Context), params[1].(*log.Logger))
	return nil, nil
}

// Logger returns a middleware handler that logs the request as it goes in and
// the response as it goes out.
func Logger() Handler {
	return LoggerInvoker(func(ctx Context, logger *log.Logger) {
		started := time.Now()

		logger = logger.WithPrefix("Logger")
		logger.Print("Started",
			"method", ctx.Request().Method,
			"path", ctx.Request().RequestURI,
			"remote", ctx.RemoteAddr(),
		)

		w := ctx.ResponseWriter()
		ctx.Next()

		logger.Print("Completed",
			"method", ctx.Request().Method,
			"path", ctx.Request().RequestURI,
			"status", w.Status(),
			"duration", time.Since(started),
		)
	})
}
