package main

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type User struct {
	Username string
	Password string
}

type ProxyHTTPRequest struct {
	ID                 string
	ClientIP           string
	ProxyAuthorization string

	Method string
	Host   string
	URL    *url.URL

	Error error
	Req   *http.Request

	Start      *time.Time
	Context    context.Context
	CancelFunc context.CancelFunc
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
	ctx, cancel := context.WithCancel(r.Context())
	return &ProxyHTTPRequest{
		Start:              &start,
		Req:                r,
		ID:                 generateTimeBasedID(start),
		ClientIP:           r.RemoteAddr,
		ProxyAuthorization: r.Header.Get("Proxy-Authorization"),
		Host:               r.Host,
		Context:            ctx,
		CancelFunc:         cancel,
	}, nil
}
