package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/maxolivera/gophis-social-network/internal/storage"
	"github.com/maxolivera/gophis-social-network/internal/storage/models"
)

type contextKey string

const (
	contextKeyPost           = contextKey("post")
	contextKeyLoggedUser     = contextKey("loggedUser")
	contextKeyRouteUser      = contextKey("routeUser")
	contextKeyLoggedUserRole = contextKey("loggedUserRole")
)

func (app *Application) middlewareRateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.Config.RateLimiter.Enabled {
			if allow, retryAfter := app.RateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimiteExceededErrorResponse(w, r, retryAfter.String())
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (app *Application) middlewareAuthToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := errors.New("authorization header is missing")
			app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			err := errors.New("authorization header is malformed")
			app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		tokenStr := parts[1]
		token, err := app.Authenticator.ValidateToken(tokenStr)
		if err != nil {
			err := fmt.Errorf("error during token validation: %v", err)
			app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		// parse user id
		claims := token.Claims.(jwt.MapClaims)
		usernameRaw, ok := claims["sub"]
		if !ok {
			err := errors.New("user ID (sub) not found in token claims")
			app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
		}

		username, ok := usernameRaw.(string)
		if !ok {
			err := errors.New("user ID (sub) in token claims is not a valid string")
			app.respondWithError(w, r, http.StatusUnauthorized, err, "Unauthorized")
		}

		user, err := app.getUser(r, username)
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

		ctx := r.Context()
		ctx = context.WithValue(ctx, contextKeyLoggedUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) middlewareBasicAuth(next http.Handler) http.Handler {
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

		next.ServeHTTP(w, r)
	})
}

func (app *Application) middlewareRouteUserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username := r.PathValue("username")
		if username == "" {
			err := fmt.Errorf("username not provided")
			app.respondWithError(w, r, http.StatusBadRequest, err, err.Error())
			return
		}

		user, err := app.getUser(r, username)
		if err != nil {
			switch err {
			case pgx.ErrNoRows:
				err = fmt.Errorf("user not found: %v", err)
				app.respondWithError(w, r, http.StatusNotFound, err, "user not found")
			default:
				err = fmt.Errorf("error fetching user: %v", err)
				app.respondWithError(w, r, http.StatusInternalServerError, err, "")
			}
			return
		}

		ctx = context.WithValue(ctx, contextKeyRouteUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) middlewarePostContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := r.PathValue("postID")
		if idStr == "" {
			err := fmt.Errorf("post_id was missing")
			app.respondWithError(w, r, http.StatusBadRequest, err, err.Error())
			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			err := fmt.Errorf("invalid post_id: %v", err)
			app.respondWithError(w, r, http.StatusBadRequest, err, "invalid post_id")
			return
		}

		post, err := app.Storage.Posts.GetByID(ctx, id)
		if err != nil {
			switch err {
			case storage.ErrNoRows:
				err := errors.New("post not found")
				app.respondWithError(w, r, http.StatusNotFound, err, "post not found")
			default:
				app.respondWithError(w, r, http.StatusInternalServerError, err, "")
			}
			return
		}
		comments, err := app.Storage.Comments.GetByPostID(ctx, post.ID)
		if err != nil {
			switch err {
			case storage.ErrNoRows:
			// NOTE(maolivera): It's ok if a post do not have comments
			default:
				app.respondWithError(w, r, http.StatusInternalServerError, err, "")
				return
			}
		}
		post.Comments = comments

		ctx = context.WithValue(ctx, contextKeyPost, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Is `userAllowed` is true, it will allow the user to perform the action "on itself", if not, it will only be allowed if role matches
func (app *Application) middlewarePostPermissions(requiredRole models.RoleType, userAllowed bool, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getLoggedUser(r)
		posts := getPost(r)

		// Check if user is the same
		if userAllowed && user.ID == posts.UserID {
			next.ServeHTTP(w, r)
			return
		}

		// If not, check if role level is enough
		role, err := app.Storage.Roles.GetByName(ctx, string(requiredRole))
		if err != nil {
			err = fmt.Errorf("error during role fetching: %v", err)
			app.respondWithError(w, r, http.StatusInternalServerError, err, "")
			return
		}

		if user.Role.Level < role.Level {
			// TODO(maolivera): Maybe add a field to the Application struct to keep the roles in memory?
			err := fmt.Errorf("role is not enough. Required '%s %d' vs '%s %d'", role.Name, role.Level, string(user.Role.Name), user.Role.Level)
			app.respondWithError(w, r, http.StatusForbidden, err, "forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getRouteUser(r *http.Request) *models.User {
	return r.Context().Value(contextKeyRouteUser).(*models.User)
}

func getLoggedUser(r *http.Request) *models.User {
	return r.Context().Value(contextKeyLoggedUser).(*models.User)
}

func getPost(r *http.Request) *models.Post {
	return r.Context().Value(contextKeyPost).(*models.Post)
}

func (app *Application) getUser(r *http.Request, username string) (*models.User, error) {
	ctx := r.Context()
	// No cache
	if !app.Config.Cache.Enabled {
		user, err := app.Storage.Users.GetByUsername(ctx, username)
		if err != nil {
			return nil, err
		}

		return user, nil
	}

	// Cache enabled
	// 1. Check cache
	var hit string
	var user *models.User
	var err error
	start := time.Now()
	user, err = app.Cache.Users.Get(ctx, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		// 2. Use DB instead
		user, err = app.Storage.Users.GetByUsername(ctx, username)
		if err != nil {
			return nil, err
		}

		// 3. Update cache
		if err = app.Cache.Users.Set(ctx, user); err != nil {
			return nil, err
		}
		hit = "miss"
	} else {
		hit = "hit"
	}

	elapsed := time.Since(start)

	app.Logger.Infow("fetching user",
		"cache", hit,
		"total time", elapsed,
		"fill", float32(app.Cache.Users.Len(ctx))/float32(app.Config.Cache.LRU.Capacity)*100,
	)

	// 4. Return user
	return user, nil
}
