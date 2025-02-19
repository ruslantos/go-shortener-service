package user

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

var secretKey = []byte("secret-key")

type User struct {
	cookie http.Cookie
}

func NewUserService() *User {
	return &User{}
}
func (u *User) UserFromContext(ctx context.Context) string {
	cookie := ctx.Value("user")
	if cookie != nil {
		logger.GetLogger().Info("get cookie", zap.String("cookie", cookie.(string)))

		userID, ok := verifyCookie(cookie.(string))
		if ok {
			return userID
		}
	}

	userID := generateuserID()
	u.cookie = createSignedCookie(userID)
	return userID
}
func createSignedCookie(userID string) http.Cookie {
	// Создаем подпись для userID
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(userID))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	// Создаем значение куки: userID + "|" + подпись
	cookieValue := fmt.Sprintf("%s|%s", userID, signature)

	// Создаем куку
	cookie := http.Cookie{
		Name:     "user",
		Value:    cookieValue,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour), // Срок действия куки
		HttpOnly: true,                           // Защита от XSS
		Secure:   true,                           // Только для HTTPS
	}

	return cookie
}
func verifyCookie(cookie string) (string, bool) {
	// Разделяем значение куки на userID и подпись
	parts := split(cookie, "|")
	if len(parts) != 2 {
		return "", false
	}

	userID := parts[0]
	signature := parts[1]

	// Проверяем подпись
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(userID))
	expectedSignature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return "", false
	}

	return userID, true
}
func split(s, sep string) []string {
	return strings.SplitN(s, sep, 2)
}
func generateuserID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
