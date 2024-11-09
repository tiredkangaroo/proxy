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
	request, err := newProxyHTTPRequest(w, r)
	if err != nil {
		env.Logger.Error("malformed request.", "error", err.Error())
		http.Error(w, fmt.Sprintf("Malformed request: %s.", err.Error()), http.StatusBadRequest)
		return
	}

	if r.Method == "CONNECT" {
		err = request.connectHTTPS()
	} else {
		err = request.connectHTTP()
	}

	if err != nil {
		data := fmt.Sprintf(InternalServerErrorHTML, err.Error())
		response := []byte(fmt.Sprintf(InternalServerErrorResponse, len(data), data))
		request.conn.Write(response)
	}
}

func main() {
	load()

	handler := new(CustomHandler)
	http.ListenAndServe(":8000", handler)
}
