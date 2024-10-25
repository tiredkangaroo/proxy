package main

import (
	"mime/multipart"
	"net/http"
)

func (proxyreq *ProxyHTTPRequest) blocked() bool {
	blocked := anyRegexMatch(env.BlockedSites, []byte(proxyreq.URL.String()))
	if blocked == true {
		return true
	}
	for _, blockerURL := range env.BlockerURLs {
		resp, err := http.DefaultClient.Do(&http.Request{
			Method: "POST",
			URL:    blockerURL,
			MultipartForm: &multipart.Form{
				Value: map[string][]string{
					"url": {proxyreq.URL.String()},
				},
			},
		})
		if err != nil {
			env.Logger.Warn("bypassing blocker due to request failure", "blocker-url", blockerURL.String(), "error", err.Error())
			return false
		}
		body := make([]byte, 5)
		_, err = resp.Body.Read(body)
		if err != nil {
			env.Logger.Warn("bypasssing blocker due to body read failure", "blocker-url", blockerURL.String(), "error", err.Error())
		}
		switch string(body) {
		case "true":
			return true // blocker wants to block the proxy request
		case "false":
			return false // blocker allows the proxy request to continue
		default:
			env.Logger.Error("blocker returned invalid data", "blocker-url", blockerURL.String())
			return false
		}
	}
	return false
}
