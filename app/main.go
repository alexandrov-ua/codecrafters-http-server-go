package main

import (
	"fmt"
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	srv := CreateServer()

	srv.Get("/", func(RequestContext) {})
	srv.Get("/echo/{str}", func(r RequestContext) {
		r.RespondWithStatusString(200, r.GetParam("str"))
	})
	srv.Get("/user-agent", func(r RequestContext) {
		r.RespondWithStatusString(200, r.GetHeader("User-Agent"))
	})

	srv.Listen("0.0.0.0:4221")

}
