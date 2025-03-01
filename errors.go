package main

import (
	"errors"
	"fmt"
)

var (
	ErrMissingCA = errors.New("PROXYCERT-AUTHORTIY-CRT and PROXYCERT-AUTHORITY-KEY env variables must be provided")
)

func InternalServerErrorResponse(id string) []byte {
	data := "HTTP/1.1 500 Internal Server Error\r\n" +
		"Content-Type: text/html\r\n" +
		"Content-Length: %d\r\n" +
		"\r\n" +
		"%s"
	body := fmt.Sprintf("<h1>Internal Server Error</h1> <p>Request ID: %s</p>", id)
	return []byte(fmt.Sprintf(data, len(body), body))
}
