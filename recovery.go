// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/charmbracelet/log"

	"github.com/flamego/flamego/inject"
)

// Recovery returns a middleware handler that recovers from any panics and
// writes a 500 status code to the response if there was one. While in
// development mode (EnvTypeDev), Recovery will also output the panic as HTML.
func Recovery() Handler {
	const html = `<html>
<head><title>PANIC: %[1]s</title>
<meta charset="utf-8" />
<style type="text/css">
html, body {
	font-family: "Roboto", sans-serif;
	color: #333333;
	background-color: #ea5343;
	margin: 0px;
}
h1 {
	color: #d04526;
	background-color: #ffffff;
	padding: 20px;
	border-bottom: 1px dashed #2b3848;
}
pre {
	margin: 20px;
	padding: 20px;
	border: 2px solid #2b3848;
	background-color: #ffffff;
	white-space: pre-wrap;       /* css-3 */
	white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
	white-space: -pre-wrap;      /* Opera 4-6 */
	white-space: -o-pre-wrap;    /* Opera 7 */
	word-wrap: break-word;       /* Internet Explorer 5.5+ */
}
</style>
</head><body>
<h1>PANIC</h1>
<pre style="font-weight: bold;">%[1]s</pre>
<pre>%[2]s</pre>
</body>
</html>`

	var (
		dunno     = []byte("???")
		centerDot = []byte("·")
		dot       = []byte(".")
		slash     = []byte("/")
	)

	// source returns a space-trimmed slice of the n'th line.
	source := func(lines [][]byte, n int) []byte {
		n-- // In a stack trace, lines are 1-indexed but our array is 0-indexed
		if n < 0 || n >= len(lines) {
			return dunno
		}
		return bytes.TrimSpace(lines[n])
	}

	// function returns, if possible, the name of the function containing the PC.
	function := func(pc uintptr) []byte {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			return dunno
		}
		name := []byte(fn.Name())
		// The name includes the path name to the package, which is unnecessary since
		// the file name is already included. Plus, it has center dots. That is, we see:
		//	runtime/debug.*T·ptrmethod
		// and want:
		//	*T.ptrmethod
		// Also the package path might contains dot (e.g. code.google.com/...), so first
		// eliminate the path prefix.
		if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
			name = name[lastSlash+1:]
		}
		if period := bytes.Index(name, dot); period >= 0 {
			name = name[period+1:]
		}
		name = bytes.ReplaceAll(name, centerDot, dot)
		return name
	}

	// stack returns a nicely formated stack frame, skipping skip frames
	stack := func(skip int) []byte {
		buf := new(bytes.Buffer)
		// As we loop, we open files and read them. These variables record the currently
		// loaded file.
		var lines [][]byte
		var lastFile string
		for i := skip; ; i++ { // Skip the expected number of frames
			pc, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			// Print this much at least.  If we can't find the source, it won't show.
			_, _ = fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
			if file != lastFile {
				data, err := os.ReadFile(file)
				if err != nil {
					continue
				}
				lines = bytes.Split(data, []byte{'\n'})
				lastFile = file
			}
			_, _ = fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
		}
		return buf.Bytes()
	}

	return LoggerInvoker(func(c Context, logger *log.Logger) {
		defer func() {
			if err := recover(); err != nil {
				stack := stack(3)
				logger.Error(fmt.Sprintf("PANIC: %s\n%s", err, stack))

				// Lookup the current ResponseWriter
				val := c.Value(inject.InterfaceOf((*http.ResponseWriter)(nil)))
				w := val.Interface().(http.ResponseWriter)

				// Respond with panic message only in development mode
				var body []byte
				if Env() == EnvTypeDev {
					w.Header().Set("Content-Type", "text/html")
					body = []byte(fmt.Sprintf(html, err, stack))
				} else {
					w.Header().Set("Content-Type", "text/plain")
					body = []byte(http.StatusText(http.StatusInternalServerError))
				}

				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write(body)
			}
		}()

		c.Next()
	})
}
