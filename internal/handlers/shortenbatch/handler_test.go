package shortenbatch

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
	"github.com/ruslantos/go-shortener-service/internal/models"
)

func TestHandler_Handle_Success(t *testing.T) {
	service := &MocklinksService{}
	linksIn := []models.Link{
		{CorrelationID: "123", OriginalURL: "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf"},
		{CorrelationID: "456", OriginalURL: "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf2"},
	}
	linksOut := []models.Link{
		{CorrelationID: "123", OriginalURL: "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf", ShortURL: "qwerty1"},
		{CorrelationID: "456", OriginalURL: "http://ivghfkudbptp.biz/qqlcxvlwy1o/pbmze/ad4hdsyf2", ShortURL: "qwerty2"},
	}
	service.EXPECT().AddBatch(context.Background(), linksIn).Return(linksOut, nil)
	h := New(service)
	in := ShortenBatchRequest{
		{CorrelationID: linksIn[0].CorrelationID, OriginalURL: linksIn[0].OriginalURL},
		{CorrelationID: linksIn[1].CorrelationID, OriginalURL: linksIn[1].OriginalURL},
	}
	out := ShortenBatchResponse{
		{CorrelationID: linksOut[0].CorrelationID, ShortURL: "http://localhost:8080/qwerty1"},
		{CorrelationID: linksOut[1].CorrelationID, ShortURL: "http://localhost:8080/qwerty2"},
	}
	marshalledIn, err := json.Marshal(in)
	assert.NoError(t, err)
	marshalledOut, err := json.Marshal(out)
	assert.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "/api/shorten/batch", io.NopCloser(bytes.NewReader(marshalledIn)))
	assert.NoError(t, err)
	rr := httptest.NewRecorder()

	h.Handle(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, string(marshalledOut), rr.Body.String())
}

// Пример использования обработчика для успешного добавления пакета ссылок
func ExampleHandler_Success() {
	// Инициализируем конфигурацию
	config.FlagShortURL = "http://short.url/"

	// Создаем мок сервиса для успешного случая
	mockService := &mockLinksService{
		addBatchFunc: func(ctx context.Context, links []models.Link) ([]models.Link, error) {
			return []models.Link{
				{OriginalURL: "http://example.com", CorrelationID: "1", ShortURL: "abc123"},
				{OriginalURL: "http://another-example.com", CorrelationID: "2", ShortURL: "def456"},
			}, nil
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	body := ShortenBatchRequest{
		{CorrelationID: "1", OriginalURL: "http://example.com"},
		{CorrelationID: "2", OriginalURL: "http://another-example.com"},
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/shorten/batch", strings.NewReader(string(bodyJSON)))
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
	// Response Body: [{"correlation_id":"1","short_url":"http://short.url/abc123"},{"correlation_id":"2","short_url":"http://short.url/def456"}]
}

// Пример использования обработчика для случая, когда ссылка уже существует
func ExampleHandler_Conflict() {
	// Инициализируем конфигурацию
	config.FlagShortURL = "http://short.url/"

	// Создаем мок сервиса для случая, когда ссылка уже существует
	mockService := &mockLinksService{
		addBatchFunc: func(ctx context.Context, links []models.Link) ([]models.Link, error) {
			return nil, internal_errors.ErrURLAlreadyExists
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	body := ShortenBatchRequest{
		{CorrelationID: "1", OriginalURL: "http://example.com"},
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/shorten/batch", strings.NewReader(string(bodyJSON)))
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
		addBatchFunc: func(ctx context.Context, links []models.Link) ([]models.Link, error) {
			return nil, errors.New("internal server error")
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	body := ShortenBatchRequest{
		{CorrelationID: "1", OriginalURL: "http://example.com"},
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/shorten/batch", strings.NewReader(string(bodyJSON)))
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
	// Response Body: add batch shorten error
}

// Пример использования обработчика для случая, когда тело запроса пустое
func ExampleHandler_EmptyBody() {
	// Создаем мок сервиса (не используется в этом случае)
	mockService := &mockLinksService{}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	req := httptest.NewRequest("POST", "/shorten/batch", strings.NewReader(""))
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
	req := httptest.NewRequest("POST", "/shorten/batch", &errorReader{})
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
	req := httptest.NewRequest("POST", "/shorten/batch", strings.NewReader("{invalid json}"))
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
	// Response Body: Unmarshalling body error
}

// Мок сервиса для тестирования
type mockLinksService struct {
	addBatchFunc func(ctx context.Context, links []models.Link) ([]models.Link, error)
}

func (m *mockLinksService) AddBatch(ctx context.Context, links []models.Link) ([]models.Link, error) {
	return m.addBatchFunc(ctx, links)
}

// errorReader для имитации ошибки чтения тела запроса
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
