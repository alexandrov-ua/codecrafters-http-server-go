package main

import "slices"

func (ctx *ServerContext) UseEncoding() {
	ctx.Use(EncodingMiddleware)
}

var supportedEncodings = []string{"gzip"}

func EncodingMiddleware(next middlewareFuncInternal, ctx *ServerContext, rctx *RequestContextImpl) {
	if enc, ok := rctx.requestHeaders["Accept-Encoding"]; ok {
		if slices.Contains[[]string](supportedEncodings, enc) {
			next(ctx, rctx)
			rctx.responseHeaders["Content-Encoding"] = enc
			return
		}
	}
	next(ctx, rctx)
}
