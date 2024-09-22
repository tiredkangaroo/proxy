package main

import (
	"fmt"
)

func checkAuth(proxyAuthorizationToken string) (err error) {
	if proxyAuthorizationToken == "" {
		return fmt.Errorf("Not authenticated.")
	}
	return nil
}

func allowRequest(request *ProxyHTTPRequest) (err error) {
	return nil
}
