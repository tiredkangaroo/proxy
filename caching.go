package main

import (
	"context"
	"crypto/tls"
	"fmt"
)

// getTLSCertFromCache gets a TLS certificate that corresponds with host from cache. If
// it fails, it will return false along with the empty certificate.
func (cs *CertificateService) getTLSCertFromCache(ctx context.Context, host string) (*tls.Certificate, error) {
	select {
	case <-acquire(cs.mx.RLock):
		cert, ok := cs.certificates[host]
		cs.mx.RUnlock()
		var err error
		if !ok {
			err = fmt.Errorf("cache miss")
		}
		return cert, err
	case <-ctx.Done():
		return nil, fmt.Errorf("operation cancelled by context")
	}
}

// setTLSCertToCache sets a TLS certificate that will correspond with host. If
// it fails, it will return an error.
func (cs *CertificateService) setTLSCertToCache(ctx context.Context, host string, cert *tls.Certificate) error {
	select {
	case <-acquire(cs.mx.Lock):
		cs.certificates[host] = cert
		cs.mx.Unlock()
		return nil
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled by context")
	}
}
