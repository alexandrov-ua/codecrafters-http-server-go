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
		if fileText, err := os.ReadFile(*dirName + r.GetParam("fileName")); err == nil {
			r.RespondWithStatusString(200, string(fileText))
		} else {
			r.RespondWithStatusString(404, "File doesn't exists")
		}
	})

	srv.Listen("0.0.0.0:4221")

}
