package functions

import (
	"net/http"
	"strings"
)

var (
	httpPrefix  = "http://"
	httpsPrefix = "https://"
)

// SchemeAndHost - получаем хоста и протокол запроса (http или https)
func SchemeAndHost(req *http.Request) string {
	if req.TLS != nil {
		var result strings.Builder
		result.Grow(len(httpsPrefix) + len(req.Host))
		result.WriteString(httpsPrefix)
		result.WriteString(req.Host)
		return result.String()
	}

	var result strings.Builder
	result.Grow(len(httpPrefix) + len(req.Host))
	result.WriteString(httpPrefix)
	result.WriteString(req.Host)
	return result.String()
}
