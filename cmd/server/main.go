package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/maxolivera/gophis-social-network/internal/api"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/env"
	"go.uber.org/zap"
)

const Version = "0.0.1"

func main() {
	// == Logger ==
	// TODO(maolivera): Change Logger type based on env
	logger := zap.Must(zap.NewDevelopment()).Sugar()
	defer logger.Sync()

	// == ENV VALUES ==
	err := godotenv.Load()
	addr, err := env.GetString("ADDR", logger)
	apiUrl, err := env.GetString("EXTERNAL_URL", logger)
	environment, err := env.GetString("ENV", logger)
	dbUrl, err := env.GetString("DB_URL", logger)
	maxOpenConns, err := env.GetInt("DB_MAX_OPEN_CONNS", logger)
	maxIdleConns, err := env.GetInt("DB_MAX_IDLE_CONNS", logger)
	maxIdleTime, err := env.GetInt("DB_MAX_IDLE_TIME", logger) // MaxIdleTime represents minutes
	if err != nil {
		logger.Fatalf("error loading env values: %v\n", err)
	}

	// == CONFIG ==
	cfg := &api.Config{
		Addr:        addr,
		Environment: environment,
		Version:     Version,
		ApiUrl:      apiUrl,
		Database: &api.DBConfig{
			Addr:               dbUrl,
			MaxOpenConnections: maxOpenConns,
			MaxIdleConnections: maxIdleConns,
			MaxIdleTime:        time.Duration(time.Duration(maxIdleTime) * time.Minute),
		},
		ExpirationTime: 3 * 24 * time.Hour,
	}

	// == DATABASE ==
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbConfig, err := pgxpool.ParseConfig(cfg.Database.Addr)
	if err != nil {
		logger.Fatalf("could not parse database_url: %v\n", err)
	}

	dbConfig.MaxConns = int32(cfg.Database.MaxOpenConnections)
	dbConfig.MinConns = int32(cfg.Database.MaxIdleConnections)
	dbConfig.MaxConnIdleTime = cfg.Database.MaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		logger.Fatalf("could not create connection pool: %v\n", err)
	}
	defer pool.Close()
	logger.Info("database connection pool established")

	queries := database.New(pool)

	// == APPLICATION ==
	app := &api.Application{
		Config:   cfg,
		Database: queries,
		Pool:     pool,
		Logger:   logger,
	}

	logger.Fatalln(app.Start())
}
