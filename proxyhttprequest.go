package main

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ProxyHTTPRequest represents a request routed through
// the proxy.
type ProxyHTTPRequest struct {
	ID                 string
	ClientIP           string
	ProxyAuthorization string
	RawHTTPRequest     []byte
	RawHTTPResponse    []byte

	Method string
	Host   string
	Port   string
	URL    *url.URL

	Error error
	Req   *http.Request

	Start                *time.Time
	UpstreamResponseTime time.Duration
	Context              context.Context
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

// newProxyHTTPRequest parses a new incomplete *ProxyHTTPRequest from
// r.
func newProxyHTTPRequest(r *http.Request) (*ProxyHTTPRequest, error) {
	start := time.Now()

	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			host = r.Host
			if r.Method == "CONNECT" {
				port = "443"
			} else {
				port = "80"
			}
		} else {
			return nil, err
		}
	}

	return &ProxyHTTPRequest{
		Start:              &start,
		Req:                r,
		ID:                 generateTimeBasedID(start),
		ClientIP:           r.RemoteAddr,
		ProxyAuthorization: r.Header.Get("Proxy-Authorization"),
		Host:               host,
		Port:               port,
		Context:            r.Context(),
	}, nil
}
