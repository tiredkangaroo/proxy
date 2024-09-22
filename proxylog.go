package main

import (
	"log/slog"
	"time"
)

func log(request *ProxyHTTPRequest) {
	slog.Debug("good", "request-id", request.ID, "url", request.URL.String(), "time", time.Since(*request.Start))
}

func logerror(request *ProxyHTTPRequest, err error) {
	var requrl string
	if request.URL == nil {
		requrl = "unknown"
	} else {
		requrl = request.URL.String()
	}
	slog.Error("bad", "request-id", request.ID, "method", request.Method, "url", request.URL.String(), "error", err, "time", time.Since(*request.Start))
}
