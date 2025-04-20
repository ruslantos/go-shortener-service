package getuserurls

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/ruslantos/go-shortener-service/internal/config"
	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

// Пример использования обработчика для успешного получения ссылок пользователя
func ExampleHandler_Success() {
	// Инициализируем конфигурацию
	config.FlagShortURL = "http://short.url/"

	// Создаем мок сервиса для успешного случая
	mockService := &mockLinksService{
		getUserUrlsFunc: func(ctx context.Context) ([]models.Link, error) {
			return []models.Link{
				{ShortURL: "abc123", OriginalURL: "http://example.com"},
				{ShortURL: "def456", OriginalURL: "http://another-example.com"},
			}, nil
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	req := httptest.NewRequest("GET", "/user/urls", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user123")
	req = req.WithContext(ctx)
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
	// Status Code: 200
	// Response Body: [{"short_url":"http://short.url/abc123","original_url":"http://example.com"},{"short_url":"http://short.url/def456","original_url":"http://another-example.com"}]
}

// Пример использования обработчика для случая, когда пользователь не авторизован
func ExampleHandler_Unauthorized() {
	// Создаем мок сервиса (не используется в этом случае)
	mockService := &mockLinksService{}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	req := httptest.NewRequest("GET", "/user/urls", nil)
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
	// Status Code: 401
	// Response Body: user not found
}

// Пример использования обработчика для случая внутренней ошибки сервера
func ExampleHandler_InternalError() {
	// Инициализируем конфигурацию
	config.FlagShortURL = "http://short.url/"

	// Создаем мок сервиса для случая внутренней ошибки сервера
	mockService := &mockLinksService{
		getUserUrlsFunc: func(ctx context.Context) ([]models.Link, error) {
			return nil, errors.New("internal server error")
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	req := httptest.NewRequest("GET", "/user/urls", nil)
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user123")
	req = req.WithContext(ctx)
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
	// Status Code: 400
	// Response Body: filed to get user urls: internal server error
}

// Мок сервиса для тестирования
type mockLinksService struct {
	getUserUrlsFunc func(ctx context.Context) ([]models.Link, error)
}

func (m *mockLinksService) GetUserUrls(ctx context.Context) ([]models.Link, error) {
	return m.getUserUrlsFunc(ctx)
}
