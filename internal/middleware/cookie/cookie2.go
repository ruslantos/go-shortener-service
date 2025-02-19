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

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем куку
		cookie, err := r.Cookie("user")
		if err != nil || cookie == nil {
			// Кука не существует, создаем новую
			userID := generateUserID()
			newCookie := createSignedCookie(userID)
			http.SetCookie(w, &newCookie)
			fmt.Printf("Новая кука создана: %s\n", userID)
			// Передаем userID в контекст запроса
			r = setUserIDToContext(r, userID)
		} else {
			// Проверяем подлинность куки
			userID, valid := verifyCookie(cookie)
			if !valid {
				// Кука не прошла проверку, создаем новую
				userID = generateUserID()
				newCookie := createSignedCookie(userID)
				http.SetCookie(w, &newCookie)
				fmt.Printf("Кука не прошла проверку, новая кука создана: %s\n", userID)
			}
			// Передаем userID в контекст запроса
			r = setUserIDToContext(r, userID)
		}

		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}

// Функция для создания подписанной куки
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

// Функция для проверки подлинности куки
func verifyCookie(cookie *http.Cookie) (string, bool) {
	// Разделяем значение куки на userID и подпись
	parts := strings.SplitN(cookie.Value, "|", 2)
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

// Функция для генерации уникального идентификатора пользователя
func generateUserID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Вспомогательная функция для передачи userID в контекст запроса
func setUserIDToContext(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), "userID", userID)
	return r.WithContext(ctx)
}

// Вспомогательная функция для получения userID из контекста запроса
func getUserIDFromContext(r *http.Request) string {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		return ""
	}
	return userID
}
