package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

// TODO(maolivera): Better struct to not send empty fields on User

type UserFeed struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
	Tags         []string  `json:"tags"`
	Author       User      `json:"author"`
	CommentCount int64     `json:"comment_count"`
}

func DBFeedToFeed(dbFeed database.GetUserFeedRow) UserFeed {
	return UserFeed{
		ID:           dbFeed.ID.Bytes,
		Title:        dbFeed.Title,
		CreatedAt:    dbFeed.CreatedAt.Time,
		Content:      dbFeed.Content,
		Tags:         dbFeed.Tags,
		Author:       User{ID: dbFeed.ID_2.Bytes, Username: dbFeed.Username.String},
		CommentCount: dbFeed.CommentCount,
	}
}

func DBFeedsToFeeds(dbFeeds []database.GetUserFeedRow) []UserFeed {
	feeds := make([]UserFeed, len(dbFeeds))
	for i, dbFeed := range dbFeeds {
		feeds[i] = DBFeedToFeed(dbFeed)
	}
	return feeds
}
