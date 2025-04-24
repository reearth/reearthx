package log

import (
	"net/http"
	"strconv"
	"strings"
)

func GetReqestID(w http.ResponseWriter, r *http.Request) string {
	if reqid := getHeader(r,
		"traceparent",
		"X-Request-ID",
		"X-Cloud-Trace-Context", // Google Cloud
		"X-Amzn-Trace-Id",       // AWS
		"X-ARR-LOG-ID",          // Azure
	); reqid != "" {
		return reqid
	}

	if w != nil {
		if reqid := w.Header().Get("X-Request-ID"); reqid != "" {
			return reqid
		}
	}

	return ""
}

func getHeader(r *http.Request, keys ...string) string {
	for _, k := range keys {
		if v := r.Header.Get(k); v != "" {
			return v
		}
		if v := r.Header.Get(strings.ToLower(k)); v != "" {
			return v
		}
	}
	return ""
}

func TraceIDFrom(trace string) string {
	if t := TraceIDFromTraceparent(trace); t != "" {
		return t
	}
	if t := TraceIDFromXCloudTraceContext(trace); t != "" {
		return t
	}
	return ""
}

func TraceIDFromTraceparent(trace string) string {
	//${version}-${trace-id}-${span-id}-${trace-flags}
	parts := strings.Split(trace, "-")
	if len(parts) < 2 {
		return ""
	}

	// https://www.w3.org/TR/trace-context/#version
	if _, err := strconv.Atoi(parts[0]); err != nil || len(parts[0]) != 2 {
		return ""
	}

	return parts[1]
}

func TraceIDFromXCloudTraceContext(trace string) string {
	// ${trace-id}/${span-id};o=${trace-flags}
	trace, _, ok := strings.Cut(trace, "/")
	if !ok {
		return ""
	}
	return trace
}
