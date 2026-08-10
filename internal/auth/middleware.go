package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type UserContextKey struct{}

var ErrNotAuthenticated error = errors.New("not authenticated")

func UserIDFromContext(ctx context.Context) *int {
	userVal := ctx.Value(UserContextKey{})
	if userVal == nil {
		slog.Error("UserIDFromContext called with no userID in context. Caused by mismatch between route protection middleware and what resolvers actually do")
		// TODO temp until real auth works
		one := 1
		return &one
	}
	userID := userVal.(int)
	return &userID

}

func AuthContextMiddleware(next http.Handler, sessionManager *scs.SessionManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userId := sessionManager.GetInt(r.Context(), "user_id")

		if userId > 0 {
			ctx := context.WithValue(r.Context(), UserContextKey{}, userId)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
