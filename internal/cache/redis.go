package cache

import (
	"github.com/redis/go-redis/v9"
)

func NewRedisStorage(r *redis.Client) *Storage {
	return &Storage{
		Users: &UserRedisStore{r},
	}
}

func NewRedisClient(address, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
}
