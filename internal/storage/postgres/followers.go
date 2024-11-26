package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxolivera/gophis-social-network/internal/database"
)

type PostgresFollowerRepository struct {
	p *pgxpool.Pool
}

func (r PostgresFollowerRepository) Follow(ctx context.Context, user, follower uuid.UUID) error {
	q := database.New(r.p)
	currentTime := time.Now().UTC()

	return q.FollowByID(ctx, database.FollowByIDParams{
		CreatedAt:  pgtype.Timestamp{Time: currentTime, Valid: true},
		UserID:     pgtype.UUID{Bytes: user, Valid: true},
		FollowerID: pgtype.UUID{Bytes: follower, Valid: true},
	})
}

func (r PostgresFollowerRepository) Unfollow(ctx context.Context, user, follower uuid.UUID) error {
	q := database.New(r.p)

	return q.UnfollowByID(ctx, database.UnfollowByIDParams{
		UserID:     pgtype.UUID{Bytes: user, Valid: true},
		FollowerID: pgtype.UUID{Bytes: follower, Valid: true},
	})
}
