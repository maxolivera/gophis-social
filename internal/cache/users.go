package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/maxolivera/gophis-social-network/internal/models"
	"github.com/redis/go-redis/v9"
)

type UserRedisStore struct {
	r *redis.Client
}

func (s UserRedisStore) Get(ctx context.Context, username string) (*models.User, error) {
	key := fmt.Sprintf("user-%s", username)

	data, err := s.r.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var user models.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s UserRedisStore) Set(ctx context.Context, user *models.User) error {
	if user.Username == "" {
		return errors.New("username is empty")
	}
	key := fmt.Sprintf("user-%s", user.Username)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.r.SetEx(ctx, key, json, UserTimeExpiration).Err()
}

func (s UserRedisStore) Delete(ctx context.Context, username string) {
	key := fmt.Sprintf("user-%s", username)
	s.r.Del(ctx, key)
}

type UserLRUCache struct {
	c *LRUCache
}

func (u UserLRUCache) Get(ctx context.Context, username string) (*models.User, error) {
	key := "user-" + username
	value, found := u.c.Get(key)
	if !found {
		return nil, nil
	}

	user, ok := value.(*models.User)
	if !ok {
		return nil, errors.New("value is not an user")
	}

	return user, nil
}

func (u UserLRUCache) Set(ctx context.Context, user *models.User) error {
	key := "user-" + user.Username
	u.c.Set(key, user)
	return nil
}

func (u UserLRUCache) Delete(ctx context.Context, username string) {
	key := "user-" + username
	u.c.Delete(key)
}
