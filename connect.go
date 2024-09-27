package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
)

// completeClientRequest fulfills HTTPS clients with the original request.
func completeClientRequest(request *ProxyHTTPRequest, client net.Conn) error {
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
		MinVersion:               tls.VersionTLS10,
		MaxVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{tlsCert},
	}
	serverWithNewTLS := tls.Server(client, tlsConfig)
	defer serverWithNewTLS.Close()

	creader := bufio.NewReader(serverWithNewTLS)

	if request.Error != nil {
		serverWithNewTLS.Write([]byte(fmt.Sprintf("HTTP/1.1 403 Forbidden\nX-Proxyrequest-Id: %s\nContent-Length: %d\n\n%s", request.ID, len(request.Error.Error()), request.Error.Error())))
		return request.Error
	}

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

	request.RawHTTPRequest, _ = httputil.DumpRequest(req, true)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log(request, err)
		return fmt.Errorf("an error occured while doing the request: %s", err.Error())
	}
	resp.Header.Add("X-ProxyRequest-ID", request.ID)

	defer resp.Body.Close()
	resp.Write(serverWithNewTLS)
	return nil
}

// getTLSKeyPair returns a TLS Key Pair either from cache based on the host or generates
// a new one if the cache is unavailable or does not have it stored. It will automatically
// cache the certificate afterwards if possible.
func getTLSKeyPair(host string, cacert string, cakey string) (tls.Certificate, error) {
	ctx := context.Background()

	// get from cache
	if tlscert, err := getFromCache(ctx, host); err == nil {
		// found cert in cache
		return tlscert, nil
	}

	// make tls certificate (not cached or cache not available)
	cert, key, err := generateMITMCertificate(host, cacert, cakey)
	if err != nil {
		return tls.Certificate{}, err
	}

	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("an error occured parsing the public private key pair: %s", err.Error())
	}

	go setTLSCertToCache(ctx, host, tlsCert)
	return tlsCert, nil
}

// connectHTTP proxies HTTP requests. It is meant to handle all methods except
// CONNECT requests.
func connectHTTP(w http.ResponseWriter, request *ProxyHTTPRequest) error {
	if request.Error != nil {
		http.Error(w, request.Error.Error(), http.StatusBadRequest)
	}

	r := request.Req
	r.Header.Del("Proxy-Authorization")
	r.Header.Del("Proxy-Connection")
	r.RequestURI = ""

	request.Method = r.Method
	request.URL = r.URL
	request.RawHTTPRequest, _ = httputil.DumpRequest(r, true)

	resp, err := http.DefaultClient.Do(r)
	w.Header().Add("X-ProxyRequest-ID", request.ID)
	if err != nil {
		http.Error(w, "an error occured with the upstream client", 502)
		return err
	}
	resp.Write(w)
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

	err = completeClientRequest(request, conn)
	return err
}
