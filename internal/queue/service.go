package queue

import (
	"context"

	"go.uber.org/zap"

	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

type linksStorage interface {
	DeleteUserUrls(ctx context.Context, ids []string, userID string) error
}

type QueueService struct {
	linksStorage linksStorage
}

func NewQueueService(linksStorage linksStorage) *QueueService {
	return &QueueService{
		linksStorage: linksStorage,
	}
}

func (q *QueueService) DeleteUserUrls(ctx context.Context, ids []string) error {
	userID := getUserIDFromContext(ctx)

	return q.linksStorage.DeleteUserUrls(ctx, ids, userID)
}

func getUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(auth.UserIDKey).(string)
	if !ok {
		return ""
	}
	logger.GetLogger().Info("get userID", zap.String("userID", userID))
	return userID
}
