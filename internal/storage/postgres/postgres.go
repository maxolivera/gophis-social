package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxolivera/gophis-social-network/internal/storage"
)

func NewPostgresStorage(p *pgxpool.Pool) *storage.Storage {
	return &storage.Storage{
		Posts:     &PostgresPostRepository{p},
		Users:     &PostgresUserRepository{p},
		Comments:  &PostgresCommentRepository{p},
		Followers: &PostgresFollowerRepository{p},
		Roles:     &PostgresRoleRepository{p},
	}
}

func withTx(pool *pgxpool.Pool, ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
