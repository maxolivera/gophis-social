package api

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

type Application struct {
	Config   *Config
	Database *database.Queries
}

type Config struct {
	Addr     string
	Database *DBConfig
}

type DBConfig struct {
	Addr               string
	MaxOpenConnections int
	MaxIdleConnections int
	MaxIdleTime        time.Duration
}

func (app *Application) Start() error {
	mux := app.GetHandlers()

	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	// TODO(maxolivera): Change to structured logging
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

	r.Route("/v1", func(r chi.Router) {
		r.Get("/healthz", app.handlerHealthz)
	})

	return r
}
