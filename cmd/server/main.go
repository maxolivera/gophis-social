package main

import (
	"log"
)

func main() {
	// Set server configuration

	cfg := config{
		addr: ":8080",
	}

	app := &application{
		config: cfg,
	}

	log.Fatalln(
		app.start(app.setHandlers()),
	)
}
