package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// Constants defining delay durations
const AccessBeforeDelayTime = time.Minute * 30
const DelayTime = time.Second * 10

// Delay holds the hostname and a cached response after delay completion.
type Delay struct {
	hostname string
	response *http.Response
}

// DelayedAccess records the time and hostname for a delayed access per client IP.
type DelayedAccess struct {
	accessedAt time.Time
	hostname   string
}

// DelayHandler manages the delay logic for specific hosts.
type DelayHandler struct {
	lastAccessedByClient map[string][]DelayedAccess // maps client IP to list of past delayed accesses
	validDelays          map[string]Delay           // maps delay ID to Delay object
	delayedHosts         []string                   // list of hosts to delay
}

// Start initializes delay handling with predefined delayed hosts.
func (dh *DelayHandler) Start() error {
	dh.lastAccessedByClient = make(map[string][]DelayedAccess)
	dh.validDelays = make(map[string]Delay)
	dh.delayedHosts = []string{"www.youtube.com", "www.instagram.com", "www.google.com"}
	return nil
}

// isDelayed checks if the given hostname is in the list of delayed hosts.
func (dh *DelayHandler) isDelayed(hostname string) bool {
	for _, delayedHost := range dh.delayedHosts {
		if delayedHost == hostname {
			return true
		}
	}
	return false
}

// Handle processes an HTTP request and applies delay if necessary.
func (dh *DelayHandler) Handle(req *http.Request, conn net.Conn) (*http.Response, error) {
	hostname := req.URL.Hostname()
	clientIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
	delayID := req.URL.Query().Get("delay-id")

	// Check if there's a cached delay response for this delay ID
	if delay, exists := dh.validDelays[delayID]; exists && delay.hostname == hostname {
		// Cached response found; delete it from cache and return it
		delete(dh.validDelays, delayID)
		return delay.response, nil
	}

	// Make request to the original host
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Return response immediately if the host is not in the delayed list
	if !dh.isDelayed(hostname) {
		return response, nil
	}

	// Skip delay if the content is not HTML (i.e., non-user-facing content)
	if !strings.Contains(response.Header.Get("Content-Type"), "text/html") {
		return response, nil
	}

	// Retrieve client's past delayed access records for this hostname
	clientAccesses, exists := dh.lastAccessedByClient[clientIP]
	if !exists {
		clientAccesses = []DelayedAccess{}
	}

	// Check if client has accessed this hostname within the allowed time window
	for _, access := range clientAccesses {
		if access.hostname == hostname && time.Since(access.accessedAt) <= AccessBeforeDelayTime {
			// Access is within allowed time; return response without delay
			return response, nil
		}
	}

	// Generate a unique delay ID using SHA-1 hash for the request
	rawDelayID := sha1.Sum([]byte(req.URL.String() + time.Now().String() + conn.RemoteAddr().String()))
	delayID = hex.EncodeToString(rawDelayID[:])

	// Update the request URL with the new delay ID
	query := req.URL.Query()
	query.Add("delay-id", delayID)
	req.URL.RawQuery = query.Encode()

	// Generate HTML page with JavaScript delay for the client
	delayPageContent := []byte(fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Delayed Request</title>
				<script>
					setTimeout(() => {
           				window.location.href = "%s";
					}, %d);
                </script>
            </head>
            <body>
            	<h1>Delayed Request</h1>
             	<pre>This request has been delayed for %s. Once you finish waiting, you will be granted access for %s.</pre>
            </body>
        </html>
	`, req.URL.String(), DelayTime.Milliseconds(), DelayTime.String(), AccessBeforeDelayTime.String()))

	// Record the delay access for the client
	dh.lastAccessedByClient[clientIP] = append(clientAccesses, DelayedAccess{
		accessedAt: time.Now(),
		hostname:   hostname,
	})

	// Cache the response to serve it after delay completion
	dh.validDelays[delayID] = Delay{
		hostname: hostname,
		response: response,
	}

	// Return the generated HTML delay page as the HTTP response
	return &http.Response{
		Status:        "200 OK",
		StatusCode:    http.StatusOK,
		Header:        http.Header{"Content-Type": []string{"text/html"}},
		ContentLength: int64(len(delayPageContent)),
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          io.NopCloser(bytes.NewBuffer(delayPageContent)),
	}, nil
}
