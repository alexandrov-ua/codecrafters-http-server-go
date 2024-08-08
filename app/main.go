package main

import (
	"flag"
	"fmt"
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
		r.RespondWithStatusString(200, r.GetHeader("User-Agent").GetFirst())
	})
	srv.UseStaticFiles("/files/", StaticFilesSettings{
		FolderPath:  *dirName,
		AllowUpload: true,
		EnableMime:  true,
	})
	srv.UseEncoding()

	srv.Listen("0.0.0.0:4221")

}
