package cache

import (
	"context"
	"errors"

	"github.com/maxolivera/gophis-social-network/internal/storage/models"
	"github.com/maxolivera/gophis-social-network/pkg/lru"
)

func NewLRUStorage(c *lru.LRUCache) *Storage {
	return &Storage{
		Users: &UserLRUCache{c},
	}
}

type UserLRUCache struct {
	c *lru.LRUCache
}

func (u UserLRUCache) Get(ctx context.Context, username string) (*models.User, error) {
	key := "user-" + username
	value, found := u.c.Get(ctx, key)
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
	u.c.Set(ctx, key, user)
	return nil
}

func (u UserLRUCache) Delete(ctx context.Context, username string) {
	key := "user-" + username
	u.c.Delete(ctx, key)
}

func (u UserLRUCache) Len(ctx context.Context) int {
	return u.c.Len(ctx)
}
