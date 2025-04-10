package functions

import "net/http"

func SchemeAndHost(req *http.Request) string {
	if req.TLS != nil {
		return "https://" + req.Host
	}
	return "https://" + req.Host
}
