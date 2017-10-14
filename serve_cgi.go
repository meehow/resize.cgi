// +build cgi

package main

import (
	"net/http"
	"net/http/cgi"
)

func Serve(handler http.Handler) error {
	return cgi.Serve(handler)
}
