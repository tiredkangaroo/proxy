package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"time"
)

func getFromCache(ctx context.Context, host string) (tls.Certificate, error) {
	resp := env.Client.HGet(ctx, "proxycerts", host)

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
	resp := env.Client.HSet(ctx, "proxycerts", host, data)
	if resp.Err() != nil {
		slog.Error("an error occured while writing the tls cert to cache", "error", resp.Err())
	}
	resp2 := env.Client.HExpireAt(ctx, "proxycerts", time.Now().Add(time.Hour*7200), host)
	if resp2.Err() != nil {
		slog.Error("an error occured while attempting to expire the tls cert in cache", "error", resp2.Err())
	}
}
