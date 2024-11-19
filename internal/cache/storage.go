package cache

import (
	"context"
	"time"

	"github.com/maxolivera/gophis-social-network/internal/models"
)

type Storage struct {
	Users interface {
		Get(context.Context, string) (*models.User, error)
		Set(context.Context, *models.User) error
		Delete(context.Context, string)
	}
}

const UserTimeExpiration = time.Minute
