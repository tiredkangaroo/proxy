package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"time"
)

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

func getFromCache(ctx context.Context, host string) (tls.Certificate, error) {
	resp := env.Client.HGet(ctx, "proxy", host)

	if resp.Err() != nil {
		return tls.Certificate{}, resp.Err()
	}

	cert, err := unmarshalTLSCertificate([]byte(resp.Val()))
	if err != nil {
		return tls.Certificate{}, err
	}

	return cert, nil
}

func setTLSCertToCache(ctx context.Context, host string, tlscert tls.Certificate) {
	data, err := marshalTLSCertificate(tlscert)
	if err != nil {
		slog.Error("an error occured while marshalling the tls certificate to write to cache", "error", err.Error())
	}
	resp := env.Client.HSet(ctx, "proxy", host, data)
	if resp.Err() != nil {
		slog.Error("an error occured while writing the tls cert to cache", "error", resp.Err())
	}
	resp2 := env.Client.HExpireAt(ctx, "proxy", time.Now().Add(time.Hour*7200), host)
	if resp2.Err() != nil {
		slog.Error("an error occured while attempting to expire the tls cert in cache", "error", resp2.Err())
	}
}

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
