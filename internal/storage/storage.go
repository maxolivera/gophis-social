package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/maxolivera/gophis-social-network/internal/storage/models"
)

var (
	ErrNoRows              = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrUsernameUnavailable = errors.New("username is unavailable")
	ErrEmailUnavailable    = errors.New("email is unavailable")
	ErrNoUser              = errors.New("user not found")
	ErrNoToken             = errors.New("token not found")
	QueryTimeDuration      = time.Second * 5
)

type Storage struct {
	Posts     PostRepository
	Users     UserRepository
	Comments  CommentRepository
	Followers FollowerRepository
	Roles     RoleRepository
}

type PostRepository interface {
	// Fetch a post by ID
	GetByID(context.Context, uuid.UUID) (*models.Post, error)
	// Stores a post
	Create(context.Context, *models.Post) error
	// Mark a post as deleted
	SoftDelete(context.Context, *models.Post) error
	// Deletes a post
	HardDelete(context.Context, *models.Post) error
	// Updates a post
	Update(context.Context, *models.Post) (*models.Post, error)
	// Retrieve feed for user. It requires sort (bool), a limit and an offset
	GetFeed(context.Context, *models.User, bool, int32, int32) ([]*models.Feed, error)
	// Search posts.
	Search(context.Context, string, []string, int32, int32, bool, *time.Time, *time.Time) ([]*models.Feed, error)
}

type UserRepository interface {
	// Fetch a user by username
	GetByUsername(context.Context, string) (*models.User, error)
	// Fetch a user by email and password. Used for log in.
	GetByEmailAndPassword(context.Context, string, string) (*models.User, error)
	// Stores a user
	Create(context.Context, *models.UserWithPassword) error
	// Stores a user and the invitation
	CreateAndInvite(context.Context, *models.UserWithPassword, []byte, time.Duration) error
	// Activates a user and deletes the invitation
	Activate(context.Context, []byte) (*models.User, error)
	// Mark a user as deleted
	SoftDelete(context.Context, uuid.UUID) error
	// Deletes a user
	HardDelete(context.Context, uuid.UUID) error
	// Updates a user. The user parameter may contain empty fields, which mean they will not change.
	Update(context.Context, *models.UserWithPassword) (*models.User, error)
}

type CommentRepository interface {
	// Create a comment on a post
	Create(context.Context, *models.Comment) error
	// Get comments from a post
	GetByPostID(context.Context, uuid.UUID) ([]*models.Comment, error)
}

type FollowerRepository interface {
	// Follows a user
	Follow(context.Context, uuid.UUID, uuid.UUID) error
	// Unfollows a user
	Unfollow(context.Context, uuid.UUID, uuid.UUID) error
}

type RoleRepository interface {
	// Get role without description nor ID.
	GetByName(context.Context, string) (*models.ReducedRole, error)
}
