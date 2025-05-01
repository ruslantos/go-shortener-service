package logger

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// responseWriter обертывает http.ResponseWriter для записи и отслеживания размера и статуса ответа.
type responseWriter struct {
	http.ResponseWriter
	body   *bytes.Buffer
	size   int
	status int
}

// Size возвращает размер ответа.
func (rw *responseWriter) Size() int {
	return rw.size
}

// Write записывает данные в ответ и отслеживает размер записанных данных.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	rw.body.Write(b)
	return n, err
}

// WriteHeader устанавливает статус ответа и вызывает оригинальный WriteHeader.
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Status возвращает статус ответа.
func (rw *responseWriter) Status() int {
	if rw.status == 0 {
		return http.StatusOK
	}
	return rw.status
}

// LoggerChi middleware для логирования HTTP-запросов и ответов.
func LoggerChi(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, body: &bytes.Buffer{}}

			cookieHeader := r.Header.Get("Cookie")

			logger.Debug("Incoming request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("cookie", cookieHeader),
			)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			var cookies []string
			cookies = append(cookies, rw.Header()["Set-Cookie"]...)

			logger.Debug("Outgoing response",
				zap.Int("status", rw.Status()),
				zap.Duration("duration", duration),
				zap.Int("response_size", rw.Size()),
				zap.String("response_body", rw.body.String()),
				zap.Strings("cookies", cookies),
			)

		})
	}
}

// GetLogger возвращает глобальный экземпляр zap.Logger.
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

// Sync синхронизирует записи логов и очищает буфер.
func Sync() error {
	return logger.Sync()
}
