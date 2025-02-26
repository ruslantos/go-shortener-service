package authheader

import (
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
	secretKey = []byte("your-secret-key") // Секретный ключ для подписи
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

		//if err != nil || cookie == nil {
		//	if cookie == nil {
		//		//logger.GetLogger().Info("Кука отсутствует в запросе")
		//	}
		//	// создаем новую
		//	userID := generateUserID()
		//	newCookie := createSignedCookie(userID)
		//	http.SetCookie(w, &newCookie)
		//	//logger.GetLogger().Info("Новая кука создана", zap.String("userID", userID))
		//	// не передаю в контекст тк userID передается из Authorization хэдера
		//} else {
		//	// проверяем
		//	userID, valid := verifyCookie(cookie)
		//	if !valid {
		//		userID = generateUserID()
		//		newCookie := createSignedCookie(userID)
		//		http.SetCookie(w, &newCookie)
		//		//logger.GetLogger().Info("Кука не прошла проверку, новая кука создана", zap.String("userID", userID))
		//
		//	}
		//	logger.GetLogger().Info("В запросе валидная кука", zap.String("userID", userID))
		//	// не передаю в контекст тк userID передается из Authorization хэдера
		//}

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

func generateUserID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
