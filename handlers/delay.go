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

const AccessBeforeDelayTime time.Duration = time.Minute * 30
const DelayTime time.Duration = time.Second * 10

type Delay struct {
	hostname string
	response *http.Response
}

type DelayedAccess struct {
	on       time.Time
	hostname string
}

type DelayHandler struct {
	lastDelayedAccess map[string][]DelayedAccess // clientip: DelayedAccess
	validDelays       map[string]Delay
	delayedhosts      []string
}

func (dh *DelayHandler) Start() error {
	dh.lastDelayedAccess = make(map[string][]DelayedAccess)
	dh.validDelays = make(map[string]Delay)
	dh.delayedhosts = []string{"www.youtube.com", "www.instagram.com", "www.google.com"}
	return nil
}

func (dh *DelayHandler) isDelayed(hostname string) bool {
	for _, host := range dh.delayedhosts {
		if host == hostname {
			return true
		}
	}
	return false
}

func (dh *DelayHandler) Handle(req *http.Request, conn net.Conn) (*http.Response, error) {
	hostname := req.URL.Hostname()
	clientip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	delayid := req.URL.Query().Get("delay-id")
	delay, ok := dh.validDelays[delayid]
	if delay.hostname == hostname { // case: delayed host but client JUST served the delay
		resp := delay.response
		delete(dh.validDelays, delayid)
		return resp, nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if !dh.isDelayed(hostname) {
		return resp, nil
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") { // case: delayed host but response is not for the user
		return resp, nil
	}

	delayedaccess, ok := dh.lastDelayedAccess[clientip]
	if !ok { // case: delayed host but client
		dh.lastDelayedAccess[clientip] = delayedaccess
	}
	for _, access := range delayedaccess {
		if access.hostname == hostname {
			if time.Since(access.on) <= AccessBeforeDelayTime { // case: delayed host but client has served delay prior
				return resp, nil
			}
			break // break because we found the hostname and client is yet to serve delay
		}
	}
	// case: delayed host but client is yet to serve delay

	rawdelayid := sha1.Sum([]byte(req.URL.String() + time.Now().String() + conn.RemoteAddr().String()))
	delayid = hex.EncodeToString(rawdelayid[:])

	qr := req.URL.Query()
	qr.Add("delay-id", delayid)
	req.URL.RawQuery = qr.Encode()

	data := []byte(fmt.Sprintf(`
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
             	<pre>This request has been delayed for %d seconds. Once you finish waiting, you will be granted access for %d minutes.</pre>
            </body>
        </html>
	`, req.URL.String(), DelayTime.Milliseconds(), int(DelayTime.Seconds()), int(AccessBeforeDelayTime.Minutes())))

	dh.lastDelayedAccess[clientip] = append(dh.lastDelayedAccess[clientip], DelayedAccess{
		on:       time.Now(),
		hostname: hostname,
	})

	dh.validDelays[delayid] = Delay{
		hostname: hostname,
		response: resp,
	}

	return &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Header:        http.Header{"Content-Type": []string{"text/html"}},
		ContentLength: int64(len(data)),
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          io.NopCloser(bytes.NewBuffer(data)),
	}, nil
}
