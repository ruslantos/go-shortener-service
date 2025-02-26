package authheader

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
)

func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		var userID string
		var valid bool

		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				userID, valid = verifyToken(parts[1])
			}
		}

		if !valid {
			userID = generateUserID()
			token := createSignedToken(userID)
			w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
			//logger.GetLogger().Info("Новый Authorization токен создан", zap.String("userID", userID))
		} else {
			//fmt.Printf("Токен Authorization прошел проверку: %s\n", userID)
		}

		// передаем userID в контекст запроса
		r = setUserIDToContext(r, userID)

		next.ServeHTTP(w, r)
	})
}

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

func setUserIDToContext(r *http.Request, userID string) *http.Request {
	logger.GetLogger().Debug("userID передан в контекст", zap.String("userID", userID))
	ctx := context.WithValue(r.Context(), UserIDKey, userID)
	return r.WithContext(ctx)
}
