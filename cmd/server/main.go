package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/maxolivera/gophis-social-network/internal/api"
	"github.com/maxolivera/gophis-social-network/internal/auth"
	"github.com/maxolivera/gophis-social-network/internal/cache"
	"github.com/maxolivera/gophis-social-network/internal/env"
	"github.com/maxolivera/gophis-social-network/internal/storage/postgres"
	lru "github.com/maxolivera/gophis-social-network/pkg/lru"
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
	cacheStruct, err := env.GetString("CACHE_STRUCT", logger)
	lruCap, err := env.GetInt("LRU_CAPACITY", logger)
	lruTTL, err := env.GetInt("LRU_TTL", logger) // Minutes
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
	}

	// == AUTH ==
	authenticator := auth.NewJWTAuthenticator(
		cfg.Authentication.Token.Secret,
		cfg.Authentication.Token.Issuer,
		cfg.Authentication.Token.Issuer,
	)

	// == CACHE ==
	var cacheStorage *cache.Storage
	cacheConfig := &api.CacheConfig{}
	if cacheStruct != "" {
		if cacheStruct == "REDIS" {
			cacheConfig.Enabled = true
			cacheConfig.Redis = &api.RedisConfig{
				Address:  redisAddr,
				Password: redisPass,
				Database: redisDb,
			}
			redisClient := cache.NewRedisClient(cacheConfig.Redis.Address, cacheConfig.Redis.Password, cacheConfig.Redis.Database)
			cacheStorage = cache.NewRedisStorage(redisClient)
		} else if cacheStruct == "LRU" {
			cacheConfig.Enabled = true
			cacheConfig.LRU = &api.LruConfig{
				Capacity: lruCap,
				TTL:      time.Duration(lruTTL) * time.Minute,
			}
			lru := lru.NewLRUCache(cacheConfig.LRU.Capacity, cacheConfig.LRU.TTL)
			cacheStorage = cache.NewLRUStorage(lru)
		}
	} else {
		cacheConfig.Enabled = false
	}
	cfg.Cache = cacheConfig

	// == STORAGE / DATABASE ==
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

	storage := postgres.NewPostgresStorage(pool)

	// == APPLICATION ==
	app := &api.Application{
		Config:        cfg,
		Storage:       storage,
		Cache:         cacheStorage,
		Pool:          pool,
		Logger:        logger,
		Authenticator: authenticator,
	}

	logger.Fatalln(app.Start())
}
