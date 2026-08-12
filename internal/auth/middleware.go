package auth

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type UserContextKey struct{}

func OptionalAuthMiddleware(next http.Handler, sessionManager *scs.SessionManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userId := sessionManager.GetInt(r.Context(), "user_id")

		// if theres a user, grab the id but if not don't worry
		if userId > 0 {
			ctx := context.WithValue(r.Context(), UserContextKey{}, userId)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "http://localhost:8080/login", http.StatusFound)

}

func RequiredAuthMiddleware(next http.Handler, sessionManager *scs.SessionManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userId := sessionManager.GetInt(r.Context(), "user_id")
		// if theres a user, grab the id
		// if not make them sign in

		if userId > 0 {
			ctx := context.WithValue(r.Context(), UserContextKey{}, userId)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		} else {
			RedirectToLogin(w, r)
		}

	})
}
