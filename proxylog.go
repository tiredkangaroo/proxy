package main

import (
	"log/slog"
	"net/url"
	"time"
)

func getURL(u *url.URL) string {
	if u == nil {
		return "unknown"
	} else {
		return u.String()
	}
}

func log(request *ProxyHTTPRequest, err error) {
	request.Error = err
	request.CancelFunc()
	if err == nil {
		slog.Debug("OK", "request-id", request.ID, "url", getURL(request.URL), "time", time.Since(*request.Start))
	} else {
		slog.Error("BAD", "request-id", request.ID, "method", request.Method, "url", getURL(request.URL), "error", err, "time", time.Since(*request.Start))
	}
}
