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
		data := fmt.Sprintf("<html><body><h1>Internal Server Error</h1><pre>%s</pre></body></html>", err.Error())
		response := []byte("HTTP/1.1 500 Internal Server Error\r\n" +
			"Content-Type: text/html\r\n" +
			fmt.Sprintf("Content-Length: %d\r\n", len(data)) +
			"\r\n" +
			data)
		request.conn.Write(response)
	}
}

func main() {
	load()

	handler := new(CustomHandler)
	http.ListenAndServe(":8000", handler)
}
