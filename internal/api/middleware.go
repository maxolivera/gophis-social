package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maxolivera/gophis-social-network/internal/models"
)

type contextKey string

const (
	contextKeyLoggedUser = contextKey("loggedUser")
	contextKeyRouteUser  = contextKey("routeUser")
)

func (app *Application) middlewareAuthToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := errors.New("authorization header is missing")
			app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			err := errors.New("authorization header is malformed")
			app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
		}

		tokenStr := parts[1]
		token, err := app.Authenticator.ValidateToken(tokenStr)
		if err != nil {
			err := fmt.Errorf("error during token validation: %v", err)
			app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
		}

		pgID := pgtype.UUID{}

		{ // parse user id
			claims := token.Claims.(jwt.MapClaims)
			userIDRaw, ok := claims["sub"]
			if !ok {
				err := errors.New("user ID (sub) not found in token claims")
				app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
			}

			userIDStr, ok := userIDRaw.(string)
			if !ok {
				err := errors.New("user ID (sub) in token claims is not a valid string")
				app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				err := fmt.Errorf("user ID (sub) is not a valid UUID: %v", err)
				app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
				return
			}

			pgID.Bytes = userID
			pgID.Valid = true
		}

		dbUser, err := app.Database.GetUserById(r.Context(), pgID)
		if err != nil {
			switch err {
			case pgx.ErrNoRows:
				err = fmt.Errorf("user not found %v", err)
			default:
				err = fmt.Errorf("error retrieving user from database: %v", err)
			}
			app.respondWithError(w, r, http.StatusUnauthorized, err, "")
			return
		}

		user := models.DBUserToUser(dbUser)

		ctx := context.WithValue(r.Context(), contextKeyLoggedUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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

func (app *Application) middlewareRouteUserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username := r.PathValue("username")
		if username == "" {
			err := fmt.Errorf("username not provided")
			// TODO(maolivera): maybe another message?
			app.respondWithError(w, r, http.StatusBadRequest, err, err.Error())
			return
		}

		dbUser, err := app.Database.GetUserByUsername(
			ctx,
			username,
		)

		if err != nil {
			switch err {
			case pgx.ErrNoRows:
				err := fmt.Errorf("username not found: %v", err)
				app.respondWithError(w, r, http.StatusNotFound, err, "user not found")
			default:
				app.respondWithError(w, r, http.StatusInternalServerError, err, "")
			}
			return
		}
		user := models.DBUserToUser(dbUser)

		ctx = context.WithValue(ctx, contextKeyRouteUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getRouteUser(r *http.Request) models.User {
	return r.Context().Value(contextKeyRouteUser).(models.User)
}

func getLoggedUser(r *http.Request) models.User {
	return r.Context().Value(contextKeyLoggedUser).(models.User)
}
