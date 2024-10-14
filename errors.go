package main

const (
	FailedToRetrieveBlockedSites string = "failed to retrieve blocked sites: %s"
	RegexCompilationFailed       string = "compiling regex failed: %s"
	RandomGenerationFailed       string = "an error occured while reading system randoms: %s"
	ProxyBlockedResponse         string = "HTTP/1.1 403 Forbidden\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Length: 93\r\n" +
		"\r\n" +
		"<h1>Request Blocked</h1>\r\n" +
		"<pre>This request has been blocked by the proxy.</pre>\r\n"
)
