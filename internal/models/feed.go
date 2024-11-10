package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

// TODO(maolivera): Better struct to not send empty fields on User

type Feed struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
	Tags         []string  `json:"tags"`
	Author       User      `json:"author"`
	CommentCount int64     `json:"comment_count"`
}

func DBFeedRowToFeed(row any) (Feed, error) {
	switch v := row.(type) {
	case database.GetUserFeedRow:
		return Feed{
			ID:           v.ID.Bytes,
			Title:        v.Title,
			CreatedAt:    v.CreatedAt.Time,
			Content:      v.Content,
			Tags:         v.Tags,
			Author:       User{ID: v.AuthorID.Bytes, Username: v.Username.String},
			CommentCount: v.CommentCount,
		}, nil
	case database.SearchPostsRow:
		return Feed{
			ID:           v.ID.Bytes,
			Title:        v.Title,
			CreatedAt:    v.CreatedAt.Time,
			Content:      v.Content,
			Tags:         v.Tags,
			Author:       User{ID: v.AuthorID.Bytes, Username: v.Username.String},
			CommentCount: v.CommentCount,
		}, nil
	default:
		return Feed{}, fmt.Errorf("unsupported row type: %T", v)
	}

}

func DBFeedsToFeeds[S ~[]E, E any](s S) ([]Feed, error) {
	feeds := make([]Feed, len(s))
	for i, dbFeed := range s {
		feed, err := DBFeedRowToFeed(dbFeed)
		if err != nil {
			return nil, err
		}
		feeds[i] = feed
	}
	return feeds, nil
}
