package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base32"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// toURL takes in a string an makes it into a *url.URL.
func toURL(s string) *url.URL {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "https://" + s
	}
	u, _ := url.Parse(s)
	return u
}

// random16Base32Characters generates 16 random base 32 characters.
func random16Base32Characters() (string, error) {
	r := make([]byte, 10)
	_, err := rand.Read(r)
	if err != nil {
		return "", fmt.Errorf(RandomGenerationFailed, err.Error())
	}
	e := base32.StdEncoding.EncodeToString(r)
	return e, nil
}

// generateTimeBasedID generates a time-based ID with 10 random bytes
// encoded in Base32
// in this format:
//
// UNIX_TIME-RandomBase32Characters
func generateTimeBasedID(t time.Time) string {
	e, err := random16Base32Characters()
	if err != nil {
		env.Logger.Error(err.Error())
	}
	return fmt.Sprintf("%d-%s", t.UnixMilli(), e)
}

// marshalTLSCertificate marshals a tls.Certificate into an []byte.
func marshalTLSCertificate(cert tls.Certificate) ([]byte, error) {
	// encode cert
	var certPEMBuffer bytes.Buffer
	for _, b := range cert.Certificate {
		pem.Encode(&certPEMBuffer, &pem.Block{Type: "CERTIFICATE", Bytes: b})
	}

	// encode private key
	var keyPEMBuffer bytes.Buffer
	kb, err := x509.MarshalECPrivateKey(cert.PrivateKey.(*ecdsa.PrivateKey))
	if err != nil {
		return []byte{}, fmt.Errorf("an error occured while marshalling the ecdsa private key: %s", err.Error())
	}
	pem.Encode(&keyPEMBuffer, &pem.Block{Type: "EC PRIVATE KEY", Bytes: kb})

	return json.Marshal(map[string][]byte{
		"cert": certPEMBuffer.Bytes(),
		"key":  keyPEMBuffer.Bytes(),
	})
}

// unmarshalTLSCertificate unmarshals data (as an []byte) into a tls.Certificate.
func unmarshalTLSCertificate(data []byte) (tls.Certificate, error) {
	var d map[string][]byte
	err := json.Unmarshal(data, &d)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("an error occured while unmarshalling the tls certificate: %s", err.Error())
	}

	cert := d["cert"]
	key := d["key"]

	certblock, _ := pem.Decode(cert)
	if certblock == nil || certblock.Type != "CERTIFICATE" {
		return tls.Certificate{}, fmt.Errorf("an error occured while attempting to decode certificate: cert provided is not a certificate")
	}

	keyblock, _ := pem.Decode(key)
	if keyblock == nil {
		return tls.Certificate{}, fmt.Errorf("an error occured while attempting to decode key: key provided is not a certificate")
	}
	pk, err := x509.ParseECPrivateKey(keyblock.Bytes)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("an error occured while parsing the ecdsa private key")
	}

	return tls.Certificate{
		Certificate: [][]byte{certblock.Bytes},
		PrivateKey:  pk,
	}, nil
}

// hijack attempts to assert w as a http.Hijacker followed
// by using the hijacker to call the Hijack function. If the
// assertion or hijacking fails, it returns an error.
func hijack(w any) (net.Conn, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("hijacking failed: not a hijackable")
	}
	conn, _, err := hijacker.Hijack()
	return conn, err
}

// slogArrayToMap converts a [key, value, key, value] array into
// a {key: value, key: value} map.
func slogArrayToMap(a []any) map[string]any {
	m := make(map[string]any)
	for i := 0; i < len(a); i += 2 {
		m[a[i].(string)] = a[i+1]
	}
	return m
}

// fetchBlockedSites fetches all blocked sites and places them into
// env.BlockedSites.
func fetchBlockedSites() (err error) {
	sites, err := getBlockedSites()
	if err != nil {
		return err
	}
	env.BlockedSites = sites
	return nil
}

// anyRegexMatch attempts to match matcher with any of the regexes.
func anyRegexMatch(regexes []*regexp.Regexp, matcher []byte) bool {
	for _, regex := range regexes {
		if regex.Match(matcher) {
			return true
		}
	}
	return false
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
