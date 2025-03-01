package main

import "net/http"

func startWebServer() error {
	// hello there starting web server here!
	http.HandleFunc("/", handler func(http.ResponseWriter, *http.Request))
}
