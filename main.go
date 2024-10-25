package main

import (
	"fmt"
	"net/http"
)

// CustomHandler provides an http.Handler in which to accept ALL request
// methods.
type CustomHandler struct{}

// ServeHTTP serves the HTTP server for the proxy.
func (_ CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request, err := newProxyHTTPRequest(r)
	if err != nil {
		env.Logger.Error("malformed request.", "error", err.Error())
		http.Error(w, fmt.Sprintf("Malformed request: %s.", err.Error()), http.StatusBadRequest)
		return
	}

	if r.Method == "CONNECT" {
		log(request, connectHTTPS(w, request))
	} else {
		log(request, connectHTTP(w, r, request))
	}
}

func main() {
	load()
	// start api and proxy servers
	go startAPI()
	handler := new(CustomHandler)
	http.ListenAndServe(":8000", handler)
}
