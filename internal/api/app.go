package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxolivera/gophis-social-network/docs"
	"github.com/maxolivera/gophis-social-network/internal/auth"
	"github.com/maxolivera/gophis-social-network/internal/cache"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/models"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type Application struct {
	Config        *Config
	Pool          *pgxpool.Pool
	Database      *database.Queries
	Cache         *cache.Storage
	Logger        *zap.SugaredLogger
	Authenticator auth.Authenticator
}

type Config struct {
	Addr           string
	Database       *DBConfig
	Environment    string
	Version        string
	ApiUrl         string
	ExpirationTime time.Duration
	Authentication *AuthConfig
	Redis          *RedisConfig
}

type RedisConfig struct {
	Enabled  bool
	Address  string
	Password string
	Database int
}

type AuthConfig struct {
	BasicAuth *BasicAuth
	Token     *TokenConfig
}

type TokenConfig struct {
	Secret         string
	ExpirationTime time.Duration
	Issuer         string
}

type BasicAuth struct {
	Username string
	Password string
}

type DBConfig struct {
	Addr               string
	MaxOpenConnections int
	MaxIdleConnections int
	MaxIdleTime        time.Duration
}

//	@title			Gophis Social API
//	@description	API for Gophis Social, the best and simplest Social Network

// @BasePath /v1

// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
func (app *Application) Start() error {
	mux := app.GetHandlers()

	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	// TODO(maolivera): Maybe move this to main?

	// == Graceful Shutdown ==
	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.Logger.Infow("signal caught", "signal", s.String())
		shutdown <- srv.Shutdown(ctx)
	}()

	app.Logger.Infow("server has started", "addr", app.Config.Addr, "env", app.Config.Environment)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.Logger.Infow("server has stopped", "addr", app.Config.Addr, "env", app.Config.Environment)

	return nil
}

func (app *Application) GetHandlers() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// timeout on request context
	r.Use(middleware.Timeout(60 * time.Second))

	// == API DOCS ==
	docsURL := fmt.Sprintf("%s/swagger/doc.json", app.Config.Addr)
	docs.SwaggerInfo.Version = app.Config.Version
	docs.SwaggerInfo.Host = app.Config.ApiUrl
	docs.SwaggerInfo.BasePath = "/v1"

	r.Route("/v1", func(r chi.Router) {
		r.With(app.middlewareBasicAuth()).Get("/healthz", app.handlerHealthz)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		// Non-auth routes
		r.Post("/register", app.handlerCreateUser)
		r.Post("/activate/{token}", app.handlerActivateUser)
		r.Post("/token", app.handlerCreateToken)

		// Add routes
		r.Route("/users", func(r chi.Router) {
			r.Use(app.middlewareAuthToken)
			r.Route("/{username}", func(r chi.Router) {
				r.Use(app.middlewareRouteUserContext)

				r.Get("/", app.handlerGetUser)
				r.Patch("/", app.handlerUpdateUser)
				// TODO(maolivera): add role to modify user
				// TODO(maolivera): add hard delete for admins
				r.Delete("/", app.handlerSoftDeleteUser)

				r.Put("/follow", app.handlerFollowUser)
				r.Put("/unfollow", app.handlerUnfollowUser)
			})
		})

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.middlewareAuthToken)

			r.Post("/", app.handlerCreatePost)
			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.middlewarePostContext)

				r.Get("/", app.handlerGetPost)
				r.Patch("/", app.middlewarePostPermissions(models.RoleModerator, true, app.handlerUpdatePost))
				r.Delete("/", app.middlewarePostPermissions(models.RoleAdmin, true, app.handlerSoftDeletePost))
				r.Delete("/hard", app.middlewarePostPermissions(models.RoleAdmin, false, app.handlerHardDeletePost))
			})
		})

		r.With(app.middlewareAuthToken).Get("/feed", app.handlerFeed)
		r.With(app.middlewareAuthToken).Get("/search", app.handlerSearch)
	})

	return r
}
