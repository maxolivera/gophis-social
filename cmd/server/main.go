package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/maxolivera/gophis-social-network/internal/handlers"
)

func main() {
	// Get env values
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("error loading .env file:", err)
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		log.Fatalln("could not find ADDR environment value")
	}
	log.Println("using %s as ADDR", addr)

	// Set handlers configuration
	cfg := handlers.ApiConfig{
		Addr: addr,
	}

	app := &application{
		api: &cfg,
	}

	log.Fatalln(app.start())
}
