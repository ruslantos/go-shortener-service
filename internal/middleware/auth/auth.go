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

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

var (
	secretKey = []byte("secret-key")
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
)

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
			userID := generateUserID()
			logger.GetLogger().Debug("Сгенерирован новый userID", zap.String("userID", userID))
			newCookie := createSignedCookie(userID)
			http.SetCookie(w, &newCookie)

			newAuthToken := createSignedAuthToken(userID)
			w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", newAuthToken))

			r = setUserIDToContext(r, userID)

		}

		next.ServeHTTP(w, r)
	})
}

// методы для Cookie
func createSignedCookie(userID string) http.Cookie {
	cookieValue := createToken(userID)

	cookie := http.Cookie{
		Name:     "user",
		Value:    cookieValue,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
	}

	return cookie
}
func verifyCookie(cookie *http.Cookie) (string, bool) {
	if cookie == nil {
		return "", false
	}

	return verifyToken(cookie.Value)
}

// методы для Auth хэдера
func createSignedAuthToken(userID string) string {
	return createToken(userID)
}
func verifyAuthToken(token string) (string, bool) {
	authToken := strings.SplitN(token, " ", 2)
	if len(authToken) != 2 && authToken[0] != "Bearer" {
		return "", false
	}

	return verifyToken(authToken[1])
}

// общие методы
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
func generateUserID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
func setUserIDToContext(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), UserIDKey, userID)
	return r.WithContext(ctx)
}
