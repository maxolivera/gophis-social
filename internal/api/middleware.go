package api

import (
	"errors"
	"net/http"
)

func (app *Application) middlewareBasicAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read auth header
			user, pass, ok := r.BasicAuth()
			if !ok {
				err := errors.New("authentication not provided")
				app.unauthorizedBasicErrorResponse(w, r, err)
			}

			// Check credentials
			if user != app.Config.Authentication.BasicAuth.Username || pass != app.Config.Authentication.BasicAuth.Password {
				err := errors.New("authentication do not match")
				app.unauthorizedBasicErrorResponse(w, r, err)
			}
		})
	}
}
