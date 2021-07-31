// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/fatih/color"

	"github.com/flamego/flamego/inject"
)

var _ inject.FastInvoker = (*LoggerInvoker)(nil)

// LoggerInvoker is an inject.FastInvoker implementation of
// `func(ctx Context, log *log.Logger)`.
type LoggerInvoker func(ctx Context, log *log.Logger)

func (invoke LoggerInvoker) Invoke(params []interface{}) ([]reflect.Value, error) {
	invoke(params[0].(Context), params[1].(*log.Logger))
	return nil, nil
}

// LoggerOptions contains options for the flamego.Logger middleware.
type LoggerOptions struct {
	// LogTimeFormat specifies the time format for the logger. The default format is
	// "2006-01-02 15:04:05".
	LogTimeFormat string
}

// Logger returns a middleware handler that logs the request as it goes in and
// the response as it goes out.
func Logger(opts ...LoggerOptions) Handler {
	var opt LoggerOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	parseLoggerOptions := func(opts LoggerOptions) LoggerOptions {
		if opts.LogTimeFormat == "" {
			opts.LogTimeFormat = "2006-01-02 15:04:05"
		}
		return opts
	}

	opt = parseLoggerOptions(opt)

	colors := map[color.Attribute]*color.Color{
		color.FgGreen:   color.New(color.FgGreen),
		color.FgHiWhite: color.New(color.FgHiWhite),
		color.FgYellow:  color.New(color.FgYellow),
		color.FgRed:     color.New(color.FgRed, color.Underline),
		color.FgHiRed:   color.New(color.FgHiRed, color.Underline, color.Bold),
		color.FgBlue:    color.New(color.FgBlue),
	}

	return LoggerInvoker(func(ctx Context, log *log.Logger) {
		started := time.Now()

		log.Printf("%s: Started %s %s for %s",
			time.Now().Format(opt.LogTimeFormat),
			ctx.Request().Method,
			ctx.Request().RequestURI,
			ctx.RemoteAddr(),
		)

		w := ctx.ResponseWriter()
		ctx.Next()

		content := fmt.Sprintf("%s: Completed %s %s %v %s in %v",
			time.Now().Format(opt.LogTimeFormat),
			ctx.Request().Method,
			ctx.Request().RequestURI,
			w.Status(),
			http.StatusText(w.Status()),
			time.Since(started),
		)
		switch w.Status() {
		case http.StatusOK, http.StatusCreated, http.StatusAccepted:
			content = colors[color.FgGreen].Sprintf("%s", content)
		case http.StatusMovedPermanently, http.StatusFound:
			content = colors[color.FgHiWhite].Sprintf("%s", content)
		case http.StatusNotModified:
			content = colors[color.FgYellow].Sprintf("%s", content)
		case http.StatusUnauthorized, http.StatusForbidden:
			content = colors[color.FgRed].Sprintf("%s", content)
		case http.StatusNotFound:
			content = colors[color.FgBlue].Sprintf("%s", content)
		case http.StatusInternalServerError:
			content = colors[color.FgHiRed].Sprintf("%s", content)
		}
		log.Println(content)
	})
}
