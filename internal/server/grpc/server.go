package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ruslantos/go-shortener-service/api"
	auth "github.com/ruslantos/go-shortener-service/internal/interceptors/authinterceptor"
	"github.com/ruslantos/go-shortener-service/internal/models"
	"github.com/ruslantos/go-shortener-service/internal/service"
)

// linksService определяет интерфейс для работы с ссылками.
type linksService interface {
	Add(ctx context.Context, long string) (string, error)
	Get(ctx context.Context, shortLink string) (string, error)
	Ping(ctx context.Context) error
	AddBatch(ctx context.Context, links []models.Link) ([]models.Link, error)
	GetUserUrls(ctx context.Context) ([]models.Link, error)
	ConsumeDeleteURLs(data service.DeletedURLs)
	GetStats(ctx context.Context) (urls int, users int, err error)
}

// Server реализует gRPC сервер для сервиса сокращения ссылок.
type Server struct {
	api.UnimplementedShortenerServer
	service linksService
}

// New создает новый экземпляр gRPC сервера с указанным сервисом для работы с ссылками.
func New(service linksService) *Server {
	return &Server{service: service}
}

// Register регистрирует сервер в gRPC сервере.
func (s *Server) Register(grpcServer *grpc.Server) {
	api.RegisterShortenerServer(grpcServer, s)
}

// CreateShortURL создает короткую ссылку для переданного оригинального URL.
func (s *Server) CreateShortURL(ctx context.Context, req *api.OriginalURL) (*api.ShortURL, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}

	shortURL, err := s.service.Add(ctx, req.Url)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.ShortURL{Url: shortURL}, nil
}

// GetOriginalURL возвращает оригинальный URL по короткой ссылке.
func (s *Server) GetOriginalURL(ctx context.Context, req *api.ShortURL) (*api.OriginalURL, error) {
	originalURL, err := s.service.Get(ctx, req.Url)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &api.OriginalURL{Url: originalURL}, nil
}

// Ping проверяет доступность сервиса.
func (s *Server) Ping(ctx context.Context, _ *api.Empty) (*api.PingResponse, error) {
	err := s.service.Ping(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &api.PingResponse{Success: true}, nil
}

// CreateShortURLBatch создает несколько коротких ссылок для переданных оригинальных URL.
func (s *Server) CreateShortURLBatch(ctx context.Context, req *api.OriginalURLsBatch) (*api.ShortURLsBatch, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}

	urls := make([]models.Link, 0, len(req.Urls))
	for _, u := range req.Urls {
		urls = append(urls, models.Link{OriginalURL: u.Url})
	}

	result, err := s.service.AddBatch(ctx, urls)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.ShortURLsBatch{Urls: make([]*api.ShortURL, 0, len(result))}
	for _, r := range result {
		resp.Urls = append(resp.Urls, &api.ShortURL{Url: r.ShortURL})
	}

	return resp, nil
}

// GetUserURLs возвращает все URL, созданные текущим пользователем.
func (s *Server) GetUserURLs(ctx context.Context, _ *api.Empty) (*api.UserURLs, error) {
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}

	urls, err := s.service.GetUserUrls(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.UserURLs{Urls: make([]*api.UserURL, 0, len(urls))}
	for _, u := range urls {
		resp.Urls = append(resp.Urls, &api.UserURL{
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
		})
	}

	return resp, nil
}

// DeleteUserURLs помечает указанные URL как удаленные для текущего пользователя.
func (s *Server) DeleteUserURLs(ctx context.Context, req *api.DeleteUserURLsRequest) (*api.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}

	for _, url := range req.Urls {
		urls := service.DeletedURLs{
			URLs:   url,
			UserID: userID,
		}
		s.service.ConsumeDeleteURLs(urls)
	}

	return &api.Empty{}, nil
}

// GetStats возвращает статистику сервиса: количество URL и пользователей.
func (s *Server) GetStats(ctx context.Context, _ *api.Empty) (*api.Stats, error) {
	urls, users, err := s.service.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.Stats{
		Urls:  int64(urls),
		Users: int64(users),
	}, nil
}

// getUserIDFromContext извлекает идентификатор пользователя из контекста.
func getUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "user not found")
	}

	return userID, nil
}
