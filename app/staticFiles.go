package main

import (
	"mime"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func (ctx *ServerContext) UseStaticFiles(urlPrefix string, settings StaticFilesSettings) {

	ctx.Use(func(next middlewareFuncInternal, ctx *ServerContext, rctx *RequestContextImpl) {
		StaticFilesMiddleware(next, ctx, rctx, settings, urlPrefix)
	})

}

type StaticFilesSettings struct {
	FolderPath  string
	AllowUpload bool
	EnableMime  bool
}

func StaticFilesMiddleware(next middlewareFuncInternal, ctx *ServerContext, rctx *RequestContextImpl, settings StaticFilesSettings, urlPrefix string) {
	if strings.HasPrefix(rctx.path, urlPrefix) {
		tmp, _ := strings.CutPrefix(rctx.path, urlPrefix)
		filePath := path.Join(settings.FolderPath, tmp)
		switch rctx.method {
		case "GET":
			HandleDownload(rctx, filePath, settings.EnableMime)
		case "POST":
			HandleUpload(rctx, filePath)
		default:
			rctx.RespondWithStatus(405)
		}
	} else {
		next(ctx, rctx)
	}
}

func HandleUpload(ctx *RequestContextImpl, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		ctx.RespondWithStatusString(500, "Failed to create file")
		return
	}
	if _, err := file.WriteString(ctx.GetBody()); err != nil {
		ctx.RespondWithStatusString(500, "Failed to write file")
		return
	}
	ctx.RespondWithStatus(201)
}

func HandleDownload(ctx *RequestContextImpl, path string, useMime bool) {

	if file, err := os.Open(path); err == nil {
		stat, _ := file.Stat()
		ctx.body = file
		ctx.status = 200
		ctx.responseHeaders["Content-Type"] = "application/octet-stream"
		ctx.responseHeaders["Content-Length"] = strconv.Itoa(int(stat.Size()))
		if useMime {
			if ct := mime.TypeByExtension(filepath.Ext(path)); ct != "" {
				ctx.responseHeaders["Content-Type"] = ct
			}

		}
	} else {
		ctx.RespondWithStatusString(404, "File doesn't exists")
	}
}
