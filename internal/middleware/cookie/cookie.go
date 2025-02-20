package cookie

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	secretKey = []byte("your-secret-key") // Секретный ключ для подписи
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user")
		if err != nil || cookie == nil {
			if cookie == nil {
				fmt.Println("Кука отсутствует в запросе")
			}
			// создаем новую
			userID := generateUserID()
			newCookie := createSignedCookie(userID)
			http.SetCookie(w, &newCookie)
			fmt.Printf("Новая кука создана: %s\n", userID)
			// передаем userID в контекст запроса
			r = setUserIDToContext(r, userID)
		} else {
			// проверяем
			userID, valid := verifyCookie(cookie)
			if !valid {
				userID = generateUserID()
				newCookie := createSignedCookie(userID)
				http.SetCookie(w, &newCookie)
				fmt.Printf("Кука не прошла проверку, новая кука создана: %s\n", userID)
			}
			// передаем userID в контекст запроса
			r = setUserIDToContext(r, userID)
		}

		next.ServeHTTP(w, r)
	})
}

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

func generateUserID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func setUserIDToContext(r *http.Request, userID string) *http.Request {
	fmt.Printf("Кука передана в контекст: %s\n", userID)
	ctx := context.WithValue(r.Context(), UserIDKey, userID)
	return r.WithContext(ctx)
}
