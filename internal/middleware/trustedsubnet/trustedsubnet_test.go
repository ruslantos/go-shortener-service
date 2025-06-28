package trustedsubnet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ruslantos/go-shortener-service/internal/config"
)

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		header         http.Header
		trustedSubnet  string
		expectedStatus int
	}{
		//{
		//	name:           "Access allowed for internal stats path with trusted IP",
		//	path:           "/api/internal/stats",
		//	header:         http.Header{"X-Real-IP": []string{"192.168.1.1"}},
		//	trustedSubnet:  "192.168.1.0/24",
		//	expectedStatus: http.StatusOK,
		//},
		{
			name:           "Access denied for internal stats path with untrusted IP",
			path:           "/api/internal/stats",
			header:         http.Header{"X-Real-IP": []string{"10.0.0.1"}},
			trustedSubnet:  "192.168.1.0/24",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Access allowed for non-internal stats path with untrusted IP",
			path:           "/api/other",
			header:         http.Header{"X-Real-IP": []string{"10.0.0.1"}},
			trustedSubnet:  "192.168.1.0/24",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Access denied for internal stats path without X-Real-IP header",
			path:           "/api/internal/stats",
			header:         http.Header{},
			trustedSubnet:  "192.168.1.0/24",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Access denied for internal stats path with invalid IP",
			path:           "/api/internal/stats",
			header:         http.Header{"X-Real-IP": []string{"invalid-ip"}},
			trustedSubnet:  "192.168.1.0/24",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Access denied for internal stats path with empty trusted subnet",
			path:           "/api/internal/stats",
			header:         http.Header{"X-Real-IP": []string{"192.168.1.1"}},
			trustedSubnet:  "",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				TrustedSubnet: tt.trustedSubnet,
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := Middleware(cfg)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header = tt.header

			rr := httptest.NewRecorder()
			middleware(next).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
