package main

import (
	"slices"
	"strings"
)

func (ctx *ServerContext) UseEncoding() {
	ctx.Use(EncodingMiddleware)
}

var supportedEncodings = []string{"gzip"}

func EncodingMiddleware(next middlewareFuncInternal, ctx *ServerContext, rctx *RequestContextImpl) {
	if ecncodingHeader, ok := rctx.requestHeaders["Accept-Encoding"]; ok {

		encodings := strings.Split(ecncodingHeader, ",")
		for _, enc := range encodings {
			enc = strings.Trim(enc, " ")
			if slices.Contains[[]string](supportedEncodings, enc) {
				next(ctx, rctx)
				rctx.responseHeaders["Content-Encoding"] = enc
				return
			}
		}
	}
	next(ctx, rctx)
}
