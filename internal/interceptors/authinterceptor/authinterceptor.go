package authinterceptor

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

type contextKey string

const (
	UserIDKey       contextKey = "userID"
	authHeaderKey   string     = "authorization"
	cookieHeaderKey string     = "grpcgateway-cookie" // Для gRPC-gateway
	userCookieName  string     = "user"
	tokenTypeBearer string     = "bearer"
)

var (
	secretKey = []byte("secret-key")
)

// AuthInterceptor возвращает grpc.UnaryServerInterceptor для аутентификации
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Получаем метаданные из контекста
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata not found")
		}

		var userID string
		var valid bool

		// Проверяем куки из метаданных (для gRPC-gateway)
		if cookieValues := md.Get(cookieHeaderKey); len(cookieValues) > 0 {
			for _, cookie := range cookieValues {
				if strings.HasPrefix(cookie, userCookieName+"=") {
					cookieValue := strings.TrimPrefix(cookie, userCookieName+"=")
					userID, valid = verifyToken(cookieValue)
					if valid {
						logger.GetLogger().Debug("Получен userID из Cookie", zap.String("userID", userID))
						break
					}
				}
			}
		}

		// Если кука невалидна, проверяем Authorization header
		if !valid {
			if authHeaders := md.Get(authHeaderKey); len(authHeaders) > 0 {
				userID, valid = verifyAuthToken(authHeaders[0])
				if valid {
					logger.GetLogger().Debug("Получен userID из Authorization токена", zap.String("userID", userID))
				}
			}
		}

		// Если ни кука, ни токен невалидны - генерируем новый userID
		if !valid {
			userID = uuid.New().String()
			logger.GetLogger().Debug("Сгенерирован новый userID", zap.String("userID", userID))
		}

		// Устанавливаем userID в контекст
		newCtx := context.WithValue(ctx, UserIDKey, userID)

		// Добавляем токен в метаданные для ответа (если нужно)
		if !valid {
			token := createToken(userID)
			header := metadata.Pairs(authHeaderKey, "Bearer "+token)
			grpc.SetHeader(newCtx, header)

			// Также можно установить куку для gRPC-gateway
			cookie := createSignedCookie(userID)
			grpc.SetHeader(newCtx, metadata.Pairs("set-cookie", cookie.String()))
		}

		return handler(newCtx, req)
	}
}

// Методы для работы с токенами (аналогичны HTTP-версии)

func createSignedCookie(userID string) *http.Cookie {
	return &http.Cookie{
		Name:     userCookieName,
		Value:    createToken(userID),
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}
}

func verifyAuthToken(token string) (string, bool) {
	authToken := strings.SplitN(token, " ", 2)
	if len(authToken) != 2 || !strings.EqualFold(authToken[0], tokenTypeBearer) {
		return "", false
	}

	return verifyToken(authToken[1])
}

func createToken(userID string) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(userID))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s|%s", userID, signature)
}

func verifyToken(token string) (string, bool) {
	parts := strings.SplitN(token, "|", 2)
	if len(parts) != 2 {
		return "", false
	}

	userID := parts[0]
	signature := parts[1]

	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(userID))
	expectedSignature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return "", false
	}

	return userID, true
}
