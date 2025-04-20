package shorten

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ruslantos/go-shortener-service/internal/config"
	internal_errors "github.com/ruslantos/go-shortener-service/internal/errors"
)

func TestHandler_Handle_Success(t *testing.T) {
	extend := "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf"
	service := &MocklinksService{}
	service.EXPECT().Add(context.Background(), extend).Return("short", nil)
	h := New(service)
	in := ShortenRequest{
		URL: extend,
	}
	marshalled, err := json.Marshal(in)
	assert.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "/api/shorten", io.NopCloser(bytes.NewReader(marshalled)))
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, `{"result":"http://localhost:8080/short"}`, rr.Body.String())
}

func TestHandler_Handle_Error(t *testing.T) {
	extend := ""
	service := &MocklinksService{}
	service.EXPECT().Add(context.Background(), extend).Return("short", errors.New("some error"))
	h := New(service)
	in := ShortenRequest{
		URL: extend,
	}
	marshalled, err := json.Marshal(in)
	assert.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "/api/shorten", io.NopCloser(bytes.NewReader(marshalled)))
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// Пример использования обработчика для успешного добавления ссылки
func ExampleHandler_Success() {
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
func ExampleHandler_Conflict() {
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

// Пример использования обработчика для случая внутренней ошибки сервера
func ExampleHandler_InternalError() {
	// Инициализируем конфигурацию
	config.FlagShortURL = "http://short.url/"

	// Создаем мок сервиса для случая внутренней ошибки сервера
	mockService := &mockLinksService{
		addFunc: func(ctx context.Context, long string) (string, error) {
			return "", errors.New("internal server error")
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
	// Status Code: 500
	// Response Body: add shorten link error
}

// Пример использования обработчика для случая, когда тело запроса пустое
func ExampleHandler_EmptyBody() {
	// Создаем мок сервиса (не используется в этом случае)
	mockService := &mockLinksService{}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	req := httptest.NewRequest("POST", "/shorten", strings.NewReader(""))
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
	// Response Body: Error reading body
}

// Пример использования обработчика для случая ошибки чтения тела запроса
func ExampleHandler_ReadBodyError() {
	// Создаем мок сервиса (не используется в этом случае)
	mockService := &mockLinksService{}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования с ошибкой чтения тела
	req := httptest.NewRequest("POST", "/shorten", &errorReader{})
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
	// Response Body: Reading body error
}

// Пример использования обработчика для случая ошибки распаковки JSON
func ExampleHandler_UnmarshalError() {
	// Создаем мок сервиса (не используется в этом случае)
	mockService := &mockLinksService{}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования с некорректным JSON
	req := httptest.NewRequest("POST", "/shorten", strings.NewReader("{invalid json}"))
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
	// Response Body: Unmarshalling error
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
