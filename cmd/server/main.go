package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/maxolivera/gophis-social-network/internal/api"
	"github.com/maxolivera/gophis-social-network/internal/auth"
	"github.com/maxolivera/gophis-social-network/internal/cache"
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
	pass, err := env.GetString("AUTH_BASIC_USER", logger)
	user, err := env.GetString("AUTH_BASIC_PASS", logger)
	redisAddr, err := env.GetString("REDIS_ADDRESS", logger)
	redisPass, err := env.GetString("REDIS_PASSWORD", logger)
	redisDb, err := env.GetInt("REDIS_DB", logger)
	secret, err := env.GetString("JWT_SECRET", logger)

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
		Authentication: &api.AuthConfig{
			BasicAuth: &api.BasicAuth{
				Username: user,
				Password: pass,
			},
			Token: &api.TokenConfig{
				Secret:         secret,
				ExpirationTime: 3 * 24 * time.Hour,
				Issuer:         "gophissocial",
			},
		},
		Redis: &api.RedisConfig{
			Address:  redisAddr,
			Password: redisPass,
			Database: redisDb,
			Enabled:  (redisAddr != ""),
		},
	}

	// == AUTH ==
	authenticator := auth.NewJWTAuthenticator(
		cfg.Authentication.Token.Secret,
		cfg.Authentication.Token.Issuer,
		cfg.Authentication.Token.Issuer,
	)

	// == CACHE ==
	var cacheStorage *cache.Storage
	if cfg.Redis.Enabled {
		redisClient := cache.NewRedisClient(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.Database)
		cacheStorage = cache.NewRedisStorage(redisClient)
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
		Config:        cfg,
		Database:      queries,
		Cache:         cacheStorage,
		Pool:          pool,
		Logger:        logger,
		Authenticator: authenticator,
	}

	logger.Fatalln(app.Start())
}
