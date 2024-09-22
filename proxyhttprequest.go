package main

import (
	"net/http"
	"net/url"
	"time"
)

type ProxyHTTPRequest struct {
	ID                 string
	HTTP               string
	Method             string
	Host               string
	URL                *url.URL
	UserAgent          string
	ProxyAuthorization string
	ClientIP           string

	Start *time.Time
}

// func (phr *ProxyHTTPRequest) HTTPRequest() *http.Request {
// 	return &http.Request{
// 		Method: phr.Method,
// 		URL:    phr.Host,
// 		Header: http.Header{
// 			"User-Agent": []string{phr.UserAgent},
// 		},
// 	}
// }

func parseRequest(r *http.Request) (*ProxyHTTPRequest, error) {
	start := time.Now()
	return &ProxyHTTPRequest{
		ID:                 generateTimeBasedID(start),
		Start:              &start,
		Host:               r.Host,
		UserAgent:          r.UserAgent(),
		ProxyAuthorization: r.Header.Get("Proxy-Authorization"),
		ClientIP:           r.RemoteAddr,
	}, nil
}
