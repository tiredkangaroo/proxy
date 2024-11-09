package main

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// ProxyHTTPRequest represents a request routed through
// the proxy.
type ProxyHTTPRequest struct {
	Host    string
	Port    string
	Req     *http.Request
	Context context.Context

	conn net.Conn
}

// newProxyHTTPRequest parses a new incomplete *ProxyHTTPRequest from
// r.
func newProxyHTTPRequest(w http.ResponseWriter, r *http.Request) (*ProxyHTTPRequest, error) {
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

	w.WriteHeader(200)
	conn, err := hijack(w)
	if err != nil {
		return nil, err
	}

	return &ProxyHTTPRequest{
		Req:     r,
		Host:    host,
		Port:    port,
		Context: r.Context(),
		conn:    conn,
	}, nil
}
