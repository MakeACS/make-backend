package auth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/crewjam/saml/samlsp"
)

type UserContextKey struct{}

type User struct {
	id int
}

func AuthContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := samlsp.SessionFromContext(r.Context())
		if session == nil {
			http.Error(w, "Unauthorized: No active session", http.StatusUnauthorized)
			return
		}

		claims := session.(samlsp.JWTSessionClaims)

		user_id, err := strconv.Atoi(claims.Attributes.Get("make_user_id"))
		if err != nil {
			http.Error(w, "Internal Server Error: Invalid user ID", http.StatusInternalServerError)
			return
		}

		user := &User{
			id: user_id,
		}

		ctx := context.WithValue(r.Context(), UserContextKey{}, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
