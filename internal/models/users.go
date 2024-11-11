package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

// Special type for Feed, do not contain CreatedAt, UpdatedAt, Email, FirstName, LastName
type ReducedUser struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

func DBUserToUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID.Bytes,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
		Email:     dbUser.Email,
		Username:  dbUser.Username,
		FirstName: dbUser.FirstName.String,
		LastName:  dbUser.LastName.String,
	}
}

func DBUsersToUser(dbUsers []database.User) []User {
	users := make([]User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = DBUserToUser(dbUser)
	}
	return users
}
