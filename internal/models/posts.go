package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

type Post struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	Comments  []Comment `json:"comments"`
	Version   int32     `json:"version"`
}

func DBPostToPost(dbPost database.Post) Post {
	return Post{
		ID:        dbPost.ID.Bytes,
		UserID:    dbPost.UserID.Bytes,
		CreatedAt: dbPost.CreatedAt.Time,
		UpdatedAt: dbPost.UpdatedAt.Time,
		Title:     dbPost.Title,
		Content:   dbPost.Content,
		Tags:      dbPost.Tags,
		Version:   dbPost.Version,
	}
}

func DBPostsToPost(dbPosts []database.Post) []Post {
	posts := make([]Post, len(dbPosts))
	for i, dbPost := range dbPosts {
		posts[i] = DBPostToPost(dbPost)
	}
	return posts
}
