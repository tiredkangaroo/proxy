package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// CertificateService represents a service that handles
// gettings certificates and the handling of the caching of
// certificates.
type CertificateService struct {
	certificates map[string]*tls.Certificate
	mx           sync.RWMutex
	sfgroup      singleflight.Group
}

// NewCertificateService creates a new certificate service.
func NewCertificateService() *CertificateService {
	return &CertificateService{
		certificates: make(map[string]*tls.Certificate),
		mx:           sync.RWMutex{},
		sfgroup:      singleflight.Group{},
	}
}

// getTLSKeyPair returns a TLS Key Pair either from cache based on the host or generates
// a new one if the cache is unavailable or does not have it stored. It will automatically
// cache the certificate afterwards if possible.
func (cs *CertificateService) getTLSKeyPair(ctx context.Context, host string, cacert string, cakey string) (*tls.Certificate, error) {
	// retrieve from cache
	if tlscert, err := cs.getTLSCertFromCache(ctx, host); err == nil { // cache hit
		return tlscert, nil
	}

	// cache miss; generate a certificate key pair and create a tls.Certificate out of the pair
	certResponse, err, _ := cs.sfgroup.Do(host, func() (interface{}, error) {
		// make tls certificate (not cached or cache not available)
		cert, key, err := generateCertificate(host, cacert, cakey)
		if err != nil {
			return nil, err
		}
		// create key pair
		tlsCert, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, fmt.Errorf("an error occured parsing the public private key pair: %s", err.Error())
		}
		return &tlsCert, nil
	})

	if err != nil {
		return nil, err
	}

	tlsCert := certResponse.(*tls.Certificate)

	go cs.setTLSCertToCache(ctx, host, tlsCert)
	return tlsCert, nil
}

// generateCertificate generates a certificate for the specified host signed by the CA Certificate and Key written
// in the files specified. It expects PEM encoded x509 certificates and PKCS8 private keys in the files.
func generateCertificate(host string, caCertFilename string, caKeyFilename string) ([]byte, []byte, error) {
	// read certificate and key files
	cacertbytes, err := os.ReadFile(caCertFilename)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to read the CA Cert File: %s", err.Error())
	}
	cakeybytes, err := os.ReadFile(caKeyFilename)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to read the CA Key File: %s", err.Error())
	}

	// decode them
	certblock, _ := pem.Decode(cacertbytes)
	cakeyblock, _ := pem.Decode(cakeybytes)

	// parse x509.Certificate out of the decoded pem.Block
	cacert, err := x509.ParseCertificate(certblock.Bytes)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to parse the CACert: %s", err.Error())
	}

	// parse private key out of the decoded pem.Block
	cakey, err := x509.ParsePKCS8PrivateKey(cakeyblock.Bytes)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to read the CA Private Key: %s", err.Error())
	}

	// generate a new private key for the new certificate
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to generate an ecdsa key: %s", err.Error())
	}

	// create a serial number for certificate
	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to generate a certificate serial number: %s", err.Error())
	}

	// create cert config
	config := &x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"N/A"},
		},
		DNSNames:              []string{host},
		NotBefore:             time.Now().Add(-(time.Hour * 7200)),
		NotAfter:              time.Now().Add(time.Hour * 7200),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// create certificate
	cert, err := x509.CreateCertificate(rand.Reader, config, cacert, &pk.PublicKey, cakey)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to create the x509 certificate: %s", err.Error())
	}

	// encode certificate and private key
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	if pemCert == nil {
		return []byte{}, []byte{}, fmt.Errorf("an unknown error occured while attempting to encode the pem cert to memory")
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to marshal the private key: %s", err.Error())
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemKey == nil {
		return []byte{}, []byte{}, fmt.Errorf("an unknown error occured while attempting to encode the private key to pem memory cert")
	}

	return pemCert, pemKey, nil
}

func addTLSToConnection(cert *tls.Certificate, conn net.Conn) *tls.Conn {
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:               tls.VersionTLS10,
		MaxVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{*cert},
	}
	return tls.Server(conn, tlsConfig)
}
