// +build fcgi

package main

import (
	"net/http"
	"net/http/fcgi"
)

func Serve(handler http.Handler) error {
	return fcgi.Serve(handler, nil)
}
