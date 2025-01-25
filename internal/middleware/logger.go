package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type responseWriter struct {
	gin.ResponseWriter
	size int
}

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: c.Writer}

		logger.Info("Incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)

		c.Writer = rw
		c.Next()

		duration := time.Since(start)

		logger.Info("Outgoing response",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.Int("response_size", rw.Size()),
		)
	}
}

func (rw *responseWriter) Size() int {
	return rw.size
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}
