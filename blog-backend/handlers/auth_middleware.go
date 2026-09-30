package handlers

import (
	"net/http"
	"strings"

	"github.com/gg582/echo-blog/blog-backend/utils"
)

// RequireAuth wraps a handler so it only runs when the request carries a
// valid Bearer token. Otherwise it responds with 401.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" || token == header || !utils.ValidateToken(token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
