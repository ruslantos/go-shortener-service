package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func BenchmarkLoggerChi(b *testing.B) {
	// Создаем тестовый логгер
	logger := zaptest.NewLogger(b)

	// Создаем тестовый обработчик, который будет обернут middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
		// Устанавливаем тестовые куки
		w.Header().Add("Set-Cookie", "test=cookie")
	})

	// Создаем middleware
	middleware := LoggerChi(logger)

	// Обертываем обработчик middleware
	wrappedHandler := middleware(handler)

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Cookie", "test=value")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Создаем recorder для каждого запроса
		rec := httptest.NewRecorder()
		rec.Body = bytes.NewBuffer(nil)

		// Выполняем запрос
		wrappedHandler.ServeHTTP(rec, req)

		// Проверяем, что ответ получен (для предотвращения оптимизаций)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status code: %d", rec.Code)
		}
	}
}

func BenchmarkLoggerChi_NoLogging(b *testing.B) {
	// Для сравнения: тест без логирования
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
		w.Header().Add("Set-Cookie", "test=cookie")
	})

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	req.Header.Set("Cookie", "test=value")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		rec.Body = bytes.NewBuffer(nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status code: %d", rec.Code)
		}
	}
}

func TestResponseWriter(t *testing.T) {
	t.Run("Test Size()", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: rec,
			body:           &bytes.Buffer{},
		}

		data := []byte("test data")
		_, err := rw.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), rw.Size())
	})

	t.Run("Test WriteHeader() and Status()", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: rec,
			body:           &bytes.Buffer{},
		}

		// Test default status
		assert.Equal(t, http.StatusOK, rw.Status())

		// Test custom status
		status := http.StatusNotFound
		rw.WriteHeader(status)
		assert.Equal(t, status, rw.Status())
	})

	t.Run("Test Write() captures body", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: rec,
			body:           &bytes.Buffer{},
		}

		data := []byte("test body")
		_, err := rw.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, data, rw.body.Bytes())
	})
}

func TestLoggerChiMiddleware(t *testing.T) {
	// Create a test logger that writes to a buffer
	//var buf bytes.Buffer
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{"stderr"}
	config.ErrorOutputPaths = []string{"stderr"}
	testLogger, err := config.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	assert.NoError(t, err)

	// Replace the global logger for testing
	logger = testLogger

	// Create a test handler that sets cookies and returns a response
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "test", Value: "cookie"})
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("test response"))
	})

	// Create a request with a cookie
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Cookie", "test=value")

	// Create a recorder to capture the response
	rec := httptest.NewRecorder()

	// Apply the middleware
	handler := LoggerChi(testLogger)(testHandler)
	handler.ServeHTTP(rec, req)

	// Verify the response
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Header().Get("Set-Cookie"), "test=cookie")
	assert.Equal(t, "test response", rec.Body.String())
}

func TestGetLogger(t *testing.T) {
	t.Run("Test singleton behavior", func(t *testing.T) {
		// Reset the global logger and once for testing
		logger = nil
		once = sync.Once{}

		// First call should initialize the logger
		l1 := GetLogger()
		assert.NotNil(t, l1)

		// Second call should return the same instance
		l2 := GetLogger()
		assert.Equal(t, l1, l2)
	})
}

func TestLoggerChiLogging(t *testing.T) {
	// Create a test logger that writes to a buffer
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	testLogger := zap.New(core)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Create a request
	req := httptest.NewRequest("GET", "/test-path", nil)
	req.Header.Set("Cookie", "session=abc123")

	// Create a recorder
	rec := httptest.NewRecorder()

	// Apply the middleware
	handler := LoggerChi(testLogger)(testHandler)
	handler.ServeHTTP(rec, req)

	// Get the logged output
	logOutput := buf.String()

	// Verify request logging
	assert.Contains(t, logOutput, `"method":"GET"`)
	assert.Contains(t, logOutput, `"path":"/test-path"`)
	assert.Contains(t, logOutput, `"cookie":"session=abc123"`)

	// Verify response logging
	assert.Contains(t, logOutput, `"status":200`)
	assert.Contains(t, logOutput, `"response_size":2`)
	assert.Contains(t, logOutput, `"response_body":"ok"`)
}
