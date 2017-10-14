// +build !cgi !fcgi

package main

import (
	"log"
	"net/http"
	"os"
)

func Serve(handler http.Handler) error {
	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = "127.0.0.1:3001"
	}
	log.Println("Listening on", addr)
	return http.ListenAndServe(addr, handler)
}
