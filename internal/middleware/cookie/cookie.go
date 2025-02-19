package cookie

import (
	"net/http"
)

const cookieKey = "user"

func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//cookie, err := r.Cookie("user")
		//if err != nil {
		//	next.ServeHTTP(w, r)
		//	return
		//}
		//
		//ctx := context.WithValue(r.Context(), cookieKey, cookie.Value)
		//r = r.WithContext(ctx)
		//
		//next.ServeHTTP(w, r)
	})
}
