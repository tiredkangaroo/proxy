package main

import (
	"bufio"
	"fmt"
	"net/http"
)

// handle fulfills HTTP and HTTPS client requests by recieving the
// HTTP request, requesting the original host server, and writing
// it back to the connection. The connection refers to the original
// HTTP connection for HTTP clients and refers to the post-CONNECT
// request and after establishing a TLS connection.
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

// connectHTTP handles HTTP clients. It is equivlent to calling
// request.handle passing in the original HTTP request.
func (request *ProxyHTTPRequest) connectHTTP() error {
	return request.handle(request.Req)
}

// connectHTTPS handles HTTPS requests with MITM certificates. It is meant to only
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
