package shorten

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/ruslantos/go-shortener-service/internal/config"
	internal_errors "github.com/ruslantos/go-shortener-service/internal/errors"
)

// Пример использования обработчика для успешного добавления ссылки
func ExampleHandler_success() {
	// Инициализируем конфигурацию
	config.FlagShortURL = "http://short.url/"

	// Создаем мок сервиса для успешного случая
	mockService := &mockLinksService{
		addFunc: func(ctx context.Context, long string) (string, error) {
			return "abc123", nil
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	body := ShortenRequest{URL: "http://example.com"}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(string(bodyJSON)))
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.Handle(w, req)

	// Проверяем результат
	resp := w.Result()
	defer resp.Body.Close()

	// Выводим результат
	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Response Body:", strings.TrimSpace(w.Body.String()))
	// Output:
	// Status Code: 201
	// Response Body: {"result":"http://short.url/abc123"}
}

// Пример использования обработчика для случая, когда ссылка уже существует
func ExampleHandler_conflict() {
	// Инициализируем конфигурацию
	config.FlagShortURL = "http://short.url/"

	// Создаем мок сервиса для случая, когда ссылка уже существует
	mockService := &mockLinksService{
		addFunc: func(ctx context.Context, long string) (string, error) {
			return "", internal_errors.ErrURLAlreadyExists
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	body := ShortenRequest{URL: "http://example.com"}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(string(bodyJSON)))
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.Handle(w, req)

	// Проверяем результат
	resp := w.Result()
	defer resp.Body.Close()

	// Выводим результат
	fmt.Println("Status Code:", resp.StatusCode)
	// Output:
	// Status Code: 409
}

// Мок сервиса для тестирования
type mockLinksService struct {
	addFunc func(ctx context.Context, long string) (string, error)
}

func (m *mockLinksService) Add(ctx context.Context, long string) (string, error) {
	return m.addFunc(ctx, long)
}

// errorReader для имитации ошибки чтения тела запроса
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
