package main

import (
	"log"
	"net/http"
	"time"
)

type application struct {
	config config
}

type config struct {
	addr string
}

func (app *application) setHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthz", app.handlerHealthz)

	return mux
}

func (app *application) start(mux *http.ServeMux) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	// TODO(maxolivera): Change to structured logging
	log.Printf("starting to listen at %s", app.config.addr)

	return srv.ListenAndServe()
}
