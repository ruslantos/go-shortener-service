package deleteuserurls

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/service"
)

// ExampleHandler_Handle демонстрирует использование метода Handle.
func ExampleHandler_Handle() {
	// Создаем мок-реализацию linkService для тестирования.
	mockService := &mockLinkService{}

	// Создаем новый Handler с мок-сервисом.
	handler := New(mockService)

	// Создаем тело запроса с URL для удаления.
	body := DeleteUserURLsRequest{"http://example.com/1", "http://example.com/2"}
	bodyBytes, _ := json.Marshal(body)

	// Создаем новый HTTP-запрос с телом.
	req, _ := http.NewRequestWithContext(context.WithValue(context.Background(), auth.UserIDKey, "user123"), "POST", "/deleteuserurls", bytes.NewBuffer(bodyBytes))

	// Создаем записыватель ответа для захвата ответа.
	rr := httptest.NewRecorder()

	// Вызываем метод Handle.
	handler.Handle(rr, req)

	// Печатаем код состояния ответа.
	fmt.Println(rr.Code)

	// Output:
	// 202
}

// mockLinkService — это мок-реализация интерфейса linkService для тестирования.
type mockLinkService struct{}

func (m *mockLinkService) ConsumeDeleteURLs(data service.DeletedURLs) {
	// Мок-реализация ничего не делает.
}
