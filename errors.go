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
	InternalServerErrorHTML     string = "<html><body><h1>Internal Server Error</h1><pre>%s</pre></body></html>"
	InternalServerErrorResponse string = "HTTP/1.1 500 Internal Server Error\r\n" +
		"Content-Type: text/html\r\n" +
		"Content-Length: %d\r\n" +
		"\r\n" +
		"%s"
)
