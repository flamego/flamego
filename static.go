// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// StaticOptions contains options for the flamego.Static middleware.
type StaticOptions struct {
	// Directory is the local directory to be used to serve static file. This value
	// is ignored when FileSystem is set. Default is "public".
	Directory string
	// FileSystem is the interface for supporting any implementation of the
	// http.FileSystem.
	FileSystem http.FileSystem
	// Prefix is the optional prefix used to serve the static directory content.
	Prefix string
	// Index specifies which file to attempt to serve as the directory index.
	// Default is "index.html".
	Index string
	// Expires is used to set the "Expires" response header for every static file
	// that is served. Default is not set.
	Expires func() string
	// SetETag indicates whether to compute and set "ETag" response header for every
	// static file that is served. File name, size and modification time are used to
	// compute the value.
	SetETag bool
	// EnableLogging indicates whether to print "[Static]" log messages whenever a
	// static file is served.
	EnableLogging bool
	// CacheControl is used to set the "Cache-Control" response header for every
	// static file that is served. Default is not set.
	CacheControl func() string
}

func generateETag(size int64, name string, modtime time.Time) string {
	value := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d%s%s", size, name, modtime.UTC().Format(http.TimeFormat))))
	return `"` + value + `"`
}

// Static returns a middleware handler that serves static files in the given
// directory.
func Static(opts ...StaticOptions) Handler {
	var opt StaticOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	parseStaticOptions := func(opts StaticOptions) StaticOptions {
		if opts.Directory == "" {
			opts.Directory = "public"
		}

		if opts.FileSystem == nil {
			opts.FileSystem = http.Dir(opts.Directory)
		}

		// Normalize the prefix when provided, we want a leading slash but no trailing
		// slash ("/").
		if opts.Prefix != "" {
			opts.Prefix = "/" + strings.Trim(opts.Prefix, "/")
		}

		if opts.Index == "" {
			opts.Index = "index.html"
		}
		return opts
	}

	opt = parseStaticOptions(opt)

	return LoggerInvoker(func(c Context, logger *log.Logger) {
		if c.Request().Method != http.MethodGet && c.Request().Method != http.MethodHead {
			return
		}

		file := c.Request().URL.Path
		// If we have a prefix set, filter files by stripping the prefix.
		if opt.Prefix != "" {
			if !strings.HasPrefix(file, opt.Prefix) {
				return
			}

			file = file[len(opt.Prefix):]
			if file != "" && file[0] != '/' {
				return
			}
		}

		// The go embed file system returns an error when the path ends with a slash or the path is empty.
		if file == "/" {
			file = "."
		} else {
			file = strings.TrimRight(file, "/")
		}

		f, err := opt.FileSystem.Open(file)
		if err != nil {
			return
		}
		defer func() { _ = f.Close() }()

		fi, err := f.Stat()
		if err != nil {
			return // File exists but failed to open.
		}

		// Try to serve index file.
		if fi.IsDir() {
			redirPath := path.Clean(c.Request().URL.Path)

			// The path.Clean removes the trailing slash, so we need to add it back when the
			// original path has it.
			if strings.HasSuffix(c.Request().URL.Path, "/") && !strings.HasSuffix(redirPath, "/") {
				redirPath += "/"
			}
			// Redirect if missing trailing slash.
			if !strings.HasSuffix(redirPath, "/") {
				http.Redirect(c.ResponseWriter(), c.Request().Request, redirPath+"/", http.StatusFound)
				return
			}

			file = path.Join(file, opt.Index)
			index, err := opt.FileSystem.Open(file)
			if err != nil {
				return
			}
			defer func() { _ = index.Close() }()

			fi, err = index.Stat()
			if err != nil || fi.IsDir() {
				return
			}

			_ = f.Close()
			f = index
		}

		if opt.EnableLogging {
			logger.Print("[Static] Serving", "file", file)
		}
		if opt.Expires != nil {
			c.ResponseWriter().Header().Set("Expires", opt.Expires())
		}
		if opt.CacheControl != nil {
			c.ResponseWriter().Header().Set("Cache-Control", opt.CacheControl())
		}

		if opt.SetETag {
			etag := generateETag(fi.Size(), fi.Name(), fi.ModTime())
			c.ResponseWriter().Header().Set("ETag", etag)
			if c.Request().Header.Get("If-None-Match") == etag {
				c.ResponseWriter().WriteHeader(http.StatusNotModified)
				return
			}
		}

		http.ServeContent(c.ResponseWriter(), c.Request().Request, file, fi.ModTime(), f)
	})
}
