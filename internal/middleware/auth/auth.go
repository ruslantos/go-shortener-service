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
	secretKey = []byte("secret-key") // Секретный ключ для подписи
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
			logger.GetLogger().Info("В запросе валидная кука", zap.String("userID", cookieUserID))
			r = setUserIDToContext(r, cookieUserID)
		// если кука невалидная, используем Authorization токен
		case authTokenValid:
			r = setUserIDToContext(r, authUserID)
		// генерим новый userID и возвращаем в куке и Authorization хэдере
		default:
			userID := generateUserID()
			newCookie := createSignedCookie(userID)
			http.SetCookie(w, &newCookie)

			newAuthToken := createSignedToken(userID)
			w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", newAuthToken))

			r = setUserIDToContext(r, userID)

		}

		next.ServeHTTP(w, r)
	})
}

func generateUserID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
func setUserIDToContext(r *http.Request, userID string) *http.Request {
	logger.GetLogger().Debug("userID передан в контекст", zap.String("userID", userID))
	ctx := context.WithValue(r.Context(), UserIDKey, userID)
	return r.WithContext(ctx)
}

// методы для Cookie
func createSignedCookie(userID string) http.Cookie {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(userID))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	cookieValue := fmt.Sprintf("%s|%s", userID, signature)

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
	parts := strings.SplitN(cookie.Value, "|", 2)
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

// методы для Auth хэдера
func createSignedToken(userID string) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(userID))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	tokenValue := fmt.Sprintf("%s|%s", userID, signature)

	return tokenValue
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
func verifyAuthToken(token string) (string, bool) {
	authToken := strings.SplitN(token, " ", 2)
	if len(authToken) != 2 && authToken[0] != "Bearer" {
		return "", false
	}

	parts := strings.SplitN(authToken[1], "|", 2)
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
