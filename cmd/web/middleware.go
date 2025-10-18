package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/yoqub-davlatov/snippetbox/pkg/models"
)

const (
	headerXSSProtection      = "X-XSS-Protection"
	headerXSSProtectionValue = "1; mode=block"

	headerFrameOptions      = "X-Frame-Options"
	headerFrameOptionsValue = "deny"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerXSSProtection, headerXSSProtectionValue)
		w.Header().Set(headerFrameOptions, headerFrameOptionsValue)

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	})

	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exists := app.session.Exists(r, authenticatedUserIDSessionKey)
		if !exists {
			next.ServeHTTP(w, r)
			return
		}

		token := app.session.GetInt(r, authenticatedUserIDSessionKey)
		user, err := app.users.Get(token)
		if err != nil && !errors.Is(err, models.ErrNoRecord) {
			app.serverError(w, err)
			return
		}
		if errors.Is(err, models.ErrNoRecord) || !user.Active {
			app.session.Remove(r, authenticatedUserIDSessionKey)
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyIsAuthenticated, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
