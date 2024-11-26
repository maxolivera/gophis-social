package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxolivera/gophis-social-network/internal/database"
	"github.com/maxolivera/gophis-social-network/internal/storage"
	"github.com/maxolivera/gophis-social-network/internal/storage/models"
)

type PostgresPostRepository struct {
	p *pgxpool.Pool
}

func (r *PostgresPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	q := database.New(r.p)
	dbPost, err := q.GetPostById(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, storage.ErrNoRows
		default:
			return nil, err
		}
	}

	post := models.DBPostToPost(dbPost)
	return post, nil
}

func (r *PostgresPostRepository) Create(ctx context.Context, p *models.Post) error {
	q := database.New(r.p)

	return q.CreatePost(ctx, database.CreatePostParams{
		ID:        pgtype.UUID{Bytes: p.ID, Valid: true},
		CreatedAt: pgtype.Timestamp{Time: p.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: p.UpdatedAt, Valid: true},
		UserID:    pgtype.UUID{Bytes: p.UserID, Valid: true},
		Title:     p.Title,
		Content:   p.Content,
		Tags:      p.Tags,
	})
}

func (r *PostgresPostRepository) SoftDelete(ctx context.Context, p *models.Post) error {
	q := database.New(r.p)

	deleted, err := q.SoftDeletePostByID(ctx, database.SoftDeletePostByIDParams{
		ID:      pgtype.UUID{Bytes: p.ID, Valid: true},
		Version: p.Version,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("post was not deleted because was not found, post_id: %v", p.ID)
		}
		return fmt.Errorf("post could not deleted: %v", err)
	}
	if !deleted {
		return fmt.Errorf("post was not deleted, post_id: %v", p.ID)
	}

	return nil
}

func (r *PostgresPostRepository) HardDelete(ctx context.Context, p *models.Post) error {
	q := database.New(r.p)
	if err := q.HardDeletePostByID(ctx, database.HardDeletePostByIDParams{
		ID:      pgtype.UUID{Bytes: p.ID, Valid: true},
		Version: p.Version,
	}); err != nil {
		if err == pgx.ErrNoRows {
			return storage.ErrNoRows
		}
		return err
	}

	return nil
}

func (r *PostgresPostRepository) Update(ctx context.Context, p *models.Post) (*models.Post, error) {
	q := database.New(r.p)
	dbPost, err := q.UpdatePost(ctx, database.UpdatePostParams{
		UpdatedAt: pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
		ID:        pgtype.UUID{Bytes: p.ID, Valid: true},
		Content:   pgtype.Text{String: p.Content, Valid: len(p.Content) > 0},
		Title:     pgtype.Text{String: p.Title, Valid: len(p.Title) > 0},
		Tags:      p.Tags,
		Version:   p.Version,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, storage.ErrNoRows
		}
		return nil, err
	}

	post := models.DBPostToPost(dbPost)

	return post, nil
}

func (r *PostgresPostRepository) GetFeed(ctx context.Context, u *models.User, sort bool, limit, offset int32) ([]*models.Feed, error) {
	q := database.New(r.p)
	dbFeed, err := q.GetUserFeed(ctx, database.GetUserFeedParams{
		UserID: pgtype.UUID{Bytes: u.ID, Valid: true},
		Sort:   sort,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, storage.ErrNoRows
		}
		return nil, err
	}

	feed, err := models.DBFeedsToFeeds(dbFeed)
	if err != nil {
		return nil, err
	}

	return feed, nil
}

func (r *PostgresPostRepository) Search(ctx context.Context, word string, tags []string, limit, offset int32, sort bool, since, until *time.Time) ([]*models.Feed, error) {
	q := database.New(r.p)
	params := database.SearchPostsParams{
		Search: "",
		Tags:   nil,
		Limit:  10,    // Default limit
		Offset: 0,     // Default offset
		Sort:   false, // Default sort order
	}

	if tags != nil {
		params.Tags = tags
	}
	if offset != -1 {
		params.Offset = offset
	}
	if limit != -1 {
		params.Limit = limit
	}
	params.Sort = sort
	params.Search = word
	if since != nil {
		params.Since = pgtype.Timestamp{Time: *since}
	}
	if until != nil {
		params.Until = pgtype.Timestamp{Time: *until}
	}

	dbFeed, err := q.SearchPosts(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, storage.ErrNoRows
		}
		return nil, err
	}

	feeds, err := models.DBFeedsToFeeds(dbFeed)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}
