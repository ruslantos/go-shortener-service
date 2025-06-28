package trustedsubnet

import (
	"context"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/config"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

type contextKey string

const TrustedSubnetKey contextKey = "trustedSubnet"

// Middleware проверяет, что IP клиента находится в доверенной подсети
func Middleware(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/internal/stats") {
				next.ServeHTTP(w, r)
				return
			}

			if cfg.TrustedSubnet == "" {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				http.Error(w, "X-Real-IP header required", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(ipStr)
			if ip == nil {
				http.Error(w, "Invalid IP address", http.StatusForbidden)
				return
			}

			// Парсим CIDR
			_, subnet, err := net.ParseCIDR(cfg.TrustedSubnet)
			if err != nil {
				logger.GetLogger().Error("Failed to parse trusted subnet",
					zap.String("subnet", cfg.TrustedSubnet),
					zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !subnet.Contains(ip) {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), TrustedSubnetKey, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
