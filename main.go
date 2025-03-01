package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
)

// CustomHandler provides an http.Handler in which to accept ALL request
// methods.
type CustomHandler struct{}

// ServeHTTP serves the HTTP server for the proxy.
func (CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request, err := newProxyHTTPRequest(w, r)
	if err != nil {
		config.Logger.Error("malformed request", "error", err.Error())
		http.Error(w, fmt.Sprintf("malformed request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if r.Method == "CONNECT" {
		err = connectHTTPS(request)
	} else {
		err = connectHTTP(request)
	}

	if err != nil {
		slog.Error("connect error", "request id", request.ID(), "err", err.Error())
		request.conn.Write(InternalServerErrorResponse(request.ID()))
	}
}

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf(err.Error())
	}

	config.Logger.Debug("starting proxy", "port", config.port)

	handler := new(CustomHandler)
	if err := http.ListenAndServe(config.port, handler); err != nil {
		log.Fatalf(err.Error())
	}
}
