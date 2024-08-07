package main

import (
	"os"
	"path"
	"strconv"
)

func (ctx *ServerContext) UseStaticFiles(url string, folderPath string, allowCreateOnPost bool) {
	ctx.Get(url+"{*path}", func(c RequestContext) {
		c.RespondWithStatusFile(200, path.Join(folderPath, c.GetParam("path")))
	})
	if allowCreateOnPost {
		ctx.Post(url+"{*path}", func(c RequestContext) {
			file, err := os.Create(path.Join(folderPath, c.GetParam("path")))
			if err != nil {
				c.RespondWithStatusString(500, "Failed to create file")
				return
			}
			if _, err := file.WriteString(c.GetBody()); err != nil {
				c.RespondWithStatusString(500, "Failed to write file")
				return
			}
			c.RespondWithStatus(201)
		})
	}
}

func (ctx *RequestContextImpl) RespondWithStatusFile(status int, path string) {

	if file, err := os.Open(path); err == nil {
		stat, _ := file.Stat()
		ctx.body = file
		ctx.status = status
		ctx.responseHeaders["Content-Type"] = "application/octet-stream"
		ctx.responseHeaders["Content-Length"] = strconv.Itoa(int(stat.Size()))
	} else {
		ctx.RespondWithStatusString(404, "File doesn't exists")
	}
}
