package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
)

var config Config

type Config struct {
	CertificateService *CertificateService
	Logger             *slog.Logger

	port string
}

func loadConfig() error {
	config.Logger = slog.Default()
	slog.SetLogLoggerLevel(slog.LevelDebug)

	crtfile := os.Getenv("PROXYCERT_AUTHORITY_CRT")
	keyfile := os.Getenv("PROXYCERT_AUTHORITY_KEY")

	if crtfile == "" || keyfile == "" {
		return ErrMissingCA
	}

	// read certificate and key files
	rawcrt, err := os.ReadFile(crtfile)
	if err != nil {
		return fmt.Errorf("ca cert file: %s", err.Error())
	}
	rawkey, err := os.ReadFile(keyfile)
	if err != nil {
		return fmt.Errorf("an error occured while attempting to read the CA Key File: %s", err.Error())
	}

	// decode them
	certblock, _ := pem.Decode(rawcrt)
	cakeyblock, _ := pem.Decode(rawkey)

	// parse x509.Certificate out of the decoded pem.Block
	cacert, err := x509.ParseCertificate(certblock.Bytes)
	if err != nil {
		return fmt.Errorf("ca cert can't be parsed: %s", err.Error())
	}

	// parse private key out of the decoded pem.Block
	cakey, err := x509.ParsePKCS8PrivateKey(cakeyblock.Bytes)
	if err != nil {
		return fmt.Errorf("an error occured while attempting to read the CA Private Key: %s", err.Error())
	}

	config.CertificateService = NewCertificateService()
	config.CertificateService.cacert = cacert
	config.CertificateService.cakey = cakey

	config.port = os.Getenv("PROXY_PORT")

	return nil
}
