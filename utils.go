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
	"log/slog"
	"net/url"
	"strings"
	"time"
)

func toURL(s string) *url.URL {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "https://" + s
	}
	u, _ := url.Parse(s)
	return u
}

func generateTimeBasedID(t time.Time) string {
	r := make([]byte, 10)
	_, err := rand.Read(r)
	if err != nil {
		slog.Error("an error occured while generating a random for a time based id", "error", err.Error())
	}
	e := base32.StdEncoding.EncodeToString(r)
	return fmt.Sprintf("%d-%s", t.UnixMilli(), e)
}

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
		return tls.Certificate{}, fmt.Errorf("an error occured while attempting to decode certificate: cert provided is not a certificate", err.Error())
	}

	keyblock, _ := pem.Decode(key)
	if keyblock == nil {
		return tls.Certificate{}, fmt.Errorf("an error occured while attempting to decode key: key provided is not a certificate", err.Error())
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
