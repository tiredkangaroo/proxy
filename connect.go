package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

// completeClientRequest fulfills HTTPS clients with the original request.
func completeClientRequest(request *ProxyHTTPRequest, client net.Conn) error {
	// get a TLS Certificate for the host (either from cache or create a new one)
	tlsCert, err := getTLSKeyPair(request, env.CACERT, env.CAKEY)
	if err != nil {
		return err
	}

	// get a TLS connection with client as if this proxy was the original site to connect to
	conn := addTLSToConnection(tlsCert, client)
	defer conn.Close()

	// read the connection
	creader := bufio.NewReader(conn)

	// parse HTTP request out of connection
	req, err := http.ReadRequest(creader)
	if err == io.EOF { // connection broken
		return nil
	} else if err != nil {
		return fmt.Errorf("an error occured reading the http request tls connection with client: %s", err.Error())
	}

	// prepare for sending request
	newURL := toURL(request.Host)
	newURL.Path = req.URL.Path
	newURL.RawQuery = req.URL.RawQuery
	req.URL = newURL
	req.RequestURI = ""

	blocked := anyRegexMatch(env.BlockedSites, []byte(req.URL.String()))
	if blocked == true {
		_, err := conn.Write([]byte(ProxyBlockedResponse))
		return err
	}

	// set information about request
	request.Method = req.Method
	request.URL = newURL

	// dump the request with body
	request.RawHTTPRequest, _ = httputil.DumpRequest(req, true)

	// do the request
	upstreamStart := time.Now()
	resp, err := http.DefaultClient.Do(req)
	request.UpstreamResponseTime = time.Since(upstreamStart)

	if err != nil {
		log(request, err)
		return fmt.Errorf("an error occured while doing the request: %s", err.Error())
	}

	// add proxy request id
	resp.Header.Add("X-ProxyRequest-ID", request.ID)

	if env.LogInfo.RawHTTPResponse { // checking here instead of at logging because RawHTTP takes a lot of memory
		request.RawHTTPResponse, _ = httputil.DumpResponse(resp, env.LogInfo.RawHTTPResponseWithBody)
	}
	defer resp.Body.Close()
	resp.Write(conn)
	return nil
}

// connectHTTP proxies HTTP requests. It is meant to handle all methods except
// CONNECT requests.
func connectHTTP(w http.ResponseWriter, r *http.Request, request *ProxyHTTPRequest) error {
	// hijack because http by default decides to send headers for no reason
	conn, err := hijack(w)
	if err != nil {
		http.Error(w, "an error occured with the hijacking of the http connection", 400)
		return err
	}
	defer conn.Close()

	// remove proxy info
	r.Header.Del("Proxy-Authorization")
	r.Header.Del("Proxy-Connection")
	r.RequestURI = ""

	// infomation for logging
	request.Method = r.Method
	request.URL = r.URL

	blocked := anyRegexMatch(env.BlockedSites, []byte(r.URL.String()))
	if blocked == true {
		_, err := conn.Write([]byte(ProxyBlockedResponse))
		return err
	}

	if env.LogInfo.RawHTTPRequest { // checking here instead of at logging because RawHTTP takes a lot of memory
		request.RawHTTPRequest, _ = httputil.DumpRequest(r, env.LogInfo.RawHTTPRequestWithBody)
	}

	// do the request
	upstreamStart := time.Now()
	resp, err := http.DefaultClient.Do(r)
	request.UpstreamResponseTime = time.Since(upstreamStart)
	if err != nil {
		http.Error(w, "an error occured with the upstream client", 502)
		return err
	}

	if env.LogInfo.RawHTTPResponse { // checking here instead of at logging because RawHTTP takes a lot of memory
		request.RawHTTPResponse, _ = httputil.DumpResponse(resp, env.LogInfo.RawHTTPResponseWithBody)
	}

	// insert proxy request id
	resp.Header.Add("X-ProxyRequest-ID", request.ID)

	resp.Write(conn)
	return nil
}

// connectHTTPS proxies HTTPS requests with MITM certificates. It is meant to only
// handle CONNECT requests.
func connectHTTPS(w http.ResponseWriter, request *ProxyHTTPRequest) error {
	// request is ok (can't write Headers after hijacking)
	w.WriteHeader(http.StatusOK)
	conn, err := hijack(w)
	if err != nil {
		return err
	}

	err = completeClientRequest(request, conn)
	return err
}
