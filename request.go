package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// Request represents a request routed through the proxy.
type Request struct {
	id   [16]byte
	conn net.Conn

	Host        string
	Port        string
	HttpRequest *http.Request
	Context     context.Context
}

// newProxyHTTPRequest parses a new incomplete *ProxyHTTPRequest from r.
func newProxyHTTPRequest(w http.ResponseWriter, r *http.Request) (*Request, error) {
	// get host and port
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

	// connect request?! (HTTPS)
	if r.Method == "CONNECT" {
		// only for CONNECT requests because the HTTPS client expects the proxy to leave after establishing a secure connection
		// for other method HTTP clients the proxy just does the request and sends it back, therefore writing the status code
		// would mask the actual status code from the upstream server and could cause other errors down the line
		w.WriteHeader(200)
	}

	conn, err := hijack(w)
	if err != nil {
		return nil, err
	}

	req := &Request{
		HttpRequest: r,
		Host:        host,
		Port:        port,
		Context:     r.Context(),
		conn:        conn,
	}
	// FIXME: handle rand error
	rand.Read(req.id[:])

	return req, nil
}

// hijack attempts to assert w as a http.Hijacker followed
// by using the hijacker to call the Hijack function. If the
// assertion or hijacking fails, it returns an error.
func hijack(w any) (net.Conn, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("hijacking failed: value passed cannot be asserted into a http.Hijacker")
	}
	conn, _, err := hijacker.Hijack()
	return conn, err
}

// ID returns the ID for the request. FIXME: handle errors +
// allow config.
func (r *Request) ID() string {
	return hex.EncodeToString(r.id[:])
}
