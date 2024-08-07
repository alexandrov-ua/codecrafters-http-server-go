package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
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
	size, _ := strconv.Atoi(rctx.responseHeaders["Content-Length"])
	buf := bytes.NewBuffer(make([]byte, size))
	w := bufio.NewWriter(gzip.NewWriter(buf))

	w.ReadFrom(rctx.body)
	w.Flush()

	rctx.body = bytes.NewReader(buf.Bytes())
	rctx.responseHeaders["Content-Length"] = strconv.Itoa(len(buf.Bytes()))
}
