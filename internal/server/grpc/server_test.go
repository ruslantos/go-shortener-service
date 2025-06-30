package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/ruslantos/go-shortener-service/api"
	auth "github.com/ruslantos/go-shortener-service/internal/interceptors/authinterceptor"
)

func TestCreateShortURL(t *testing.T) {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := api.NewShortenerClient(conn)

	resp, err := client.CreateShortURL(context.Background(), &api.OriginalURL{
		Url: "https://example.com",
	})

	if err != nil {
		t.Fatalf("CreateShortURL failed: %v", err)
	}

	if resp.GetUrl() == "" {
		t.Error("expected non-empty URL")
	}
}

func TestGetOriginalURL(t *testing.T) {
	client, ctx, cleanup := setupClient(t)
	defer cleanup()

	// Сначала создаем ссылку для теста
	createResp, err := client.CreateShortURL(ctx, &api.OriginalURL{
		Url: "https://example.com/get-test",
	})
	if err != nil {
		t.Fatalf("CreateShortURL failed: %v", err)
	}

	// Тест получения оригинальной ссылки
	resp, err := client.GetOriginalURL(ctx, &api.ShortURL{
		Url: createResp.GetUrl(),
	})
	if err != nil {
		t.Fatalf("GetOriginalURL failed: %v", err)
	}

	assert.Equal(t, "https://example.com/get-test", resp.GetUrl())
}

func TestPing(t *testing.T) {
	client, ctx, cleanup := setupClient(t)
	defer cleanup()

	resp, err := client.Ping(ctx, &api.Empty{})
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	assert.True(t, resp.GetSuccess())
}

func TestCreateShortURLBatch(t *testing.T) {
	client, ctx, cleanup := setupClient(t)
	defer cleanup()

	batchResp, err := client.CreateShortURLBatch(ctx, &api.OriginalURLsBatch{
		Urls: []*api.OriginalURL{
			{Url: "https://example.com/batch1"},
			{Url: "https://example.com/batch2"},
		},
	})
	if err != nil {
		t.Fatalf("CreateShortURLBatch failed: %v", err)
	}

	assert.Len(t, batchResp.GetUrls(), 2)
	for _, url := range batchResp.GetUrls() {
		assert.NotEmpty(t, url.GetUrl())
	}
}

func TestGetUserURLs(t *testing.T) {
	client, ctx, cleanup := setupClient(t)
	defer cleanup()

	// Создаем несколько ссылок для пользователя
	_, err := client.CreateShortURL(ctx, &api.OriginalURL{Url: "https://example.com/user1"})
	if err != nil {
		t.Fatalf("CreateShortURL failed: %v", err)
	}
	_, err = client.CreateShortURL(ctx, &api.OriginalURL{Url: "https://example.com/user2"})
	if err != nil {
		t.Fatalf("CreateShortURL failed: %v", err)
	}

	// Получаем ссылки пользователя
	resp, err := client.GetUserURLs(ctx, &api.Empty{})
	if err != nil {
		t.Fatalf("GetUserURLs failed: %v", err)
	}

	assert.GreaterOrEqual(t, len(resp.GetUrls()), 2)
	for _, url := range resp.GetUrls() {
		assert.NotEmpty(t, url.GetShortUrl())
		assert.NotEmpty(t, url.GetOriginalUrl())
	}
}

//func TestDeleteUserURLs(t *testing.T) {
//	client, ctx, cleanup := setupClient(t)
//	defer cleanup()
//
//	// Создаем ссылку для удаления
//	createResp, err := client.CreateShortURL(ctx, &api.OriginalURL{
//		Url: "https://example.com/to-delete",
//	})
//	if err != nil {
//		t.Fatalf("CreateShortURL failed: %v", err)
//	}
//
//	_, err = client.DeleteUserURLs(ctx, &api.DeleteUserURLsRequest{
//		Urls: []string{createResp.GetUrl()},
//	})
//	if err != nil {
//		t.Fatalf("DeleteUserURLs failed: %v", err)
//	}
//
//	// Проверяем, что ссылка больше не доступна
//	res, err := client.GetOriginalURL(ctx, &api.ShortURL{
//		Url: createResp.GetUrl(),
//	})
//	fmt.Println(res)
//	assert.Equal(t, codes.NotFound, status.Code(err))
//}

//func TestGetStats(t *testing.T) {
//	client, ctx, cleanup := setupClient(t)
//	defer cleanup()
//
//	// Создаем несколько ссылок для статистики
//	for i := 0; i < 3; i++ {
//		_, err := client.CreateShortURL(ctx, &api.OriginalURL{
//			Url: "https://example.com/stats" + string(rune(i)),
//		})
//		if err != nil {
//			t.Fatalf("CreateShortURL failed: %v", err)
//		}
//	}
//
//	// Получаем статистику
//	resp, err := client.GetStats(ctx, &api.Empty{})
//	if err != nil {
//		t.Fatalf("GetStats failed: %v", err)
//	}
//
//	assert.GreaterOrEqual(t, resp.GetUrls(), int64(3))
//	assert.GreaterOrEqual(t, resp.GetUsers(), int64(1))
//}
//
//func TestUnauthorizedAccess(t *testing.T) {
//	client, _, cleanup := setupClient(t)
//	defer cleanup()
//
//	// Контекст без авторизации
//	ctx := context.Background()
//
//	// Проверяем все методы на отсутствие доступа
//	_, err := client.CreateShortURL(ctx, &api.OriginalURL{Url: "https://example.com"})
//	assert.Equal(t, codes.Unauthenticated, status.Code(err))
//
//	_, err = client.GetUserURLs(ctx, &api.Empty{})
//	assert.Equal(t, codes.Unauthenticated, status.Code(err))
//
//	_, err = client.DeleteUserURLs(ctx, &api.DeleteUserURLsRequest{Urls: []string{"test"}})
//	assert.Equal(t, codes.Unauthenticated, status.Code(err))
//}

func setupClient(t *testing.T) (api.ShortenerClient, context.Context, func()) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}

	// Создаем контекст с тестовым userID
	userID := "test-user-id"
	md := metadata.Pairs(string(auth.UserIDKey), userID)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	cleanup := func() {
		conn.Close()
	}

	return api.NewShortenerClient(conn), ctx, cleanup
}
