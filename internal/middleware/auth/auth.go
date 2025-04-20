package authheader

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

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

var (
	secretKey = []byte("secret-key")
)

type contextKey string

// UserIDKey ключ для хранения userID в контексте запроса.
const UserIDKey contextKey = "userID"

// CookieMiddleware middleware для обработки аутентификации через куки и Authorization header.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user")
		cookieUserID, cookieValid := verifyCookie(cookie)

		authHeader := r.Header.Get("Authorization")
		authUserID, authTokenValid := verifyAuthToken(authHeader)

		switch {
		// если кука валидная - используем ее
		case cookieValid && err == nil:
			logger.GetLogger().Debug("Получен userID из Cookie", zap.String("userID", cookieUserID))
			r = setUserIDToContext(r, cookieUserID)

		// если кука невалидная, используем Authorization токен
		case authTokenValid:
			logger.GetLogger().Debug("Получен userID из Authorization токена", zap.String("userID", authUserID))
			r = setUserIDToContext(r, authUserID)

		// генерим новый userID и возвращаем в куке и Authorization хэдере
		default:
			userID := uuid.New().String()
			logger.GetLogger().Debug("Сгенерирован новый userID", zap.String("userID", userID))
			newCookie := createSignedCookie(userID)
			http.SetCookie(w, &newCookie)

			r = setUserIDToContext(r, userID)

		}

		next.ServeHTTP(w, r)
	})
}

// методы для Cookie

// createSignedCookie создает подписанную куку с userID.
func createSignedCookie(userID string) http.Cookie {
	cookieValue := createToken(userID)

	cookie := http.Cookie{
		Name:     "user",
		Value:    cookieValue,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	}

	return cookie
}

// verifyCookie проверяет подписанную куку и возвращает userID и флаг валидности.
func verifyCookie(cookie *http.Cookie) (string, bool) {
	if cookie == nil {
		return "", false
	}

	return verifyToken(cookie.Value)
}

// методы для Auth хэдера

// createSignedAuthToken создает подписанный Authorization токен с userID.
func createSignedAuthToken(userID string) string {
	return createToken(userID)
}

// verifyAuthToken проверяет Authorization токен и возвращает userID и флаг валидности.
func verifyAuthToken(token string) (string, bool) {
	authToken := strings.SplitN(token, " ", 2)
	if len(authToken) != 2 && authToken[0] != "Bearer" {
		return "", false
	}

	return verifyToken(authToken[1])
}

// общие методы

// createToken создает подписанную строку с userID.
func createToken(userID string) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(userID))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s|%s", userID, signature)
}

// verifyToken проверяет подписанную строку и возвращает userID и флаг валидности.
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

// setUserIDToContext устанавливает userID в контекст запроса.
func setUserIDToContext(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), UserIDKey, userID)
	return r.WithContext(ctx)
}
