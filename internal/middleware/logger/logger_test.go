package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

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
