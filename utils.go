package main

import (
	"net/url"
	"strings"
)

// toURL takes in a string and parses it using url.Parse. If the the original string did not specify the
// http/https prefix, it adds it before parsing, allowing the parse to be valid.
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
