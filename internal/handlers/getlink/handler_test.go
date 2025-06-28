package getlink

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	internal_erors "github.com/ruslantos/go-shortener-service/internal/errors"
)

func TestHandler_Handle_Success(t *testing.T) {
	service := &mocklinksService{}
	service.EXPECT().Get(context.Background(), "short").Return("extend", nil)
	h := New(service)
	req, err := http.NewRequest(http.MethodGet, "short", nil)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	assert.Equal(t, "extend", rr.Header().Get("Location"))
}

func TestHandler_Handle_BadRequest(t *testing.T) {
	storage := &mocklinksService{}
	storage.EXPECT().Get(context.Background(), "short").Return("", errors.New("some error"))
	h := New(storage)
	req, err := http.NewRequest(http.MethodGet, "short", nil)
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "failed to get original_url: some error\n", rr.Body.String())
}

// Пример использования обработчика для успешного редиректа
func ExampleHandler_success() {
	// Создаем мок сервиса для успешного случая
	mockService := &mockLinksService{
		getFunc: func(ctx context.Context, shortLink string) (string, error) {
			return "http://example.com", nil
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.Handle(w, req)

	// Проверяем результат
	resp := w.Result()
	defer resp.Body.Close()

	// Выводим результат
	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Location Header:", resp.Header.Get("Location"))
	// Output:
	// Status Code: 307
	// Location Header: http://example.com
}

// Пример использования обработчика для случая, когда ссылка не найдена
func ExampleHandler_notFound() {
	// Создаем мок сервиса для случая, когда ссылка не найдена
	mockService := &mockLinksService{
		getFunc: func(ctx context.Context, shortLink string) (string, error) {
			return "", internal_erors.ErrURLNotFound
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.Handle(w, req)

	// Проверяем результат
	resp := w.Result()
	defer resp.Body.Close()

	// Выводим результат
	fmt.Println("Status Code:", resp.StatusCode)
	// Output:
	// Status Code: 404
}

// Пример использования обработчика для случая, когда ссылка удалена
func ExampleHandler_gone() {
	// Создаем мок сервиса для случая, когда ссылка удалена
	mockService := &mockLinksService{
		getFunc: func(ctx context.Context, shortLink string) (string, error) {
			return "", internal_erors.ErrURLDeleted
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.Handle(w, req)

	// Проверяем результат
	resp := w.Result()
	defer resp.Body.Close()

	// Выводим результат
	fmt.Println("Status Code:", resp.StatusCode)
	// Output:
	// Status Code: 410
}

// Мок сервиса для тестирования
type mockLinksService struct {
	getFunc func(ctx context.Context, shortLink string) (string, error)
}

func (m *mockLinksService) Get(ctx context.Context, shortLink string) (string, error) {
	return m.getFunc(ctx, shortLink)
}
