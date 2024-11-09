package main

import (
	"bufio"
	"net"
	"net/http"
)

// ResponseHandler defines an interface of an extension that
// handles response for the proxy.
type ResponseHandler interface {
	// Start
	Start() error
	Handle(*http.Request, net.Conn) (*http.Response, error)
}

// DefaultResponseHandler is an implementation of ResponseHandler.
type DefaultResponseHandler struct {
}

// Start is an implementation of ResponseHandler Start function.
func (_ *DefaultResponseHandler) Start() error {
	return nil
}

// Handle is an implementation of ResponseHandler's Handle function. This implementation is
// equivlent to calling http.DefaultClient.Do, while passing in req.
func (_ *DefaultResponseHandler) Handle(req *http.Request, _ net.Conn) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

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

	resp, err := env.ResponseHandler.Handle(req, request.conn)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return resp.Write(request.conn)
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
