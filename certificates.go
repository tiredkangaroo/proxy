package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

func generateMITMCertificate(host string, CACertFilename string, CAKeyFilename string) ([]byte, []byte, error) {
	cacertbytes, err := os.ReadFile(CACertFilename)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to read the CA Cert File: %s", err.Error())
	}
	cakeybytes, err := os.ReadFile(CAKeyFilename)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to read the CA Key File: %s", err.Error())
	}
	certblock, _ := pem.Decode(cacertbytes)
	cakeyblock, _ := pem.Decode(cakeybytes)
	cacert, err := x509.ParseCertificate(certblock.Bytes)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to parse the CACert: %s", err.Error())
	}
	cakey, err := x509.ParsePKCS8PrivateKey(cakeyblock.Bytes)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to read the CA Private Key: %s", err.Error())
	}
	// consists of certificate, private key (pk), serial number (sn)
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to generate an ecdsa key: %s", err.Error())
	}
	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to generate a certificate serial number: %s", err.Error())
	}
	cert := x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"N/A"},
		},
		DNSNames:              []string{host},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 7200),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, &cert, cacert, &pk.PublicKey, cakey)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to create the x509 certificate: %s", err.Error())
	}

	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if pemCert == nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to encode the pem cert to memory: %s", err.Error())
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to marshal the private key: %s", err.Error())
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemCert == nil {
		return []byte{}, []byte{}, fmt.Errorf("an error occured while attempting to encode the private key to pem memory cert: %s", err.Error())
	}
	return pemCert, pemKey, nil
}
