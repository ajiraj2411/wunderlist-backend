package auth

import (
	"net"
	"net/http"
	"strings"
)

func ExtractUserAgent(r *http.Request) string {
	if ua := r.Header.Get("User-Agent"); ua != "" {
		return ua
	}
	return "unknown"
}

func ExtractClientIP(r *http.Request) string {
	headers := []string{
		"X-Real-IP",
		"X-Forwarded-For",
		"CF-Connecting-IP",
	}

	for _, h := range headers {
		if v := r.Header.Get(h); v != "" {
			return strings.Split(v, ",")[0]
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return ip
	}
	return r.RemoteAddr
}
