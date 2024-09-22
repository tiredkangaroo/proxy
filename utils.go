package main

import (
	"crypto/rand"
	"encoding/base32"
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
