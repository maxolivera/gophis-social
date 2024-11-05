package main

import (
	"log"
	"net/http"
	"time"

	"github.com/maxolivera/gophis-social-network/internal/handlers"
)

type application struct {
	api *handlers.ApiConfig
}

func (app *application) start() error {
	mux := app.api.GetHandlers()

	srv := &http.Server{
		Addr:         app.api.Addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}


	// TODO(maxolivera): Change to structured logging
	log.Printf("starting to listen at %s", app.api.Addr)

	return srv.ListenAndServe()
}
