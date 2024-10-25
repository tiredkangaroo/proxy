package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var urls = []string{
	"https://api.github.com/users/tiredkangaroo",
	"https://api.github.com/users/nikumar1206",
	"https://api.github.com/users/octocat",
	"https://randomuser.me/api/",
}

// MeasureTime makes a GET request to a URL and returns the time taken to complete the request.
func MeasureTime(client *http.Client, url string) (time.Duration, error) {
	start := time.Now()
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return time.Since(start), nil
}

// AverageTime calculates the average time taken for a list of URLs by testing each URL `reps` times using the provided client.
func AverageTime(client *http.Client, urls []string, reps int) (time.Duration, error) {
	var totalDuration time.Duration
	totalRequests := reps * len(urls)

	for i := 0; i < reps; i++ {
		for _, u := range urls {
			duration, err := MeasureTime(client, u)
			if err != nil {
				return 0, err
			}
			totalDuration += duration
		}
	}
	avgDuration := totalDuration / time.Duration(totalRequests)
	return avgDuration, nil
}

func main() {
	// Number of times to repeat tests for each URL
	repetitions := 20

	// Without Proxy
	nonProxyClient := &http.Client{}

	// With Proxy
	proxyURL, err := url.Parse("http://localhost:8000")
	if err != nil {
		fmt.Printf("Error parsing proxy URL: %v\n", err)
		return
	}

	proxyClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	avgProxyTime, err := AverageTime(proxyClient, urls, repetitions)
	if err != nil {
		fmt.Printf("Error measuring proxy time: %v\n", err)
		return
	}
	fmt.Printf("Average proxy connection time: %v\n", avgProxyTime)

	avgNonProxyTime, err := AverageTime(nonProxyClient, urls, repetitions)
	if err != nil {
		fmt.Printf("Error measuring non-proxy time: %v\n", err)
		return
	}
	fmt.Printf("Average non-proxy connection time: %v\n", avgNonProxyTime)

	// Calculate and print the average time added by the proxy
	avgTimeAdded := avgProxyTime - avgNonProxyTime
	fmt.Printf("Average time added by the proxy: %v\n", avgTimeAdded)
}
