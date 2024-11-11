package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/maxolivera/gophis-social-network/docs"
	"github.com/maxolivera/gophis-social-network/internal/database"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Application struct {
	Config   *Config
	Database *database.Queries
}

type Config struct {
	Addr        string
	Database    *DBConfig
	Environment string
	Version     string
	ApiUrl      string
}

type DBConfig struct {
	Addr               string
	MaxOpenConnections int
	MaxIdleConnections int
	MaxIdleTime        time.Duration
}

//	@title			Gophis Social API
//	@description	API for Gophis Social, the best and simplest Social Network

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

	// TODO(maolivera): Change to structured logging
	log.Printf("starting to listen at %s", app.Config.Addr)

	return srv.ListenAndServe()
}

func (app *Application) GetHandlers() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// timeout on reuqest context
	r.Use(middleware.Timeout(60 * time.Second))

	// == API DOCS ==
	docsURL := fmt.Sprintf("%s/swagger/doc.json", app.Config.Addr)
	docs.SwaggerInfo.Version = app.Config.Version
	docs.SwaggerInfo.Host = app.Config.ApiUrl
	docs.SwaggerInfo.BasePath = "/v2"

	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthz", app.handlerHealthz)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		// TODO(maolivera): Add timeout within context
		r.Route("/users", func(r chi.Router) {
			// TODO(maolivera): Change / to be directly the base path (not under "/"), if authenticated
			r.Post("/", app.handlerCreateUser)

			r.Route("/{username}", func(r chi.Router) {
				r.Use(app.middlewareUserContext)

				r.Get("/", app.handlerGetUser)
				r.Patch("/", app.handlerUpdateUser)
				// TODO(maolivera): add hard delete for admins
				r.Delete("/", app.handlerSoftDeleteUser)

				r.Put("/follow", app.handlerFollowUser)
				r.Put("/unfollow", app.handlerUnfollowUser)
			})
		})

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.handlerCreatePost)
			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.middlewarePostContext)

				r.Get("/", app.handlerGetPost)
				r.Patch("/", app.handlerUpdatePost)
				// TODO(maolivera): add hard delete for admins
				r.Delete("/", app.handlerSoftDeletePost)
			})
		})

		// TODO(maolivera): Change to auth
		r.Get("/", app.handlerFeed)
		r.Get("/search", app.handlerSearch)
	})

	return r
}
