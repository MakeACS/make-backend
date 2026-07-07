package auth

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

func AuthContextMiddleware(next http.Handler, sessionManager *scs.SessionManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userId := sessionManager.GetInt(r.Context(), "user_id")

		if userId > 0 {

			ctx := context.WithValue(r.Context(), "user_id", userId)
			r = r.WithContext(ctx)
		} else {
			http.Error(w, "Unathorized: No valid session found", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
