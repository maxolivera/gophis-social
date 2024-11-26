package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

type Comment struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"post_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Content   string    `json:"content"`
	User      *User     `json:"user"`
}

func DBCommentToComment(dbComment database.Comment) *Comment {
	return &Comment{
		ID:        dbComment.ID.Bytes,
		PostID:    dbComment.PostID.Bytes,
		CreatedAt: dbComment.CreatedAt.Time,
		UpdatedAt: dbComment.UpdatedAt.Time,
		Content:   dbComment.Content,
		User:      &User{ID: dbComment.UserID.Bytes},
	}
}

func DBCommentsToComments(dbComments []database.Comment) []*Comment {
	comments := make([]*Comment, len(dbComments))
	for i, dbComment := range dbComments {
		comments[i] = DBCommentToComment(dbComment)
	}
	return comments
}

func DBCommentWithUser(dbComment database.GetCommentsByPostRow) *Comment {
	return &Comment{
		ID:        dbComment.ID.Bytes,
		CreatedAt: dbComment.CreatedAt.Time,
		Content:   dbComment.Content,
		User: &User{
			Username:  dbComment.Username.String,
			FirstName: dbComment.Username.String,
			Email:     dbComment.Email.String,
		},
	}
}

// Maps returned query
func DBCommentsWithUser(dbComments []database.GetCommentsByPostRow) []*Comment {
	comments := make([]*Comment, len(dbComments))
	for i, dbComment := range dbComments {
		comments[i] = DBCommentWithUser(dbComment)
	}
	return comments
}
