package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

// Do not hold the password
type User struct {
	ID        uuid.UUID   `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Email     string      `json:"email"`
	Username  string      `json:"username"`
	FirstName string      `json:"first_name,omitempty"`
	LastName  string      `json:"last_name,omitempty"`
	Role      ReducedRole `json:"role"`
}

// It has the real password. Should never be used besides on storage layers.
type UserWithPassword struct {
	User     User
	Password string
}

// Special type for Feed, do not contain CreatedAt, UpdatedAt, Email, FirstName, LastName
type ReducedUser struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

func DBUserToUser(dbUser database.User) *User {
	return &User{
		ID:        dbUser.ID.Bytes,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
		Email:     dbUser.Email,
		Username:  dbUser.Username,
		FirstName: dbUser.FirstName.String,
		LastName:  dbUser.LastName.String,
	}
}

func DBUsersToUser(dbUsers []database.User) []*User {
	users := make([]*User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = DBUserToUser(dbUser)
	}
	return users
}

func DBUserWithRoleToUser(dbUser database.GetUserByUsernameRow) *User {
	return &User{
		ID:        dbUser.ID.Bytes,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
		Email:     dbUser.Email,
		Username:  dbUser.Username,
		FirstName: dbUser.FirstName.String,
		LastName:  dbUser.LastName.String,
		Role: ReducedRole{
			Level: int(dbUser.Level),
			Name:  RoleType(dbUser.Name),
		},
	}
}
