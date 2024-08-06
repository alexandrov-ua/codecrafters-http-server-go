package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var dirName = flag.String("directory", "/tmp/", "Dirictory name")
	flag.Parse()

	fmt.Println("Logs from your program will appear here!")
	srv := CreateServer()

	srv.Get("/", func(RequestContext) {})
	srv.Get("/echo/{str}", func(r RequestContext) {
		r.RespondWithStatusString(200, r.GetParam("str"))
	})
	srv.Get("/user-agent", func(r RequestContext) {
		r.RespondWithStatusString(200, r.GetHeader("User-Agent"))
	})
	srv.Get("/files/{fileName}", func(r RequestContext) {
		r.RespondWithStatusFile(200, *dirName+r.GetParam("fileName"))
	})
	srv.Post("/files/{fileName}", func(r RequestContext) {
		file, err := os.Create(*dirName + r.GetParam("fileName"))
		if err != nil {
			r.RespondWithStatusString(500, "Failed to create file")
			return
		}
		if _, err := file.WriteString(r.GetBody()); err != nil {
			r.RespondWithStatusString(500, "Failed to write file")
			return
		}
		r.RespondWithStatus(201)
	})

	srv.Listen("0.0.0.0:4221")

}
