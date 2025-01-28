package middleware

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	once   sync.Once
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
	size int
}

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}

		logger.Info("Incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		c.Writer = rw
		c.Next()

		duration := time.Since(start)

		logger.Info("Outgoing response",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.Int("response_size", rw.Size()),
			zap.String("response_body", rw.body.String()),
		)
	}
}

func (rw *responseWriter) Size() int {
	return rw.size
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	rw.body.Write(b)
	return n, err
}

func LoggerChi(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a gin context to use the existing Logger middleware
			ginContext, _ := gin.CreateTestContext(w)
			ginContext.Request = r

			rw := &responseWriter{ResponseWriter: ginContext.Writer, body: &bytes.Buffer{}}

			logger.Info("Incoming request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			ginContext.Writer = rw
			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			logger.Info("Outgoing response",
				zap.Int("status", rw.Status()),
				zap.Duration("duration", duration),
				zap.Int("response_size", rw.Size()),
				zap.String("response_body", rw.body.String()),
			)
		})
	}
}

func GetLogger() *zap.Logger {
	once.Do(func() {
		// Configure the logger
		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		var err error
		logger, err = config.Build()
		if err != nil {
			panic("failed to initialize logger: " + err.Error())
		}
	})
	return logger
}

func Sync() error {
	return logger.Sync()
}
