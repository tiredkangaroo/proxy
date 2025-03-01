package main

import (
	"bufio"
	"log/slog"
	"net/http"
)

// handle fulfills the request. With HTTP clients, req is the original request
// retrieved from the original connection to the proxy server. With HTTPS clients, req is
// retrieved finishing the proxy-side of the connection (by writing a 200), retrieving a self-signed
// certificate issued to the original host, establishing a TLS connection and reading from it (the request
// originally intended).
func handle(proxyreq *Request) error {
	defer proxyreq.conn.Close()

	newURL, err := toURL(proxyreq.Host, proxyreq.https)
	if err != nil {
		slog.Error("toURL", "id", proxyreq.ID(), "error", err.Error(), "host", proxyreq.Host)
		proxyreq.conn.Write(InternalServerErrorResponse(proxyreq.ID()))
		return err
	}

	newURL.Path = proxyreq.HttpRequest.URL.Path
	newURL.RawQuery = proxyreq.HttpRequest.URL.RawQuery

	proxyreq.HttpRequest.URL = newURL
	proxyreq.HttpRequest.RequestURI = ""
	proxyreq.HttpRequest.Header.Del("Proxy-Authorization")
	proxyreq.HttpRequest.Header.Del("Proxy-Connection")

	resp, err := http.DefaultClient.Do(proxyreq.HttpRequest)
	if err != nil {
		slog.Error("http request do", "id", proxyreq.ID(), "error", err.Error())
		proxyreq.conn.Write(InternalServerErrorResponse(proxyreq.ID()))
		return err
	}

	defer resp.Body.Close()
	return resp.Write(proxyreq.conn)
}

// connectHTTP handles HTTP clients. It is equivlent to calling
// request.handle passing in the original HTTP request.
func connectHTTP(req *Request) error {
	return handle(req)
}

// connectHTTPS handles HTTPS requests with MITM certificates. It is meant to only
// handle CONNECT requests.
func connectHTTPS(request *Request) error {
	// get a TLS Certificate for the host (either from cache or create a new one)
	tlsCert, err := config.CertificateService.getTLSKeyPair(request.Context, request.Host)
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
	request.HttpRequest = req

	err = handle(request)
	return err
}
