package main

import (
	"bufio"
	"fmt"
	"net/http"
)

// completeClientRequest fulfills HTTPS clients with the original request.
func (request *ProxyHTTPRequest) handle(req *http.Request) error {
	defer request.conn.Close()

	newURL, err := toURL(req.Host)
	if err != nil {
		return err
	}

	newURL.Path = req.URL.Path
	newURL.RawQuery = req.URL.RawQuery

	req.URL = newURL
	req.RequestURI = ""
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Proxy-Connection")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("an error occured while doing the request: %s", err.Error())
	}

	defer resp.Body.Close()
	resp.Write(request.conn)
	return nil
}

// connectHTTP proxies HTTP requests. It is meant to handle all methods except
// CONNECT requests.
func (request *ProxyHTTPRequest) connectHTTP() error {
	return request.handle(request.Req)
}

// connectHTTPS proxies HTTPS requests with MITM certificates. It is meant to only
// handle CONNECT requests.
func (request *ProxyHTTPRequest) connectHTTPS() error {
	// get a TLS Certificate for the host (either from cache or create a new one)
	tlsCert, err := env.CertificateService.getTLSKeyPair(request.Context, request.Host, env.CACERT, env.CAKEY)
	if err != nil {
		return err
	}

	// get a TLS connection with client as if this proxy was the original site to connect to
	tlsconn := addTLSToConnection(tlsCert, request.conn)
	request.conn = tlsconn

	// read the connection
	creader := bufio.NewReader(request.conn)

	req, err := http.ReadRequest(creader)
	if err != nil {
		return err
	}

	err = request.handle(req)
	return err
}
