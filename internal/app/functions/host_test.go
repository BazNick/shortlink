package functions

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemeAndHost(t *testing.T) {
	tests := []struct {
		name     string
		request  *http.Request
		expected string
	}{
		{
			name: "HTTP request",
			request: &http.Request{
				Host: "localhost:8080",
				TLS:  nil,
			},
			expected: "http://localhost:8080",
		},
		{
			name: "HTTPS request",
			request: &http.Request{
				Host: "example.com",
				TLS:  &tls.ConnectionState{},
			},
			expected: "https://example.com",
		},
		{
			name: "HTTP request with custom host",
			request: &http.Request{
				Host: "api.example.com:443",
				TLS:  nil,
			},
			expected: "http://api.example.com:443",
		},
		{
			name: "HTTPS request with custom host",
			request: &http.Request{
				Host: "secure.example.com:8443",
				TLS:  &tls.ConnectionState{},
			},
			expected: "https://secure.example.com:8443",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SchemeAndHost(test.request)
			assert.Equal(t, test.expected, result)
		})
	}
}
