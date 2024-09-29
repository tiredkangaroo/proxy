package main

import (
	"log/slog"
	"strconv"
	"time"
)

type LogInfo struct {
	// If set to true, the ID will be logged in the logging system(s) used.
	ID bool
	// If set to true, the ClientIP will be logged in the logging system(s) used.
	ClientIP bool
	// If set to true, the ProxyAuthorization will be logged in the logging system(s) used (not recommended).
	ProxyAuthorization bool
	// If set to true, the RawHTTPRequest will be logged in the logging system(s) used.
	RawHTTPRequest bool
	// If set to true, the body of the request will be logged as well.
	// If RawHTTPRequest is set to false, the body will not be logged regardless
	// of this value.
	RawHTTPRequestWithBody bool
	// If set to true, the RawHTTPResponse will be logged in the logging system(s) used.
	RawHTTPResponse bool
	// If set to true, the body of the response will be logged as well.
	// If RawHTTPResponse is set to false, the body will not be logged regardless
	// of this value.
	RawHTTPResponseWithBody bool
	// If set to true, the Method of the request (eg. GET, POST) will be logged
	// in the logging system(s) used.
	Method bool
	// If set to true, the URL will be logged in the logging system(s) used. This
	// includes paths and query params.
	URL bool
	// If set to true, the time will be logged in the logging system(s) used. Note:
	// if ID is set to true, the time of the request can still be made out from the ID
	// as the ID is based on the time of the request.
	Time bool
	// If set to true, the ProcessingTime will be logged in the logging system(s) used.
	ProcessingTime bool
	// If set to true, the UpstreamResponseTime will be logged in the logging system(s) used.
	UpstreamResponseTime bool
}

// generateLogArgs generates arguments for slog logging based on whether or not
// the information should be logged.
func generateLogArgs(request *ProxyHTTPRequest, loginfo *LogInfo) []any {
	args := []any{}
	if loginfo.ID {
		args = append(args, "id", request.ID)
	}
	if loginfo.ClientIP {
		args = append(args, "clientIP", request.ClientIP)
	}
	if loginfo.ProxyAuthorization {
		args = append(args, "proxyAuthorization", request.ProxyAuthorization)
	}
	if request.Error != nil {
		args = append(args, "error", request.Error.Error())
	}
	if loginfo.Method {
		args = append(args, "method", request.Method)
	}
	if loginfo.URL && request.URL != nil {
		args = append(args, "url", request.URL.String())
	}
	if loginfo.Time && request.Start != nil {
		args = append(args, "time", request.Start.Unix())
	}
	if loginfo.ProcessingTime {
		args = append(args, "processingTime", time.Since(*request.Start).Milliseconds())
	}
	if loginfo.UpstreamResponseTime {
		args = append(args, "upstreamResponseTime", request.UpstreamResponseTime.Milliseconds())
	}
	if loginfo.RawHTTPRequest {
		args = append(args, "rawHTTPRequest", request.RawHTTPRequest)
	}
	if loginfo.RawHTTPResponse {
		args = append(args, "rawHTTPResponse", request.RawHTTPResponse)
	}
	return args
}

func parseLogInfo(loginfo string) *LogInfo {
	defaultInfo := &LogInfo{
		ID:                      true,
		ClientIP:                true,
		ProxyAuthorization:      false,
		RawHTTPRequest:          false,
		RawHTTPRequestWithBody:  false,
		RawHTTPResponse:         false,
		RawHTTPResponseWithBody: false,
		Method:                  true,
		URL:                     true,
		Time:                    true,
		ProcessingTime:          true,
	}
	if loginfo == "0" {
		return nil
	}
	if len(loginfo) != 11 {
		slog.Warn("length must be exactly 11: using default log")
		return defaultInfo
	}
	info := []bool{}
	for _, c := range loginfo {
		b, err := strconv.Atoi(string(c))
		if err != nil || (b != 0 && b != 1) {
			slog.Warn("error with Atoi or digit is not binary: using default log")
			return defaultInfo
		}
		t := binaryToBool(b)
		info = append(info, t)
	}
	return &LogInfo{
		ID:                      info[0],
		ClientIP:                info[1],
		ProxyAuthorization:      info[2],
		RawHTTPRequest:          info[3],
		RawHTTPRequestWithBody:  info[4],
		RawHTTPResponse:         info[5],
		RawHTTPResponseWithBody: info[6],
		Method:                  info[7],
		URL:                     info[8],
		Time:                    info[9],
		ProcessingTime:          info[10],
	}
}

func log(request *ProxyHTTPRequest, err error) {
	if env.LogInfo == nil {
		return
	}
	request.Error = err
	a := generateLogArgs(request, env.LogInfo)
	go func() {
		err := saveProxyRequest(slogArrayToMap(a))
		if err != nil {
			slog.Error("saving proxy request", "error", err)
		}
	}()

	if err == nil {
		env.Logger.Debug("PROXY REQUEST", a...)
	} else {
		env.Logger.Error("PROXY REQUEST", a...)
	}
}
