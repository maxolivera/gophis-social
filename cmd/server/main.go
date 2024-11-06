package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/maxolivera/gophis-social-network/internal/api"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/env"
)

func main() {
	// == ENV VALUES ==
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("error loading .env file:", err)
	}

	addr, err := env.GetString("ADDR")
	if err != nil {
		log.Fatalln(err)
	}

	dbUrl, err := env.GetString("DB_URL")
	if dbUrl == "" {
		log.Fatalln("could not find DB_URL environment value")
	}
	log.Println("Databse url found", dbUrl)

	maxOpenConns, err := env.GetInt("DB_MAX_OPEN_CONNS")
	if err != nil {
		log.Fatalln(err)
	}

	maxIdleConns, err := env.GetInt("DB_MAX_IDLE_CONNS")
	if err != nil {
		log.Fatalln(err)
	}

	// MaxIdleTime represents minutes
	maxIdleTime, err := env.GetInt("DB_MAX_IDLE_TIME")
	if err != nil {
		log.Fatalln(err)
	}

	// == CONFIG ==
	cfg := &api.Config{
		Addr: addr,
		Database: &api.DBConfig{
			Addr:               dbUrl,
			MaxOpenConnections: maxOpenConns,
			MaxIdleConnections: maxIdleConns,
			MaxIdleTime:        time.Duration(time.Duration(maxIdleTime) * time.Minute),
		},
	}

	// == DATABASE ==
	/*  TODO(maolivera): look how to configure and create a pool with SQLC and PGX
	dbConfig, err := pgxpool.ParseConfig(cfg.Database.Addr)
	if err != nil {
		log.Fatalf("could not parse database_url: %v\n", err)
	}

	dbConfig.MaxConns = int32(cfg.Database.MaxOpenConnections)
	dbConfig.MinConns = int32(cfg.Database.MaxIdleConnections)
	dbConfig.MaxConnIdleTime = cfg.Database.MaxIdleTime

	dbpool, err := pgxpool.NewWithConfig(
		ctx,
		dbConfig,
	)
	if err != nil {
		log.Fatalln("could not create connection pool:", err)
	}
	defer dbpool.Close()
	log.Println("database connection pool established")
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, cfg.Database.Addr)
	if err != nil {
		log.Fatalln("could not connect to database: ", err)
	}

	db := database.New(conn)

	// == APPLICATION ==
	app := &api.Application{
		Config:   cfg,
		Database: db,
	}

	log.Fatalln(app.Start())
}
