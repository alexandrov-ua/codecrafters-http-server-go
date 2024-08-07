package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"strconv"
	"strings"
)

func (ctx *ServerContext) UseEncoding() {
	ctx.Use(EncodingMiddleware)
}

var supportedEncodings = map[string]func(*RequestContextImpl){"gzip": HandleGzipCompression}

func EncodingMiddleware(next middlewareFuncInternal, ctx *ServerContext, rctx *RequestContextImpl) {
	if ecncodingHeader, ok := rctx.requestHeaders["Accept-Encoding"]; ok {

		encodings := strings.Split(ecncodingHeader, ",")
		for _, enc := range encodings {
			enc = strings.Trim(enc, " ")
			if prov, ok := supportedEncodings[enc]; ok {
				next(ctx, rctx)
				prov(rctx)
				return
			}
		}
	}
	next(ctx, rctx)
}

func HandleGzipCompression(rctx *RequestContextImpl) {
	rctx.responseHeaders["Content-Encoding"] = "gzip"
	buf := bytes.Buffer{}
	w := gzip.NewWriter(&buf)

	io.Copy(w, rctx.body)
	w.Flush()
	w.Close()

	rctx.body = bytes.NewReader(buf.Bytes())
	rctx.responseHeaders["Content-Length"] = strconv.Itoa(len(buf.Bytes()))
}
