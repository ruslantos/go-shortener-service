package authheader

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTokenAndVerifyToken(t *testing.T) {
	userID := "test-user-id"
	token := createToken(userID)

	// Проверяем, что токен создается корректно
	parts := strings.SplitN(token, "|", 2)
	require.Len(t, parts, 2, "Token should have two parts separated by |")
	assert.Equal(t, userID, parts[0], "First part should be userID")

	// Проверяем верификацию токена
	verifiedUserID, valid := verifyToken(token)
	assert.True(t, valid, "Token should be valid")
	assert.Equal(t, userID, verifiedUserID, "UserID should match")

	// Проверяем невалидные токены
	testCases := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"No separator", "invalidtoken"},
		{"Invalid signature", "userID|invalidsignature"},
		{"Tampered userID", "tampered|signature"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, valid := verifyToken(tc.token)
			assert.False(t, valid, "Token should be invalid")
		})
	}
}

func TestCreateSignedCookieAndVerifyCookie(t *testing.T) {
	userID := "test-user-id"
	cookie := createSignedCookie(userID)

	// Проверяем создание куки
	assert.Equal(t, "user", cookie.Name, "Cookie name should be 'user'")
	assert.Equal(t, "/", cookie.Path, "Cookie path should be /")
	assert.True(t, cookie.Expires.After(time.Now()), "Cookie should expire in the future")
	assert.True(t, cookie.HttpOnly, "Cookie should be HttpOnly")

	// Проверяем верификацию куки
	verifiedUserID, valid := verifyCookie(&cookie)
	assert.True(t, valid, "Cookie should be valid")
	assert.Equal(t, userID, verifiedUserID, "UserID should match")

	// Проверяем невалидные куки
	t.Run("Nil cookie", func(t *testing.T) {
		_, valid := verifyCookie(nil)
		assert.False(t, valid, "Nil cookie should be invalid")
	})

	t.Run("Invalid cookie value", func(t *testing.T) {
		invalidCookie := http.Cookie{
			Name:  "user",
			Value: "invalid|signature",
		}
		_, valid := verifyCookie(&invalidCookie)
		assert.False(t, valid, "Cookie with invalid value should be invalid")
	})
}

func TestCreateSignedAuthTokenAndVerifyAuthToken(t *testing.T) {
	userID := "test-user-id"
	token := createSignedAuthToken(userID)

	// Проверяем верификацию Authorization токена
	verifiedUserID, valid := verifyAuthToken("Bearer " + token)
	assert.True(t, valid, "Auth token should be valid")
	assert.Equal(t, userID, verifiedUserID, "UserID should match")

	// Проверяем невалидные Authorization токены
	testCases := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"No Bearer prefix", token},
		{"Invalid Bearer token", "Bearer invalid"},
		{"Malformed token", "Bearer userID|invalidsignature"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, valid := verifyAuthToken(tc.token)
			assert.False(t, valid, "Auth token should be invalid")
		})
	}
}

func TestCookieMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey)
		assert.NotEmpty(t, userID, "UserID should be set in context")
		w.WriteHeader(http.StatusOK)
	})

	t.Run("No cookie or auth header - should generate new userID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		middleware := CookieMiddleware(testHandler)
		middleware.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.NotEmpty(t, res.Header.Get("Set-Cookie"), "Should set new cookie")
	})

	t.Run("Valid cookie - should use cookie userID", func(t *testing.T) {
		userID := "test-user-id"
		cookie := createSignedCookie(userID)

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&cookie)
		rec := httptest.NewRecorder()

		middleware := CookieMiddleware(testHandler)
		middleware.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Empty(t, res.Header.Get("Set-Cookie"), "Should not set new cookie")
	})

	t.Run("Invalid cookie but valid auth header - should use auth header userID", func(t *testing.T) {
		userID := "test-user-id"
		token := createSignedAuthToken(userID)

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "user", Value: "invalid"})
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		middleware := CookieMiddleware(testHandler)
		middleware.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Empty(t, res.Header.Get("Set-Cookie"), "Should not set new cookie")
	})

	t.Run("Invalid cookie and invalid auth header - should generate new userID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "user", Value: "invalid"})
		req.Header.Set("Authorization", "Bearer invalid")
		rec := httptest.NewRecorder()

		middleware := CookieMiddleware(testHandler)
		middleware.ServeHTTP(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.NotEmpty(t, res.Header.Get("Set-Cookie"), "Should set new cookie")
	})
}

func TestSetUserIDToContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	userID := "test-user-id"

	newReq := setUserIDToContext(req, userID)
	ctxValue := newReq.Context().Value(UserIDKey)

	assert.Equal(t, userID, ctxValue, "UserID should be set in context")
}

func TestContextKeyType(t *testing.T) {
	// Проверяем, что UserIDKey имеет правильный тип
	var key contextKey = "userID"
	assert.IsType(t, key, UserIDKey, "UserIDKey should be of type contextKey")
}
