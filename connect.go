package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
)

func sendHostConnToClientConn(request *ProxyHTTPRequest, client net.Conn) error {
	host, _, err := net.SplitHostPort(request.Host)
	if err != nil {
		return fmt.Errorf("an error occured while parsing the host: %s", err.Error())
	}

	tlsCert, err := getTLSKeyPair(host, env.CACERT, env.CAKEY)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{tlsCert},
	}
	serverWithNewTLS := tls.Server(client, tlsConfig)
	defer serverWithNewTLS.Close()

	creader := bufio.NewReader(serverWithNewTLS)
	for {
		req, err := http.ReadRequest(creader)
		if err == io.EOF { // connection broken
			return nil
		} else if err != nil {
			return fmt.Errorf("an error occured reading the http request tls connection with client: %s", err.Error())
		}
		request.Method = req.Method
		// prepare for sending request
		newURL := toURL(request.Host)
		newURL.Path = req.URL.Path
		newURL.RawQuery = req.URL.RawQuery

		req.URL = newURL
		req.RequestURI = ""

		request.URL = newURL

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logerror(request, err)
			return fmt.Errorf("an error occured while doing the request: %s", err.Error())
		}
		resp.Header.Add("X-ProxyRequest-ID", request.ID)

		defer resp.Body.Close()
		resp.Write(serverWithNewTLS)
	}
	return nil
}

// connectHTTPS proxies HTTPS requests with MITM certificates. It is meant to only
// handle CONNECT requests.
func connectHTTPS(w http.ResponseWriter, request *ProxyHTTPRequest) error {
	// request is ok (can't write Headers after hijacking)
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("hijacking failed")
	}
	// hijacking so we can directly write the intended server to connect to
	// to the client
	conn, _, err := hijacker.Hijack()
	if err != nil {
		return err
	}

	err = sendHostConnToClientConn(request, conn)
	return err
}
