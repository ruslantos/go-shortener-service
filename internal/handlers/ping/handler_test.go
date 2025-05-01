package ping

import (
	"context"
	"fmt"
	"net/http/httptest"
)

// Пример использования обработчика для успешного пинга
func ExampleHandler_success() {
	// Создаем мок сервиса для успешного случая
	mockService := &mockLinksService{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}

	// Создаем обработчик с мок сервисом
	handler := New(mockService)

	// Создаем запрос и запись для тестирования
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.Handle(w, req)

	// Проверяем результат
	resp := w.Result()
	defer resp.Body.Close()

	// Выводим результат
	fmt.Println("Status Code:", resp.StatusCode)
	// Output:
	// Status Code: 200
}

// Мок сервиса для тестирования
type mockLinksService struct {
	pingFunc func(ctx context.Context) error
}

func (m *mockLinksService) Ping(ctx context.Context) error {
	return m.pingFunc(ctx)
}
