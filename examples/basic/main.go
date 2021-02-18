package main

import (
	"net/http"

	"github.com/usvc/go-server"
)

func main() {
	options := server.NewHTTPOptions()
	mux := http.NewServeMux()
	s := server.NewHTTP(options, mux)
	s.Start()
}
