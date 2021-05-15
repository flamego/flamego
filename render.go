// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package flamego

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// Render is a thin wrapper to render content to the ResponseWriter.
type Render interface {
	// JSON encodes given value in JSON format with given status code to the
	// ResponseWriter.
	JSON(status int, v interface{})
	// XML encodes given value in XML format with given status code to the
	// ResponseWriter.
	XML(status int, v interface{})
	// Binary writes binary data with given status code to the ResponseWriter.
	Binary(status int, v []byte)
	// PlainText writes string with given status code to the ResponseWriter.
	PlainText(status int, s string)
}

type render struct {
	opts           RenderOptions
	responseWriter ResponseWriter // The ResponseWriter to write rendered content.
}

// RenderOptions contains options for the flamego.Renderer middleware.
type RenderOptions struct {
	// Charset specifies the value of the "charset" to be responded with the
	// "Content-Type" header. Default is "utf-8".
	Charset string
	// JSONIndent specifies the indent value when encoding content in JSON format.
	// Default is no indentation.
	JSONIndent string
	// XMLIndent specifies the indent value when encoding content in XML format.
	// Default is no indentation.
	XMLIndent string
}

func (r *render) JSON(status int, v interface{}) {
	r.responseWriter.Header().Set("Content-Type", "application/json; charset="+r.opts.Charset)
	r.responseWriter.WriteHeader(status)

	enc := json.NewEncoder(r.responseWriter)
	if r.opts.JSONIndent != "" {
		enc.SetIndent("", r.opts.JSONIndent)
	}

	err := enc.Encode(v)
	if err != nil {
		http.Error(r.responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (r *render) XML(status int, v interface{}) {
	r.responseWriter.Header().Set("Content-Type", "text/xml; charset="+r.opts.Charset)
	r.responseWriter.WriteHeader(status)

	enc := xml.NewEncoder(r.responseWriter)
	if r.opts.XMLIndent != "" {
		enc.Indent("", r.opts.XMLIndent)
	}

	err := enc.Encode(v)
	if err != nil {
		http.Error(r.responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (r *render) Binary(status int, v []byte) {
	r.responseWriter.Header().Set("Content-Type", "application/octet-stream")
	r.responseWriter.WriteHeader(status)
	_, _ = r.responseWriter.Write(v)
}

func (r *render) PlainText(status int, s string) {
	r.responseWriter.Header().Set("Content-Type", "text/plain; charset="+r.opts.Charset)
	r.responseWriter.WriteHeader(status)
	_, _ = r.responseWriter.Write([]byte(s))
}

// Renderer returns a middleware handler that injects flamego.Render into the
// request context, which is used for rendering content to the ResponseWriter.
func Renderer(opts ...RenderOptions) Handler {
	var opt RenderOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	parseRenderOptions := func(opts RenderOptions) RenderOptions {
		if opts.Charset == "" {
			opts.Charset = "utf-8"
		}
		return opts
	}

	opt = parseRenderOptions(opt)

	return ContextInvoker(func(c Context) {
		r := &render{
			opts:           opt,
			responseWriter: c.ResponseWriter(),
		}
		c.MapTo(r, (*Render)(nil))
	})
}
