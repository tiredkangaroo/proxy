package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

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

// toURL takes in a string and parses it using url.Parse.
// If the the original string did not specify the http/https prefix,
// it adds it before parsing, allowing the parse to be valid.
func toURL(s string) (*url.URL, error) {
	// prepare for sending request
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "https://" + s
	}
	return url.Parse(s)
}

// acquire attempts to acquire a lock with the function of the lock passed
// in. it will close the channel once the lock has been successfully acquired.
func acquire(lockfunc func()) chan struct{} {
	c := make(chan struct{})
	go func() {
		lockfunc()
		c <- struct{}{}
		close(c)
	}()
	return c
}
