package authheader

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
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
			fmt.Printf("Новый Authorization токен создан: %s\n", userID)
		} else {
			fmt.Printf("Токен Authorization прошел проверку: %s\n", userID)
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

func setUserIDToContext(r *http.Request, userID string) *http.Request {
	fmt.Printf("userID передан в контекст: %s\n", userID)
	ctx := context.WithValue(r.Context(), UserIDKey, userID)
	return r.WithContext(ctx)
}
